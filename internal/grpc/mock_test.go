// internal/grpc/mock_test.go
package grpc

import (
	"context"
	"sync"
	"time"

	"github.com/as-tanais/observy/internal/audit"
	models "github.com/as-tanais/observy/internal/model"
)

// MockService имитирует сервис метрик для тестов
type MockService struct {
	metrics   []models.Metrics
	ipAddress string
	mu        sync.Mutex
	auditChan chan AuditCall
}

type AuditCall struct {
	Metrics []string
	IP      string
}

func NewMockService() *MockService {
	return &MockService{
		auditChan: make(chan AuditCall, 10),
	}
}

func (m *MockService) UpdateBatch(ctx context.Context, metrics []models.Metrics, ipAddress string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metrics = metrics
	m.ipAddress = ipAddress
	return nil
}

func (m *MockService) NotifyAudit(ctx context.Context, metricNames []string, ipAddress string) {
	m.auditChan <- AuditCall{
		Metrics: metricNames,
		IP:      ipAddress,
	}
}

func (m *MockService) GetMetrics() []models.Metrics {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.metrics
}

// MockObserver имитирует observer для аудита
type MockObserver struct {
	events []audit.AuditEvent
	mu     sync.Mutex
}

func NewMockObserver() *MockObserver {
	return &MockObserver{
		events: make([]audit.AuditEvent, 0),
	}
}

func (m *MockObserver) Notify(ctx context.Context, metrics []string, ipAddr string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, audit.AuditEvent{
		TS:        time.Now().Unix(),
		Metrics:   metrics,
		IPAddress: ipAddr,
	})
}

func (m *MockObserver) GetEvents() []audit.AuditEvent {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.events
}
