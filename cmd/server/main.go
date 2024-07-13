package main

import (
	"github.com/dglazkoff/go-metrics/cmd/server/config"
	"github.com/dglazkoff/go-metrics/internal/logger"
)

func main() {
	cfg := config.ParseConfig()
	err := logger.Initialize()

	if err != nil {
		panic(err)
	}

	if err := Run(&cfg); err != nil {
		logger.Log.Debug("Error on starting server", "err", err)
	}
}
