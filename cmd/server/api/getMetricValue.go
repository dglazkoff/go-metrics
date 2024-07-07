package api

import (
	"encoding/json"
	"fmt"
	"github.com/dglazkoff/go-metrics/cmd/server/logger"
	"github.com/dglazkoff/go-metrics/internal/models"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func (a API) GetMetricValueInRequest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")

		if metricType != "gauge" && metricType != "counter" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		value, err := a.metricsService.Get(metricName)

		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if value.Delta != nil {
			w.Write([]byte(fmt.Sprint(*value.Delta)))
		}

		if value.Value != nil {
			w.Write([]byte(fmt.Sprint(*value.Value)))
		}

		/*
			fmt.Sprint делает преобразование "всего что угодно" в строку, и тут могут быть проблемы в постедствии, когда структура хранения усложнится
			лучше всегда использовать явное преобразование, что бы читать кода всегда видел из какого типа в какой идет преобрзование, в данном случае подойдет fmt.Sprintf("%d", value) тут явным образом ожидается число
		*/

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
	}
}

func (a API) GetMetricValueInBody() http.HandlerFunc {
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

		value, err := a.metricsService.Get(metric.ID)

		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if err := enc.Encode(value); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}
}
