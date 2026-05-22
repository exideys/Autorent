package config

import (
	"crypto/tls"
	"fmt"
	"os"

	"github.com/go-sql-driver/mysql"
)

type Config struct {
	DatabaseName string
	DatabaseDSN  string
	ServerDSN    string
}

func Load() (*Config, error) {
	dbUser := getEnv("DB_USER", "root")
	dbPassword := getEnv("DB_PASSWORD", "")
	dbHost := getEnv("DB_HOST", "gateway01.eu-central-1.prod.aws.tidbcloud.com")
	dbPort := getEnv("DB_PORT", "4000")
	dbName := getEnv("DB_NAME", "autorent")

	if err := mysql.RegisterTLSConfig("tidb", &tls.Config{
		MinVersion: tls.VersionTLS12,
		ServerName: dbHost,
	}); err != nil {
		return nil, err
	}

	dbConfig := mysql.NewConfig()
	dbConfig.User = dbUser
	dbConfig.Passwd = dbPassword
	dbConfig.Net = "tcp"
	dbConfig.Addr = fmt.Sprintf("%s:%s", dbHost, dbPort)
	dbConfig.DBName = dbName
	dbConfig.ParseTime = true
	dbConfig.TLSConfig = "tidb"

	serverConfig := *dbConfig
	serverConfig.DBName = ""

	return &Config{
		DatabaseName: dbName,
		DatabaseDSN:  dbConfig.FormatDSN(),
		ServerDSN:    serverConfig.FormatDSN(),
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}
