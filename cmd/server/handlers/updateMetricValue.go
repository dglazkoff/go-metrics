package handlers

import (
	"github.com/dglazkoff/go-metrics/cmd/server/storage"
	"net/http"
)

func UpdateMetricValue(store *storage.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		metricType := r.PathValue("metricType")
		metricName := r.PathValue("metricName")
		metricValue := r.PathValue("metricValue")

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
