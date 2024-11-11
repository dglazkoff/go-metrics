package router

import (
	"context"

	constants "github.com/dglazkoff/go-metrics/internal/const"
	"github.com/dglazkoff/go-metrics/internal/logger"
	"github.com/dglazkoff/go-metrics/internal/models"
	pb "github.com/dglazkoff/go-metrics/internal/models/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type metric interface {
	UpdateList(ctx context.Context, metric []models.Metrics) error
}

type MetricsServer struct {
	pb.UnimplementedMetricsServer

	metricService metric
}

func NewMetricsServer(metricService metric) *MetricsServer {
	return &MetricsServer{metricService: metricService}
}

func (ms *MetricsServer) UpdateMetrics(ctx context.Context, in *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	var metrics []models.Metrics

	for _, metric := range in.Metrics {
		if metric.Type == pb.Metric_Gauge {
			value := metric.GetValue()
			m := models.Metrics{
				ID:    metric.Id,
				MType: constants.MetricTypeGauge,
				Value: &value,
			}

			metrics = append(metrics, m)
		}

		if metric.Type == pb.Metric_Counter {
			delta := metric.GetDelta()
			m := models.Metrics{
				ID:    metric.Id,
				MType: constants.MetricTypeCounter,
				Delta: &delta,
			}

			metrics = append(metrics, m)
		}
	}

	err := ms.metricService.UpdateList(ctx, metrics)

	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "error on update metrics: %v", err)
	}

	logger.Log.Debug("Metrics updated")
	return nil, nil
}
