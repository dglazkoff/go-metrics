package handlers

import (
	"bytes"
	"encoding/json"
	"github.com/dglazkoff/go-metrics/cmd/server/logger"
	"github.com/dglazkoff/go-metrics/cmd/server/storage"
	"github.com/dglazkoff/go-metrics/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

// покрыл не все кейсы так как мало времени и все эти кейсы должны проверяться инкрементными тестами
func TestUpdateMetricValue(t *testing.T) {
	type want struct {
		store      storage.MemStorage
		statusCode int
	}

	var deltaValue int64 = 1
	var value float64 = 101

	tests := []struct {
		name    string
		method  string
		store   storage.MemStorage
		metric  models.Metrics
		request string
		want    want
	}{
		{
			name:   "success test",
			method: http.MethodPost,
			store:  storage.MemStorage{CounterMetrics: map[string]int64{}},
			metric: models.Metrics{ID: "value", MType: "counter", Delta: &deltaValue},
			want: want{
				store:      storage.MemStorage{CounterMetrics: map[string]int64{"value": deltaValue}},
				statusCode: http.StatusOK,
			},
		},
		{
			name:   "invalid method GET",
			method: http.MethodGet,
			store:  storage.MemStorage{},
			metric: models.Metrics{ID: "value", MType: "counter", Delta: &deltaValue},
			want: want{
				store:      storage.MemStorage{},
				statusCode: http.StatusMethodNotAllowed,
			},
		},
		{
			name:   "add counter to previous result",
			method: http.MethodPost,
			store:  storage.MemStorage{CounterMetrics: map[string]int64{"value": 1}},
			metric: models.Metrics{ID: "value", MType: "counter", Delta: &deltaValue},
			want: want{
				store:      storage.MemStorage{CounterMetrics: map[string]int64{"value": 2}},
				statusCode: http.StatusOK,
			},
		},
		{
			name:   "update gauge metric",
			method: http.MethodPost,
			store:  storage.MemStorage{GaugeMetrics: map[string]float64{"value": 1}},
			metric: models.Metrics{ID: "value", MType: "gauge", Value: &value},
			want: want{
				store:      storage.MemStorage{GaugeMetrics: map[string]float64{"value": value}},
				statusCode: http.StatusOK,
			},
		},
		{
			name:   "add gauge metric",
			method: http.MethodPost,
			store:  storage.MemStorage{GaugeMetrics: map[string]float64{"value": 1}},
			metric: models.Metrics{ID: "value1", MType: "gauge", Value: &value},
			want: want{
				store:      storage.MemStorage{GaugeMetrics: map[string]float64{"value1": value, "value": 1}},
				statusCode: http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := logger.Initialize()
			require.NoError(t, err)

			ts := httptest.NewServer(Router(&tt.store))
			defer ts.Close()

			var b bytes.Buffer
			err = json.NewEncoder(&b).Encode(tt.metric)
			request, err := http.NewRequest(tt.method, ts.URL+"/update", &b)
			require.NoError(t, err)

			result, err := ts.Client().Do(request)
			require.NoError(t, err)

			assert.Equal(t, tt.want.statusCode, result.StatusCode)

			err = result.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.want.store, tt.store)
		})
	}
}
