package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/dglazkoff/go-metrics/cmd/server/logger"
	"github.com/dglazkoff/go-metrics/cmd/server/storage"
	"github.com/dglazkoff/go-metrics/internal/models"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func getMetricValueInRequest(store *storage.MemStorage) http.HandlerFunc {
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

func getMetricValueInBody(store *storage.MemStorage) http.HandlerFunc {
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

		// почему тут выставление заголовка заработало а ниже не работало?
		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)

		if metric.MType == "gauge" {
			value, ok := store.GaugeMetrics[metric.ID]

			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			if err := enc.Encode(models.Metrics{ID: metric.ID, MType: metric.MType, Value: &value}); err != nil {
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

			if err := enc.Encode(models.Metrics{ID: metric.ID, MType: metric.MType, Delta: &value}); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		// w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}
}
