package handlers

import (
	"encoding/json"
	"github.com/dglazkoff/go-metrics/cmd/server/logger"
	"github.com/dglazkoff/go-metrics/cmd/server/storage"
	"github.com/dglazkoff/go-metrics/internal/models"
	"net/http"
)

func getMetricValue(store *storage.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var metric models.Metrics
		if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
			logger.Log.Debug("Error while decode", err)
		}

		if metric.MType != "gauge" && metric.MType != "counter" {
			logger.Log.Debug("Wrong type")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		enc := json.NewEncoder(w)

		if metric.MType == "gauge" {
			value, ok := store.GaugeMetrics[metric.ID]

			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			if err := enc.Encode(value); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		if metric.MType == "counter" {
			value, ok := store.CounterMetrics[metric.ID]

			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			if err := enc.Encode(value); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
	}
}
