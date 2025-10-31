package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

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

				r.Body = io.NopCloser(gzipReader)

				r.Header.Del("Content-Encoding")

				r.ContentLength = -1
			}

			next.ServeHTTP(w, r)
		})
	}
}

func GzipCompressResponse() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				next.ServeHTTP(w, r)
				return
			}

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

type gzipResponseWriter struct {
	http.ResponseWriter
	gzipWriter *gzip.Writer
	written    bool
}

func (w *gzipResponseWriter) WriteHeader(statusCode int) {

	contentType := w.Header().Get("Content-Type")

	if !strings.HasPrefix(contentType, "application/json") &&
		!strings.HasPrefix(contentType, "text/html") &&
		!strings.HasPrefix(contentType, "text/plain") {

		w.ResponseWriter.WriteHeader(statusCode)
		return
	}

	w.ResponseWriter.WriteHeader(statusCode)
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
