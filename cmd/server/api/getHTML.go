package api

import (
	"context"
	"net/http"

	"github.com/dglazkoff/go-metrics/cmd/server/html"
	"github.com/dglazkoff/go-metrics/internal/logger"
)

// GetHTML - хендлер получения html страницы с метриками
func (a API) GetHTML() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")

		metrics, err := a.metricsService.GetAll(r.Context())

		if err != nil {
			logger.Log.Debug("Error while get all metrics: ", err)
		}

		component := html.Metrics(metrics)
		component.Render(context.Background(), w)

		w.WriteHeader(http.StatusOK)
	}
}
