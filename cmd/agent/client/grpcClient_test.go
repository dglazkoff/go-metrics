package client

import (
	"context"
	"net"
	"testing"

	constants "github.com/dglazkoff/go-metrics/internal/const"
	"github.com/dglazkoff/go-metrics/internal/logger"
	"github.com/dglazkoff/go-metrics/internal/models"
	pb "github.com/dglazkoff/go-metrics/internal/models/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

type MockMetricsClient struct {
	mock.Mock
}

func (m *MockMetricsClient) UpdateMetrics(ctx context.Context, req *pb.UpdateMetricsRequest, opts ...grpc.CallOption) (*pb.UpdateMetricsResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*pb.UpdateMetricsResponse), args.Error(1)
}

func TestSendMetricsByGRPC(t *testing.T) {
	err := logger.Initialize()
	assert.NoError(t, err)

	lis := bufconn.Listen(1024 * 1024)
	s := grpc.NewServer()
	defer s.Stop()

	go func() {
		err = s.Serve(lis)
		assert.NoError(t, err)
	}()

	conn, err := grpc.NewClient("bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	mockClient := &MockMetricsClient{}
	metrics := []models.Metrics{
		{
			ID:    "gauge1",
			MType: constants.MetricTypeGauge,
			Value: func(v float64) *float64 { return &v }(123.45),
		},
		{
			ID:    "counter1",
			MType: constants.MetricTypeCounter,
			Delta: func(d int64) *int64 { return &d }(10),
		},
	}

	mockResponse := &pb.UpdateMetricsResponse{}
	mockClient.On("UpdateMetrics", mock.Anything, mock.Anything).Return(mockResponse, nil)

	grpcClient := NewMetricsClient(mockClient)
	grpcClient.SendMetricsByGRPC(metrics)

	mockClient.AssertCalled(t, "UpdateMetrics", mock.Anything, mock.Anything)
}
