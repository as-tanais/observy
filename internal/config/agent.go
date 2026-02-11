package config

import (
	"flag"
	"fmt"
	"time"
)

// AgentConfig содержит конфигурацию агента сбора метрик
// Значение получаются из переменных окружения(ПРИОРИТЕТ) или флагов командной строки
type AgentConfig struct {
	// Address - адрес и порт куда будут отправляться метрики
	// default : "localhost:8080"
	ServerAddress string

	// PollInterval — интервал сбора метрик (в секундах).
	// default: 2 секунд.
	PollInterval time.Duration

	// ReportInterval — интервал отправки метрик на сервер (в секундах).
	// default: 2 секунд.
	ReportInterval time.Duration

	// Key — секретный ключ для подписи запросов между сервером и агентом(и).
	// Если нет то проверки нет
	Key string

	// RateLimit — ограничение количества одновременных запросов к серверу.
	// По умолчанию: 1.
	// Минимальное значение: 1.
	RateLimit int
}

// NewAgentConfig Возращает указатель на конфиг или ошибку если не удалось получить обязательные параметры
// Пример использования:
//
//		cfg, err := config.NewAgentConfig()
//		if err != nil {
//		    log.Fatal(err)
//		}
//		fmt.Printf("Agent will collect metrics every %s and send to %s every %s\n",
//	   cfg.PollInterval, cfg.ServerAddress, cfg.ReportInterval)
func NewAgentConfig() (*AgentConfig, error) {
	cfg := &AgentConfig{}

	addrFlag := flag.String("a", "localhost:8080", "Server address host:port")
	pollFlag := flag.Int("p", 2, "Poll interval in seconds")
	reportFlag := flag.Int("r", 10, "Report interval in seconds")
	keyFlag := flag.String("k", "", "Secret key for request singing")
	limitFlag := flag.Int("l", 1, "Rate limit")

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

	cfg.Key = GetEnvOrDefault("KEY", *keyFlag)

	rateLimit, err := GetEnvIntOrDefault("RATE_LIMIT", *limitFlag)
	if err != nil {
		return nil, fmt.Errorf("invalid rate limit: %w", err)
	}

	cfg.RateLimit = rateLimit

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate проверяет корректность конфигурации.
//
// Возвращает ошибку, если:
//   - адрес сервера пустой
//   - другие критические параметры имеют недопустимые значения
//
// Используется автоматически в NewAgentConfig после загрузки параметров.
func (c *AgentConfig) Validate() error {
	if c.PollInterval <= 0 {
		return fmt.Errorf("poll interval must be positive")
	}
	if c.ReportInterval <= 0 {
		return fmt.Errorf("report interval must be positive")
	}
	if c.RateLimit <= 0 {
		return fmt.Errorf("rate limit must be positive, got %d", c.RateLimit)
	}
	return nil
}

func (c *AgentConfig) ServerURL() string {
	return "http://" + c.ServerAddress
}

// PollsPerReport вычисляет количество сборов метрик между отправками на сервер.
//
// Возвращает:
//   - целое число >= 1 (минимум 1 сбор за период отправки)
//
// Пример:
//
//	cfg := &AgentConfig{
//	    PollInterval:   2 * time.Second,
//	    ReportInterval: 10 * time.Second,
//	}
//	fmt.Println(cfg.PollsPerReport()) // Выведет: 5
func (c *AgentConfig) PollsPerReport() int {
	polls := int(c.ReportInterval / c.PollInterval)
	if polls == 0 {
		return 1
	}
	return polls
}
