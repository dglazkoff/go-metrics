package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"github.com/dglazkoff/go-metrics/internal/const"
	"github.com/dglazkoff/go-metrics/internal/logger"
	"github.com/dglazkoff/go-metrics/internal/models"
	"github.com/go-resty/resty/v2"
	"math/rand"
	"reflect"
	"runtime"
	"time"
)

type GaugeMetrics struct {
	runtime.MemStats
	RandomValue float64
}

type CounterMetrics struct {
	PollCount int64
}

var client = resty.New()

func updateGaugeMetric(name string, value float64) {
	body, err := json.Marshal(models.Metrics{MType: constants.MetricTypeGauge, ID: name, Value: &value})

	if err != nil {
		logger.Log.Debug("Error while marshal data: ", err)
		return
	}

	buf := bytes.NewBuffer(nil)
	zb := gzip.NewWriter(buf)
	_, err = zb.Write([]byte(body))

	if err != nil {
		logger.Log.Debug("Error on write gzip data: ", err)
		return
	}

	err = zb.Close()

	if err != nil {
		logger.Log.Debug("Error on close gzip writer: ", err)
		return
	}

	logger.Log.Debug("Do request to /update/")
	_, err = client.R().SetBody(buf).SetHeader("Content-Encoding", "gzip").SetHeader("Content-Type", "application/json").Post("/update/")

	if err != nil {
		logger.Log.Debug("Error on request: ", err)
		return
	}
}

func updateCounterMetric(name string, value int64) {
	body, err := json.Marshal(models.Metrics{MType: constants.MetricTypeCounter, ID: name, Delta: &value})

	if err != nil {
		logger.Log.Debug("Error while marshal data: ", err)
		return
	}

	buf := bytes.NewBuffer(nil)
	zb := gzip.NewWriter(buf)
	_, err = zb.Write([]byte(body))

	if err != nil {
		logger.Log.Debug("Error on write gzip data: ", err)
		return
	}

	err = zb.Close()

	if err != nil {
		logger.Log.Debug("Error on close gzip writer: ", err)
		return
	}

	logger.Log.Debug("Do request to /update/")
	_, err = client.R().SetBody(buf).SetHeader("Content-Encoding", "gzip").SetHeader("Content-Type", "application/json").Post("/update/")

	if err != nil {
		logger.Log.Debug("Error on request: ", err)
		return
	}
}

func updateMetrics(gm *GaugeMetrics, cm *CounterMetrics, cfg *Config) {
	writeMetricsInterval := time.Duration(cfg.reportInterval) * time.Second

	for {
		time.Sleep(writeMetricsInterval)

		valuesGm := reflect.ValueOf(*gm)
		typesGm := valuesGm.Type()
		for i := 0; i < valuesGm.NumField(); i++ {
			if typesGm.Field(i).Name == "MemStats" {
				continue
			}

			updateGaugeMetric(typesGm.Field(i).Name, valuesGm.Field(i).Float())
		}

		valuesGmMemStats := reflect.ValueOf((*gm).MemStats)
		typesGmMemStats := valuesGmMemStats.Type()

		for i := 0; i < valuesGmMemStats.NumField(); i++ {
			field := valuesGmMemStats.Field(i)

			switch field.Kind() {
			case reflect.Float32, reflect.Float64:
				updateGaugeMetric(typesGmMemStats.Field(i).Name, field.Float())
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				updateGaugeMetric(typesGmMemStats.Field(i).Name, float64(field.Uint()))
			default:
				continue
			}
		}

		valuesCm := reflect.ValueOf(*cm)
		typesCm := valuesCm.Type()
		for i := 0; i < valuesCm.NumField(); i++ {
			updateCounterMetric(typesCm.Field(i).Name, valuesCm.Field(i).Int())
		}
	}
}

func writeMetrics(gm *GaugeMetrics, cm *CounterMetrics, cfg *Config) {
	writeMetricsInterval := time.Duration(cfg.pollInterval) * time.Second

	for {
		time.Sleep(writeMetricsInterval)

		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		gm.MemStats = memStats
		gm.RandomValue = rand.Float64()

		cm.PollCount += 1
	}
}

func main() {
	cfg := parseConfig()

	err := logger.Initialize()

	if err != nil {
		panic(err)
	}

	client.SetBaseURL("http://" + cfg.runAddr)

	var gm = GaugeMetrics{}
	var cm = CounterMetrics{}

	go writeMetrics(&gm, &cm, &cfg)
	updateMetrics(&gm, &cm, &cfg)
}
