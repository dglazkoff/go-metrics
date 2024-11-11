package storage

import (
	"testing"

	"github.com/dglazkoff/go-metrics/cmd/server/config"
	"github.com/dglazkoff/go-metrics/internal/logger"
	"github.com/stretchr/testify/assert"
)

//func TestInitStorages_WithDatabaseDSN(t *testing.T) {
//	// Create a mock database connection
//	db, mock, err := sqlmock.New()
//	if err != nil {
//		t.Fatalf("Error creating mock database: %v", err)
//	}
//	defer db.Close()
//
//	// Mock the bootstrap function if needed
//	mock.ExpectQuery("SELECT 1").WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
//
//	// Create a mock config
//	cfg := &config.Config{
//		DatabaseDSN: "mockDSN",
//	}
//
//	// Call InitStorages
//	store, fileStorage := InitStorages(cfg, db)
//
//	// Assertions
//	assert.NotNil(t, store, "Expected MetricsStorage to be initialized")
//	assert.NotNil(t, fileStorage, "Expected FileStorage to be initialized")
//
//	// Ensure all expectations were met
//	if err := mock.ExpectationsWereMet(); err != nil {
//		t.Errorf("Not all SQL expectations were met: %v", err)
//	}
//}

func TestInitStorages_WithoutDatabaseDSN(t *testing.T) {
	err := logger.Initialize()
	assert.NoError(t, err)

	cfg := &config.Config{
		DatabaseDSN: "",
	}

	store, fileStorage := InitStorages(cfg)

	assert.NotNil(t, store, "Expected MetricsStorage to be initialized")
	assert.NotNil(t, fileStorage, "Expected FileStorage to be initialized")
}
