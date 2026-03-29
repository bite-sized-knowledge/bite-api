package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("DB_HOST", "")
	t.Setenv("JWT_ACCESS_EXPIRY", "")

	cfg := Load()

	if cfg.Port != "8080" {
		t.Fatalf("expected default port 8080, got %s", cfg.Port)
	}
	if cfg.DBHost != "127.0.0.1" {
		t.Fatalf("expected default DB host, got %s", cfg.DBHost)
	}
	if cfg.JWTAccessExpiry != 15*time.Minute {
		t.Fatalf("expected default access expiry, got %v", cfg.JWTAccessExpiry)
	}
}

func TestLoadOverrides(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("DB_HOST", "mysql")
	t.Setenv("JWT_ACCESS_EXPIRY", "30m")
	t.Setenv("JWT_REFRESH_EXPIRY", "48h")
	t.Setenv("RECSYS_BASE_URL", "http://recsys:8001")

	cfg := Load()

	if cfg.Port != "9090" || cfg.DBHost != "mysql" {
		t.Fatalf("env overrides not applied: %+v", cfg)
	}
	if cfg.JWTAccessExpiry != 30*time.Minute {
		t.Fatalf("expected 30m access expiry, got %v", cfg.JWTAccessExpiry)
	}
	if cfg.JWTRefreshExpiry != 48*time.Hour {
		t.Fatalf("expected 48h refresh expiry, got %v", cfg.JWTRefreshExpiry)
	}
	if cfg.RecsysBaseURL != "http://recsys:8001" {
		t.Fatalf("expected recsys override, got %s", cfg.RecsysBaseURL)
	}
	_ = os.Getenv("PORT")
}
