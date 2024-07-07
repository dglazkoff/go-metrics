package api

import (
	"github.com/dglazkoff/go-metrics/cmd/server/config"
	"github.com/dglazkoff/go-metrics/internal/models"
)

type metric interface {
	Get(name string) (models.Metrics, error)
	GetAll() []models.Metrics
	Update(metric models.Metrics) error
}

// не делаем экспортируемых полей чтобы скрыть
type API struct {
	metricsService metric
	cfg            *config.Config
}

func NewAPI(m metric, cfg *config.Config) API {
	return API{metricsService: m, cfg: cfg}
}
