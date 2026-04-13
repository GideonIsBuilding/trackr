package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	// Server
	Port string

	// Database
	DatabaseURL string

	// Auth
	JWTSecret        string
	JWTExpiry        time.Duration
	JWTRefreshExpiry time.Duration

	// Reminder engine
	ReminderCheckInterval time.Duration

	// Email (SMTP)
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	FromAddress  string
	FromName     string
	AppURL       string
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:                  getEnv("PORT", "8080"),
		DatabaseURL:           mustGetEnv("DATABASE_URL"),
		JWTSecret:             mustGetEnv("JWT_SECRET"),
		JWTExpiry:             getDurationEnv("JWT_EXPIRY_HOURS", 24) * time.Hour,
		JWTRefreshExpiry:      getDurationEnv("JWT_REFRESH_EXPIRY_DAYS", 7) * 24 * time.Hour,
		ReminderCheckInterval: getDurationEnv("REMINDER_CHECK_INTERVAL_MINUTES", 60) * time.Minute,

		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     getEnv("SMTP_PORT", "587"),
		SMTPUsername: getEnv("SMTP_USERNAME", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		FromAddress:  getEnv("FROM_ADDRESS", "noreply@trackr.app"),
		FromName:     getEnv("FROM_NAME", "Trackr"),
		AppURL:       getEnv("APP_URL", "http://localhost:5173"),
	}

	if len(cfg.JWTSecret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func mustGetEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("required environment variable %q is not set", key))
	}
	return v
}

func getDurationEnv(key string, fallback int64) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return time.Duration(fallback)
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return time.Duration(fallback)
	}
	return time.Duration(n)
}
