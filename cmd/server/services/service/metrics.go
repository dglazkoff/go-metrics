// слой service
package service

import (
	"context"
	"errors"

	"github.com/dglazkoff/go-metrics/cmd/server/config"
	"github.com/dglazkoff/go-metrics/cmd/server/storage"
	constants "github.com/dglazkoff/go-metrics/internal/const"
	"github.com/dglazkoff/go-metrics/internal/logger"
	"github.com/dglazkoff/go-metrics/internal/models"
)

type fileStorage interface {
	WriteMetrics(isLoop bool)
}

type metricStorage interface {
	ReadMetric(ctx context.Context, name string) (models.Metrics, error)
	ReadMetrics(ctx context.Context) ([]models.Metrics, error)
	UpdateMetric(ctx context.Context, metric models.Metrics) error
	PingDB(ctx context.Context) error
}

type service struct {
	storage     metricStorage
	fileStorage fileStorage
	cfg         *config.Config
}

// New - метод для создания сервиса
func New(s storage.MetricsStorage, f fileStorage, cfg *config.Config) service {
	return service{storage: s, cfg: cfg, fileStorage: f}
}

// Get - метод для получения метрики по имени
func (s service) Get(ctx context.Context, name string) (models.Metrics, error) {
	return s.storage.ReadMetric(ctx, name)
}

// GetAll - метод для получения всех метрик
func (s service) GetAll(ctx context.Context) ([]models.Metrics, error) {
	return s.storage.ReadMetrics(ctx)
}

// Update - метод для обновления метрики
func (s service) Update(ctx context.Context, metric models.Metrics) error {
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

	err := s.storage.UpdateMetric(ctx, metric)
	if s.cfg.StoreInterval == 0 && err == nil {
		s.fileStorage.WriteMetrics(false)
	}

	return err
}

// UpdateList - метод для обновления списка метрик
func (s service) UpdateList(ctx context.Context, metrics []models.Metrics) error {
	for _, metric := range metrics {
		err := s.Update(ctx, metric)
		if err != nil {
			logger.Log.Debug("Error while updating metric ", err)
			return err
		}
	}

	return nil
}

// PingDB - метод для проверки соединения с БД
func (s service) PingDB(ctx context.Context) error {
	return s.storage.PingDB(ctx)
}
