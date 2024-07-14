package api

import (
	"context"
	"github.com/dglazkoff/go-metrics/cmd/server/html"
	"net/http"
)

func (a API) GetHTML() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// @tmvrus почему если писать тот же код как ниже, то не вставлялся заголовок??
		w.Header().Set("Content-Type", "text/html")

		component := html.Metrics(a.metricsService.GetAll())
		component.Render(context.Background(), w)
		// w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
	}
}