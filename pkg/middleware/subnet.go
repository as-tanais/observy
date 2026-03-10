package middleware

import (
	"net"
	"net/http"

	"go.uber.org/zap"
)

func TrustedSubnetMiddleware(trustedSubnet *net.IPNet, log *zap.Logger) func(http.Handler) http.Handler {

	if trustedSubnet == nil {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			realIP := r.Header.Get("X-Real-IP")
			if realIP == "" {
				http.Error(w, "Missing X-Real-IP header", http.StatusForbidden)
				return
			}

			ip := net.ParseIP(realIP)
			if ip == nil {
				http.Error(w, "Invalid IP format", http.StatusForbidden)
				return
			}

			if !trustedSubnet.Contains(ip) {
				http.Error(w, "IP not in trusted subnet", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
