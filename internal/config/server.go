package config

import (
	"flag"
	"fmt"
)

type ServerConfig struct {
	Address string
}

func NewServerConfig() (*ServerConfig, error) {
	cfg := &ServerConfig{}

	addrFlag := flag.String("a", "localhost:8080", "Server address host:port")

	flag.Parse()

	cfg.Address = GetEnvOrDefault("ADDRESS", *addrFlag)

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
