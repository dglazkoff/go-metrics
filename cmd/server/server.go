package main

import (
	"github.com/dglazkoff/go-metrics/cmd/server/handlers"
	"github.com/dglazkoff/go-metrics/cmd/server/storage"
	"net/http"
)

func Run() error {
	store := storage.MemStorage{GaugeMetrics: make(map[string]float64), CounterMetrics: make(map[string]int64)}

	mux := http.NewServeMux()
	mux.HandleFunc("/update/{metricType}/{metricName}/{metricValue}", handlers.UpdateMetricValue(&store))

	return http.ListenAndServe(":8080", mux)
}
