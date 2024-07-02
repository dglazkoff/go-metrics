package handlers

import (
	"github.com/dglazkoff/go-metrics/cmd/server/gzip"
	"github.com/dglazkoff/go-metrics/cmd/server/logger"
	"github.com/dglazkoff/go-metrics/cmd/server/storage"
	"github.com/go-chi/chi/v5"
)

func Router(store *storage.MemStorage) chi.Router {
	r := chi.NewRouter()

	r.Post("/update/", logger.Log.Request(gzip.GzipHandle(updateMetricValueInBody(store))))
	r.Post("/update/{metricType}/{metricName}/{metricValue}", logger.Log.Request(gzip.GzipHandle(updateMetricValueInRequest(store))))

	r.Post("/value/", logger.Log.Request(gzip.GzipHandle(getMetricValueInBody(store))))
	r.Get("/value/{metricType}/{metricName}", logger.Log.Request(gzip.GzipHandle(getMetricValueInRequest(store))))

	r.Get("/", logger.Log.Request(getHTML(store)))

	return r
}
