package main

import (
	"net/http"
	"strconv"
)

type MemStorage struct {
	gaugeMetrics   GaugeMetrics
	counterMetrics CounterMetrics
}

type GaugeMetrics map[string]float64

func (gm GaugeMetrics) Save(name, value string) error {
	floatValue, err := strconv.ParseFloat(value, 64)

	if err != nil {
		return err
	}

	gm[name] = floatValue

	return nil
}

type CounterMetrics map[string]int64

func (cm CounterMetrics) Save(name, value string) error {
	intValue, err := strconv.ParseInt(value, 10, 64)

	if err != nil {
		return err
	}

	cm[name] += intValue

	return nil
}

// можно ли как-то объявить все поля не nil сразу а не прописывать определение каждого?
var storage = MemStorage{gaugeMetrics: make(map[string]float64), counterMetrics: make(map[string]int64)}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/update/{metricType}/{metricName}/{metricValue}", updateMetricValue)

	return http.ListenAndServe(":8080", mux)
}

func updateMetricValue(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	metricType := r.PathValue("metricType")
	metricName := r.PathValue("metricName")
	metricValue := r.PathValue("metricValue")

	if metricType != "gauge" && metricType != "counter" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if metricType == "gauge" {
		err := storage.gaugeMetrics.Save(metricName, metricValue)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	if metricType == "counter" {
		err := storage.counterMetrics.Save(metricName, metricValue)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}
