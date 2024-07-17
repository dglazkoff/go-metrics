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

func sendBody(body []byte) {
	buf := bytes.NewBuffer(nil)
	zb := gzip.NewWriter(buf)
	_, err := zb.Write(body)

	if err != nil {
		logger.Log.Debug("Error on write gzip data: ", err)
		return
	}

	err = zb.Close()

	if err != nil {
		logger.Log.Debug("Error on close gzip writer: ", err)
		return
	}

	logger.Log.Debug("Do request to /updates/")
	_, err = client.R().SetBody(buf).SetHeader("Content-Encoding", "gzip").SetHeader("Content-Type", "application/json").Post("/updates/")

	if err != nil {
		logger.Log.Debug("Error on request: ", err)
		return
	}
}

func updateMetrics(gm *GaugeMetrics, cm *CounterMetrics, cfg *Config) {
	writeMetricsInterval := time.Duration(cfg.reportInterval) * time.Second

	for {
		time.Sleep(writeMetricsInterval)

		var metrics []models.Metrics
		valuesGm := reflect.ValueOf(*gm)
		typesGm := valuesGm.Type()
		for i := 0; i < valuesGm.NumField(); i++ {
			if typesGm.Field(i).Name == "MemStats" {
				continue
			}
			value := valuesGm.Field(i).Float()
			metrics = append(metrics, models.Metrics{MType: constants.MetricTypeGauge, ID: typesGm.Field(i).Name, Value: &value})
		}

		valuesGmMemStats := reflect.ValueOf((*gm).MemStats)
		typesGmMemStats := valuesGmMemStats.Type()

		for i := 0; i < valuesGmMemStats.NumField(); i++ {
			field := valuesGmMemStats.Field(i)
			var value float64

			switch field.Kind() {
			case reflect.Float32, reflect.Float64:
				value = field.Float()
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				value = float64(field.Uint())
			default:
				continue
			}

			metrics = append(metrics, models.Metrics{MType: constants.MetricTypeGauge, ID: typesGmMemStats.Field(i).Name, Value: &value})
		}

		valuesCm := reflect.ValueOf(*cm)
		typesCm := valuesCm.Type()
		for i := 0; i < valuesCm.NumField(); i++ {
			delta := valuesCm.Field(i).Int()
			metrics = append(metrics, models.Metrics{MType: constants.MetricTypeCounter, ID: typesCm.Field(i).Name, Delta: &delta})
		}

		body, err := json.Marshal(metrics)

		if err != nil {
			logger.Log.Debug("Error while marshal data: ", err)
			return
		}

		sendBody(body)
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

//func writeMetrics(metrics *[]models.Metrics, cfg *Config) {
//	writeMetricsInterval := time.Duration(cfg.pollInterval) * time.Second
//
//	for {
//		time.Sleep(writeMetricsInterval)
//
//		var memStats runtime.MemStats
//		runtime.ReadMemStats(&memStats)
//
//		valuesGmMemStats := reflect.ValueOf(memStats)
//		typesGmMemStats := valuesGmMemStats.Type()
//		for i := 0; i < valuesGmMemStats.NumField(); i++ {
//			field := valuesGmMemStats.Field(i)
//			var metric models.Metrics
//
//			switch field.Kind() {
//			case reflect.Float32, reflect.Float64:
//				value := field.Float()
//				metric = models.Metrics{MType: constants.MetricTypeGauge, ID: typesGmMemStats.Field(i).Name, Value: &value}
//			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
//				value := float64(field.Uint())
//				metric = models.Metrics{MType: constants.MetricTypeGauge, ID: typesGmMemStats.Field(i).Name, Value: &value}
//			default:
//				continue
//			}
//
//			*metrics = append(*metrics, metric)
//		}
//
//		randomValue := rand.Float64()
//		countDelta := int64(1)
//		*metrics = append(*metrics, models.Metrics{MType: constants.MetricTypeGauge, ID: "RandomValue", Value: &randomValue})
//		*metrics = append(*metrics, models.Metrics{MType: constants.MetricTypeCounter, ID: "PollCount", Delta: &countDelta})
//		cm.PollCount += 1
//	}
//}

func main() {
	cfg := parseConfig()

	err := logger.Initialize()

	if err != nil {
		panic(err)
	}

	client.SetBaseURL("http://" + cfg.runAddr)

	gm := GaugeMetrics{}
	cm := CounterMetrics{}

	go writeMetrics(&gm, &cm, &cfg)
	updateMetrics(&gm, &cm, &cfg)
}
