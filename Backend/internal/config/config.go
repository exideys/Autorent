package config

import (
	"crypto/tls"
	"fmt"
	"os"
	"strings"

	"github.com/go-sql-driver/mysql"
)

type Config struct {
	DatabaseDSN  string
	GeminiAPIKey string
	GeminiModel  string
	UseMockCars  bool
}

func Load() (*Config, error) {
	dbUser := getEnvAny([]string{"DB_USER", "DB_USERNAME"}, "root")
	dbPassword := getEnv("DB_PASSWORD", "")
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "4000")
	dbName := getEnvAny([]string{"DB_NAME", "DB_DATABASE"}, "autorent")
	geminiAPIKey := strings.TrimSpace(getEnv("GEMINI_API_KEY", ""))
	geminiModel := strings.TrimSpace(getEnv("GEMINI_MODEL", "gemini-2.5-flash"))
	useMockCars := getEnvBool("USE_MOCK_CARS", false)

	if err := mysql.RegisterTLSConfig("tidb", &tls.Config{
		MinVersion: tls.VersionTLS12,
		ServerName: dbHost,
	}); err != nil {
		return nil, err
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&tls=tidb", dbUser, dbPassword, dbHost, dbPort, dbName)
	/*dsnConfig := mysql.Config{
		User:      dbUser,
		Passwd:    dbPassword,
		Net:       "tcp",
		Addr:      fmt.Sprintf("%s:%s", dbHost, dbPort),
		DBName:    dbName,
		ParseTime: true,
		TLSConfig: "tidb",
	}

	dsn := dsnConfig.FormatDSN()*/

	return &Config{
		DatabaseDSN:  dsn,
		GeminiAPIKey: geminiAPIKey,
		GeminiModel:  geminiModel,
		UseMockCars:  useMockCars,
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
