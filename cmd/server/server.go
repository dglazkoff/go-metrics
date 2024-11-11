package main

import (
	"errors"
	"net/http"
	_ "net/http/pprof"

	"github.com/dglazkoff/go-metrics/cmd/server/config"
	"github.com/dglazkoff/go-metrics/cmd/server/router"
	"github.com/dglazkoff/go-metrics/cmd/server/storage"
)

func RunHTTPServer(cfg *config.Config, errChan chan<- error) *http.Server {
	store, fileStorage, err := storage.InitStorages(cfg)

	if err != nil {
		errChan <- err
		return nil
	}

	// logger.Log.Infow("Starting HTTP Server on ", "addr", cfg.RunAddr)

	if cfg.StoreInterval != 0 {
		go fileStorage.WriteMetrics(true)
	}

	server := &http.Server{
		Addr:    cfg.RunAddr,
		Handler: router.Router(store, fileStorage, cfg),
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
	}()

	return server
}
