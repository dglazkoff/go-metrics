package handlers

import (
	"github.com/dglazkoff/go-metrics/cmd/server/storage"
	"net/http"
)

func UpdateMetricValue(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	metricType := r.PathValue("metricType")
	metricName := r.PathValue("metricName")
	metricValue := r.PathValue("metricValue")

	r.Body.Close()

	if metricType != "gauge" && metricType != "counter" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if metricType == "gauge" {
		err := storage.Storage.GaugeMetrics.Save(metricName, metricValue)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	if metricType == "counter" {
		err := storage.Storage.CounterMetrics.Save(metricName, metricValue)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}
