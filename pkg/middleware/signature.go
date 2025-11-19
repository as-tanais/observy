package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"net/http"
)

func SignatureMiddleware(secretKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if secretKey != "" {

				if r.ContentLength > 0 {
					body, err := io.ReadAll(r.Body)
					if err != nil {
						http.Error(w, "failed to read request body", http.StatusBadRequest)
						return
					}
					r.Body.Close()

					receivedHash := r.Header.Get("HashSHA256")
					if receivedHash == "" {
						http.Error(w, "missing HashSHA256 header", http.StatusBadRequest)
						return
					}

					// Считаем ожидаемый хеш от тела + секретного ключа
					mac := hmac.New(sha256.New, []byte(secretKey))
					mac.Write(body)
					expectedHash := base64.StdEncoding.EncodeToString(mac.Sum(nil))

					if receivedHash != expectedHash {
						http.Error(w, "invalid HashSHA256", http.StatusBadRequest)
						return
					}

					r.Body = io.NopCloser(bytes.NewReader(body))
				}

			}

			if secretKey != "" {

				recorder := newResponseRecorder(w)

				next.ServeHTTP(recorder, r)

				mac := hmac.New(sha256.New, []byte(secretKey))
				mac.Write(recorder.buf.Bytes())
				hash := base64.StdEncoding.EncodeToString(mac.Sum(nil))

				for k, vv := range recorder.Header() {
					for _, v := range vv {
						w.Header().Add(k, v)
					}
				}

				w.Header().Set("HashSHA256", hash)

				w.WriteHeader(recorder.statusCode)
				w.Write(recorder.buf.Bytes())
			} else {

				next.ServeHTTP(w, r)
			}
		})
	}
}

type responseRecorder struct {
	http.ResponseWriter
	buf        *bytes.Buffer
	statusCode int
}

func newResponseRecorder(w http.ResponseWriter) *responseRecorder {
	return &responseRecorder{
		ResponseWriter: w,
		buf:            &bytes.Buffer{},
		statusCode:     http.StatusOK,
	}
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	return r.buf.Write(b)
}

func (r *responseRecorder) Header() http.Header {
	return r.ResponseWriter.Header()
}
