package handlers

import (
	"context"
	"github.com/dglazkoff/go-metrics/cmd/server/html"
	"github.com/dglazkoff/go-metrics/cmd/server/storage"
	"net/http"
)

func getHTML(store *storage.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// w.Write([]byte(html.Hello("John")))
		component := html.Metrics(store)
		component.Render(context.Background(), w)
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
	}
}
