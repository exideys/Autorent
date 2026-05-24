package config

import (
	"crypto/tls"
	"fmt"
	"os"

	"github.com/go-sql-driver/mysql"
)

type Config struct {
	DatabaseDSN string
}

func Load() (*Config, error) {
	dbUser := getEnv("DB_USER", "root")
	dbPassword := getEnv("DB_PASSWORD", "")
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "4000")
	dbName := getEnv("DB_NAME", "autorent")

	if err := mysql.RegisterTLSConfig("tidb", &tls.Config{
		MinVersion: tls.VersionTLS12,
		ServerName: dbHost,
	}); err != nil {
		return nil, err
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&tls=tidb", dbUser, dbPassword, dbHost, dbPort, dbName)
	return &Config{DatabaseDSN: dsn}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
