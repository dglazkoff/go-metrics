package api

import (
	"net/http"

	"github.com/dglazkoff/go-metrics/internal/logger"
)

func (a API) PingDB() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := a.metricsService.PingDB(r.Context())

		if err != nil {
			logger.Log.Debug("Error on ping db ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
