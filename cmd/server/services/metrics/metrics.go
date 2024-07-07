package metrics

import (
	"errors"
	"github.com/dglazkoff/go-metrics/cmd/server/config"
	"github.com/dglazkoff/go-metrics/cmd/server/logger"
	"github.com/dglazkoff/go-metrics/cmd/server/storage"
	"github.com/dglazkoff/go-metrics/internal/models"
)

type metricStorage interface {
	ReadMetric(name string) (models.Metrics, error)
	ReadMetrics() []models.Metrics
	UpdateMetric(metric models.Metrics) error
}

type service struct {
	storage metricStorage
	cfg     *config.Config
}

func New(s storage.MetricsStorage, cfg *config.Config) service {
	return service{storage: s, cfg: cfg}
}

func (s service) Get(name string) (models.Metrics, error) {
	return s.storage.ReadMetric(name)
}

func (s service) GetAll() []models.Metrics {
	return s.storage.ReadMetrics()
}

func (s service) Update(metric models.Metrics) error {
	if metric.MType != "gauge" && metric.MType != "counter" {
		logger.Log.Debug("Wrong type")
		return errors.New("wrong type")
	}

	if metric.MType == "gauge" {
		// протестировать что не передал и протестировать неправильный формат
		if metric.Value == nil {
			logger.Log.Debug("Required Value field for gauge metric type")
			return errors.New("required Value field for gauge metric type")
		}
	}

	if metric.MType == "counter" {
		if metric.Delta == nil {
			logger.Log.Debug("Required Delta field for counter metric type")
			return errors.New("required Delta field for counter metric type")
		}
	}

	return s.storage.UpdateMetric(metric)

	//if cfg.StoreInterval == 0 {
	//	file.WriteMetrics(store, false, cfg)
	//}
}
