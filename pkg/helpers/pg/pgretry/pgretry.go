package pgretry

import (
	"context"
	"errors"
	"net"

	pgerrors "github.com/as-tanais/observy/pkg/helpers/pg/errors"
	"github.com/jackc/pgx/v5"
)

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
