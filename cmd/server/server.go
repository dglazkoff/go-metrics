package main

import (
	"github.com/dglazkoff/go-metrics/cmd/server/config"
	"github.com/dglazkoff/go-metrics/cmd/server/router"
	"github.com/dglazkoff/go-metrics/cmd/server/storage/file"
	"github.com/dglazkoff/go-metrics/cmd/server/storage/metrics"
	"github.com/dglazkoff/go-metrics/internal/logger"
	"github.com/dglazkoff/go-metrics/internal/models"
	"net/http"
)

func Run(cfg *config.Config) error {
	store := metrics.New([]models.Metrics{})

	fileStorage := file.New(&store, cfg)
	fileStorage.ReadMetrics()

	logger.Log.Infow("Starting Server on ", "addr", cfg.RunAddr)

	if cfg.StoreInterval != 0 {
		go fileStorage.WriteMetrics(true)
	}

	return http.ListenAndServe(cfg.RunAddr, router.Router(&store, &fileStorage, cfg))
}
