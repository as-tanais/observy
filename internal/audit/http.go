package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type HTTPSub struct {
	url     string
	client  *http.Client
	timeout time.Duration
}

func NewHTTPSub(url string) *HTTPSub {
	return &HTTPSub{
		url:     url,
		client:  &http.Client{Timeout: 5 * time.Second},
		timeout: 5 * time.Second,
	}
}

func (h *HTTPSub) Send(ctx context.Context, event AuditEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", h.url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	return nil
}
