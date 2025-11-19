package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"log"
	"net/http"
)

func SignatureMiddleware(secretKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			shouldVerify := secretKey != "" && (r.URL.Path == "/update/" || r.URL.Path == "/updates/")

			if shouldVerify {
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

				log.Printf("SERVER DEBUG: PATH=%s", r.URL.Path)
				log.Printf("SERVER DEBUG: JSON=%s", string(body))
				log.Printf("SERVER DEBUG: KEY=%q", secretKey)
				log.Printf("SERVER DEBUG: RECEIVED_HASH=%s", receivedHash)

				mac := hmac.New(sha256.New, []byte(secretKey))
				mac.Write(body)
				expectedHash := base64.StdEncoding.EncodeToString(mac.Sum(nil))

				log.Printf("SERVER DEBUG: EXPECTED_HASH=%s", expectedHash)

				if receivedHash != expectedHash {
					http.Error(w, "invalid HashSHA256", http.StatusBadRequest)
					return
				}

				r.Body = io.NopCloser(bytes.NewReader(body))
			}

			if secretKey != "" {

				recorder := newResponseRecorder(w)

				next.ServeHTTP(recorder, r)

				mac := hmac.New(sha256.New, []byte(secretKey))
				mac.Write(recorder.buf.Bytes())
				hash := base64.StdEncoding.EncodeToString(mac.Sum(nil))

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
