package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
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

	models "github.com/as-tanais/observy/internal/model"
	"github.com/as-tanais/observy/pkg/helpers/retry"
)

var client = &http.Client{
	Timeout: 10 * time.Second,
}

func Send(metrics []models.Metrics, serverAddress string, key string) {
	for _, metric := range metrics {
		if err := sendMetricJSON(metric, serverAddress, key); err != nil {
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

func sendMetricJSON(metric models.Metrics, serverAddress string, key string) error {
	if err := validateMetric(metric); err != nil {
		return err
	}

	jsonData, err := json.Marshal(metric)
	if err != nil {
		return fmt.Errorf("json marshal error: %w", err)
	}

	hash := calculateHashSHA256(jsonData, key)

	log.Printf("AGENT DEBUG: JSON=%s", string(jsonData))
	log.Printf("AGENT DEBUG: KEY=%q", key)
	log.Printf("AGENT DEBUG: HASH=%s", hash)

	// Сжимаем JSON-данные с помощью gzip
	var compressedBuffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressedBuffer)

	if _, err := gzipWriter.Write(jsonData); err != nil {
		return fmt.Errorf("gzip write error: %w", err)
	}

	if err := gzipWriter.Close(); err != nil {
		return fmt.Errorf("gzip close error: %w", err)
	}

	compressedData := compressedBuffer.Bytes()
	url := fmt.Sprintf("%s/update/", serverAddress)

	return retry.WithBackoff(func() error {

		req, err := newCompressedJSONRequest(http.MethodPost, url, compressedData)
		if err != nil {
			return fmt.Errorf("creating request error: %w", err)
		}

		if hash != "" {
			req.Header.Set("HashSHA256", hash)
			log.Printf("SET HashSHA256: %s", hash)
		} else {
			log.Printf("NO HashSHA256 (key=%q)", key)
		}

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

func SendBatchMetrics(metrics []models.Metrics, serverAddress string, key string) error {
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

	var compressedBuffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressedBuffer)
	if _, err := gzipWriter.Write(jsonData); err != nil {
		return fmt.Errorf("gzip write error: %w", err)
	}
	if err := gzipWriter.Close(); err != nil {
		return fmt.Errorf("gzip close error: %w", err)
	}

	compressedData := compressedBuffer.Bytes()
	url := fmt.Sprintf("%s/updates/", serverAddress)

	return retry.WithBackoff(func() error {
		req, err := newCompressedJSONRequest(http.MethodPost, url, compressedData)
		if err != nil {
			return fmt.Errorf("creating request error: %w", err)
		}

		if hash != "" {
			req.Header.Set("HashSHA256", hash)
		}

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
