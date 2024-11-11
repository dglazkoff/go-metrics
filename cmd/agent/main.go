package main

import (
	"fmt"
	"log"
	mathRand "math/rand"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/dglazkoff/go-metrics/cmd/agent/client"
	"github.com/dglazkoff/go-metrics/cmd/agent/config"
	constants "github.com/dglazkoff/go-metrics/internal/const"
	"github.com/dglazkoff/go-metrics/internal/logger"
	"github.com/dglazkoff/go-metrics/internal/models"
	pb "github.com/dglazkoff/go-metrics/internal/models/proto"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	BuildVersion = "N/A"
	BuildDate    = "N/A"
	BuildCommit  = "N/A"
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

func updateMetricsWorkerPool(gm *GaugeMetrics, cm *CounterMetrics, cfg *config.Config) {
	workersChan := make(chan struct{}, cfg.RateLimit)
	var wg sync.WaitGroup

	wg.Add(cfg.RateLimit)
	go func() {
		writeMetricsInterval := time.Duration(cfg.ReportInterval) * time.Second
		defer close(workersChan)
		for {
			time.Sleep(writeMetricsInterval)
			workersChan <- struct{}{}
		}
	}()

	for i := 0; i < cfg.RateLimit; i++ {
		go func() {
			for range workersChan {
				updateMetrics(gm, cm, cfg)
			}
			wg.Done()
		}()
	}

	wg.Wait()
	fmt.Println(2)
}

func parseMetrics(gm *GaugeMetrics, cm *CounterMetrics) []models.Metrics {
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

	return metrics
}

func updateMetrics(gm *GaugeMetrics, cm *CounterMetrics, cfg *config.Config) {
	metrics := parseMetrics(gm, cm)

	if cfg.IsGRPC {
		conn, err := grpc.NewClient(cfg.RunAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()
		mc := pb.NewMetricsClient(conn)

		grpcClient := client.NewMetricsClient(mc)
		grpcClient.SendMetricsByGRPC(metrics)
		return
	}

	httpClient := client.NewClient([]time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second})
	httpClient.SendMetricsByHTTP(metrics, cfg)
}

func writeMetricsOnce(
	gm *GaugeMetrics,
	cm *CounterMetrics,
	memStatProvider func() (*mem.VirtualMemoryStat, error),
	cpuCountProvider func(bool) (int, error),
) {
	var memStats runtime.MemStats
	v, err := memStatProvider()

	if err != nil {
		logger.Log.Debug("Error while get memory stats: ", err)
	} else {
		gm.TotalMemory = float64(v.Total)
		gm.FreeMemory = float64(v.Free)
	}

	c, err := cpuCountProvider(false)

	if err != nil {
		logger.Log.Debug("Error while get cpu counts: ", err)
	} else {
		gm.CPUutilization1 = float64(c)
	}

	runtime.ReadMemStats(&memStats)
	gm.MemStats = memStats
	gm.RandomValue = mathRand.Float64()

	cm.PollCount += 1
}

func writeMetrics(gm *GaugeMetrics, cm *CounterMetrics, cfg *config.Config) {
	writeMetricsInterval := time.Duration(cfg.PollInterval) * time.Second

	for {
		select {
		case <-time.After(writeMetricsInterval):
			writeMetricsOnce(gm, cm, mem.VirtualMemory, cpu.Counts)
		}
	}
}

// go run -ldflags "-X main.BuildVersion=v1.0.1 -X 'main.BuildDate=$(date +'%Y/%m/%d %H:%M:%S')'" ./cmd/agent
func main() {
	err := logger.Initialize()

	if err != nil {
		panic(err)
	}

	cfg := config.ParseConfig()

	fmt.Printf("Build version: %s\n", BuildVersion)
	fmt.Printf("Build date: %s\n", BuildDate)
	fmt.Printf("Build commit: %s\n", BuildCommit)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	gm := GaugeMetrics{}
	cm := CounterMetrics{}

	go func() {
		sig := <-sigs
		logger.Log.Debug("Signal: ", sig)
		updateMetrics(&gm, &cm, &cfg)
		os.Exit(0)
	}()

	go writeMetrics(&gm, &cm, &cfg)

	updateMetricsWorkerPool(&gm, &cm, &cfg)
}
