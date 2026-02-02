package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// ServerConfig содержит конфигурацию сервера метрик
// Значение получаются из переменных окружения(ПРИОРИТЕТ) или флагов командной строки
type ServerConfig struct {
	// Address - адрес и порт для запуска сервера
	// default : "localhost:8080"
	Address string

	// LogLevel — уровень логирования (info, debub ...)
	// default : "info".
	LogLevel string

	// StoreInterval — интервал автосохранения метрик в файл (в секундах).
	// Значение 0 означает синхронное сохранение после каждой операции.
	// default: 300 секунд.
	StoreInterval time.Duration

	// FileStoragePath — путь к файлу для персистентного хранения метрик.
	// default: "./metrics-backup.json" в рабочей директории.
	FileStoragePath string

	// Restore — флаг восстановления метрик из файла при старте приложения.
	// Если true — метрики загружаются из файла, указанного в FileStoragePath.
	Restore bool

	// Key — секретный ключ для подписи запросов между сервером и агентом(и).
	// Если нет то проверки нет
	Key string
	DBConfig

	// AuditFile — путь к файлу для записи аудит-логов операций с метриками.
	// Если пустой — файловый аудит отключён.
	AuditFile string

	// AuditURL — URL для отправки аудит-логов по HTTP
	// Если пустой — HTTP-аудит отключён.
	AuditURL string
}

// DBConfig содержит параметры подключения к базе данных PostgreSQL.
type DBConfig struct {

	// DSN : "postgresql://user:password@host:port/dbname?sslmode=disable"
	DSN string
}

// NewServerConfig создает новый ServerConfig
// Возращает указатель на конфиг или ошибку если не удалось получить обязательные параметры
// Пример использования:
//
//	cfg, err := config.NewServerConfig()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("Server will start on", cfg.Address)
func NewServerConfig() (*ServerConfig, error) {
	cfg := &ServerConfig{}

	addrFlag := flag.String("a", "localhost:8080", "Server address host:port")
	logFlag := flag.String("l", "info", "Log level (debug, info, warn, error)")
	dbFlag := flag.String("d", "", "Database dsn")

	storeIntervalFlag := flag.Int("i", 300, "Store interval in seconds (0 = synchronous)")
	fileStoragePathFlag := flag.String("f", getDefaultFilePath(), "File storage path")
	restoreFlag := flag.Bool("r", true, "Restore previously saved values on startup")

	keyFlag := flag.String("k", "", "Secret key for request singing")

	auditFileFlag := flag.String("audit-file", "", "file audit logs path")
	auditURLFlag := flag.String("audit-url", "", "url audit logs address")

	flag.Parse()

	cfg.Address = GetEnvOrDefault("ADDRESS", *addrFlag)
	cfg.LogLevel = GetEnvOrDefault("LOG_LEVEL", *logFlag)
	cfg.DSN = GetEnvOrDefault("DATABASE_DSN", *dbFlag)

	cfg.AuditFile = GetEnvOrDefault("AUDIT_FILE", *auditFileFlag)
	cfg.AuditURL = GetEnvOrDefault("AUDIT_URL", *auditURLFlag)

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

	cfg.Key = GetEnvOrDefault("KEY", *keyFlag)

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
// Используется автоматически в NewServerConfig после загрузки параметров.
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
