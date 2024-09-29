package api

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/dglazkoff/go-metrics/cmd/server/config"
	"github.com/dglazkoff/go-metrics/cmd/server/services/service"
	"github.com/dglazkoff/go-metrics/cmd/server/storage/file"
	"github.com/dglazkoff/go-metrics/cmd/server/storage/metrics"
	constants "github.com/dglazkoff/go-metrics/internal/const"
	"github.com/dglazkoff/go-metrics/internal/logger"
	"github.com/dglazkoff/go-metrics/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPI_UpdateMetrics(t *testing.T) {
	type want struct {
		store  []models.Metrics
		status int
	}

	cfg := config.Config{
		RunAddr:         ":8080",
		FileStoragePath: "/tmp/metrics-db.json",
		StoreInterval:   300,
		IsRestore:       true,
		DatabaseDSN:     "",
		SecretKey:       "",
	}
	var deltaValue int64 = 1
	var deltaResultValue int64 = 2
	var value float64 = 101

	clean := func() {
		deltaValue = 1
		deltaResultValue = 2
		value = 101
	}

	t.Cleanup(clean)

	tests := []struct {
		name    string
		store   []models.Metrics
		metrics []models.Metrics
		request string
		want    want
	}{
		{
			name:  "simple test",
			store: []models.Metrics{},
			metrics: []models.Metrics{
				{ID: "valueCounter", MType: constants.MetricTypeCounter, Delta: &deltaValue},
				{ID: "valueGauge", MType: constants.MetricTypeGauge, Value: &value},
			},
			want: want{
				store: []models.Metrics{
					{ID: "valueCounter", MType: constants.MetricTypeCounter, Delta: &deltaValue},
					{ID: "valueGauge", MType: constants.MetricTypeGauge, Value: &value},
				},
				status: http.StatusOK,
			},
		},
		{
			name: "update previous metrics",
			store: []models.Metrics{
				{ID: "valueCounter", MType: constants.MetricTypeCounter, Delta: &deltaResultValue},
				{ID: "valueGauge", MType: constants.MetricTypeGauge, Value: &value},
			},
			metrics: []models.Metrics{
				{ID: "valueCounter", MType: constants.MetricTypeCounter, Delta: &deltaValue},
				{ID: "valueGauge", MType: constants.MetricTypeGauge, Value: &value},
			},
			want: want{
				store: []models.Metrics{
					{ID: "valueCounter", MType: constants.MetricTypeCounter, Delta: &deltaResultValue},
					{ID: "valueGauge", MType: constants.MetricTypeGauge, Value: &value},
				},
				status: http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := logger.Initialize()
			require.NoError(t, err)

			store := metrics.New(tt.store)
			fileStore := file.New(store, &cfg)
			metricService := service.New(store, fileStore, &cfg)
			newAPI := NewAPI(metricService, &cfg)

			ts := httptest.NewServer(newAPI.UpdateList())
			defer ts.Close()

			var b bytes.Buffer
			err = json.NewEncoder(&b).Encode(tt.metrics)
			require.NoError(t, err)

			request, err := http.NewRequest(http.MethodPost, ts.URL, &b)
			require.NoError(t, err)

			result, err := ts.Client().Do(request)
			require.NoError(t, err)

			err = result.Body.Close()
			require.NoError(t, err)

			res, err := store.ReadMetrics(context.Background())
			require.NoError(t, err)
			assert.Equal(t, tt.want.store, res)
		})
	}
}
