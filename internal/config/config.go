package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Host           string
	Port           int
	DBPath         string
	APIKey         string
	AdminPassword  string
	MaxSessions    int
	LogLevel       string
	PublicURL      string
}

func Load() (*Config, error) {
	envFile := os.Getenv("WHATSAR_ENV_FILE")
	if envFile == "" {
		envFile = ".env"
	}
	_ = godotenv.Load(envFile)

	port, _ := strconv.Atoi(envOr("WHATSAR_PORT", "8080"))
	maxSessions, _ := strconv.Atoi(envOr("WHATSAR_MAX_SESSIONS", "5"))

	return &Config{
		Host:          envOr("WHATSAR_HOST", "127.0.0.1"),
		Port:          port,
		DBPath:        envOr("WHATSAR_DB_PATH", "./data/whatsar.db"),
		APIKey:        envOr("WHATSAR_API_KEY", "dev-api-key"),
		AdminPassword: envOr("WHATSAR_ADMIN_PASSWORD", "admin"),
		MaxSessions:   maxSessions,
		LogLevel:      envOr("WHATSAR_LOG_LEVEL", "info"),
		PublicURL:     envOr("WHATSAR_PUBLIC_URL", ""),
	}, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}