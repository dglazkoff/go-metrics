package metrics

import (
	"context"
	"fmt"
	"github.com/dglazkoff/go-metrics/internal/const"
	"github.com/dglazkoff/go-metrics/internal/models"
)

type storage struct {
	metrics []models.Metrics
}

func New(metrics []models.Metrics) *storage {
	storeMetrics := append([]models.Metrics{}, metrics...)

	return &storage{
		metrics: storeMetrics,
	}
}

func (s *storage) ReadMetrics() ([]models.Metrics, error) {
	return s.metrics, nil
}

func (s *storage) ReadMetric(name string) (models.Metrics, error) {
	for _, metric := range s.metrics {
		if metric.ID == name {
			return metric, nil
		}
	}

	return models.Metrics{}, fmt.Errorf("metric not found by name %s", name)
}

func (s *storage) UpdateMetric(metric models.Metrics) error {
	for i, m := range s.metrics {
		if m.ID == metric.ID {
			if metric.MType == constants.MetricTypeGauge {

				s.metrics[i] = metric
				return nil
			}

			if metric.MType == constants.MetricTypeCounter {
				*s.metrics[i].Delta += *metric.Delta
				return nil
			}

			return fmt.Errorf("unknown metric type %s", metric.MType)
		}
	}

	s.metrics = append(s.metrics, metric)
	return nil
}

func (s *storage) SaveMetrics(metrics []models.Metrics) {
	s.metrics = append(s.metrics, metrics...)
}

func (s *storage) PingDB(ctx context.Context) error {
	return nil
}
