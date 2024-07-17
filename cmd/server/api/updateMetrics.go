package api

import (
	"encoding/json"
	"github.com/dglazkoff/go-metrics/internal/logger"
	"github.com/dglazkoff/go-metrics/internal/models"
	"net/http"
)

func (a API) UpdateList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var metrics []models.Metrics

		if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
			logger.Log.Debug("Error while decode: ", err)
			return
		}

		err := a.metricsService.UpdateList(metrics)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
