package handlers

import (
	"encoding/json"
	"github.com/dglazkoff/go-metrics/cmd/server/flags"
	"github.com/dglazkoff/go-metrics/cmd/server/logger"
	"github.com/dglazkoff/go-metrics/cmd/server/storage"
	"github.com/dglazkoff/go-metrics/internal/models"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

// разбить на хендлеры и сервисы

func updateMetricValueInRequest(store *storage.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")
		metricValue := chi.URLParam(r, "metricValue")

		var model models.Metrics

		if metricType == "gauge" {
			floatValue, err := strconv.ParseFloat(metricValue, 64)

			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				logger.Log.Debug("Error when parse metric value", err)
				return
			}

			model = models.Metrics{ID: metricName, MType: metricType, Value: &floatValue}
		}

		if metricType == "counter" {
			intValue, err := strconv.ParseInt(metricValue, 10, 64)

			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				logger.Log.Debug("Error when parse metric value", err)
				return
			}

			model = models.Metrics{ID: metricName, MType: metricType, Delta: &intValue}
		}

		updateMetricValue(store, model, w, r)
	}
}

func updateMetricValueInBody(store *storage.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var metric models.Metrics

		if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
			logger.Log.Debug("Error while decode: ", err)
			return
		}

		updateMetricValue(store, metric, w, r)
	}
}

func updateMetricValue(store *storage.MemStorage, metric models.Metrics, w http.ResponseWriter, r *http.Request) {
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

	if flags.FlagStoreInterval == 0 {
		storage.WriteMetrics(store, false)
	}

	w.WriteHeader(http.StatusOK)
}
