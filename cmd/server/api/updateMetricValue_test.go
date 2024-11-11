package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/dglazkoff/go-metrics/cmd/server/config"
	"github.com/dglazkoff/go-metrics/cmd/server/services/service"
	"github.com/dglazkoff/go-metrics/cmd/server/storage/file"
	"github.com/dglazkoff/go-metrics/cmd/server/storage/metrics"
	constants "github.com/dglazkoff/go-metrics/internal/const"
	"github.com/dglazkoff/go-metrics/internal/logger"
	"github.com/dglazkoff/go-metrics/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPI_UpdateMetricValue(t *testing.T) {
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
		name        string
		store       []models.Metrics
		metric      models.Metrics
		metricValue string
		want        want
	}{
		{
			name:   "success test",
			store:  []models.Metrics{},
			metric: models.Metrics{ID: "value", MType: constants.MetricTypeCounter, Delta: &deltaValue},
			want: want{
				store:  []models.Metrics{{ID: "value", MType: constants.MetricTypeCounter, Delta: &deltaValue}},
				status: http.StatusOK,
			},
		},
		{
			name:   "wrong metric type",
			store:  []models.Metrics{},
			metric: models.Metrics{ID: "value", MType: "wrong", Delta: &deltaValue},
			want: want{
				store:  []models.Metrics{},
				status: http.StatusBadRequest,
			},
		},
		{
			name:   "add counter to previous result",
			store:  []models.Metrics{{ID: "value", MType: constants.MetricTypeCounter, Delta: &deltaValue}},
			metric: models.Metrics{ID: "value", MType: constants.MetricTypeCounter, Delta: &deltaValue},
			want: want{
				store:  []models.Metrics{{ID: "value", MType: constants.MetricTypeCounter, Delta: &deltaResultValue}},
				status: http.StatusOK,
			},
		},
		{
			name:   "update gauge metric",
			store:  []models.Metrics{{ID: "value", MType: constants.MetricTypeGauge, Value: &value}},
			metric: models.Metrics{ID: "value", MType: constants.MetricTypeGauge, Value: &value},
			want: want{
				store:  []models.Metrics{{ID: "value", MType: constants.MetricTypeGauge, Value: &value}},
				status: http.StatusOK,
			},
		},
		{
			name:   "add gauge metric",
			store:  []models.Metrics{{ID: "value", MType: constants.MetricTypeGauge, Value: &value}},
			metric: models.Metrics{ID: "value1", MType: constants.MetricTypeGauge, Value: &value},
			want: want{
				store:  []models.Metrics{{ID: "value", MType: constants.MetricTypeGauge, Value: &value}, {ID: "value1", MType: constants.MetricTypeGauge, Value: &value}},
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

			ts := httptest.NewServer(newAPI.UpdateMetricValueInBody())
			defer ts.Close()

			var b bytes.Buffer
			err = json.NewEncoder(&b).Encode(tt.metric)
			require.NoError(t, err)

			request, err := http.NewRequest(http.MethodPost, ts.URL, &b)
			require.NoError(t, err)

			result, err := ts.Client().Do(request)
			require.NoError(t, err)

			err = result.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.want.status, result.StatusCode)

			res, err := store.ReadMetrics(context.Background())
			require.NoError(t, err)
			assert.Equal(t, tt.want.store, res)

			clean()
		})

		t.Run(tt.name, func(t *testing.T) {
			err := logger.Initialize()
			require.NoError(t, err)

			store := metrics.New(tt.store)
			fileStore := file.New(store, &cfg)
			metricService := service.New(store, fileStore, &cfg)
			newAPI := NewAPI(metricService, &cfg)

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("metricType", tt.metric.MType)
				rctx.URLParams.Add("metricName", tt.metric.ID)

				if tt.metric.Value != nil {
					if tt.metricValue != "" {
						rctx.URLParams.Add("metricValue", tt.metricValue)
					} else {
						rctx.URLParams.Add("metricValue", strconv.FormatFloat(*tt.metric.Value, 'f', -1, 64))
					}
				}

				if tt.metric.Delta != nil {
					rctx.URLParams.Add("metricValue", strconv.FormatInt(*tt.metric.Delta, 10))
				}

				r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

				newAPI.UpdateMetricValueInRequest()(w, r)
			}))
			defer ts.Close()

			request, err := http.NewRequest(http.MethodPost, ts.URL, nil)
			require.NoError(t, err)

			result, err := ts.Client().Do(request)
			require.NoError(t, err)

			err = result.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.want.status, result.StatusCode)

			res, err := store.ReadMetrics(context.Background())
			require.NoError(t, err)
			assert.Equal(t, tt.want.store, res)
		})
	}
}

func TestUpdateMetricValueInRequest_Error(t *testing.T) {
	cfg := config.Config{
		RunAddr:         ":8080",
		FileStoragePath: "/tmp/metrics-db.json",
		StoreInterval:   300,
		IsRestore:       true,
		DatabaseDSN:     "",
		SecretKey:       "",
	}
	var deltaValue int64 = 1
	var value float64 = 101

	type want struct {
		status int
	}

	tests := []struct {
		name        string
		metric      models.Metrics
		metricValue string
		want        want
	}{
		{
			name:        "wrong gauge metric value",
			metric:      models.Metrics{ID: "value1", MType: constants.MetricTypeGauge, Value: &value},
			metricValue: "wrong",
			want: want{
				status: http.StatusBadRequest,
			},
		},
		{
			name:        "wrong counter metric value",
			metric:      models.Metrics{ID: "value1", MType: constants.MetricTypeCounter, Delta: &deltaValue},
			metricValue: "wrong",
			want: want{
				status: http.StatusBadRequest,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := logger.Initialize()
			require.NoError(t, err)

			store := metrics.New([]models.Metrics{})
			fileStore := file.New(store, &cfg)
			metricService := service.New(store, fileStore, &cfg)
			newAPI := NewAPI(metricService, &cfg)

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("metricType", tt.metric.MType)
				rctx.URLParams.Add("metricName", tt.metric.ID)

				if tt.metric.Value != nil {
					if tt.metricValue != "" {
						rctx.URLParams.Add("metricValue", tt.metricValue)
					} else {
						rctx.URLParams.Add("metricValue", strconv.FormatFloat(*tt.metric.Value, 'f', -1, 64))
					}
				}

				if tt.metric.Delta != nil {
					if tt.metricValue != "" {
						rctx.URLParams.Add("metricValue", tt.metricValue)
					} else {
						rctx.URLParams.Add("metricValue", strconv.FormatInt(*tt.metric.Delta, 10))
					}
				}

				r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

				newAPI.UpdateMetricValueInRequest()(w, r)
			}))
			defer ts.Close()

			request, err := http.NewRequest(http.MethodPost, ts.URL, nil)
			require.NoError(t, err)

			result, err := ts.Client().Do(request)
			require.NoError(t, err)

			err = result.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.want.status, result.StatusCode)
		})
	}
}
