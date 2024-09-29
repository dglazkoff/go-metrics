package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	constants "github.com/dglazkoff/go-metrics/internal/const"
	"github.com/dglazkoff/go-metrics/internal/logger"
	"github.com/dglazkoff/go-metrics/internal/models"
	"github.com/go-chi/chi/v5"
)

// GetMetricValueInRequest - хендлер для получения метрики по данным в URLParams
func (a API) GetMetricValueInRequest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")

		if metricType != constants.MetricTypeGauge && metricType != constants.MetricTypeCounter {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		value, err := a.metricsService.Get(r.Context(), metricName)

		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		/*
			fmt.Sprint делает преобразование "всего что угодно" в строку, и тут могут быть проблемы в постедствии, когда структура хранения усложнится
			лучше всегда использовать явное преобразование, что бы читать кода всегда видел из какого типа в какой идет преобрзование, в данном случае подойдет fmt.Sprintf("%d", value) тут явным образом ожидается число
		*/

		w.Header().Set("Content-Type", "text/plain")
		// Fprintf вызывает Write и после этого нельзя проставлять заголовок. даже вызвав WriteHeader

		if value.Delta != nil {
			fmt.Fprintf(w, "%d", *value.Delta)
		}

		if value.Value != nil {
			fmt.Fprintf(w, "%g", *value.Value)
		}
	}
}

// GetMetricValueInBody - хендлер для получения метрики по данным в body
func (a API) GetMetricValueInBody() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var metric models.Metrics
		if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
			logger.Log.Debug("Error while decode", err)
			return
		}

		if metric.MType != constants.MetricTypeGauge && metric.MType != constants.MetricTypeCounter {
			logger.Log.Debug("Wrong type")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		enc := json.NewEncoder(w)

		value, err := a.metricsService.Get(r.Context(), metric.ID)

		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if err = enc.Encode(value); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return

		}
	}
}
