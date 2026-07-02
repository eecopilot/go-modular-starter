package config

import (
	"strings"
	"testing"
	"time"
)

func lookup(values map[string]string) LookupFunc {
	return func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}
}

func TestLoadFromLookupDefaults(t *testing.T) {
	cfg, err := LoadFromLookup(lookup(nil))
	if err != nil {
		t.Fatalf("LoadFromLookup() error = %v", err)
	}

	if cfg.App.Name != "go-modular-starter" {
		t.Fatalf("unexpected app name: %q", cfg.App.Name)
	}
	if cfg.HTTP.Addr != ":8080" {
		t.Fatalf("unexpected http addr: %q", cfg.HTTP.Addr)
	}
	if cfg.HTTP.ShutdownTimeout != 15*time.Second {
		t.Fatalf("unexpected shutdown timeout: %s", cfg.HTTP.ShutdownTimeout)
	}
	if cfg.Userkit.Enabled {
		t.Fatal("userkit should be disabled by default")
	}
}

func TestLoadFromLookupValidatesUserkitWhenEnabled(t *testing.T) {
	_, err := LoadFromLookup(lookup(map[string]string{
		"USERKIT_ENABLED": "true",
	}))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "USERKIT_DATABASE_URL") {
		t.Fatalf("expected database URL error, got %v", err)
	}
	if !strings.Contains(err.Error(), "USERKIT_JWT_SECRET") {
		t.Fatalf("expected JWT secret error, got %v", err)
	}
}

func TestLoadFromLookupRejectsDefaultJWTSecretInProduction(t *testing.T) {
	_, err := LoadFromLookup(lookup(map[string]string{
		"APP_ENV":              "production",
		"USERKIT_ENABLED":      "true",
		"USERKIT_DATABASE_URL": "postgres://user:pass@localhost/app",
		"USERKIT_JWT_SECRET":   "change-me-to-a-long-random-secret",
	}))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "USERKIT_JWT_SECRET must be changed in production") {
		t.Fatalf("expected production JWT secret error, got %v", err)
	}
}

func TestLoadFromLookupParsesValues(t *testing.T) {
	cfg, err := LoadFromLookup(lookup(map[string]string{
		"APP_NAME":                  "billing-api",
		"HTTP_ADDR":                 ":9090",
		"HTTP_SHUTDOWN_TIMEOUT":     "3s",
		"HTTP_CORS_ALLOWED_ORIGINS": "https://app.example, https://admin.example",
		"LOG_LEVEL":                 "debug",
		"LOG_FORMAT":                "json",
		"USERKIT_ENABLED":           "true",
		"USERKIT_DATABASE_URL":      "postgres://user:pass@localhost/app",
		"USERKIT_JWT_SECRET":        "0123456789abcdef",
		"USERKIT_TOKEN_TTL":         "2h",
	}))
	if err != nil {
		t.Fatalf("LoadFromLookup() error = %v", err)
	}

	if cfg.App.Name != "billing-api" || cfg.HTTP.Addr != ":9090" {
		t.Fatalf("unexpected config: %#v", cfg)
	}
	if cfg.HTTP.ShutdownTimeout != 3*time.Second {
		t.Fatalf("unexpected shutdown timeout: %s", cfg.HTTP.ShutdownTimeout)
	}
	if len(cfg.HTTP.CORSAllowedOrigins) != 2 {
		t.Fatalf("unexpected CORS origins: %#v", cfg.HTTP.CORSAllowedOrigins)
	}
	if cfg.Userkit.TokenTTL != 2*time.Hour {
		t.Fatalf("unexpected token ttl: %s", cfg.Userkit.TokenTTL)
	}
}
