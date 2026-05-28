package config

import (
	"crypto/tls"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-sql-driver/mysql"
)

type Config struct {
	DatabaseDSN     string
	JWTSecret       string
	JWTTokenTTL     time.Duration
	AdminSetupToken string
	GeminiAPIKey    string
	GeminiModel     string
	UseMockCars     bool
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
	geminiAPIKey := strings.TrimSpace(getEnv("GEMINI_API_KEY", ""))
	geminiModel := strings.TrimSpace(getEnv("GEMINI_MODEL", "gemini-2.5-flash"))
	useMockCars := getEnvBool("USE_MOCK_CARS", false)

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
		DatabaseDSN:     mysqlConfig.FormatDSN(),
		JWTSecret:       jwtSecret,
		JWTTokenTTL:     jwtTTL,
		AdminSetupToken: adminSetupToken,
		GeminiAPIKey:    geminiAPIKey,
		GeminiModel:     geminiModel,
		UseMockCars:     useMockCars,
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

func getEnvBool(key string, defaultValue bool) bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if value == "" {
		return defaultValue
	}

	return value == "true" || value == "1" || value == "yes"
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
