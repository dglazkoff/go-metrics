package main

import (
	"github.com/dglazkoff/go-metrics/cmd/server/config"
	"github.com/dglazkoff/go-metrics/cmd/server/logger"
	"github.com/dglazkoff/go-metrics/cmd/server/router"
	"github.com/dglazkoff/go-metrics/cmd/server/services/file"
	"github.com/dglazkoff/go-metrics/cmd/server/storage/metrics"
	"github.com/dglazkoff/go-metrics/internal/models"
	"net/http"
)

func Run(cfg *config.Config) error {
	store := metrics.New([]models.Metrics{})

	fileService := file.New(&store, cfg)
	fileService.ReadMetrics()

	logger.Log.Infow("Starting Server on ", "addr", cfg.RunAddr)

	if cfg.StoreInterval != 0 {
		go fileService.WriteMetrics(true)
	}

	return http.ListenAndServe(cfg.RunAddr, router.Router(&store, cfg))
}
