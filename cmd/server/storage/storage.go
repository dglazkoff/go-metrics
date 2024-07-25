package storage

import (
	"context"
	"github.com/dglazkoff/go-metrics/internal/models"
)

type MetricsStorage interface {
	ReadMetric(ctx context.Context, name string) (models.Metrics, error)
	ReadMetrics(ctx context.Context) ([]models.Metrics, error)
	UpdateMetric(ctx context.Context, metric models.Metrics) error
	SaveMetrics(ctx context.Context, metrics []models.Metrics) error
	PingDB(ctx context.Context) error
}

type FileStorage interface {
	WriteMetrics(isLoop bool)
	ReadMetrics()
}
