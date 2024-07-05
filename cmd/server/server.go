package main

import (
	"github.com/dglazkoff/go-metrics/cmd/server/flags"
	"github.com/dglazkoff/go-metrics/cmd/server/handlers"
	"github.com/dglazkoff/go-metrics/cmd/server/logger"
	"github.com/dglazkoff/go-metrics/cmd/server/storage"
	"net/http"
)

func Run() error {
	store := storage.MemStorage{GaugeMetrics: make(map[string]float64), CounterMetrics: make(map[string]int64)}
	storage.ReadMetrics(&store)

	logger.Log.Infow("Starting Server on ", "addr", flags.FlagRunAddr)

	if flags.FlagStoreInterval != 0 {
		go storage.WriteMetrics(&store, true)
	}

	return http.ListenAndServe(flags.FlagRunAddr, handlers.Router(&store))
}
