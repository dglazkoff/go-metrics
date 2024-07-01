package handlers

import (
	"encoding/json"
	"github.com/dglazkoff/go-metrics/cmd/server/logger"
	"github.com/dglazkoff/go-metrics/cmd/server/storage"
	"github.com/dglazkoff/go-metrics/internal/models"
	"net/http"
)

// разбить на хендлеры и сервисы

func updateMetricValue(store *storage.MemStorage) http.HandlerFunc {
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

		if metric.MType == "gauge" {
			// протестировать что не передал и протестировать неправильный формат
			if metric.Value == nil {
				logger.Log.Debug("Required Value field for gauge metric type")
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			store.GaugeMetrics.Save(metric.ID, metric.Value)
		}

		if metric.MType == "counter" {
			if metric.Delta == nil {
				logger.Log.Debug("Required Delta field for counter metric type")
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			store.CounterMetrics.Save(metric.ID, metric.Delta)
		}

		w.WriteHeader(http.StatusOK)
	}
}
