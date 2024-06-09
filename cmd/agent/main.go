package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"reflect"
	"runtime"
	"time"
)

const (
	pollInterval   = 2
	reportInterval = 10
)

type GaugeMetrics struct {
	runtime.MemStats
	RandomValue float64
}

type CounterMetrics struct {
	PollCount int64
}

func updateMetric(metricType, name, value string) {
	r, err := http.Post("http://localhost:8080/update/"+metricType+"/"+name+"/"+value, "text/plain", nil)

	if r != nil {
		err = r.Body.Close()
	}

	fmt.Println(err)
}

func updateMetrics(gm *GaugeMetrics, cm *CounterMetrics) {
	for {
		time.Sleep(reportInterval * time.Second)

		valuesGm := reflect.ValueOf(*gm)
		typesGm := valuesGm.Type()
		for i := 0; i < valuesGm.NumField(); i++ {
			if typesGm.Field(i).Name == "MemStats" {
				continue
			}

			updateMetric("gauge", typesGm.Field(i).Name, fmt.Sprint(valuesGm.Field(i)))
		}

		valuesGmMemStats := reflect.ValueOf((*gm).MemStats)
		typesGmMemStats := valuesGmMemStats.Type()
		for i := 0; i < valuesGmMemStats.NumField(); i++ {
			updateMetric("gauge", typesGmMemStats.Field(i).Name, fmt.Sprint(valuesGmMemStats.Field(i)))
		}

		valuesCm := reflect.ValueOf(*cm)
		typesCm := valuesCm.Type()
		for i := 0; i < valuesCm.NumField(); i++ {
			updateMetric("counter", typesCm.Field(i).Name, fmt.Sprint(valuesCm.Field(i)))
		}
	}
}

func writeMetrics(gm *GaugeMetrics, cm *CounterMetrics) {
	for {
		time.Sleep(pollInterval * time.Second)

		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		gm.MemStats = memStats
		gm.RandomValue = rand.Float64()

		cm.PollCount += 1
	}
}

func main() {
	var gm = GaugeMetrics{}
	var cm = CounterMetrics{}

	go writeMetrics(&gm, &cm)
	go updateMetrics(&gm, &cm)

	select {}
}
