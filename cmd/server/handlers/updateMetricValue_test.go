package handlers

import (
	"github.com/dglazkoff/go-metrics/cmd/server/storage"
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
	tests := []struct {
		name   string
		method string
		store  storage.MemStorage
		//metricType  string
		//metricName  string
		//metricValue string
		path    string
		request string
		want    want
	}{
		{
			name:   "success test",
			method: http.MethodPost,
			store:  storage.MemStorage{CounterMetrics: map[string]int64{}},
			path:   "/update/counter/value/1",
			want: want{
				store:      storage.MemStorage{CounterMetrics: map[string]int64{"value": 1}},
				statusCode: http.StatusOK,
			},
		},
		{
			name:   "invalid method GET",
			method: http.MethodGet,
			store:  storage.MemStorage{},
			path:   "/update/counter/value/1",
			want: want{
				store:      storage.MemStorage{},
				statusCode: http.StatusMethodNotAllowed,
			},
		},
		{
			name:   "add counter to previous result",
			method: http.MethodPost,
			store:  storage.MemStorage{CounterMetrics: map[string]int64{"value": 1}},
			path:   "/update/counter/value/1",
			want: want{
				store:      storage.MemStorage{CounterMetrics: map[string]int64{"value": 2}},
				statusCode: http.StatusOK,
			},
		},
		{
			name:   "update gauge metric",
			method: http.MethodPost,
			store:  storage.MemStorage{GaugeMetrics: map[string]float64{"value": 1}},
			path:   "/update/gauge/value/101",
			want: want{
				store:      storage.MemStorage{GaugeMetrics: map[string]float64{"value": 101}},
				statusCode: http.StatusOK,
			},
		},
		{
			name:   "add gauge metric",
			method: http.MethodPost,
			store:  storage.MemStorage{GaugeMetrics: map[string]float64{"value": 1}},
			path:   "/update/gauge/value1/101",
			want: want{
				store:      storage.MemStorage{GaugeMetrics: map[string]float64{"value1": 101, "value": 1}},
				statusCode: http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(Router(&tt.store))
			defer ts.Close()

			request, err := http.NewRequest(tt.method, ts.URL+tt.path, nil)
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
