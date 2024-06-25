package handlers

import (
	"github.com/dglazkoff/go-metrics/cmd/server/storage"
	"github.com/go-chi/chi/v5"
	"net/http"
)

// разбить на хендлеры и сервисы

func updateMetricValue(store *storage.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")
		metricValue := chi.URLParam(r, "metricValue")

		if metricType != "gauge" && metricType != "counter" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if metricType == "gauge" {
			err := store.GaugeMetrics.Save(metricName, metricValue)

			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}

		if metricType == "counter" {
			err := store.CounterMetrics.Save(metricName, metricValue)

			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}
