package config

import (
	"crypto/tls"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-sql-driver/mysql"
)

type Config struct {
	DatabaseDSN                  string
	JWTSecret                    string
	JWTTokenTTL                  time.Duration
	AdminSetupToken              string
	GeminiAPIKey                 string
	GeminiModel                  string
	DeepLAPIKey                  string
	DeepLAPIURL                  string
	GoogleDriveOAuthClientID     string
	GoogleDriveOAuthClientSecret string
	GoogleDriveOAuthRefreshToken string
	GoogleDriveCarsFolderID      string
	GoogleDriveNewsFolderID      string
	GoogleDriveSupportFolderID   string
	GoogleAuthClientID           string
	ImageUploadMaxBytes          int64
	SupportAttachmentMaxBytes    int64
}

var (
	tidbTLSConfigOnce sync.Once
	tidbTLSConfigErr  error
)

func Load() (*Config, error) {
	dbUser := getEnvAny([]string{"DB_USER", "DB_USERNAME"}, "root")
	dbPassword := getEnv("DB_PASSWORD", "")
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "4000")
	dbName := getEnvAny([]string{"DB_NAME", "DB_DATABASE"}, "autorent")
	dbTLS := getEnv("DB_TLS", "tidb")
	jwtSecret := getEnv("JWT_SECRET", "change-me-in-production")
	jwtTTL := getEnvDuration("JWT_TOKEN_TTL", 24*time.Hour)
	adminSetupToken := getEnv("ADMIN_SETUP_TOKEN", "")
	geminiAPIKey := getEnv("GEMINI_API_KEY", "")
	geminiModel := getEnv("GEMINI_MODEL", "gemini-2.5-flash")
	deepLAPIKey := getEnv("DEEPL_API_KEY", "")
	deepLAPIURL := getEnv("DEEPL_API_URL", "https://api-free.deepl.com")
	googleDriveOAuthClientID := getEnv("GOOGLE_DRIVE_OAUTH_CLIENT_ID", "")
	googleDriveOAuthClientSecret := getEnv("GOOGLE_DRIVE_OAUTH_CLIENT_SECRET", "")
	googleDriveOAuthRefreshToken := getEnv("GOOGLE_DRIVE_OAUTH_REFRESH_TOKEN", "")
	googleDriveCarsFolderID := normalizeGoogleDriveFolderID(getEnvAny([]string{"GOOGLE_DRIVE_CARS_FOLDER_ID", "GOOGLE_DRIVE_CARS_FOLDER_URL"}, ""))
	googleDriveNewsFolderID := normalizeGoogleDriveFolderID(getEnvAny([]string{"GOOGLE_DRIVE_NEWS_FOLDER_ID", "GOOGLE_DRIVE_NEWS_FOLDER_URL"}, ""))
	googleDriveSupportFolderID := normalizeGoogleDriveFolderID(getEnvAny([]string{"GOOGLE_DRIVE_SUPPORT_FOLDER_ID", "GOOGLE_DRIVE_SUPPORT_FOLDER_URL"}, ""))
	googleAuthClientID := getEnv("GOOGLE_AUTH_CLIENT_ID", "")
	imageUploadMaxBytes := getEnvInt64("IMAGE_UPLOAD_MAX_BYTES", 10*1024*1024)
	supportAttachmentMaxBytes := getEnvInt64("SUPPORT_ATTACHMENT_MAX_BYTES", imageUploadMaxBytes)

	tlsConfigName, err := resolveTLSConfig(dbTLS, dbHost)
	if err != nil {
		return nil, err
	}

	mysqlConfig := mysql.Config{
		User:                 dbUser,
		Passwd:               dbPassword,
		Net:                  "tcp",
		Addr:                 net.JoinHostPort(dbHost, dbPort),
		DBName:               dbName,
		ParseTime:            true,
		Loc:                  time.Local,
		AllowNativePasswords: true,
		TLSConfig:            tlsConfigName,
		Params: map[string]string{
			"charset": "utf8mb4",
		},
	}

	return &Config{
		DatabaseDSN:                  mysqlConfig.FormatDSN(),
		JWTSecret:                    jwtSecret,
		JWTTokenTTL:                  jwtTTL,
		AdminSetupToken:              adminSetupToken,
		GeminiAPIKey:                 geminiAPIKey,
		GeminiModel:                  geminiModel,
		DeepLAPIKey:                  deepLAPIKey,
		DeepLAPIURL:                  deepLAPIURL,
		GoogleDriveOAuthClientID:     googleDriveOAuthClientID,
		GoogleDriveOAuthClientSecret: googleDriveOAuthClientSecret,
		GoogleDriveOAuthRefreshToken: googleDriveOAuthRefreshToken,
		GoogleDriveCarsFolderID:      googleDriveCarsFolderID,
		GoogleDriveNewsFolderID:      googleDriveNewsFolderID,
		GoogleDriveSupportFolderID:   googleDriveSupportFolderID,
		GoogleAuthClientID:           googleAuthClientID,
		ImageUploadMaxBytes:          imageUploadMaxBytes,
		SupportAttachmentMaxBytes:    supportAttachmentMaxBytes,
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAny(keys []string, defaultValue string) string {
	for _, key := range keys {
		if value := os.Getenv(key); value != "" {
			return value
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}

	return duration
}

func getEnvInt64(key string, defaultValue int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed <= 0 {
		return defaultValue
	}

	return parsed
}

func normalizeGoogleDriveFolderID(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	parsedURL, err := url.Parse(trimmed)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return trimmed
	}

	if folderID := folderIDFromPath(parsedURL.Path); folderID != "" {
		return folderID
	}
	if folderID := strings.TrimSpace(parsedURL.Query().Get("id")); folderID != "" {
		return folderID
	}

	return trimmed
}

func folderIDFromPath(path string) string {
	parts := strings.Split(path, "/")
	for index, part := range parts {
		if part == "folders" && index+1 < len(parts) {
			return strings.TrimSpace(parts[index+1])
		}
	}

	return ""
}

func resolveTLSConfig(value string, dbHost string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "0", "false", "disable", "disabled", "off", "none":
		return "", nil
	case "tidb":
		if err := registerTiDBTLSConfig(dbHost); err != nil {
			return "", err
		}
		return "tidb", nil
	default:
		return value, nil
	}
}

func registerTiDBTLSConfig(dbHost string) error {
	tidbTLSConfigOnce.Do(func() {
		tidbTLSConfigErr = mysql.RegisterTLSConfig("tidb", &tls.Config{
			MinVersion: tls.VersionTLS12,
			ServerName: dbHost,
		})
	})

	return tidbTLSConfigErr
}
