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
	newAPI := api.NewAPI(metricService, cfg)

	r.Post("/update/", logger.Log.Request(gzip.GzipHandle(newAPI.UpdateMetricValueInBody(), false)))
	r.Post("/update/{metricType}/{metricName}/{metricValue}", logger.Log.Request(gzip.GzipHandle(newAPI.UpdateMetricValueInRequest(), false)))

	r.Post("/value/", logger.Log.Request(gzip.GzipHandle(newAPI.GetMetricValueInBody(), false)))
	r.Get("/value/{metricType}/{metricName}", logger.Log.Request(gzip.GzipHandle(newAPI.GetMetricValueInRequest(), false)))

	r.Get("/", logger.Log.Request(gzip.GzipHandle(newAPI.GetHTML(), true)))

	return r
}
