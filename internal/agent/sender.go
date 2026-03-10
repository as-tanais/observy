package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"syscall"
	"time"

	"github.com/as-tanais/observy/internal/crypto"
	models "github.com/as-tanais/observy/internal/model"
	"github.com/as-tanais/observy/pkg/helpers/retry"
)

var client = &http.Client{
	Timeout: 10 * time.Second,
}

func Send(metrics []models.Metrics, serverAddress string, key string, publicKey *rsa.PublicKey) {
	for _, metric := range metrics {
		if err := sendMetricJSON(metric, serverAddress, key, publicKey); err != nil {
			log.Printf("sending error %s: %v", metric.ID, err)
		}
	}
}

func sendMetric(metric models.Metrics, serverAddress string) error {
	var value string

	switch metric.MType {
	case models.Gauge:
		if metric.Value == nil {
			return fmt.Errorf("gauge metric %s: value nil", metric.ID)
		}
		value = fmt.Sprintf("%f", *metric.Value)
	case models.Counter:
		if metric.Delta == nil {
			return fmt.Errorf("counter metric %s: value nil", metric.ID)
		}
		value = fmt.Sprintf("%d", *metric.Delta)
	default:
		return fmt.Errorf("unknown type of metric: %s", metric.MType)
	}

	url := fmt.Sprintf("%s/update/%s/%s/%s", serverAddress, metric.MType, metric.ID, value)

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("creating request error: %w", err)
	}

	req.Header.Set("Content-Type", "text/plain")
	addRealIPHeader(req)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("sending error: %w", err)
	}

	defer resp.Body.Close()

	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unknown error %d", resp.StatusCode)
	}

	return nil
}

func sendMetricJSON(metric models.Metrics, serverAddress string, key string, publicKey *rsa.PublicKey) error {
	if err := validateMetric(metric); err != nil {
		return err
	}

	jsonData, err := json.Marshal(metric)
	if err != nil {
		return fmt.Errorf("json marshal error: %w", err)
	}

	hash := calculateHashSHA256(jsonData, key)

	// Сжимаем JSON-данные
	var compressedBuffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressedBuffer)
	if _, err := gzipWriter.Write(jsonData); err != nil {
		return fmt.Errorf("gzip write error: %w", err)
	}
	if err := gzipWriter.Close(); err != nil {
		return fmt.Errorf("gzip close error: %w", err)
	}

	dataToSend := compressedBuffer.Bytes()
	contentEncoding := "gzip"

	// Шифруем, если есть ключ
	if publicKey != nil {
		encryptedData, err := crypto.Encrypt(dataToSend, publicKey)
		if err != nil {
			return fmt.Errorf("encryption error: %w", err)
		}
		dataToSend = encryptedData
		contentEncoding = "rsa"
	}

	url := fmt.Sprintf("%s/update/", serverAddress)

	return retry.WithBackoff(func() error {
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(dataToSend))
		if err != nil {
			return fmt.Errorf("creating request error: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Content-Encoding", contentEncoding)

		if hash != "" {
			req.Header.Set("HashSHA256", hash)
		}

		addRealIPHeader(req)

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("sending error: %w", err)
		}
		defer resp.Body.Close()
		_, _ = io.Copy(io.Discard, resp.Body)

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("server returned status %d", resp.StatusCode)
		}

		return nil
	}, isRetriable)
}

func validateMetric(metric models.Metrics) error {
	if metric.ID == "" {
		return fmt.Errorf("metric ID cannot be empty")
	}

	switch metric.MType {
	case models.Gauge:
		if metric.Value == nil {
			return fmt.Errorf("gauge metric %s: value is nil", metric.ID)
		}
	case models.Counter:
		if metric.Delta == nil {
			return fmt.Errorf("counter metric %s: delta is nil", metric.ID)
		}
	default:
		return fmt.Errorf("unknown metric type: %s", metric.MType)
	}

	return nil
}

func SendBatchMetrics(metrics []models.Metrics, serverAddress string, key string, publicKey *rsa.PublicKey) error {
	if len(metrics) == 0 {
		return nil
	}

	for _, metric := range metrics {
		if err := validateMetric(metric); err != nil {
			return err
		}
	}

	jsonData, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("json marshal error: %w", err)
	}

	hash := calculateHashSHA256(jsonData, key)

	// 1. Сжимаем данные
	var compressedBuffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressedBuffer)
	if _, err := gzipWriter.Write(jsonData); err != nil {
		return fmt.Errorf("gzip write error: %w", err)
	}
	if err := gzipWriter.Close(); err != nil {
		return fmt.Errorf("gzip close error: %w", err)
	}

	dataToSend := compressedBuffer.Bytes()
	contentEncoding := "gzip"

	// 2. Если есть публичный ключ - шифруем сжатые данные
	if publicKey != nil {
		encryptedData, err := crypto.Encrypt(dataToSend, publicKey)
		if err != nil {
			return fmt.Errorf("encryption error: %w", err)
		}
		dataToSend = encryptedData
		contentEncoding = "rsa" // Сервер ожидает этот заголовок
	}

	url := fmt.Sprintf("%s/updates/", serverAddress)

	return retry.WithBackoff(func() error {
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(dataToSend))
		if err != nil {
			return fmt.Errorf("creating request error: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Content-Encoding", contentEncoding)
		req.Header.Set("Accept-Encoding", "gzip")

		if hash != "" {
			req.Header.Set("HashSHA256", hash)
		}

		addRealIPHeader(req)

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("sending error: %w", err)
		}
		defer resp.Body.Close()

		io.Copy(io.Discard, resp.Body)

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("server returned status %d", resp.StatusCode)
		}

		return nil
	}, isRetriable)
}

func newCompressedJSONRequest(method string, url string, data []byte) (*http.Request, error) {
	req, err := http.NewRequest(method, url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")

	addRealIPHeader(req)

	return req, nil
}

func isRetriable(err error) bool {
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	if errors.Is(err, syscall.ECONNREFUSED) || errors.Is(err, syscall.ECONNRESET) {
		return true
	}

	return false
}

func calculateHashSHA256(data []byte, key string) string {
	if key == "" {
		return ""
	}
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write(data)
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
