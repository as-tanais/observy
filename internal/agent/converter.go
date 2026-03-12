package agent

import (
	"fmt"
	"log"

	models "github.com/as-tanais/observy/internal/model"
	pb "github.com/as-tanais/observy/internal/proto/metrics"
)

func ToProtoMetric(metric models.Metrics) (*pb.Metric, error) {
	pbMetric := &pb.Metric{
		Id: metric.ID,
	}

	switch metric.MType {
	case models.Gauge:
		pbMetric.Type = pb.Metric_GAUGE
		if metric.Value != nil {
			pbMetric.Value = *metric.Value
		} else {
			return nil, fmt.Errorf("gauge metric %s has nil value", metric.ID)
		}
	case models.Counter:
		pbMetric.Type = pb.Metric_COUNTER
		if metric.Delta != nil {
			pbMetric.Delta = *metric.Delta
		} else {
			return nil, fmt.Errorf("counter metric %s has nil delta", metric.ID)
		}
	default:
		return nil, fmt.Errorf("unknown metric type: %s", metric.MType)
	}

	return pbMetric, nil
}

func ToProtoMetrics(metrics []models.Metrics) ([]*pb.Metric, error) {
	result := make([]*pb.Metric, 0, len(metrics))

	for _, m := range metrics {
		pbMetric, err := ToProtoMetric(m)
		if err != nil {

			log.Printf("Skipping metric %s: %v", m.ID, err)
			continue
		}
		result = append(result, pbMetric)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no valid metrics to send")
	}

	return result, nil
}
