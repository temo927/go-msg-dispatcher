package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port            string
	DBDSN           string
	BatchSize       int
	TickInterval    time.Duration
	MaxMessageChars int
	MaxRetries      int

	WebhookURL        string
	WebhookAuthHeader string
	WebhookAuthValue  string
	AcceptAny2xx      bool

	RedisAddr     string
	RedisPassword string
	RedisDB       int
	RedisTTL      time.Duration
}

func Load() *Config {
	cfg := &Config{}

	cfg.Port = getEnv("PORT", "8080")
	cfg.DBDSN = getEnv("DB_DSN", "postgres://postgres:postgres@postgres:5432/msgsvc?sslmode=disable")
	cfg.BatchSize = getEnvInt("BATCH_SIZE", 2)
	cfg.TickInterval = getEnvDuration("TICK_INTERVAL", 2*time.Minute)
	cfg.MaxMessageChars = getEnvInt("MAX_MESSAGE_CHARS", 1000)
	cfg.MaxRetries = getEnvInt("MAX_RETRIES", 5)

	cfg.WebhookURL = getEnv("WEBHOOK_URL", "")
	cfg.WebhookAuthHeader = getEnv("WEBHOOK_AUTH_HEADER", "")
	cfg.WebhookAuthValue = getEnv("WEBHOOK_AUTH_VALUE", "")
	cfg.AcceptAny2xx = getEnvBool("ACCEPT_ANY_2XX", false)

	cfg.RedisAddr = getEnv("REDIS_ADDR", "redis:6379")
	cfg.RedisPassword = os.Getenv("REDIS_PASSWORD")
	cfg.RedisDB = getEnvInt("REDIS_DB", 0)
	cfg.RedisTTL = getEnvDuration("REDIS_SENT_META_TTL", 7*24*time.Hour)

	return cfg
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func getEnvBool(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		return v == "true" || v == "1"
	}
	return def
}

func getEnvDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
