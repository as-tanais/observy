package agent

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	models "github.com/as-tanais/observy/internal/model"
)

var client = &http.Client{
	Timeout: 10 * time.Second,
}

func Send(metrics []models.Metrics, serverAddress string) {
	for _, metric := range metrics {
		if err := sendMetric(metric, serverAddress); err != nil {
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
