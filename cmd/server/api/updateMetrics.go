package api

import (
	"encoding/json"
	"net/http"

	"github.com/dglazkoff/go-metrics/internal/logger"
	"github.com/dglazkoff/go-metrics/internal/models"
)

func (a API) UpdateList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var metrics []models.Metrics

		if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
			logger.Log.Debug("Error while decode: ", err)
			return
		}

		err := a.metricsService.UpdateList(r.Context(), metrics)

		/*
			UpdateList использует метод Update. там где я его вызываю тоже возвращаю StatusBadRequest.
			Думаю сделал потому что могут прислать MType неверный, и в таком случае кажется 400 ошибка подходит.
			Но также может от БД придти ошибка, и тогда вероятно всего 500 должно вернуть. Как такие случаи обрабатывать?
			Из сервиса же не вернешь код ошибки, заводить какие-то уникальные ошибки если MType неверный и если словили такую ошибку то возвращать 400, в противном случае 500?


			именно так, каждому типу ошибки должен соответствовать свой статус
		*/
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
