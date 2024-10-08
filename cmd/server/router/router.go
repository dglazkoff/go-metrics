package router

import (
	"github.com/dglazkoff/go-metrics/cmd/server/api"
	"github.com/dglazkoff/go-metrics/cmd/server/bodyhash"
	"github.com/dglazkoff/go-metrics/cmd/server/config"
	"github.com/dglazkoff/go-metrics/cmd/server/gzip"
	"github.com/dglazkoff/go-metrics/cmd/server/services/metrics"
	"github.com/dglazkoff/go-metrics/cmd/server/storage"
	"github.com/dglazkoff/go-metrics/internal/logger"
	"github.com/go-chi/chi/v5"
)

func Router(store storage.MetricsStorage, fs storage.FileStorage, cfg *config.Config) chi.Router {
	r := chi.NewRouter()

	metricService := metrics.New(store, fs, cfg)
	newAPI := api.NewAPI(metricService, cfg)
	bh := bodyhash.Initialize(cfg)

	r.Post("/update/", logger.Log.Request(bh.BodyHash(gzip.GzipHandle(newAPI.UpdateMetricValueInBody(), false))))
	r.Post("/update/{metricType}/{metricName}/{metricValue}", logger.Log.Request(bh.BodyHash(gzip.GzipHandle(newAPI.UpdateMetricValueInRequest(), false))))

	r.Post("/updates/", logger.Log.Request(bh.BodyHash(gzip.GzipHandle(newAPI.UpdateList(), false))))

	r.Post("/value/", logger.Log.Request(bh.BodyHash(gzip.GzipHandle(newAPI.GetMetricValueInBody(), false))))
	r.Get("/value/{metricType}/{metricName}", logger.Log.Request(bh.BodyHash(gzip.GzipHandle(newAPI.GetMetricValueInRequest(), false))))

	r.Get("/", logger.Log.Request(bh.BodyHash(gzip.GzipHandle(newAPI.GetHTML(), true))))

	r.Get("/ping", logger.Log.Request(newAPI.PingDB()))

	return r
}
