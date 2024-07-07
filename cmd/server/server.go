package main

import (
	"github.com/dglazkoff/go-metrics/cmd/server/config"
	"github.com/dglazkoff/go-metrics/cmd/server/handlers"
	"github.com/dglazkoff/go-metrics/cmd/server/logger"
	"github.com/dglazkoff/go-metrics/cmd/server/storage"
	"net/http"
)

func Run(cfg *config.Config) error {
	store := storage.MemStorage{GaugeMetrics: make(map[string]float64), CounterMetrics: make(map[string]int64)}
	storage.ReadMetrics(&store, cfg)

	logger.Log.Infow("Starting Server on ", "addr", cfg.RunAddr)

	if cfg.StoreInterval != 0 {
		go storage.WriteMetrics(&store, true, cfg)
	}

	return http.ListenAndServe(cfg.RunAddr, handlers.Router(&store, cfg))
}
