package handlers

import (
	"fmt"
	"github.com/dglazkoff/go-metrics/cmd/server/storage"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func getMetricValue(store *storage.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")

		if metricType != "gauge" && metricType != "counter" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var result string

		if metricType == "gauge" {
			value, ok := store.GaugeMetrics[metricName]

			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			result = fmt.Sprint(value)
		}

		if metricType == "counter" {
			value, ok := store.CounterMetrics[metricName]

			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			result = fmt.Sprint(value)
		}

		w.Write([]byte(result))
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
	}
}
