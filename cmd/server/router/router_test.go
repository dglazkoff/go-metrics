package router

import (
	"context"
	"net/http"

	"github.com/dglazkoff/go-metrics/cmd/server/api"
	"github.com/dglazkoff/go-metrics/cmd/server/config"
	"github.com/dglazkoff/go-metrics/internal/models"
)

type metricService struct{}

func (m *metricService) Get(ctx context.Context, name string) (models.Metrics, error) {
	return models.Metrics{}, nil
}

func (m *metricService) GetAll(ctx context.Context) ([]models.Metrics, error) {
	return []models.Metrics{}, nil
}

func (m *metricService) Update(ctx context.Context, metric models.Metrics) error {
	return nil
}

func (m *metricService) UpdateList(ctx context.Context, metric []models.Metrics) error {
	return nil
}

func (m *metricService) PingDB(ctx context.Context) error {
	return nil
}

func Example() {
	cfg := &config.Config{}
	ms := &metricService{}
	newAPI := api.NewAPI(ms, cfg)

	http.Handle("/updates/", newAPI.UpdateList())
	http.Handle("/update/", newAPI.UpdateMetricValueInBody())
	http.Handle("/update/{metricType}/{metricName}/{metricValue}", newAPI.UpdateMetricValueInRequest())
	http.Handle("/value/", newAPI.GetMetricValueInBody())
	http.Handle("/value/{metricType}/{metricName}", newAPI.GetMetricValueInRequest())
}
