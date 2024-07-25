package api

import (
	"context"
	"github.com/dglazkoff/go-metrics/cmd/server/html"
	"github.com/dglazkoff/go-metrics/internal/logger"
	"net/http"
)

func (a API) GetHTML() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// @tmvrus почему если писать тот же код как ниже, то не вставлялся заголовок??
		w.Header().Set("Content-Type", "text/html")

		metrics, err := a.metricsService.GetAll(r.Context())

		if err != nil {
			logger.Log.Debug("Error while get all metrics: ", err)
		}

		component := html.Metrics(metrics)
		component.Render(context.Background(), w)
		// w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
	}
}
