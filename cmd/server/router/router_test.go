package router

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dglazkoff/go-metrics/cmd/server/config"
	"github.com/dglazkoff/go-metrics/cmd/server/storage/file"
	"github.com/dglazkoff/go-metrics/cmd/server/storage/metrics"
	"github.com/dglazkoff/go-metrics/internal/logger"
	"github.com/dglazkoff/go-metrics/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type metricService struct{}

func (m *metricService) Get(ctx context.Context, name string) (models.Metrics, error) {
	return models.Metrics{}, nil
}

func (m *metricService) GetAll(ctx context.Context) ([]models.Metrics, error) {
	return []models.Metrics{}, nil
}

func (m *metricService) Update(ctx context.Context, metric models.Metrics) error {
	return nil
}

func (m *metricService) UpdateList(ctx context.Context, metric []models.Metrics) error {
	return nil
}

func (m *metricService) PingDB(ctx context.Context) error {
	return nil
}

func TestRouter(t *testing.T) {
	err := logger.Initialize()
	require.NoError(t, err)

	cfg := &config.Config{}

	store := metrics.New([]models.Metrics{})
	fileStore := file.New(store, cfg)

	router := Router(store, fileStore, cfg)

	tests := []struct {
		name         string
		method       string
		url          string
		body         []byte
		expectedCode int
	}{
		{
			name:         "POST /update/",
			method:       http.MethodPost,
			url:          "/update/",
			body:         []byte(`{"type":"gauge","id":"test_metric","value":100.5}`),
			expectedCode: http.StatusOK,
		},
		{
			name:         "POST /update/{metricType}/{metricName}/{metricValue}",
			method:       http.MethodPost,
			url:          "/update/gauge/test_metric/100.5",
			expectedCode: http.StatusOK,
		},
		{
			name:         "POST /updates/",
			method:       http.MethodPost,
			url:          "/updates/",
			body:         []byte(`[{"type":"gauge","id":"test_metric","value":100.5}]`),
			expectedCode: http.StatusOK,
		},
		{
			name:         "POST /value/",
			method:       http.MethodPost,
			url:          "/value/",
			body:         []byte(`{"type":"gauge","id":"test_metric"}`),
			expectedCode: http.StatusOK,
		},
		{
			name:         "GET /value/{metricType}/{metricName}",
			method:       http.MethodGet,
			url:          "/value/gauge/test_metric",
			expectedCode: http.StatusOK,
		},
		{
			name:         "GET /",
			method:       http.MethodGet,
			url:          "/",
			expectedCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.url, bytes.NewReader(tt.body))
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedCode, rec.Code)
		})
	}
}
