package config

import (
	"testing"
	"time"
)

func TestLoadSuccess(t *testing.T) {
	t.Setenv("DB_DSN", "postgres://user:pass@localhost:5432/db?sslmode=disable")
	t.Setenv("ADMIN_SECRET", "admin")
	t.Setenv("USER_SECRET", "user")
	t.Setenv("RATE_LIMIT_REQUESTS", "5")
	t.Setenv("RATE_LIMIT_INTERVAL", "2s")
	t.Setenv("IDEMPOTENCY_TTL", "30s")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if cfg.Database.DSN == "" || cfg.Auth.AdminSecret != "admin" || cfg.Auth.UserSecret != "user" {
		t.Fatalf("unexpected config %+v", cfg)
	}
	if cfg.RateLimit.Requests != 5 || cfg.RateLimit.Interval != 2*time.Second {
		t.Fatalf("unexpected rate limit %+v", cfg.RateLimit)
	}
	if cfg.Idempotency.TTL != 30*time.Second {
		t.Fatalf("unexpected ttl %v", cfg.Idempotency.TTL)
	}
}

func TestLoadMissingRequired(t *testing.T) {
	t.Setenv("DB_DSN", "")
	t.Setenv("ADMIN_SECRET", "")
	t.Setenv("USER_SECRET", "")
	if _, err := Load(); err == nil {
		t.Fatalf("expected error for missing envs")
	}
}

func TestLoadInvalidDuration(t *testing.T) {
	t.Setenv("DB_DSN", "postgres://user:pass@localhost/db")
	t.Setenv("ADMIN_SECRET", "admin")
	t.Setenv("USER_SECRET", "user")
	t.Setenv("HTTP_READ_TIMEOUT", "not-duration")
	if _, err := Load(); err == nil {
		t.Fatalf("expected error for invalid duration")
	}
}

func TestLoadInvalidInt(t *testing.T) {
	t.Setenv("DB_DSN", "postgres://user:pass@localhost/db")
	t.Setenv("ADMIN_SECRET", "admin")
	t.Setenv("USER_SECRET", "user")
	t.Setenv("RATE_LIMIT_REQUESTS", "abc")
	if _, err := Load(); err == nil {
		t.Fatalf("expected error for invalid int")
	}
}
