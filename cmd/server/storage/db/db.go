package db

import (
	"database/sql"
	"github.com/dglazkoff/go-metrics/cmd/server/config"
	"github.com/dglazkoff/go-metrics/cmd/server/storage"
	"github.com/dglazkoff/go-metrics/internal/logger"
	"github.com/dglazkoff/go-metrics/internal/models"
	"time"
)

type db interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
}

type metricStorage interface {
	SaveMetrics(metrics []models.Metrics)
	ReadMetrics() []models.Metrics
}

type dbStorage struct {
	db      db
	storage metricStorage
	cfg     *config.Config
}

func New(db db, s storage.MetricsStorage, cfg *config.Config) dbStorage {
	return dbStorage{storage: s, cfg: cfg, db: db}
}

func (d dbStorage) ReadMetrics() {
	if !d.cfg.IsRestore {
		return
	}

	var metrics []models.Metrics
	rows, err := d.db.Query("SELECT id, type, value, delta from metrics")

	if err != nil {
		logger.Log.Debug("error while reading metrics ", err)
		return
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

	d.storage.SaveMetrics(metrics)
}

func (d dbStorage) WriteMetrics(isLoop bool) {
	for {
		time.Sleep(time.Duration(d.cfg.StoreInterval) * time.Second)
		metrics := d.storage.ReadMetrics()

		for _, metric := range metrics {
			_, err := d.db.Exec("TRUNCATE TABLE metrics")

			if err != nil {
				logger.Log.Debug("error while truncate the table ", err)
				return
			}

			_, err = d.db.Exec("INSERT INTO metrics (id, type, value, delta) VALUES($1, $2, $3, $4)", metric.ID, metric.MType, metric.Value, metric.Delta)

			if err != nil {
				logger.Log.Debug("error while insert value ", err)
			}
		}

		if !isLoop {
			break
		}
	}
}
