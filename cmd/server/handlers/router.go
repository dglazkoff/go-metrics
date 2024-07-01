package handlers

import (
	"github.com/dglazkoff/go-metrics/cmd/server/logger"
	"github.com/dglazkoff/go-metrics/cmd/server/storage"
	"github.com/go-chi/chi/v5"
)

func Router(store *storage.MemStorage) chi.Router {
	r := chi.NewRouter()

	r.Post("/update", logger.Log.Request(updateMetricValue(store)))
	r.Post("/value", logger.Log.Request(getMetricValue(store)))
	r.Get("/", logger.Log.Request(getHTML(store)))

	return r
}
