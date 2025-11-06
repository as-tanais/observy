package server

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
)

func dbMigrate(dsn string) error {
	m, err := migrate.New("file://./migrations", dsn)
	if err != nil {
		return fmt.Errorf("new migrate: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}
