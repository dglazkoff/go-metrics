package handlers

import (
	"github.com/dglazkoff/go-metrics/cmd/server/logger"
	"github.com/dglazkoff/go-metrics/cmd/server/storage"
	"github.com/go-chi/chi/v5"
)

func Router(store *storage.MemStorage) chi.Router {
	r := chi.NewRouter()

	r.Post("/update", logger.Log.Request(updateMetricValueInBody(store)))
	r.Post("/update/{metricType}/{metricName}/{metricValue}", logger.Log.Request(updateMetricValueInRequest(store)))

	r.Post("/value", logger.Log.Request(getMetricValueInBody(store)))
	r.Get("/value/{metricType}/{metricName}", logger.Log.Request(getMetricValueInRequest(store)))

	r.Get("/", logger.Log.Request(getHTML(store)))

	return r
}
