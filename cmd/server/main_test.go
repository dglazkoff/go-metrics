package main

import (
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/dglazkoff/go-metrics/cmd/server/config"
	"github.com/dglazkoff/go-metrics/internal/logger"
	"github.com/stretchr/testify/assert"
)

func TestRunApp_HTTPServer(t *testing.T) {
	err := logger.Initialize()
	assert.NoError(t, err)

	cfg := &config.Config{IsGRPC: false}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	go func() {
		err := runApp(cfg)
		assert.NoError(t, err)
	}()

	time.Sleep(100 * time.Millisecond)
	process, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("Failed to find current process: %v", err)
	}
	process.Signal(syscall.SIGTERM)

	time.Sleep(100 * time.Millisecond)
}

func TestRunApp_GRPCServer(t *testing.T) {
	err := logger.Initialize()
	assert.NoError(t, err)

	cfg := &config.Config{IsGRPC: true}

	go func() {
		err := runApp(cfg)
		assert.NoError(t, err)
	}()

	time.Sleep(100 * time.Millisecond)
	process, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("Failed to find current process: %v", err)
	}
	process.Signal(syscall.SIGTERM)

	time.Sleep(100 * time.Millisecond)
}
