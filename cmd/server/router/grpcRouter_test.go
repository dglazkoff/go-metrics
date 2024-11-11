package router

import (
	"context"
	"testing"

	constants "github.com/dglazkoff/go-metrics/internal/const"
	"github.com/dglazkoff/go-metrics/internal/logger"
	"github.com/dglazkoff/go-metrics/internal/models"
	pb "github.com/dglazkoff/go-metrics/internal/models/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock для интерфейса metric
type mockMetricService struct {
	mock.Mock
}

func (m *mockMetricService) UpdateList(ctx context.Context, metric []models.Metrics) error {
	args := m.Called(ctx, metric)
	return args.Error(0)
}

func TestUpdateMetrics(t *testing.T) {
	err := logger.Initialize()
	assert.NoError(t, err)

	mockService := new(mockMetricService)
	server := NewMetricsServer(mockService)

	ctx := context.Background()
	req := &pb.UpdateMetricsRequest{
		Metrics: []*pb.Metric{
			{
				Id:   "metric1",
				Type: pb.Metric_Gauge,
				MetricValue: &pb.Metric_Value{
					Value: 1.23,
				},
			},
			{
				Id:   "metric2",
				Type: pb.Metric_Counter,
				MetricValue: &pb.Metric_Delta{
					Delta: 5,
				},
			},
		},
	}

	expectedMetrics := []models.Metrics{
		{
			ID:    "metric1",
			MType: constants.MetricTypeGauge,
			Value: float64Pointer(1.23),
		},
		{
			ID:    "metric2",
			MType: constants.MetricTypeCounter,
			Delta: int64Pointer(5),
		},
	}

	mockService.On("UpdateList", ctx, expectedMetrics).Return(nil)

	resp, err := server.UpdateMetrics(ctx, req)

	assert.NoError(t, err)
	assert.Nil(t, resp)
	mockService.AssertExpectations(t)
}

func float64Pointer(v float64) *float64 {
	return &v
}

func int64Pointer(v int64) *int64 {
	return &v
}
