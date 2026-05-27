package config

import (
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
)

func TestLoadBuildsDatabaseAndAuthConfigFromEnv(t *testing.T) {
	t.Setenv("DB_USER", "app_user")
	t.Setenv("DB_PASSWORD", "app_password")
	t.Setenv("DB_HOST", "db.example.com")
	t.Setenv("DB_PORT", "3307")
	t.Setenv("DB_NAME", "autorent_test")
	t.Setenv("DB_TLS", "false")
	t.Setenv("JWT_SECRET", "jwt-secret")
	t.Setenv("JWT_TOKEN_TTL", "2h")
	t.Setenv("ADMIN_SETUP_TOKEN", "setup-token")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	dsn, err := mysql.ParseDSN(cfg.DatabaseDSN)
	if err != nil {
		t.Fatalf("failed to parse dsn: %v", err)
	}

	if dsn.User != "app_user" {
		t.Fatalf("expected db user app_user, got %q", dsn.User)
	}
	if dsn.Passwd != "app_password" {
		t.Fatalf("expected configured db password")
	}
	if dsn.Addr != "db.example.com:3307" {
		t.Fatalf("expected db.example.com:3307, got %q", dsn.Addr)
	}
	if dsn.DBName != "autorent_test" {
		t.Fatalf("expected autorent_test database, got %q", dsn.DBName)
	}
	if dsn.TLSConfig != "" {
		t.Fatalf("expected disabled TLS config, got %q", dsn.TLSConfig)
	}
	if !dsn.ParseTime {
		t.Fatal("expected parseTime to be enabled")
	}
	if cfg.JWTSecret != "jwt-secret" {
		t.Fatalf("expected configured jwt secret, got %q", cfg.JWTSecret)
	}
	if cfg.JWTTokenTTL != 2*time.Hour {
		t.Fatalf("expected 2h token ttl, got %s", cfg.JWTTokenTTL)
	}
	if cfg.AdminSetupToken != "setup-token" {
		t.Fatalf("expected configured admin setup token")
	}
}

func TestLoadUsesLegacyDatabaseEnvNames(t *testing.T) {
	t.Setenv("DB_USER", "")
	t.Setenv("DB_NAME", "")
	t.Setenv("DB_USERNAME", "legacy_user")
	t.Setenv("DB_DATABASE", "legacy_db")
	t.Setenv("DB_TLS", "false")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	dsn, err := mysql.ParseDSN(cfg.DatabaseDSN)
	if err != nil {
		t.Fatalf("failed to parse dsn: %v", err)
	}

	if dsn.User != "legacy_user" {
		t.Fatalf("expected legacy_user, got %q", dsn.User)
	}
	if dsn.DBName != "legacy_db" {
		t.Fatalf("expected legacy_db, got %q", dsn.DBName)
	}
}

func TestLoadDefaultsToTiDBTLS(t *testing.T) {
	t.Setenv("DB_HOST", "tidb.example.com")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	dsn, err := mysql.ParseDSN(cfg.DatabaseDSN)
	if err != nil {
		t.Fatalf("failed to parse dsn: %v", err)
	}

	if dsn.TLSConfig != "tidb" {
		t.Fatalf("expected tidb TLS config, got %q", dsn.TLSConfig)
	}
}

func TestGetEnvDurationFallsBackOnInvalidValue(t *testing.T) {
	t.Setenv("JWT_TOKEN_TTL", "not-a-duration")

	duration := getEnvDuration("JWT_TOKEN_TTL", 30*time.Minute)
	if duration != 30*time.Minute {
		t.Fatalf("expected fallback duration, got %s", duration)
	}
}
