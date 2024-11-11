// Пакет storage предоставляет интерфейсы для работы с хранилищами
package storage

import (
	"context"
	"database/sql"

	"github.com/dglazkoff/go-metrics/cmd/server/config"
	"github.com/dglazkoff/go-metrics/cmd/server/storage/db"
	"github.com/dglazkoff/go-metrics/cmd/server/storage/file"
	"github.com/dglazkoff/go-metrics/cmd/server/storage/metrics"
	"github.com/dglazkoff/go-metrics/internal/logger"
	"github.com/dglazkoff/go-metrics/internal/models"
)

// MetricsStorage - интерфейс для работы с хранилищем метрик
type MetricsStorage interface {
	// ReadMetric - метод для получения метрики по имени
	ReadMetric(ctx context.Context, name string) (models.Metrics, error)
	// ReadMetrics - метод для получения всех метрик
	ReadMetrics(ctx context.Context) ([]models.Metrics, error)
	// UpdateMetric - метод для обновления метрики
	UpdateMetric(ctx context.Context, metric models.Metrics) error
	// SaveMetrics - метод для добавления списка метрик
	SaveMetrics(ctx context.Context, metrics []models.Metrics) error
	// PingDB - метод для проверки соединения с БД
	PingDB(ctx context.Context) error
}

// FileStorage - интерфейс для работы с файловым хранилищем
type FileStorage interface {
	// WriteMetrics - метод для записи метрик в файл
	// isLoop - проставляется false, если нужно единожды записать метрики
	WriteMetrics(isLoop bool)
	// ReadMetrics - метод для чтения метрик из файла
	ReadMetrics()
}

func InitStorages(cfg *config.Config) (MetricsStorage, FileStorage) {
	var store MetricsStorage

	if cfg.DatabaseDSN != "" {
		pgDB, err := sql.Open("pgx", cfg.DatabaseDSN)

		if err != nil {
			logger.Log.Debug("Error on open db", "err", err)
			panic(err)
		}
		defer pgDB.Close()

		dbStore := db.New(pgDB, db.RetryIntervals)
		err = db.Bootstrap(dbStore)

		if err != nil {
			logger.Log.Debug("Error on bootstrap db ", err)
			panic(err)
		}

		store = dbStore
	} else {
		store = metrics.New([]models.Metrics{})
	}

	fileStorage := file.New(store, cfg)

	fileStorage.ReadMetrics()

	return store, fileStorage
}
