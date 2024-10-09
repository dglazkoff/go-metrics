// Пакет storage предоставляет интерфейсы для работы с хранилищами
package storage

import (
	"context"

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
