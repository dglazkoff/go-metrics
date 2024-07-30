package api

import (
	"encoding/json"
	"fmt"
	"github.com/dglazkoff/go-metrics/internal/const"
	"github.com/dglazkoff/go-metrics/internal/logger"
	"github.com/dglazkoff/go-metrics/internal/models"
	"github.com/go-chi/chi/v5"
	"net/http"
)

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

		// не нравится что логика работы со значениями метрики в хендлере

		/*
			fmt.Sprint делает преобразование "всего что угодно" в строку, и тут могут быть проблемы в постедствии, когда структура хранения усложнится
			лучше всегда использовать явное преобразование, что бы читать кода всегда видел из какого типа в какой идет преобрзование, в данном случае подойдет fmt.Sprintf("%d", value) тут явным образом ожидается число
		*/

		w.Header().Set("Content-Type", "text/plain")
		// Fprintf вызывает Write и после этого нельзя проставлять заголовок. даже вызвав WriteHeader

		if value.Delta != nil {
			fmt.Fprintf(w, "%d", *value.Delta)
		}
		//
		if value.Value != nil {
			fmt.Fprintf(w, "%g", *value.Value)
		}

		// почему тут заголовок не устанавливается ?
		// w.Header().Set("Content-Type", "text/plain")

		// If WriteHeader is not called explicitly, the first call to Write
		// will trigger an implicit WriteHeader(http.StatusOK).
		// w.Header().Set("HashSHA256", "hash")
		// даже статус не выводится нужный

		// по сути если есть Write то WriteHeader бесполезно вызывать
		// w.WriteHeader(http.StatusOK)
	}
}

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

		// @tmvrus почему тут выставление заголовка заработало, а ниже не работало?

		// @tmvrus: на сколько я помню все хедеры должны быть записаны до вызова Write
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

		// w.Header().Set("Content-Type", "application/json")
		// w.WriteHeader(http.StatusOK)
	}
}
