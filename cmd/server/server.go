package main

import (
	"database/sql"
	"github.com/dglazkoff/go-metrics/cmd/server/config"
	"github.com/dglazkoff/go-metrics/cmd/server/router"
	"github.com/dglazkoff/go-metrics/cmd/server/storage"
	"github.com/dglazkoff/go-metrics/cmd/server/storage/db"
	"github.com/dglazkoff/go-metrics/cmd/server/storage/file"
	"github.com/dglazkoff/go-metrics/cmd/server/storage/metrics"
	"github.com/dglazkoff/go-metrics/internal/logger"
	"github.com/dglazkoff/go-metrics/internal/models"
	"net/http"
)

func Run(cfg *config.Config) error {
	pgDB, err := sql.Open("pgx", cfg.DatabaseDSN)

	if err != nil {
		logger.Log.Debug("Error on starting db", "err", err)
	}
	defer pgDB.Close()

	store := metrics.New([]models.Metrics{}, pgDB)

	var staticStorage storage.StaticStorage

	if cfg.DatabaseDSN != "" {
		_, err = pgDB.Exec("CREATE TABLE IF NOT EXISTS metrics (id VARCHAR(250) PRIMARY KEY, type VARCHAR(250) NOT NULL, value DOUBLE PRECISION, delta INTEGER)")

		if err != nil {
			logger.Log.Debug("error while creating table ", err)
		}

		dbStorage := db.New(pgDB, &store, cfg)
		staticStorage = dbStorage
	} else {
		fileStorage := file.New(&store, cfg)
		staticStorage = fileStorage
	}

	staticStorage.ReadMetrics()

	logger.Log.Infow("Starting Server on ", "addr", cfg.RunAddr)

	if cfg.StoreInterval != 0 {
		go staticStorage.WriteMetrics(true)
	}

	return http.ListenAndServe(cfg.RunAddr, router.Router(&store, staticStorage, cfg))
}
