package metrics

import (
	"context"
	"errors"
	"github.com/dglazkoff/go-metrics/cmd/server/config"
	"github.com/dglazkoff/go-metrics/cmd/server/storage"
	"github.com/dglazkoff/go-metrics/internal/const"
	"github.com/dglazkoff/go-metrics/internal/logger"
	"github.com/dglazkoff/go-metrics/internal/models"
)

type fileStorage interface {
	WriteMetrics(isLoop bool)
}

type metricStorage interface {
	ReadMetric(name string) (models.Metrics, error)
	ReadMetrics() []models.Metrics
	UpdateMetric(metric models.Metrics) error
	PingDB(ctx context.Context) error
}

type service struct {
	storage     metricStorage
	fileStorage fileStorage
	cfg         *config.Config
}

func New(s storage.MetricsStorage, f fileStorage, cfg *config.Config) service {
	return service{storage: s, cfg: cfg, fileStorage: f}
}

func (s service) Get(name string) (models.Metrics, error) {
	return s.storage.ReadMetric(name)
}

func (s service) GetAll() []models.Metrics {
	return s.storage.ReadMetrics()
}

func (s service) Update(metric models.Metrics) error {
	if metric.MType != constants.MetricTypeGauge && metric.MType != constants.MetricTypeCounter {
		logger.Log.Debug("Wrong type")
		return errors.New("wrong type")
	}

	if metric.MType == constants.MetricTypeGauge {
		// протестировать что не передал и протестировать неправильный формат
		if metric.Value == nil {
			logger.Log.Debug("Required Value field for gauge metric type")
			return errors.New("required Value field for gauge metric type")
		}
	}

	if metric.MType == constants.MetricTypeCounter {
		if metric.Delta == nil {
			logger.Log.Debug("Required Delta field for counter metric type")
			return errors.New("required Delta field for counter metric type")
		}
	}

	err := s.storage.UpdateMetric(metric)
	if s.cfg.StoreInterval == 0 && err == nil {
		s.fileStorage.WriteMetrics(false)
	}

	return err
}

func (s service) PingDB(ctx context.Context) error {
	return s.storage.PingDB(ctx)
}
