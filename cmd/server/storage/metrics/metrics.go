package metrics

import (
	"fmt"
	"github.com/dglazkoff/go-metrics/internal/models"
)

type storage []models.Metrics

func New(metrics []models.Metrics) storage {
	stor := append([]models.Metrics{}, metrics...)

	return stor
}

func (s *storage) ReadMetrics() []models.Metrics {
	return *s
}

func (s *storage) ReadMetric(name string) (models.Metrics, error) {
	for _, metric := range *s {
		if metric.ID == name {
			return metric, nil
		}
	}

	return models.Metrics{}, fmt.Errorf("metric not found by name %s", name)
}

func (s *storage) UpdateMetric(metric models.Metrics) error {
	for i, m := range *s {
		if m.ID == metric.ID {
			if metric.MType == "gauge" {
				(*s)[i] = metric
				return nil
			}

			if metric.MType == "counter" {
				*(*s)[i].Delta += *metric.Delta
				return nil
			}

			return fmt.Errorf("unknown metric type %s", metric.MType)
		}
	}

	*s = append(*s, metric)
	return nil
}
