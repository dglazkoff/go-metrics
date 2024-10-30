package main

import (
	"database/sql"
	"errors"
	"net/http"
	_ "net/http/pprof"

	"github.com/dglazkoff/go-metrics/cmd/server/config"
	"github.com/dglazkoff/go-metrics/cmd/server/router"
	"github.com/dglazkoff/go-metrics/cmd/server/storage"
	"github.com/dglazkoff/go-metrics/cmd/server/storage/db"
	"github.com/dglazkoff/go-metrics/cmd/server/storage/file"
	"github.com/dglazkoff/go-metrics/cmd/server/storage/metrics"
	"github.com/dglazkoff/go-metrics/internal/logger"
	"github.com/dglazkoff/go-metrics/internal/models"
)

func Run(cfg *config.Config, errChan chan<- error) *http.Server {
	pgDB, err := sql.Open("pgx", cfg.DatabaseDSN)

	if err != nil {
		logger.Log.Debug("Error on open db", "err", err)
		panic(err)
	}
	defer pgDB.Close()

	var store storage.MetricsStorage

	if cfg.DatabaseDSN != "" {
		dbStore := db.New(pgDB, cfg)
		err = db.Bootstrap(dbStore)

		if err != nil {
			logger.Log.Debug("Error on bootstrap db ", err)
			panic(err)
		}

		store = dbStore
	} else {
		store = metrics.New([]models.Metrics{})
	}

	fileStorage := file.New(store, cfg)

	fileStorage.ReadMetrics()

	logger.Log.Infow("Starting Server on ", "addr", cfg.RunAddr)

	if cfg.StoreInterval != 0 {
		go fileStorage.WriteMetrics(true)
	}

	server := &http.Server{
		Addr:    cfg.RunAddr,
		Handler: router.Router(store, fileStorage, cfg),
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Log.Debug("Error on running server", "err", err)
			errChan <- err
		}
	}()

	return server
}
