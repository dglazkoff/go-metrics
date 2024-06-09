package handlers

import (
	"github.com/dglazkoff/go-metrics/cmd/server/storage"
	"github.com/go-chi/chi/v5"
)

func Router(store *storage.MemStorage) chi.Router {
	r := chi.NewRouter()

	r.Post("/update/{metricType}/{metricName}/{metricValue}", UpdateMetricValue(store))
	r.Get("/value/{metricType}/{metricName}", GetMetricValue(store))
	r.Get("/", GetHTML(store))

	return r
}
