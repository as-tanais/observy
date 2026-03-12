package agent

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	models "github.com/as-tanais/observy/internal/model"
	pb "github.com/as-tanais/observy/internal/proto/metrics"
)

// GRPCClient представляет gRPC клиент для отправки метрик
type GRPCClient struct {
	conn    *grpc.ClientConn
	client  pb.MetricsClient
	localIP string
	timeout time.Duration
}

// NewGRPCClient создает новый gRPC клиент
func NewGRPCClient(serverAddr string) (*GRPCClient, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.NewClient(
		serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client: %w", err)
	}

	if err := waitForConnection(ctx, conn); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	localIP := getLocalIP()
	if localIP == "" {
		log.Println("Warning: Could not determine local IP, x-real-ip header will be empty")
	} else {
		log.Printf("Local IP detected for gRPC: %s", localIP)
	}

	client := pb.NewMetricsClient(conn)

	return &GRPCClient{
		conn:    conn,
		client:  client,
		localIP: localIP,
		timeout: 10 * time.Second,
	}, nil
}

// waitForConnection ожидает готовности соединения
func waitForConnection(ctx context.Context, conn *grpc.ClientConn) error {
	for {
		state := conn.GetState()
		if state == connectivity.Ready {
			return nil
		}
		if state == connectivity.Shutdown {
			return fmt.Errorf("connection is shut down")
		}

		// Ждем изменения состояния или таймаута
		if !conn.WaitForStateChange(ctx, state) {
			return ctx.Err()
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}
}

// Close закрывает соединение
func (c *GRPCClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// SendMetrics отправляет метрики через gRPC
func (c *GRPCClient) SendMetrics(ctx context.Context, metrics []models.Metrics) error {
	if len(metrics) == 0 {
		return nil
	}

	pbMetrics, err := ToProtoMetrics(metrics)
	if err != nil {
		return fmt.Errorf("failed to convert metrics: %w", err)
	}

	md := metadata.New(map[string]string{
		"x-real-ip": c.localIP,
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	req := &pb.UpdateMetricsRequest{
		Metrics: pbMetrics,
	}

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	start := time.Now()
	resp, err := c.client.UpdateMetrics(ctx, req)
	if err != nil {
		return fmt.Errorf("gRPC call failed: %w", err)
	}

	log.Printf("Sent %d metrics via gRPC to %s (took %v)",
		len(pbMetrics), c.localIP, time.Since(start))

	_ = resp
	return nil
}

// SendMetricsBatch отправляет метрики
func (c *GRPCClient) SendMetricsBatch(ctx context.Context, metrics []models.Metrics) error {
	return c.SendMetrics(ctx, metrics)
}

// WithTimeout устанавливает таймаут для запросов
func (c *GRPCClient) WithTimeout(timeout time.Duration) *GRPCClient {
	c.timeout = timeout
	return c
}

// IsConnected проверяет состояние соединения
func (c *GRPCClient) IsConnected() bool {
	if c.conn == nil {
		return false
	}
	return c.conn.GetState() == connectivity.Ready
}
