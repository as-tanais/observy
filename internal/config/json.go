package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"
)

// JSONServerConfig представляет структуру JSON-файла конфигурации сервера
type JSONServerConfig struct {
	Address         string `json:"address"`
	LogLevel        string `json:"log_level"`
	StoreInterval   string `json:"store_interval"`
	FileStoragePath string `json:"file_storage_path"`
	Restore         bool   `json:"restore"`
	Key             string `json:"key"`
	DSN             string `json:"database_dsn"`
	AuditFile       string `json:"audit_file"`
	AuditURL        string `json:"audit_url"`
	CryptoKeyPath   string `json:"crypto_key"`
}

type JSONAgentConfig struct {
	Address        string `json:"address"`
	LogLevel       string `json:"log_level"`
	PollInterval   string `json:"poll_interval"`
	ReportInterval string `json:"report_interval"`
	Key            string `json:"key"`
	RateLimit      int    `json:"rate_limit"`
	CryptoKeyPath  string `json:"crypto_key"`
}

func LoadFromJSON(path string) (*JSONServerConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg JSONServerConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (j *JSONServerConfig) ApplyToServerConfig(cfg *ServerConfig) {
	if cfg.Address == "" && j.Address != "" {
		cfg.Address = j.Address
	}

	if cfg.LogLevel == "" && j.LogLevel != "" {
		cfg.LogLevel = j.LogLevel
	}

	if j.StoreInterval != "" {
		if interval, err := time.ParseDuration(j.StoreInterval); err == nil && cfg.StoreInterval == 0 {
			cfg.StoreInterval = interval
		}
	}

	if cfg.FileStoragePath == "" && j.FileStoragePath != "" {
		cfg.FileStoragePath = j.FileStoragePath
	}

	if !cfg.Restore && j.Restore {
		cfg.Restore = j.Restore
	}

	if cfg.Key == "" && j.Key != "" {
		cfg.Key = j.Key
	}

	if cfg.DSN == "" && j.DSN != "" {
		cfg.DSN = j.DSN
	}

	if cfg.AuditFile == "" && j.AuditFile != "" {
		cfg.AuditFile = j.AuditFile
	}

	if cfg.AuditURL == "" && j.AuditURL != "" {
		cfg.AuditURL = j.AuditURL
	}

	if cfg.CryptoKeyPath == "" && j.CryptoKeyPath != "" {
		cfg.CryptoKeyPath = j.CryptoKeyPath
	}
}

// Для АГЕНТА

// LoadAgentFromJSON загружает конфигурацию агента из JSON-файла
func LoadAgentFromJSON(path string) (*JSONAgentConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg JSONAgentConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse JSON config: %w", err)
	}

	return &cfg, nil
}

// ApplyToAgentConfig применяет значения из JSON к основному конфигу агента
// (только если они не были установлены через флаги или env)
func (j *JSONAgentConfig) ApplyToAgentConfig(cfg *AgentConfig) {
	// ServerAddress
	if cfg.ServerAddress == "" && j.Address != "" {
		cfg.ServerAddress = j.Address
	}

	// PollInterval
	if j.PollInterval != "" && cfg.PollInterval == 0 {
		if interval, err := parseDuration(j.PollInterval); err == nil {
			cfg.PollInterval = interval
		}
	}

	// ReportInterval
	if j.ReportInterval != "" && cfg.ReportInterval == 0 {
		if interval, err := parseDuration(j.ReportInterval); err == nil {
			cfg.ReportInterval = interval
		}
	}

	// Key
	if cfg.Key == "" && j.Key != "" {
		cfg.Key = j.Key
	}

	// RateLimit (только если не установлен и значение > 0)
	if cfg.RateLimit == 0 && j.RateLimit > 0 {
		cfg.RateLimit = j.RateLimit
	}

	// CryptoKey
	if cfg.CryptoKey == "" && j.CryptoKeyPath != "" {
		cfg.CryptoKey = j.CryptoKeyPath
	}
}

// Вспомогательная функция для парсинга длительности из строки
// Поддерживает форматы: "10s", "1m", "300" (как секунды)
func parseDuration(value string) (time.Duration, error) {
	// Сначала пробуем распарсить как стандартную длительность (с суффиксом)
	if duration, err := time.ParseDuration(value); err == nil {
		return duration, nil
	}

	// Если не получилось, пробуем как число секунд
	seconds, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid duration format: %s", value)
	}

	return time.Duration(seconds) * time.Second, nil
}
