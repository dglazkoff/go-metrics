package main

import (
	"database/sql"
	"github.com/dglazkoff/go-metrics/cmd/server/config"
	"github.com/dglazkoff/go-metrics/cmd/server/router"
	"github.com/dglazkoff/go-metrics/cmd/server/storage/file"
	"github.com/dglazkoff/go-metrics/cmd/server/storage/metrics"
	"github.com/dglazkoff/go-metrics/internal/logger"
	"github.com/dglazkoff/go-metrics/internal/models"
	"net/http"
)

func Run(cfg *config.Config) error {
	db, err := sql.Open("pgx", cfg.DatabaseDSN)

	if err != nil {
		logger.Log.Debug("Error on starting db", "err", err)
	}
	defer db.Close()

	store := metrics.New([]models.Metrics{}, db)

	fileStorage := file.New(&store, cfg)
	fileStorage.ReadMetrics()

	logger.Log.Infow("Starting Server on ", "addr", cfg.RunAddr)

	if cfg.StoreInterval != 0 {
		go fileStorage.WriteMetrics(true)
	}

	return http.ListenAndServe(cfg.RunAddr, router.Router(&store, &fileStorage, cfg))
}
