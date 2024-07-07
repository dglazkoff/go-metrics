package router

import (
	"github.com/dglazkoff/go-metrics/cmd/server/api"
	"github.com/dglazkoff/go-metrics/cmd/server/config"
	"github.com/dglazkoff/go-metrics/cmd/server/gzip"
	"github.com/dglazkoff/go-metrics/cmd/server/logger"
	"github.com/dglazkoff/go-metrics/cmd/server/services/metrics"
	"github.com/dglazkoff/go-metrics/cmd/server/storage"
	"github.com/go-chi/chi/v5"
)

func Router(store storage.MetricsStorage, cfg *config.Config) chi.Router {
	r := chi.NewRouter()
	metricService := metrics.New(store, cfg)
	newApi := api.NewAPI(metricService, cfg)

	r.Post("/update/", logger.Log.Request(gzip.GzipHandle(newApi.UpdateMetricValueInBody())))
	r.Post("/update/{metricType}/{metricName}/{metricValue}", logger.Log.Request(gzip.GzipHandle(newApi.UpdateMetricValueInRequest())))

	r.Post("/value/", logger.Log.Request(gzip.GzipHandle(newApi.GetMetricValueInBody())))
	r.Get("/value/{metricType}/{metricName}", logger.Log.Request(gzip.GzipHandle(newApi.GetMetricValueInRequest())))

	r.Get("/", logger.Log.Request(newApi.GetHTML()))

	return r
}
