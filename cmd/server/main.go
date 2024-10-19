package main

import (
	"fmt"

	"github.com/dglazkoff/go-metrics/cmd/server/config"
	"github.com/dglazkoff/go-metrics/internal/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	BuildVersion = "N/A"
	BuildDate    = "N/A"
	BuildCommit  = "N/A"
)

// go run -ldflags "-X main.BuildVersion=v1.0.1 -X 'main.BuildDate=$(date +'%Y/%m/%d %H:%M:%S')'" ./cmd/server
func main() {
	cfg := config.ParseConfig()
	err := logger.Initialize()

	if err != nil {
		panic(err)
	}

	fmt.Printf("Build version: %s\n", BuildVersion)
	fmt.Printf("Build date: %s\n", BuildDate)
	fmt.Printf("Build commit: %s\n", BuildCommit)

	if err := Run(&cfg); err != nil {
		logger.Log.Debug("Error on starting server", "err", err)
	}
}
