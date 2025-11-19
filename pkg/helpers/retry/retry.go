package retry

import (
	"fmt"
	"time"
)

// Количество повторов должно быть ограничено тремя дополнительными попытками.
// Интервалы между повторами должны увеличиваться: 1s, 3s, 5s.

const retryAttempts = 3

var retryDelays = []time.Duration{
	1 * time.Second,
	3 * time.Second,
	5 * time.Second,
}

type IsRetriable func(err error) bool

func WithBackoff(fn func() error, isRetriable IsRetriable) error {
	var err error

	for attempt := 0; attempt <= retryAttempts; attempt++ {
		err = fn()
		if err == nil {
			return nil
		}

		if !isRetriable(err) {
			return fmt.Errorf("non-retriable error: %w", err)
		}

		if attempt < retryAttempts {
			time.Sleep(retryDelays[attempt])
		}
	}

	return fmt.Errorf("failed after %d retries: %w", retryAttempts, err)
}
