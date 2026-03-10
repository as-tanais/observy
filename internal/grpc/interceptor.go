package grpc

import (
	"context"
	"net"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func IPInterceptor(trustedSubnet *net.IPNet, log *zap.Logger) grpc.UnaryServerInterceptor {

	if trustedSubnet == nil {
		return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

			ip := extractIPFromMetadata(ctx, log)
			if ip != "" {

				ctx = context.WithValue(ctx, "audit_ip", ip)
				log.Debug("IP saved for audit", zap.String("ip", ip))
			}
			return handler(ctx, req)
		}
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

		realIP := extractIPFromMetadata(ctx, log)
		if realIP == "" {
			log.Warn("Missing x-real-ip header in metadata")
			return nil, status.Errorf(codes.PermissionDenied, "missing x-real-ip header")
		}

		ip := net.ParseIP(realIP)
		if ip == nil {
			log.Warn("Invalid IP format", zap.String("ip", realIP))
			return nil, status.Errorf(codes.PermissionDenied, "invalid IP format")
		}

		if !trustedSubnet.Contains(ip) {
			log.Warn("IP not in trusted subnet",
				zap.String("ip", realIP),
				zap.String("trusted_subnet", trustedSubnet.String()))
			return nil, status.Errorf(codes.PermissionDenied, "IP not in trusted subnet")
		}

		ctx = context.WithValue(ctx, "audit_ip", realIP)
		log.Debug("IP verified and saved for audit", zap.String("ip", realIP))

		return handler(ctx, req)
	}
}

func extractIPFromMetadata(ctx context.Context, log *zap.Logger) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Debug("No metadata in request")
		return ""
	}

	ips := md.Get("x-real-ip")
	if len(ips) == 0 {
		log.Debug("Missing x-real-ip header in metadata")
		return ""
	}

	realIP := ips[0]

	if strings.Contains(realIP, ":") {
		host, _, err := net.SplitHostPort(realIP)
		if err != nil {
			log.Debug("Failed to split host port", zap.String("ip", realIP), zap.Error(err))
			return ""
		}
		realIP = host
	}

	return realIP
}
