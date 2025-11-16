package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Database struct {
		DSN                string
		ConnectRetries     int
		ConnectRetryInterval time.Duration
	}
	HTTP struct {
		Addr            string
		ReadTimeout     time.Duration
		WriteTimeout    time.Duration
		ShutdownTimeout time.Duration
	}
	Auth struct {
		AdminSecret string
		UserSecret  string
	}
	RateLimit struct {
		Requests           int
		Interval           time.Duration
		TrustForwardHeader bool
	}
	Idempotency struct {
		TTL time.Duration
	}
}

func Load() (Config, error) {
	var cfg Config
	var err error

	if cfg.Database.DSN, err = requiredEnv("DB_DSN"); err != nil {
		return cfg, err
	}
	if cfg.Database.ConnectRetries, err = intOrDefault("DB_CONNECT_RETRIES", 5); err != nil {
		return cfg, err
	}
	if cfg.Database.ConnectRetryInterval, err = durationOrDefault("DB_CONNECT_RETRY_INTERVAL", time.Second); err != nil {
		return cfg, err
	}

	cfg.HTTP.Addr = envOrDefault("HTTP_ADDR", ":8080")
	if cfg.HTTP.ReadTimeout, err = durationOrDefault("HTTP_READ_TIMEOUT", 5*time.Second); err != nil {
		return cfg, err
	}
	if cfg.HTTP.WriteTimeout, err = durationOrDefault("HTTP_WRITE_TIMEOUT", 5*time.Second); err != nil {
		return cfg, err
	}
	if cfg.HTTP.ShutdownTimeout, err = durationOrDefault("HTTP_SHUTDOWN_TIMEOUT", 5*time.Second); err != nil {
		return cfg, err
	}

	if cfg.Auth.AdminSecret, err = requiredEnv("ADMIN_SECRET"); err != nil {
		return cfg, err
	}
	if cfg.Auth.UserSecret, err = requiredEnv("USER_SECRET"); err != nil {
		return cfg, err
	}

	if cfg.RateLimit.Requests, err = intOrDefault("RATE_LIMIT_REQUESTS", 10); err != nil {
		return cfg, err
	}
	if cfg.RateLimit.Interval, err = durationOrDefault("RATE_LIMIT_INTERVAL", time.Second); err != nil {
		return cfg, err
	}
	if cfg.RateLimit.TrustForwardHeader, err = boolOrDefault("RATE_LIMIT_TRUST_FORWARD", false); err != nil {
		return cfg, err
	}

	if cfg.Idempotency.TTL, err = durationOrDefault("IDEMPOTENCY_TTL", time.Minute); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func requiredEnv(key string) (string, error) {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val, nil
	}
	return "", fmt.Errorf("missing required env %s", key)
}

func envOrDefault(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}

func durationOrDefault(key string, fallback time.Duration) (time.Duration, error) {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		d, err := time.ParseDuration(val)
		if err != nil {
			return 0, fmt.Errorf("env %s has invalid duration %q: %w", key, val, err)
		}
		return d, nil
	}
	return fallback, nil
}

func intOrDefault(key string, fallback int) (int, error) {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		n, err := strconv.Atoi(val)
		if err != nil {
			return 0, fmt.Errorf("env %s has invalid int %q: %w", key, val, err)
		}
		return n, nil
	}
	return fallback, nil
}

func boolOrDefault(key string, fallback bool) (bool, error) {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		b, err := strconv.ParseBool(val)
		if err != nil {
			return false, fmt.Errorf("env %s has invalid bool %q: %w", key, val, err)
		}
		return b, nil
	}
	return fallback, nil
}
