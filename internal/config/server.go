package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type ServerConfig struct {
	Address         string
	LogLevel        string
	StoreInterval   time.Duration
	FileStoragePath string
	Restore         bool
	DBConfig
}

type DBConfig struct {
	DSN string
}

func NewServerConfig() (*ServerConfig, error) {
	cfg := &ServerConfig{}

	addrFlag := flag.String("a", "localhost:8080", "Server address host:port")
	logFlag := flag.String("l", "info", "Log level (debug, info, warn, error)")
	dbFlag := flag.String("d", "", "Database dsn")

	storeIntervalFlag := flag.Int("i", 300, "Store interval in seconds (0 = synchronous)")
	fileStoragePathFlag := flag.String("f", getDefaultFilePath(), "File storage path")
	restoreFlag := flag.Bool("r", true, "Restore previously saved values on startup")

	flag.Parse()

	cfg.Address = GetEnvOrDefault("ADDRESS", *addrFlag)
	cfg.LogLevel = GetEnvOrDefault("LOG_LEVEL", *logFlag)
	cfg.DSN = GetEnvOrDefault("DATABASE_DSN", *dbFlag)

	if envInterval := os.Getenv("STORE_INTERVAL"); envInterval != "" {
		interval, err := strconv.Atoi(envInterval)
		if err != nil {
			return nil, fmt.Errorf("invalid STORE_INTERVAL: %w", err)
		}
		cfg.StoreInterval = time.Duration(interval) * time.Second
	} else {
		cfg.StoreInterval = time.Duration(*storeIntervalFlag) * time.Second
	}

	cfg.FileStoragePath = GetEnvOrDefault("FILE_STORAGE_PATH", *fileStoragePathFlag)

	if envRestore := os.Getenv("RESTORE"); envRestore != "" {
		restore, err := strconv.ParseBool(envRestore)
		if err != nil {
			return nil, fmt.Errorf("invalid RESTORE value: %w", err)
		}
		cfg.Restore = restore
	} else {
		cfg.Restore = *restoreFlag
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *ServerConfig) Validate() error {
	if c.Address == "" {
		return fmt.Errorf("server address cannot be empty")
	}
	return nil
}

func getDefaultFilePath() string {

	wd, err := os.Getwd()
	if err != nil {

		return "/tmp/metrics-backup.json"
	}

	return filepath.Join(wd, "metrics-backup.json")
}
