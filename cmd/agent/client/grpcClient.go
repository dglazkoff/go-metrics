package client

import (
	"context"
	"fmt"

	constants "github.com/dglazkoff/go-metrics/internal/const"
	"github.com/dglazkoff/go-metrics/internal/models"
	pb "github.com/dglazkoff/go-metrics/internal/models/proto"
)

type GRPCMetricsClient struct {
	client pb.MetricsClient
}

func NewMetricsClient(conn pb.MetricsClient) *GRPCMetricsClient {
	return &GRPCMetricsClient{client: conn}
}

func (c *GRPCMetricsClient) SendMetricsByGRPC(metrics []models.Metrics) {
	protoMetrics := make([]*pb.Metric, 0, len(metrics))

	for _, metric := range metrics {
		if metric.MType == constants.MetricTypeGauge {
			protoMetrics = append(protoMetrics, &pb.Metric{
				Id:   metric.ID,
				Type: pb.Metric_Gauge,
				MetricValue: &pb.Metric_Value{
					Value: *metric.Value,
				},
			})
		}

		if metric.MType == constants.MetricTypeCounter {
			protoMetrics = append(protoMetrics, &pb.Metric{
				Id:   metric.ID,
				Type: pb.Metric_Counter,
				MetricValue: &pb.Metric_Delta{
					Delta: *metric.Delta,
				},
			})
		}
	}

	res, _ := c.client.UpdateMetrics(context.Background(), &pb.UpdateMetricsRequest{
		Metrics: protoMetrics,
	})

	fmt.Println(res)
}
