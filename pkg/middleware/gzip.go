package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// GzipDecompressRequest распаковывает входящие сжатые запросы
// Проверяет заголовок Content-Encoding: gzip
func GzipDecompressRequest() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Проверяем, сжато ли тело запроса
			if r.Header.Get("Content-Encoding") == "gzip" {
				gzipReader, err := gzip.NewReader(r.Body)
				if err != nil {
					http.Error(w, "Failed to decompress request", http.StatusBadRequest)
					return
				}
				defer gzipReader.Close()

				// Заменяем тело запроса на распакованный reader
				r.Body = io.NopCloser(gzipReader)

				// Удаляем заголовок, так как теперь body уже распакован
				r.Header.Del("Content-Encoding")

				// Обновляем Content-Length, так как размер изменился
				r.ContentLength = -1
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GzipCompressResponse сжимает ответы клиентам, которые это поддерживают
// Проверяет заголовок Accept-Encoding: gzip
func GzipCompressResponse() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Проверяем, поддерживает ли клиент gzip
			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				next.ServeHTTP(w, r)
				return
			}

			// Оборачиваем ResponseWriter в gzip writer
			gzipWriter := &gzipResponseWriter{
				ResponseWriter: w,
				gzipWriter:     gzip.NewWriter(w),
			}
			defer gzipWriter.Close()

			gzipWriter.Header().Set("Content-Encoding", "gzip")

			gzipWriter.Header().Del("Content-Length")

			next.ServeHTTP(gzipWriter, r)
		})
	}
}

// gzipResponseWriter оборачивает http.ResponseWriter для поддержки gzip
type gzipResponseWriter struct {
	http.ResponseWriter
	gzipWriter *gzip.Writer
	written    bool
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {

	if !w.written {
		w.written = true
		contentType := w.Header().Get("Content-Type")

		if !strings.HasPrefix(contentType, "application/json") &&
			!strings.HasPrefix(contentType, "text/html") &&
			!strings.HasPrefix(contentType, "text/plain") {

			return w.ResponseWriter.Write(b)
		}
	}

	return w.gzipWriter.Write(b)
}

func (w *gzipResponseWriter) Close() error {
	if w.gzipWriter != nil {
		return w.gzipWriter.Close()
	}
	return nil
}
