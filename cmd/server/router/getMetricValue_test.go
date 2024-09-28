package router

import (
	"github.com/dglazkoff/go-metrics/cmd/server/config"
	"github.com/dglazkoff/go-metrics/cmd/server/storage/file"
	"github.com/dglazkoff/go-metrics/cmd/server/storage/metrics"
	constants "github.com/dglazkoff/go-metrics/internal/const"
	"github.com/dglazkoff/go-metrics/internal/logger"
	"github.com/dglazkoff/go-metrics/internal/models"
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkGetMetricValue(b *testing.B) {
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

	store := metrics.New([]models.Metrics{{ID: "value", MType: constants.MetricTypeCounter, Delta: &deltaValue}})
	fileStore := file.New(store, &cfg)
	ts := httptest.NewServer(Router(store, &fileStore, &cfg))
	defer ts.Close()

	request, _ := http.NewRequest(http.MethodGet, ts.URL+"/value/"+constants.MetricTypeCounter+"/value", nil)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		response, _ := ts.Client().Do(request)

		if response != nil {
			response.Body.Close()
		}
	}
}
