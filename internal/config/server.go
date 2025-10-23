package config

import (
	"flag"
	"fmt"
)

type ServerConfig struct {
	Address  string
	LogLevel string
}

func NewServerConfig() (*ServerConfig, error) {
	cfg := &ServerConfig{}

	addrFlag := flag.String("a", "localhost:8080", "Server address host:port")
	logFlag := flag.String("l", "info", "Log level (debug, info, warn, error)")

	flag.Parse()

	cfg.Address = GetEnvOrDefault("ADDRESS", *addrFlag)
	cfg.LogLevel = GetEnvOrDefault("LOG_LEVEL", *logFlag)

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
