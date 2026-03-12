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
	// Убираем контекст с таймаутом здесь - соединение будет в фоне
	conn, err := grpc.NewClient(
		serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client: %w", err)
	}

	// НЕ ждем соединения здесь - пусть устанавливается в фоне
	log.Printf("gRPC client created for %s (connecting in background)", serverAddr)

	// Получаем локальный IP для заголовка
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
		timeout: 10 * time.Second, // Таймаут для запросов
	}, nil
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

	// Проверяем состояние соединения
	state := c.conn.GetState()
	if state == connectivity.Shutdown {
		return fmt.Errorf("connection is shut down")
	}

	// Если соединение не готово, пытаемся подождать немного
	if state != connectivity.Ready {
		log.Printf("Connection state: %s, waiting for ready...", state)

		// Ждем готовности соединения (но не долго)
		waitCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		for state != connectivity.Ready {
			if !c.conn.WaitForStateChange(waitCtx, state) {
				// Таймаут ожидания - пробуем отправить, может работать
				log.Printf("Timeout waiting for connection, attempting send anyway...")
				break
			}
			state = c.conn.GetState()
		}
	}

	// Конвертируем метрики в protobuf формат
	pbMetrics, err := ToProtoMetrics(metrics)
	if err != nil {
		return fmt.Errorf("failed to convert metrics: %w", err)
	}

	// Создаем контекст с метаданными (IP агента)
	md := metadata.New(map[string]string{
		"x-real-ip": c.localIP,
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	// Создаем запрос
	req := &pb.UpdateMetricsRequest{
		Metrics: pbMetrics,
	}

	// Таймаут через контекст для конкретного запроса
	callCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// Отправляем запрос
	start := time.Now()
	resp, err := c.client.UpdateMetrics(callCtx, req)
	if err != nil {
		return fmt.Errorf("gRPC call failed: %w", err)
	}

	log.Printf("Sent %d metrics via gRPC to %s (took %v)",
		len(pbMetrics), c.localIP, time.Since(start))

	_ = resp
	return nil
}

// SendMetricsBatch отправляет метрики (аналог SendBatchMetrics для HTTP)
func (c *GRPCClient) SendMetricsBatch(ctx context.Context, metrics []models.Metrics) error {
	return c.SendMetrics(ctx, metrics)
}

// IsReady проверяет, готово ли соединение
func (c *GRPCClient) IsReady() bool {
	if c.conn == nil {
		return false
	}
	return c.conn.GetState() == connectivity.Ready
}

// WithTimeout устанавливает таймаут для запросов
func (c *GRPCClient) WithTimeout(timeout time.Duration) *GRPCClient {
	c.timeout = timeout
	return c
}
