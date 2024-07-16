package db

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/dglazkoff/go-metrics/cmd/server/config"
	constants "github.com/dglazkoff/go-metrics/internal/const"
	"github.com/dglazkoff/go-metrics/internal/logger"
	"github.com/dglazkoff/go-metrics/internal/models"
)

type dbStorage struct {
	db  *sql.DB
	cfg *config.Config
}

func New(db *sql.DB, cfg *config.Config) *dbStorage {
	return &dbStorage{cfg: cfg, db: db}
}

func Bootstrap(d *dbStorage) error {
	_, err := d.db.Exec("CREATE TABLE IF NOT EXISTS metrics (id VARCHAR(250) PRIMARY KEY, type VARCHAR(250) NOT NULL, value DOUBLE PRECISION, delta INTEGER)")

	if err != nil {
		logger.Log.Debug("error while creating table ", err)
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
	row := d.db.QueryRow("SELECT id, type, value, delta from metrics WHERE id = $1", id)

	err := row.Scan(&metric.ID, &metric.MType, &metric.Value, &metric.Delta)

	if err != nil {
		logger.Log.Debug("error while reading metric ", err)
		return models.Metrics{}, fmt.Errorf("error while reading metric")
	}

	return metric, err
}

func (d *dbStorage) UpdateMetric(metric models.Metrics) error {
	if metric.MType == constants.MetricTypeGauge {
		_, err := d.db.Exec("INSERT INTO metrics (id, type, value, delta) VALUES($1, $2, $3, $4)", metric.ID, metric.MType, metric.Value, metric.Delta)
		return err
	}

	if metric.MType == constants.MetricTypeCounter {
		dbMetric, err := d.ReadMetric(metric.ID)

		if err != nil {
			_, err = d.db.Exec("INSERT INTO metrics (id, type, value, delta) VALUES($1, $2, $3, $4)", metric.ID, metric.MType, metric.Value, metric.Delta)
			return err
		}

		_, err = d.db.Exec("UPDATE metrics SET delta = $1 WHERE id = $2", *dbMetric.Delta+*metric.Delta, metric.ID)
		return err
	}

	return fmt.Errorf("unknown metric type %s", metric.MType)
}

func (d *dbStorage) SaveMetrics(metrics []models.Metrics) {
	// можно сделать транзакцию для ускорения
	for _, metric := range metrics {
		_, err := d.db.Exec("INSERT INTO metrics (id, type, value, delta) VALUES($1, $2, $3, $4)", metric.ID, metric.MType, metric.Value, metric.Delta)

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
