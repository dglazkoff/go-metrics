package main

import (
	"encoding/json"
	"fmt"
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
	body, err := json.Marshal(models.Metrics{MType: "gauge", ID: name, Value: &value})

	if err == nil {
		_, err = client.R().SetBody(body).Post("/update/")
	}

	fmt.Println(err)
}

func updateCounterMetric(name string, value int64) {
	body, err := json.Marshal(models.Metrics{MType: "counter", ID: name, Delta: &value})

	if err == nil {
		_, err = client.R().SetBody(body).Post("/update/")
	}

	fmt.Println(err)
}

func updateMetrics(gm *GaugeMetrics, cm *CounterMetrics) {
	for {
		time.Sleep(time.Duration(flagReportInterval) * time.Second)

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

func writeMetrics(gm *GaugeMetrics, cm *CounterMetrics) {
	for {
		time.Sleep(time.Duration(flagPollInterval) * time.Second)

		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		gm.MemStats = memStats
		gm.RandomValue = rand.Float64()

		cm.PollCount += 1
	}
}

func main() {
	parseFlags()

	client.SetBaseURL("http://" + flagRunAddr)

	var gm = GaugeMetrics{}
	var cm = CounterMetrics{}

	go writeMetrics(&gm, &cm)
	go updateMetrics(&gm, &cm)

	select {}
}
