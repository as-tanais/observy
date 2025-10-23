package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// responseWriter обёртка для захвата статуса и размера ответа
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

// newResponseWriter создаёт обёртку над ResponseWriter
func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK, // По умолчанию 200
		size:           0,
	}
}

// WriteHeader перехватывает код статуса
func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// Write перехватывает размер ответа
func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

// WithLogging добавляет логирование запросов и ответов
func WithLogging(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Запоминаем время начала
			start := time.Now()

			// Оборачиваем ResponseWriter для перехвата данных
			rw := newResponseWriter(w)

			// Вызываем следующий handler
			next.ServeHTTP(rw, r)

			// Вычисляем длительность
			duration := time.Since(start)

			// Логируем запрос и ответ
			logger.Info("HTTP request",
				zap.String("method", r.Method),
				zap.String("uri", r.RequestURI),
				zap.Int("status", rw.statusCode),
				zap.Int("size", rw.size),
				zap.Duration("duration", duration),
			)
		})
	}
}
