package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	sigs := make(chan os.Signal, 1)
	errChan := make(chan error)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	fmt.Printf("Build version: %s\n", BuildVersion)
	fmt.Printf("Build date: %s\n", BuildDate)
	fmt.Printf("Build commit: %s\n", BuildCommit)

	server := Run(&cfg, errChan)

	select {
	case err := <-errChan:
		logger.Log.Debug("Server error occurred: ", err)
	case sig := <-sigs:
		logger.Log.Debug("Signal: ", sig)
	}

	// увидел такой способ, но не понимаю, как он работает
	// посмотреть пример из шатдауна, там немного другое флоу, разобраться в чем разница и как он работает
	// почему у меня ListenAndServe не завершал работу программы?
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Log.Debug("Server Shutdown Failed", "err", err)
	}

	logger.Log.Infow("Server exited properly")
}
