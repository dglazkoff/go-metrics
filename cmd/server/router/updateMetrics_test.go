package router

import (
	"bytes"
	"encoding/json"
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

func BenchmarkUpdateMetrics(b *testing.B) {
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

	metricsToUpdate := make([]models.Metrics, 0, 5)

	for i := 0; i < 5; i++ {
		metricsToUpdate = append(metricsToUpdate, models.Metrics{ID: "value", MType: constants.MetricTypeCounter, Delta: &deltaValue})
	}

	body, _ := json.Marshal(metricsToUpdate)
	buf := bytes.NewBuffer(nil)
	buf.Write(body)

	request, _ := http.NewRequest(http.MethodPost, ts.URL+"/updates/", buf)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		result, _ := ts.Client().Do(request)
		result.Body.Close()
	}
}
