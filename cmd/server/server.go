package main

import (
	"fmt"
	"github.com/dglazkoff/go-metrics/cmd/server/handlers"
	"github.com/dglazkoff/go-metrics/cmd/server/storage"
	"net/http"
)

func Run() error {
	store := storage.MemStorage{GaugeMetrics: make(map[string]float64), CounterMetrics: make(map[string]int64)}

	fmt.Println("Running server on ", flagRunAddr)
	return http.ListenAndServe(flagRunAddr, handlers.Router(&store))
}
