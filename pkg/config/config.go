package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Database DatabaseConfig
	Redis    RedisConfig
	App      AppConfig
	Message  MessageConfig
	Webhook  WebhookConfig
	Seed     SeedConfig
}

type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
	CacheTTL time.Duration
}

type AppConfig struct {
	Port                    string
	Env                     string
	LogLevel                string
	GracefulShutdownTimeout time.Duration
	APIToken                string
}

type MessageConfig struct {
	BatchSize       int
	IntervalSeconds int
	MaxRetries      int
	CharLimit       int
	WorkerCount     int
}

type WebhookConfig struct {
	URL                 string
	AuthKey             string
	TimeoutSeconds      int
	MaxRetries          int
	RateLimitPerSecond  int
}

type SeedConfig struct {
	MessageCount int
}

func Load() (*Config, error) {
	cfg := &Config{
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "messaging_user"),
			Password:        getEnv("DB_PASSWORD", "secure_password_123"),
			Name:            getEnv("DB_NAME", "messaging_db"),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
			CacheTTL: getEnvAsDuration("REDIS_CACHE_TTL", 168*time.Hour),
		},
		App: AppConfig{
			Port:                    getEnv("APP_PORT", "8080"),
			Env:                     getEnv("APP_ENV", "development"),
			LogLevel:                getEnv("LOG_LEVEL", "info"),
			GracefulShutdownTimeout: getEnvAsDuration("GRACEFUL_SHUTDOWN_TIMEOUT", 30*time.Second),
			APIToken:                getEnv("API_TOKEN", ""),
		},
		Message: MessageConfig{
			BatchSize:       getEnvAsInt("MESSAGE_BATCH_SIZE", 2),
			IntervalSeconds: getEnvAsInt("MESSAGE_INTERVAL_SECONDS", 10),
			MaxRetries:      getEnvAsInt("MESSAGE_MAX_RETRIES", 3),
			CharLimit:       getEnvAsInt("MESSAGE_CHAR_LIMIT", 160),
			WorkerCount:     getEnvAsInt("MESSAGE_WORKER_COUNT", 5),
		},
		Webhook: WebhookConfig{
			URL:                getEnv("WEBHOOK_URL", "https://webhook.site/c3f13233-1ed4-429e-9649-8133b3b9c9cd"),
			AuthKey:            getEnv("WEBHOOK_AUTH_KEY", "INS.me1x9uMcyYGlhKKQVPoc.bO3j9aZwRTOcA2Ywo"),
			TimeoutSeconds:     getEnvAsInt("WEBHOOK_TIMEOUT_SECONDS", 30),
			MaxRetries:         getEnvAsInt("WEBHOOK_MAX_RETRIES", 3),
			RateLimitPerSecond: getEnvAsInt("WEBHOOK_RATE_LIMIT_PER_SECOND", 10),
		},
		Seed: SeedConfig{
			MessageCount: getEnvAsInt("SEED_MESSAGE_COUNT", 100),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.Database.Host == "" {
		return fmt.Errorf("DB_HOST is required")
	}
	if c.Database.User == "" {
		return fmt.Errorf("DB_USER is required")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("DB_NAME is required")
	}
	if c.Webhook.URL == "" {
		return fmt.Errorf("WEBHOOK_URL is required")
	}
	if c.Webhook.AuthKey == "" {
		return fmt.Errorf("WEBHOOK_AUTH_KEY is required")
	}
	if c.Message.BatchSize < 1 {
		return fmt.Errorf("MESSAGE_BATCH_SIZE must be at least 1")
	}
	if c.Message.IntervalSeconds < 1 {
		return fmt.Errorf("MESSAGE_INTERVAL_SECONDS must be at least 1")
	}
	if c.Message.CharLimit < 1 {
		return fmt.Errorf("MESSAGE_CHAR_LIMIT must be at least 1")
	}
	return nil
}

func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

func (c *RedisConfig) Address() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	return defaultValue
}
