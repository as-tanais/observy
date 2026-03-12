// internal/grpc/server.go
package grpc

import (
	"context"
	"fmt"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	models "github.com/as-tanais/observy/internal/model"
	pb "github.com/as-tanais/observy/internal/proto/metrics"
	"github.com/as-tanais/observy/internal/service"
)

// MetricsServer реализует gRPC интерфейс
type MetricsServer struct {
	pb.UnimplementedMetricsServer
	service *service.MetricsService
	logger  *zap.Logger
}

// NewMetricsServer создает новый gRPC сервер
func NewMetricsServer(service *service.MetricsService, logger *zap.Logger) *MetricsServer {
	return &MetricsServer{
		service: service,
		logger:  logger,
	}
}

func (s *MetricsServer) UpdateMetrics(ctx context.Context, req *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	if len(req.Metrics) == 0 {
		s.logger.Debug("Received empty metrics batch")
		return &pb.UpdateMetricsResponse{}, nil
	}

	auditIP, _ := ctx.Value(AuditIPKey).(string)

	s.logger.Debug("Processing gRPC metrics request",
		zap.String("audit_ip", auditIP),
		zap.Int("count", len(req.Metrics)))

	metrics := make([]models.Metrics, 0, len(req.Metrics))

	for _, pbMetric := range req.Metrics {
		metric := models.Metrics{
			ID: pbMetric.Id,
		}

		switch pbMetric.Type {
		case pb.Metric_GAUGE:
			metric.MType = models.Gauge
			metric.Value = &pbMetric.Value
		case pb.Metric_COUNTER:
			metric.MType = models.Counter
			metric.Delta = &pbMetric.Delta
		default:
			s.logger.Warn("Unknown metric type, skipping",
				zap.String("id", pbMetric.Id),
				zap.String("type", pbMetric.Type.String()))
			continue
		}

		metrics = append(metrics, metric)
	}

	if len(metrics) == 0 {
		s.logger.Warn("No valid metrics to save")
		return &pb.UpdateMetricsResponse{}, nil
	}

	if err := s.service.UpdateBatch(ctx, metrics, auditIP); err != nil {
		s.logger.Error("Failed to save metrics batch",
			zap.Error(err),
			zap.Int("count", len(metrics)))
		return nil, status.Errorf(codes.Internal, "failed to save metrics: %v", err)
	}

	s.logger.Info("Successfully saved metrics batch via gRPC",
		zap.Int("count", len(metrics)))

	return &pb.UpdateMetricsResponse{}, nil
}

// StartGRPCServer запускает gRPC сервер
func StartGRPCServer(addr string, service *service.MetricsService, trustedSubnet *net.IPNet, logger *zap.Logger) (*grpc.Server, net.Listener, error) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to listen: %w", err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(IPInterceptor(trustedSubnet, logger)),
	)

	metricsServer := NewMetricsServer(service, logger)
	pb.RegisterMetricsServer(grpcServer, metricsServer)

	logger.Info("gRPC server configured",
		zap.String("addr", addr),
		zap.Bool("subnet_protection", trustedSubnet != nil))

	return grpcServer, lis, nil
}
