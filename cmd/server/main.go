package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dglazkoff/go-metrics/cmd/server/config"
	"github.com/dglazkoff/go-metrics/internal/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/grpc"
)

var (
	BuildVersion = "N/A"
	BuildDate    = "N/A"
	BuildCommit  = "N/A"
)

func runApp(cfg *config.Config) error {
	err := logger.Initialize()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	fmt.Printf("Build version: %s\n", BuildVersion)
	fmt.Printf("Build date: %s\n", BuildDate)
	fmt.Printf("Build commit: %s\n", BuildCommit)

	sigs := make(chan os.Signal, 1)
	errChan := make(chan error, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	var grpcServer *grpc.Server
	var httpServer *http.Server

	if cfg.IsGRPC {
		grpcServer = RunGRPCServer(cfg, errChan)
	} else {
		httpServer = RunHTTPServer(cfg, errChan)
	}

	select {
	case err := <-errChan:
		logger.Log.Debug("Server error occurred: ", err)
		return err
	case sig := <-sigs:
		logger.Log.Debug("Signal: ", sig)
	}

	if httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(ctx); err != nil {
			logger.Log.Debug("Server Shutdown Failed", "err", err)
		}
	}

	if grpcServer != nil {
		grpcServer.GracefulStop()
	}

	logger.Log.Infow("Server exited properly")
	return nil
}

// go run -ldflags "-X main.BuildVersion=v1.0.1 -X 'main.BuildDate=$(date +'%Y/%m/%d %H:%M:%S')'" ./cmd/server
func main() {
	cfg := config.ParseConfig()
	if err := runApp(&cfg); err != nil {
		panic(err)
	}
}
