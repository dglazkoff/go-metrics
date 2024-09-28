package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dglazkoff/go-metrics/cmd/server/config"
	constants "github.com/dglazkoff/go-metrics/internal/const"
	"github.com/dglazkoff/go-metrics/internal/logger"
	"github.com/dglazkoff/go-metrics/internal/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

var retryIntervals = []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

type dbStorage struct {
	db  *sql.DB
	cfg *config.Config
}

func New(db *sql.DB, cfg *config.Config) *dbStorage {
	return &dbStorage{cfg: cfg, db: db}
}

func dbExecute(exec func() (sql.Result, error)) (res sql.Result, err error) {
	for index, interval := range retryIntervals {
		logger.Log.Debug("execute db. retry number: ", index)
		res, err = exec()

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgerrcode.IsConnectionException(pgErr.Code) {
				if index == len(retryIntervals)-1 {
					return nil, fmt.Errorf("no connection to database")
				}

				time.Sleep(interval)
				continue
			}
		}

		break
	}

	return res, err
}

func dbQueryRow(query func() (*sql.Row, error)) (res *sql.Row, err error) {
	for index, interval := range retryIntervals {
		logger.Log.Debug("reading data from db. retry number: ", index)
		res, err = query()

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgerrcode.IsConnectionException(pgErr.Code) {
				if index == len(retryIntervals)-1 {
					return nil, fmt.Errorf("no connection to database")
				}

				time.Sleep(interval)
				continue
			}
		}

		break
	}

	return res, err
}

func Bootstrap(d *dbStorage) error {
	_, err := dbExecute(func() (sql.Result, error) {
		return d.db.Exec("CREATE TABLE IF NOT EXISTS metrics (id VARCHAR(250) PRIMARY KEY, type VARCHAR(250) NOT NULL, value DOUBLE PRECISION, delta BIGINT)")
	})

	//  @tmvrus как вообще понимать какого рода ошибка упала ? читать код библиотек и понимать какие ошибки они выкидывают?
	// Потому что я отключаю БД и ожидаю что у меня упадет ошибка pgconn.PgError, но у меня падает видимо какая-то другая, более низкоуровневая ошибка

	// я пытался воспроизвести так: отключаю совсем БД, но это была ошибка не типа pgconn.PgError и я не знаю какого
	// вот как это понять?
	//if err != nil {
	//	var pgErr *pgconn.PgError
	//	if errors.As(err, &pgErr) {
	//		fmt.Println(1)
	//		fmt.Println(pgErr.Message) // => syntax error at end of input
	//		fmt.Println(pgErr.Code)    // => 42601
	//	}
	//}

	/*
		@tmvrus:
			нужно смотреть что возвращает тебе код при каждой ошибке, если убить ДБ то драйвер скорее всего даст ошибку сетевого стека,
			это будет net.OpError с различными кодами, либо ошибка самого драйвера если он обрабатывает net.OpError
	*/

	if err != nil {
		logger.Log.Debug("error while creating table: ", err)
		return err
	}

	return nil
}

func (d *dbStorage) ReadMetrics(ctx context.Context) ([]models.Metrics, error) {
	var metrics []models.Metrics
	rows, err := d.db.QueryContext(ctx, "SELECT id, type, value, delta from metrics")

	if err != nil {
		logger.Log.Debug("error while reading metrics ", err)
		return nil, fmt.Errorf("error while reading metrics")
	}

	defer rows.Close()

	for rows.Next() {
		var metric models.Metrics
		err = rows.Scan(&metric.ID, &metric.MType, &metric.Value, &metric.Delta)

		if err != nil {
			logger.Log.Debug("error while scan metric ", err)
			continue
		}

		metrics = append(metrics, metric)
	}

	if rows.Err() != nil {
		logger.Log.Debug("error from rows ", err)
	}

	return metrics, nil
}

func (d *dbStorage) ReadMetric(ctx context.Context, id string) (models.Metrics, error) {
	var metric models.Metrics
	_, err := dbQueryRow(func() (*sql.Row, error) {
		row := d.db.QueryRowContext(ctx, "SELECT id, type, value, delta from metrics WHERE id = $1", id)

		err := row.Scan(&metric.ID, &metric.MType, &metric.Value, &metric.Delta)

		return row, err
	})

	if err != nil {
		logger.Log.Debug("error while reading metric ", err)
		return models.Metrics{}, fmt.Errorf("error while reading metric")
	}

	return metric, nil
}

func (d *dbStorage) UpdateMetric(ctx context.Context, metric models.Metrics) error {
	if metric.MType == constants.MetricTypeGauge {
		_, err := dbExecute(func() (sql.Result, error) {
			return d.db.ExecContext(
				ctx,
				"INSERT INTO metrics (id, type, value, delta) VALUES($1, $2, $3, $4) ON CONFLICT (id) DO UPDATE SET value = $3",
				metric.ID, metric.MType, metric.Value, metric.Delta,
			)
		})
		return err
	}

	if metric.MType == constants.MetricTypeCounter {
		dbMetric, err := d.ReadMetric(ctx, metric.ID)

		if err != nil {
			_, err = dbExecute(func() (sql.Result, error) {
				return d.db.ExecContext(ctx, "INSERT INTO metrics (id, type, value, delta) VALUES($1, $2, $3, $4)", metric.ID, metric.MType, metric.Value, metric.Delta)
			})

			return err
		}

		newDelta := *dbMetric.Delta + *metric.Delta
		_, err = dbExecute(func() (sql.Result, error) {
			return d.db.ExecContext(ctx, "UPDATE metrics SET delta = $1 WHERE id = $2", &newDelta, metric.ID)
		})
		return err
	}

	return fmt.Errorf("unknown metric type %s", metric.MType)
}

func (d *dbStorage) SaveMetrics(ctx context.Context, metrics []models.Metrics) error {
	var notSavedMetricsIds []string
	// можно сделать транзакцию для ускорения
	for _, metric := range metrics {
		err := d.UpdateMetric(ctx, metric)

		if err != nil {
			notSavedMetricsIds = append(notSavedMetricsIds, metric.ID)
			logger.Log.Debug("error while insert value ", err)
		}
	}

	if len(notSavedMetricsIds) > 0 {
		return fmt.Errorf("metrics was not saved: %s", strings.Join(notSavedMetricsIds, ", "))
	}

	return nil
}

func (d *dbStorage) PingDB(ctx context.Context) error {
	if err := d.db.PingContext(ctx); err != nil {
		return fmt.Errorf("no connection to database %w", err)
	}

	return nil
}
