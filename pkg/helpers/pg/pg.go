package pg

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPoolConfig(dsn string) (*pgxpool.Config, error) {
	return pgxpool.ParseConfig(dsn)
}

func NewConnection(ctx context.Context, poolConfig *pgxpool.Config) (*pgxpool.Pool, error) {
	return pgxpool.NewWithConfig(ctx, poolConfig)
}
