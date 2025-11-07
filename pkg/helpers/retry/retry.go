package retry

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	pgerrors "github.com/as-tanais/observy/pkg/helpers/pg/errors"
	"github.com/jackc/pgx/v5"
)

// Количество повторов должно быть ограничено тремя дополнительными попытками.
// Интервалы между повторами должны увеличиваться: 1s, 3s, 5s.

const RetryAttempts = 3

var RetryDelays = []time.Duration{
	1 * time.Second,
	3 * time.Second,
	5 * time.Second,
}

func WithBackoff(fn func() error) error {
	var err error

	for attempt := 0; attempt <= RetryAttempts; attempt++ {
		err = fn()
		if err == nil {
			return nil
		}

		if !IsRetriable(err) {
			return fmt.Errorf("non-retriable error: %w", err)
		}

		if attempt < RetryAttempts {
			time.Sleep(RetryDelays[attempt])
		}
	}

	return fmt.Errorf("failed after %d retries: %w", RetryAttempts, err)
}

func IsRetriable(err error) bool {
	if err == nil {
		return false
	}

	if pgerrors.IsRetriable(err) {
		return true
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return false
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	return false
}
