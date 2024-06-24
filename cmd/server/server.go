package main

import (
	"github.com/dglazkoff/go-metrics/cmd/server/handlers"
	"github.com/dglazkoff/go-metrics/cmd/server/logger"
	"github.com/dglazkoff/go-metrics/cmd/server/storage"
	"net/http"
)

func Run() error {
	store := storage.MemStorage{GaugeMetrics: make(map[string]float64), CounterMetrics: make(map[string]int64)}

	// fmt.Println("Running server on ", flagRunAddr)
	logger.Log.Infow("Starting Server on ", "addr", flagRunAddr)

	return http.ListenAndServe(flagRunAddr, handlers.Router(&store))
}
