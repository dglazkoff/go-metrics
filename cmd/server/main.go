package main

import (
	"github.com/dglazkoff/go-metrics/cmd/server/flags"
	"github.com/dglazkoff/go-metrics/cmd/server/logger"
)

func main() {
	flags.ParseFlags()
	err := logger.Initialize()

	if err != nil {
		panic(err)
	}

	if err := Run(); err != nil {
		logger.Log.Infow("Error on starting server", "err", err)
		panic(err)
	}
}
