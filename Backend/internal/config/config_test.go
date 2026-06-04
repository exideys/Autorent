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
	t.Setenv("GEMINI_API_KEY", "gemini-key")
	t.Setenv("GEMINI_MODEL", "gemini-test-model")
	t.Setenv("DEEPL_API_KEY", "deepl-key")
	t.Setenv("DEEPL_API_URL", "https://example.deepl.test")
	t.Setenv("GOOGLE_DRIVE_OAUTH_CLIENT_ID", "oauth-client-id")
	t.Setenv("GOOGLE_DRIVE_OAUTH_CLIENT_SECRET", "oauth-client-secret")
	t.Setenv("GOOGLE_DRIVE_OAUTH_REFRESH_TOKEN", "oauth-refresh-token")
	t.Setenv("GOOGLE_DRIVE_CARS_FOLDER_ID", "cars-folder")
	t.Setenv("GOOGLE_DRIVE_NEWS_FOLDER_ID", "https://drive.google.com/drive/folders/news-folder?usp=sharing")
	t.Setenv("GOOGLE_AUTH_CLIENT_ID", "google-auth-client-id")
	t.Setenv("IMAGE_UPLOAD_MAX_BYTES", "123456")

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
	if cfg.GeminiAPIKey != "gemini-key" {
		t.Fatalf("expected configured gemini api key")
	}
	if cfg.GeminiModel != "gemini-test-model" {
		t.Fatalf("expected configured gemini model")
	}
	if cfg.DeepLAPIKey != "deepl-key" {
		t.Fatalf("expected configured deepl api key")
	}
	if cfg.DeepLAPIURL != "https://example.deepl.test" {
		t.Fatalf("expected configured deepl api url, got %q", cfg.DeepLAPIURL)
	}
	if cfg.GoogleDriveOAuthClientID != "oauth-client-id" {
		t.Fatalf("expected configured google drive oauth client id")
	}
	if cfg.GoogleDriveOAuthClientSecret != "oauth-client-secret" {
		t.Fatalf("expected configured google drive oauth client secret")
	}
	if cfg.GoogleDriveOAuthRefreshToken != "oauth-refresh-token" {
		t.Fatalf("expected configured google drive oauth refresh token")
	}
	if cfg.GoogleDriveCarsFolderID != "cars-folder" {
		t.Fatalf("expected configured cars folder id, got %q", cfg.GoogleDriveCarsFolderID)
	}
	if cfg.GoogleDriveNewsFolderID != "news-folder" {
		t.Fatalf("expected parsed news folder id, got %q", cfg.GoogleDriveNewsFolderID)
	}
	if cfg.GoogleAuthClientID != "google-auth-client-id" {
		t.Fatalf("expected configured Google auth client id")
	}
	if cfg.ImageUploadMaxBytes != 123456 {
		t.Fatalf("expected configured upload limit, got %d", cfg.ImageUploadMaxBytes)
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

func TestGetEnvInt64FallsBackOnInvalidValue(t *testing.T) {
	t.Setenv("IMAGE_UPLOAD_MAX_BYTES", "not-a-number")

	value := getEnvInt64("IMAGE_UPLOAD_MAX_BYTES", 99)
	if value != 99 {
		t.Fatalf("expected fallback integer value, got %d", value)
	}
}

func TestNormalizeGoogleDriveFolderIDSupportsAlternateURL(t *testing.T) {
	value := normalizeGoogleDriveFolderID("https://drive.google.com/open?id=alternate-folder")

	if value != "alternate-folder" {
		t.Fatalf("expected alternate-folder, got %q", value)
	}
}

func TestLoadDefaultsGeminiModel(t *testing.T) {
	t.Setenv("DB_TLS", "false")
	t.Setenv("GEMINI_MODEL", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.GeminiModel != "gemini-2.5-flash" {
		t.Fatalf("expected default gemini model, got %q", cfg.GeminiModel)
	}
}

func TestLoadDefaultsDeepLAPIURL(t *testing.T) {
	t.Setenv("DB_TLS", "false")
	t.Setenv("DEEPL_API_URL", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.DeepLAPIURL != "https://api-free.deepl.com" {
		t.Fatalf("expected default deepl api url, got %q", cfg.DeepLAPIURL)
	}
}
