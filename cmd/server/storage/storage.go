package storage

import "github.com/dglazkoff/go-metrics/internal/models"

type MetricsStorage interface {
	ReadMetric(name string) (models.Metrics, error)
	ReadMetrics() []models.Metrics
	UpdateMetric(metric models.Metrics) error
	SaveMetrics(metrics []models.Metrics)
}

type FileStorage interface {
	WriteMetrics(isLoop bool)
	ReadMetrics()
}
