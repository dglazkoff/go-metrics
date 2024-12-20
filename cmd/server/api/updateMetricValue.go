package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	constants "github.com/dglazkoff/go-metrics/internal/const"
	"github.com/dglazkoff/go-metrics/internal/logger"
	"github.com/dglazkoff/go-metrics/internal/models"
	"github.com/go-chi/chi/v5"
)

// UpdateMetricValueInRequest - хендлер обновления метрики, передаваемой в URLParams
func (a API) UpdateMetricValueInRequest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")
		metricValue := chi.URLParam(r, "metricValue")

		var model models.Metrics

		if metricType == constants.MetricTypeGauge {
			floatValue, err := strconv.ParseFloat(metricValue, 64)

			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				logger.Log.Debug("Error when parse metric value", err)
				return
			}

			model = models.Metrics{ID: metricName, MType: metricType, Value: &floatValue}
		}

		if metricType == constants.MetricTypeCounter {
			intValue, err := strconv.ParseInt(metricValue, 10, 64)

			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				logger.Log.Debug("Error when parse metric value", err)
				return
			}

			model = models.Metrics{ID: metricName, MType: metricType, Delta: &intValue}
		}

		err := a.metricsService.Update(r.Context(), model)

		if err != nil {
			logger.Log.Debug("Error while update metric: ", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// UpdateMetricValueInBody - хендлер обновления метрики, передаваемой в body
func (a API) UpdateMetricValueInBody() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var metric models.Metrics

		if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
			logger.Log.Debug("Error while decode: ", err)
			return
		}

		err := a.metricsService.Update(r.Context(), metric)

		if err != nil {
			logger.Log.Debug("Error while update metric: ", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
