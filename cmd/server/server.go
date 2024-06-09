package main

import (
	"github.com/dglazkoff/go-metrics/cmd/server/handlers"
	"net/http"
)

func Run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/update/{metricType}/{metricName}/{metricValue}", handlers.UpdateMetricValue)

	return http.ListenAndServe(":8080", mux)
}
