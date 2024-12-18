package file

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/dglazkoff/go-metrics/cmd/server/config"
	"github.com/dglazkoff/go-metrics/internal/logger"
	"github.com/dglazkoff/go-metrics/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockStorage struct {
	metrics []models.Metrics
	err     error
}

func (m *MockStorage) SaveMetrics(ctx context.Context, metrics []models.Metrics) error {
	m.metrics = metrics
	return m.err
}

func (m *MockStorage) ReadMetrics(ctx context.Context) ([]models.Metrics, error) {
	return m.metrics, m.err
}

func TestReadMetrics(t *testing.T) {
	err := logger.Initialize()
	require.NoError(t, err)

	floatValue := 42.5
	intValue := int64(5)

	tests := []struct {
		name           string
		cfg            *config.Config
		metrics        []models.Metrics
		expectSaveCall bool
		err            error
	}{
		{
			name: "Successful read and save",
			cfg: &config.Config{
				IsRestore:       true,
				FileStoragePath: "test.json",
			},
			metrics: []models.Metrics{
				{ID: "metric1", MType: "gauge", Value: &floatValue},
				{ID: "metric2", MType: "counter", Delta: &intValue},
			},
			expectSaveCall: true,
		},
		{
			name: "IsRestore is false",
			cfg: &config.Config{
				IsRestore:       false,
				FileStoragePath: "test.json",
			},
			metrics: []models.Metrics{
				{ID: "metric1", MType: "gauge", Value: &floatValue},
				{ID: "metric2", MType: "counter", Delta: &intValue},
			},
			expectSaveCall: false,
		},
		{
			name: "Finish function on error",
			err:  errors.New("error"),
			cfg: &config.Config{
				IsRestore:       false,
				FileStoragePath: "test.json",
			},
			metrics: []models.Metrics{
				{ID: "metric1", MType: "gauge", Value: &floatValue},
				{ID: "metric2", MType: "counter", Delta: &intValue},
			},
			expectSaveCall: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := MockStorage{err: tt.err}
			dir, _ := os.Getwd()
			path := filepath.Join(dir, tt.cfg.FileStoragePath)

			if tt.metrics != nil {
				file, _ := os.Create(path)
				err = json.NewEncoder(file).Encode(tt.metrics)
				assert.NoError(t, err)
				file.Close()

				defer os.Remove(path)
			}

			s := fileStorage{cfg: tt.cfg, storage: &mockStorage}
			s.ReadMetrics()

			if tt.expectSaveCall {
				assert.Equal(t, tt.metrics, mockStorage.metrics)
			} else {
				assert.Nil(t, mockStorage.metrics)
			}
		})
	}
}

func TestWriteMetrics(t *testing.T) {
	err := logger.Initialize()
	require.NoError(t, err)

	floatValue := 42.5
	intValue := int64(5)

	tests := []struct {
		name            string
		cfg             *config.Config
		storage         *MockStorage
		expectedMetrics []models.Metrics
		expectFile      bool
	}{
		{
			name: "Successful write metrics",
			cfg: &config.Config{
				FileStoragePath: "test.json",
				StoreInterval:   1,
			},
			storage: &MockStorage{
				metrics: []models.Metrics{
					{ID: "metric1", MType: "gauge", Value: &floatValue},
					{ID: "metric2", MType: "counter", Delta: &intValue},
				},
			},
			expectFile: true,
			expectedMetrics: []models.Metrics{
				{ID: "metric1", MType: "gauge", Value: &floatValue},
				{ID: "metric2", MType: "counter", Delta: &intValue},
			},
		},
		{
			name: "Successful if no FileStoragePath",
			cfg: &config.Config{
				FileStoragePath: "",
				StoreInterval:   1,
			},
			storage:         &MockStorage{},
			expectFile:      false,
			expectedMetrics: []models.Metrics{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, _ := os.Getwd()
			path := filepath.Join(dir, tt.cfg.FileStoragePath)
			defer os.Remove(path)

			s := fileStorage{cfg: tt.cfg, storage: tt.storage}
			s.WriteMetrics(false)

			if tt.expectFile {
				file, err := os.Open(path)
				assert.NoError(t, err)
				defer file.Close()

				var metrics []models.Metrics
				err = json.NewDecoder(file).Decode(&metrics)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedMetrics, metrics)
			} else if tt.cfg.FileStoragePath != "" {
				_, err := os.Stat(path)
				assert.True(t, os.IsNotExist(err))
			}
		})
	}
}

func TestWriteMetrics_ReadMetricsError(t *testing.T) {
	err := logger.Initialize()
	require.NoError(t, err)

	mockStorage := MockStorage{err: errors.New("error")}

	s := fileStorage{
		cfg:     &config.Config{FileStoragePath: "mock/path"},
		storage: &mockStorage,
	}

	s.WriteMetrics(false)
}
