package main

import (
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"testing"
	"time"

	constants "github.com/dglazkoff/go-metrics/internal/const"
	"github.com/dglazkoff/go-metrics/internal/logger"
	"github.com/dglazkoff/go-metrics/internal/models"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestParseMetrics(t *testing.T) {
	err := logger.Initialize()
	assert.NoError(t, err)

	floatValue := 42.5
	var intValue int64 = 5
	var uintValue uint64 = 10

	var uintToFloatValue float64 = 10

	tests := []struct {
		name     string
		gm       GaugeMetrics
		cm       CounterMetrics
		expected []models.Metrics
	}{
		{
			name: "basic parse",
			gm: GaugeMetrics{
				RandomValue: floatValue,
				MemStats: runtime.MemStats{
					TotalAlloc: uintValue,
				},
			},
			cm: CounterMetrics{
				PollCount: intValue,
			},
			expected: []models.Metrics{
				{MType: constants.MetricTypeGauge, ID: "RandomValue", Value: &floatValue},
				{MType: constants.MetricTypeGauge, ID: "TotalAlloc", Value: &uintToFloatValue},
				{MType: constants.MetricTypeCounter, ID: "PollCount", Delta: &intValue},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseMetrics(&tt.gm, &tt.cm)
			assert.Subset(t, result, tt.expected)
		})
	}
}

type MockMem struct {
	mock.Mock
}

func (m *MockMem) VirtualMemory() (*mem.VirtualMemoryStat, error) {
	args := m.Called()
	return args.Get(0).(*mem.VirtualMemoryStat), args.Error(1)
}

type MockCPU struct {
	mock.Mock
}

func (c *MockCPU) Counts(logical bool) (int, error) {
	args := c.Called(logical)
	return args.Int(0), args.Error(1)
}

func TestWriteMetricsOnce(t *testing.T) {
	gm := &GaugeMetrics{}
	cm := &CounterMetrics{}

	mockMem := &MockMem{}
	mockCPU := &MockCPU{}
	mockMem.On("VirtualMemory").Return(&mem.VirtualMemoryStat{Total: 8000000, Free: 2000000}, nil)
	mockCPU.On("Counts", false).Return(4, nil)

	writeMetricsOnce(gm, cm, mockMem.VirtualMemory, mockCPU.Counts)

	assert.Equal(t, 8000000.0, gm.TotalMemory, "TotalMemory should match the mocked value")
	assert.Equal(t, 2000000.0, gm.FreeMemory, "FreeMemory should match the mocked value")
	assert.Equal(t, 4.0, gm.CPUutilization1, "CPUutilization1 should match the mocked value")
	assert.Equal(t, int64(1), cm.PollCount, "PollCount should be incremented")

	mockMem.AssertExpectations(t)
	mockCPU.AssertExpectations(t)
}

func TestRunApp_Success(t *testing.T) {
	err := logger.Initialize()
	assert.NoError(t, err)

	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()
	os.Args = []string{
		"cmd/agent",
		"-p", "10",
		"-l", "1",
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	go func() {
		err := runApp()
		assert.NoError(t, err)
	}()

	time.Sleep(100 * time.Millisecond)
	process, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("Failed to find current process: %v", err)
	}
	process.Signal(syscall.SIGTERM)

	time.Sleep(100 * time.Millisecond)

	assert.NoError(t, err, "runApp should complete without error")
}
