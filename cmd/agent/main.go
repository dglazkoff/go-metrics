package main

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/dglazkoff/go-metrics/internal/const"
	"github.com/dglazkoff/go-metrics/internal/logger"
	"github.com/dglazkoff/go-metrics/internal/models"
	"github.com/go-resty/resty/v2"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"math/rand"
	"net/url"
	"reflect"
	"runtime"
	"sync"
	"time"
)

type GaugeMetrics struct {
	runtime.MemStats
	RandomValue     float64
	TotalMemory     float64
	FreeMemory      float64
	CPUutilization1 float64
}

type CounterMetrics struct {
	PollCount int64
}

var client = resty.New()
var retryIntervals = []time.Duration{1, 3, 5}

func sendRequest(body interface{}, hash []byte, retryNumber int) {
	logger.Log.Debug("Do request to /updates/")
	request := client.R().SetBody(body).SetHeader("Content-Encoding", "gzip").SetHeader("Content-Type", "application/json")

	if hash != nil {
		request.SetHeader("HashSHA256", hex.EncodeToString(hash))
	}

	_, err := request.Post("/updates/")

	if err != nil {
		logger.Log.Debug("Error on request: ", err)

		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			if retryNumber == 3 {
				return
			}

			time.Sleep(retryIntervals[retryNumber] * time.Second)
			sendRequest(body, hash, retryNumber+1)
		}
	}
}

func sendBody(body []byte, cfg *Config) {
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

	var hash []byte
	if cfg.secretKey != "" {
		logger.Log.Debug("Encoding body")
		h := hmac.New(sha256.New, []byte(cfg.secretKey))
		h.Write(buf.Bytes())
		hash = h.Sum(nil)
	}

	sendRequest(buf, hash, 0)
}

func updateMetricsWorkerPool(gm *GaugeMetrics, cm *CounterMetrics, cfg *Config) {
	workersChan := make(chan struct{}, cfg.rateLimit)
	var wg sync.WaitGroup

	wg.Add(cfg.rateLimit)
	go func() {
		writeMetricsInterval := time.Duration(cfg.reportInterval) * time.Second
		defer close(workersChan)
		for {
			time.Sleep(writeMetricsInterval)
			workersChan <- struct{}{}
		}
	}()

	for i := 0; i < cfg.rateLimit; i++ {
		go func() {
			for range workersChan {
				updateMetrics(gm, cm, cfg)
			}
			wg.Done()
		}()
	}

	wg.Wait()
}

func updateMetrics(gm *GaugeMetrics, cm *CounterMetrics, cfg *Config) {
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

	sendBody(body, cfg)
}

func writeMetrics(gm *GaugeMetrics, cm *CounterMetrics, cfg *Config) {
	writeMetricsInterval := time.Duration(cfg.pollInterval) * time.Second

	for {
		time.Sleep(writeMetricsInterval)

		var memStats runtime.MemStats
		v, err := mem.VirtualMemory()

		if err != nil {
			logger.Log.Debug("Error while get memory stats: ", err)
		} else {
			gm.TotalMemory = float64(v.Total)
			gm.FreeMemory = float64(v.Free)
		}

		c, err := cpu.Counts(false)

		if err != nil {
			logger.Log.Debug("Error while get cpu counts: ", err)
		} else {
			gm.CPUutilization1 = float64(c)
		}

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

	gm := GaugeMetrics{}
	cm := CounterMetrics{}

	go writeMetrics(&gm, &cm, &cfg)

	updateMetricsWorkerPool(&gm, &cm, &cfg)
}
