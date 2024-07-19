package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/dglazkoff/go-metrics/cmd/server/config"
	"github.com/dglazkoff/go-metrics/internal/const"
	"github.com/dglazkoff/go-metrics/internal/logger"
	"github.com/dglazkoff/go-metrics/internal/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"time"
)

var retryIntervals = []time.Duration{1, 3, 5}

type dbStorage struct {
	db  *sql.DB
	cfg *config.Config
}

func New(db *sql.DB, cfg *config.Config) *dbStorage {
	return &dbStorage{cfg: cfg, db: db}
}

func dbExecute(exec func() (sql.Result, error), retryNumber int) (sql.Result, error) {
	logger.Log.Debug("execute db. retry number: ", retryNumber)
	_, err := exec()

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgerrcode.IsConnectionException(pgErr.Code) {
			if retryNumber == 3 {
				return nil, fmt.Errorf("no connection to database")
			}

			time.Sleep(retryIntervals[retryNumber] * time.Second)
			return dbExecute(exec, retryNumber+1)
		}
	}

	return nil, err
}

func dbQueryRow(query func() (*sql.Row, error), retryNumber int) (*sql.Row, error) {
	logger.Log.Debug("reading data from db. retry number: ", retryNumber)
	_, err := query()

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgerrcode.IsConnectionException(pgErr.Code) {
			if retryNumber == 3 {
				return nil, fmt.Errorf("no connection to database")
			}

			time.Sleep(retryIntervals[retryNumber] * time.Second)
			return dbQueryRow(query, retryNumber+1)
		}
	}

	return nil, err
}

func Bootstrap(d *dbStorage) error {
	_, err := dbExecute(func() (sql.Result, error) {
		return d.db.Exec("CREATE TABLE IF NOT EXISTS metrics (id VARCHAR(250) PRIMARY KEY, type VARCHAR(250) NOT NULL, value DOUBLE PRECISION, delta BIGINT)")
	}, 0)

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

	if err != nil {
		logger.Log.Debug("error while creating table: ", err)
		return err
	}

	return nil
}

func (d *dbStorage) ReadMetrics() ([]models.Metrics, error) {
	var metrics []models.Metrics
	rows, err := d.db.Query("SELECT id, type, value, delta from metrics")

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
		}

		metrics = append(metrics, metric)
	}

	if rows.Err() != nil {
		logger.Log.Debug("error from rows ", err)
	}

	return metrics, nil
}

func (d *dbStorage) ReadMetric(id string) (models.Metrics, error) {
	var metric models.Metrics
	_, err := dbQueryRow(func() (*sql.Row, error) {
		row := d.db.QueryRow("SELECT id, type, value, delta from metrics WHERE id = $1", id)

		err := row.Scan(&metric.ID, &metric.MType, &metric.Value, &metric.Delta)

		return row, err
	}, 0)

	if err != nil {
		logger.Log.Debug("error while reading metric ", err)
		return models.Metrics{}, fmt.Errorf("error while reading metric")
	}

	return metric, err
}

func (d *dbStorage) UpdateMetric(metric models.Metrics) error {
	dbMetric, err := d.ReadMetric(metric.ID)

	if err != nil {
		_, err = dbExecute(func() (sql.Result, error) {
			return d.db.Exec("INSERT INTO metrics (id, type, value, delta) VALUES($1, $2, $3, $4)", metric.ID, metric.MType, metric.Value, metric.Delta)
		}, 0)

		return err
	}

	if metric.MType == constants.MetricTypeGauge {
		_, err = dbExecute(func() (sql.Result, error) {
			return d.db.Exec("UPDATE metrics SET value = $1 WHERE id = $2", metric.Value, metric.ID)
		}, 0)
		return err
	}

	if metric.MType == constants.MetricTypeCounter {
		newDelta := *dbMetric.Delta + *metric.Delta
		_, err = dbExecute(func() (sql.Result, error) {
			return d.db.Exec("UPDATE metrics SET delta = $1 WHERE id = $2", &newDelta, metric.ID)
		}, 0)
		return err
	}

	return fmt.Errorf("unknown metric type %s", metric.MType)
}

func (d *dbStorage) SaveMetrics(metrics []models.Metrics) {
	// можно сделать транзакцию для ускорения
	for _, metric := range metrics {
		_, err := dbExecute(func() (sql.Result, error) {
			return d.db.Exec("INSERT INTO metrics (id, type, value, delta) VALUES($1, $2, $3, $4)", metric.ID, metric.MType, metric.Value, metric.Delta)
		}, 0)

		if err != nil {
			logger.Log.Debug("error while insert value ", err)
		}
	}
}

func (d *dbStorage) PingDB(ctx context.Context) error {
	if err := d.db.PingContext(ctx); err != nil {
		return fmt.Errorf("no connection to database")
	}

	return nil
}

//func (d dbStorage) ReadMetrics() {
//	if !d.cfg.IsRestore {
//		return
//	}
//
//	var metrics []models.Metrics
//	rows, err := d.db.Query("SELECT id, type, value, delta from metrics")
//
//	if err != nil {
//		logger.Log.Debug("error while reading metrics ", err)
//		return
//	}
//
//	defer rows.Close()
//
//	for rows.Next() {
//		var metric models.Metrics
//		err = rows.Scan(&metric.ID, &metric.MType, &metric.Value, &metric.Delta)
//
//		if err != nil {
//			logger.Log.Debug("error while scan metric ", err)
//		}
//
//		metrics = append(metrics, metric)
//	}
//
//	if rows.Err() != nil {
//		logger.Log.Debug("error from rows ", err)
//	}
//
//	d.storage.SaveMetrics(metrics)
//}

//func (d dbStorage) WriteMetrics(isLoop bool) {
//	for {
//		time.Sleep(time.Duration(d.cfg.StoreInterval) * time.Second)
//		metrics := d.storage.ReadMetrics()
//
//		for _, metric := range metrics {
//			_, err := d.db.Exec("TRUNCATE TABLE metrics")
//
//			if err != nil {
//				logger.Log.Debug("error while truncate the table ", err)
//				return
//			}
//
//			_, err = d.db.Exec("INSERT INTO metrics (id, type, value, delta) VALUES($1, $2, $3, $4)", metric.ID, metric.MType, metric.Value, metric.Delta)
//
//			if err != nil {
//				logger.Log.Debug("error while insert value ", err)
//			}
//		}
//
//		if !isLoop {
//			break
//		}
//	}
//}
