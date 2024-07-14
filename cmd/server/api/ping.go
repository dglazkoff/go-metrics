package api

import (
	"net/http"
)

func (a API) PingDB() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := a.metricsService.PingDB(r.Context())

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
