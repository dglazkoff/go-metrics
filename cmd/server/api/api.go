package api

import (
	"context"
	"github.com/dglazkoff/go-metrics/cmd/server/config"
	"github.com/dglazkoff/go-metrics/internal/models"
)

type metric interface {
	Get(name string) (models.Metrics, error)
	GetAll() ([]models.Metrics, error)
	Update(metric models.Metrics) error
	PingDB(ctx context.Context) error
}

// не делаем экспортируемых полей чтобы скрыть
type API struct {
	metricsService metric
	cfg            *config.Config
}

func NewAPI(m metric, cfg *config.Config) API {
	return API{metricsService: m, cfg: cfg}
}
