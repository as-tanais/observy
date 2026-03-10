// internal/grpc/interceptor_test.go
package grpc

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestIPInterceptor_NoSubnet(t *testing.T) {
	logger := zap.NewNop()
	interceptor := IPInterceptor(nil, logger)

	ctx := context.Background()
	md := metadata.New(map[string]string{
		"x-real-ip": "192.168.1.100",
	})
	ctx = metadata.NewIncomingContext(ctx, md)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		// Проверяем, что IP сохранился в контексте
		ip, ok := ctx.Value("audit_ip").(string)
		assert.True(t, ok, "audit_ip should be in context")
		assert.Equal(t, "192.168.1.100", ip)
		return "response", nil
	}

	resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, handler)
	assert.NoError(t, err)
	assert.Equal(t, "response", resp)
}

func TestIPInterceptor_WithSubnet_ValidIP(t *testing.T) {
	logger := zap.NewNop()
	_, trustedNet, _ := net.ParseCIDR("192.168.1.0/24")
	interceptor := IPInterceptor(trustedNet, logger)

	ctx := context.Background()
	md := metadata.New(map[string]string{
		"x-real-ip": "192.168.1.100",
	})
	ctx = metadata.NewIncomingContext(ctx, md)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		ip, ok := ctx.Value("audit_ip").(string)
		assert.True(t, ok)
		assert.Equal(t, "192.168.1.100", ip)
		return "response", nil
	}

	resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, handler)
	assert.NoError(t, err)
	assert.Equal(t, "response", resp)
}

func TestIPInterceptor_WithSubnet_InvalidIP(t *testing.T) {
	logger := zap.NewNop()
	_, trustedNet, _ := net.ParseCIDR("192.168.1.0/24")
	interceptor := IPInterceptor(trustedNet, logger)

	ctx := context.Background()
	md := metadata.New(map[string]string{
		"x-real-ip": "10.0.0.1",
	})
	ctx = metadata.NewIncomingContext(ctx, md)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "response", nil
	}

	_, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, handler)
	assert.Error(t, err)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.PermissionDenied, st.Code())
}

func TestIPInterceptor_MissingHeader(t *testing.T) {
	logger := zap.NewNop()
	_, trustedNet, _ := net.ParseCIDR("192.168.1.0/24")
	interceptor := IPInterceptor(trustedNet, logger)

	ctx := context.Background()
	ctx = metadata.NewIncomingContext(ctx, metadata.New(map[string]string{}))

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "response", nil
	}

	_, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, handler)
	assert.Error(t, err)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.PermissionDenied, st.Code())
}

func TestIPInterceptor_WithPort(t *testing.T) {
	logger := zap.NewNop()
	_, trustedNet, _ := net.ParseCIDR("192.168.1.0/24")
	interceptor := IPInterceptor(trustedNet, logger)

	ctx := context.Background()
	md := metadata.New(map[string]string{
		"x-real-ip": "192.168.1.100:8080",
	})
	ctx = metadata.NewIncomingContext(ctx, md)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		ip, ok := ctx.Value("audit_ip").(string)
		assert.True(t, ok)
		assert.Equal(t, "192.168.1.100", ip)
		return "response", nil
	}

	resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, handler)
	assert.NoError(t, err)
	assert.Equal(t, "response", resp)
}
