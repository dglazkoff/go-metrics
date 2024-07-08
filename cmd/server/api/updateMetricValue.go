package api

import (
	"encoding/json"
	"github.com/dglazkoff/go-metrics/cmd/server/logger"
	"github.com/dglazkoff/go-metrics/internal/models"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

func (a API) UpdateMetricValueInRequest() http.HandlerFunc {
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

		err := a.metricsService.Update(model)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func (a API) UpdateMetricValueInBody() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var metric models.Metrics

		if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
			logger.Log.Debug("Error while decode: ", err)
			return
		}

		err := a.metricsService.Update(metric)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
