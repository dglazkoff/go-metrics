package router

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dglazkoff/go-metrics/cmd/server/config"
	"github.com/dglazkoff/go-metrics/cmd/server/storage/file"
	"github.com/dglazkoff/go-metrics/cmd/server/storage/metrics"
	constants "github.com/dglazkoff/go-metrics/internal/const"
	"github.com/dglazkoff/go-metrics/internal/logger"
	"github.com/dglazkoff/go-metrics/internal/models"
)

//func TestUpdateMetricValue(t *testing.T) {
//	type want struct {
//		store      []models.Metrics
//		statusCode int
//	}
//
//	cfg := config.ParseConfig()
//	var deltaValue int64 = 1
//	var deltaResultValue int64 = 2
//	var value float64 = 101
//
//	tests := []struct {
//		name    string
//		method  string
//		store   []models.Metrics
//		metric  models.Metrics
//		request string
//		want    want
//	}{
//		{
//			name:   "success test",
//			method: http.MethodPost,
//			store:  []models.Metrics{},
//			metric: models.Metrics{ID: "value", MType: constants.MetricTypeCounter, Delta: &deltaValue},
//			want: want{
//				store:      []models.Metrics{{ID: "value", MType: constants.MetricTypeCounter, Delta: &deltaValue}},
//				statusCode: http.StatusOK,
//			},
//		},
//		{
//			name:   "invalid method GET",
//			method: http.MethodGet,
//			store:  []models.Metrics{},
//			metric: models.Metrics{ID: "value", MType: constants.MetricTypeCounter, Delta: &deltaValue},
//			want: want{
//				store:      []models.Metrics{},
//				statusCode: http.StatusMethodNotAllowed,
//			},
//		},
//		{
//			name:   "add counter to previous result",
//			method: http.MethodPost,
//			store:  []models.Metrics{{ID: "value", MType: constants.MetricTypeCounter, Delta: &deltaValue}},
//			metric: models.Metrics{ID: "value", MType: constants.MetricTypeCounter, Delta: &deltaValue},
//			want: want{
//				store:      []models.Metrics{{ID: "value", MType: constants.MetricTypeCounter, Delta: &deltaResultValue}},
//				statusCode: http.StatusOK,
//			},
//		},
//		{
//			name:   "update gauge metric",
//			method: http.MethodPost,
//			store:  []models.Metrics{{ID: "value", MType: constants.MetricTypeGauge, Value: &value}},
//			metric: models.Metrics{ID: "value", MType: constants.MetricTypeGauge, Value: &value},
//			want: want{
//				store:      []models.Metrics{{ID: "value", MType: constants.MetricTypeGauge, Value: &value}},
//				statusCode: http.StatusOK,
//			},
//		},
//		{
//			name:   "add gauge metric",
//			method: http.MethodPost,
//			store:  []models.Metrics{{ID: "value", MType: constants.MetricTypeGauge, Value: &value}},
//			metric: models.Metrics{ID: "value1", MType: constants.MetricTypeGauge, Value: &value},
//			want: want{
//				store:      []models.Metrics{{ID: "value", MType: constants.MetricTypeGauge, Value: &value}, {ID: "value1", MType: constants.MetricTypeGauge, Value: &value}},
//				statusCode: http.StatusOK,
//			},
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			err := logger.Initialize()
//			require.NoError(t, err)
//
//			store := metrics.New(tt.store)
//			fileStore := file.New(store, &cfg)
//			ts := httptest.NewServer(Router(store, &fileStore, &cfg))
//			defer ts.Close()
//
//			var b bytes.Buffer
//			err = json.NewEncoder(&b).Encode(tt.metric)
//			require.NoError(t, err)
//			request, err := http.NewRequest(tt.method, ts.URL+"/update/", &b)
//			require.NoError(t, err)
//
//			result, err := ts.Client().Do(request)
//			require.NoError(t, err)
//
//			assert.Equal(t, tt.want.statusCode, result.StatusCode)
//
//			err = result.Body.Close()
//			require.NoError(t, err)
//
//			res, err := store.ReadMetrics(context.Background())
//			require.NoError(t, err)
//			assert.Equal(t, tt.want.store, res)
//		})
//	}
//}

func BenchmarkUpdateMetricValue(b *testing.B) {
	cfg := config.Config{
		RunAddr:         ":8080",
		FileStoragePath: "/tmp/metrics-db.json",
		StoreInterval:   300,
		IsRestore:       true,
		DatabaseDSN:     "",
		SecretKey:       "",
	}
	var deltaValue int64 = 1

	logger.Initialize()

	store := metrics.New([]models.Metrics{})
	fileStore := file.New(store, &cfg)
	ts := httptest.NewServer(Router(store, &fileStore, &cfg))
	defer ts.Close()

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(models.Metrics{ID: "value", MType: constants.MetricTypeCounter, Delta: &deltaValue})

	request, _ := http.NewRequest(http.MethodPost, ts.URL+"/update/", &buf)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		result, _ := ts.Client().Do(request)
		result.Body.Close()
	}
}
