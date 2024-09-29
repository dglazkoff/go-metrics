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
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestAPI_GetMetricValue(t *testing.T) {
	type want struct {
		metric models.Metrics
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

	clean := func() {
		deltaValue = 1
	}

	t.Cleanup(clean)

	tests := []struct {
		name    string
		store   []models.Metrics
		metric  models.Metrics
		request string
		want    want
	}{
		{
			name:   "simple test",
			store:  []models.Metrics{{ID: "value", MType: constants.MetricTypeCounter, Delta: &deltaValue}},
			metric: models.Metrics{ID: "value", MType: constants.MetricTypeCounter},
			want: want{
				metric: models.Metrics{ID: "value", MType: constants.MetricTypeCounter, Delta: &deltaValue},
				status: http.StatusOK,
			},
		},
		{
			name:   "wrong metric type",
			store:  []models.Metrics{{ID: "value", MType: constants.MetricTypeCounter, Delta: &deltaValue}},
			metric: models.Metrics{ID: "value", MType: "wrong", Delta: &deltaValue},
			want: want{
				metric: models.Metrics{},
				status: http.StatusBadRequest,
			},
		},
		{
			name:   "not found metric",
			store:  []models.Metrics{{ID: "value", MType: constants.MetricTypeCounter, Delta: &deltaValue}},
			metric: models.Metrics{ID: "value1", MType: constants.MetricTypeCounter},
			want: want{
				metric: models.Metrics{},
				status: http.StatusNotFound,
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

			ts := httptest.NewServer(newAPI.GetMetricValueInBody())
			defer ts.Close()

			var b bytes.Buffer
			err = json.NewEncoder(&b).Encode(tt.metric)
			require.NoError(t, err)

			request, err := http.NewRequest(http.MethodPost, ts.URL, &b)
			require.NoError(t, err)

			result, err := ts.Client().Do(request)
			require.NoError(t, err)

			defer result.Body.Close()
			assert.Equal(t, tt.want.status, result.StatusCode)

			var resultMetric models.Metrics

			if result.Body == http.NoBody {
				if (tt.want.metric != models.Metrics{}) {
					t.Fatalf("Expected non-empty response body")
				}

				return
			}

			err = json.NewDecoder(result.Body).Decode(&resultMetric)
			require.NoError(t, err)

			//err = result.Body.Close()
			// require.NoError(t, err)

			assert.Equal(t, tt.want.metric, resultMetric)

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

				r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

				newAPI.GetMetricValueInRequest()(w, r)
			}))
			defer ts.Close()

			request, err := http.NewRequest(http.MethodPost, ts.URL, nil)
			require.NoError(t, err)

			result, err := ts.Client().Do(request)
			require.NoError(t, err)

			defer result.Body.Close()
			assert.Equal(t, tt.want.status, result.StatusCode)

			if result.Body == http.NoBody {
				if (tt.want.metric != models.Metrics{}) {
					t.Errorf("Expected non-empty response body")
				}

				return
			}

			var resultMetric models.Metrics

			if result.Header.Get("Content-Type") == "text/plain" {
				responseData, err := io.ReadAll(result.Body)
				assert.NoError(t, err)

				responseInt, err := strconv.Atoi(string(responseData))
				assert.NoError(t, err)

				assert.Equal(t, int64(responseInt), *tt.want.metric.Delta)
				return
			}

			err = json.NewDecoder(result.Body).Decode(&resultMetric)
			require.NoError(t, err)

			assert.Equal(t, tt.want.metric, resultMetric)
		})
	}
}
