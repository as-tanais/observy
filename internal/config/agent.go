package config

import (
	"flag"
	"fmt"
	"time"
)

type AgentConfig struct {
	ServerAddress  string
	PollInterval   time.Duration
	ReportInterval time.Duration
}

func NewAgentConfig() (*AgentConfig, error) {
	cfg := &AgentConfig{}

	addrFlag := flag.String("a", "localhost:8080", "Server address host:port")
	pollFlag := flag.Int("p", 2, "Poll interval in seconds")
	reportFlag := flag.Int("r", 10, "Report interval in seconds")

	flag.Parse()

	cfg.ServerAddress = GetEnvOrDefault("ADDRESS", *addrFlag)

	pollSec, err := GetEnvIntOrDefault("POLL_INTERVAL", *pollFlag)
	if err != nil {
		return nil, fmt.Errorf("invalid POLL_INTERVAL: %w", err)
	}
	cfg.PollInterval = time.Duration(pollSec) * time.Second

	reportSec, err := GetEnvIntOrDefault("REPORT_INTERVAL", *reportFlag)
	if err != nil {
		return nil, fmt.Errorf("invalid REPORT_INTERVAL: %w", err)
	}
	cfg.ReportInterval = time.Duration(reportSec) * time.Second

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *AgentConfig) Validate() error {
	if c.PollInterval <= 0 {
		return fmt.Errorf("poll interval must be positive")
	}
	if c.ReportInterval <= 0 {
		return fmt.Errorf("report interval must be positive")
	}
	return nil
}

func (c *AgentConfig) ServerURL() string {
	return "http://" + c.ServerAddress
}

func (c *AgentConfig) PollsPerReport() int {
	polls := int(c.ReportInterval / c.PollInterval)
	if polls == 0 {
		return 1
	}
	return polls
}
