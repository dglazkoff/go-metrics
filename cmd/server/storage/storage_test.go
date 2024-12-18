package storage

import (
	"testing"

	"github.com/dglazkoff/go-metrics/cmd/server/config"
	"github.com/dglazkoff/go-metrics/internal/logger"
	"github.com/stretchr/testify/assert"
)

func TestInitStorages_WithDatabaseDSN(t *testing.T) {
	err := logger.Initialize()
	assert.NoError(t, err)

	cfg := &config.Config{
		DatabaseDSN: "wrong_dsn",
	}

	store, fileStorage, err := InitStorages(cfg)

	assert.Nil(t, store)
	assert.Nil(t, fileStorage)
	assert.Error(t, err)
}

func TestInitStorages_WithoutDatabaseDSN(t *testing.T) {
	err := logger.Initialize()
	assert.NoError(t, err)

	cfg := &config.Config{
		DatabaseDSN: "",
	}

	store, fileStorage, err := InitStorages(cfg)

	assert.NotNil(t, store, "Expected MetricsStorage to be initialized")
	assert.NotNil(t, fileStorage, "Expected FileStorage to be initialized")
	assert.NoError(t, err)
}
