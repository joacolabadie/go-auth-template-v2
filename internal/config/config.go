package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type DatabaseConfig struct {
	ConnectionString string
	MaxConns         int32
	MinConns         int32
	MaxConnLifetime  time.Duration
}

type AppConfig struct {
	Database DatabaseConfig
}

func loadEnv() error {
	err := godotenv.Load("../../.env")
	if err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}

	requiredVars := []string{"DATABASE_URL"}
	for _, v := range requiredVars {
		if os.Getenv(v) == "" {
			return fmt.Errorf("required environment variable %s is not set", v)
		}
	}

	return nil
}

func getEnvInt32(key string, fallback int32) int32 {
	valStr := os.Getenv(key)
	if valStr == "" {
		return fallback
	}

	val, err := strconv.Atoi(valStr)
	if err != nil {
		return fallback
	}

	return int32(val)
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	valStr := os.Getenv(key)
	if valStr == "" {
		return fallback
	}

	val, err := time.ParseDuration(valStr)
	if err != nil {
		return fallback
	}

	return val
}

func LoadConfig() (AppConfig, error) {
	err := loadEnv()
	if err != nil {
		return AppConfig{}, err
	}

	var cfg AppConfig

	cfg.Database = DatabaseConfig{
		ConnectionString: os.Getenv("DATABASE_URL"),
		MaxConns:         getEnvInt32("DB_MAX_CONNS", 10),
		MinConns:         getEnvInt32("DB_MIN_CONNS", 2),
		MaxConnLifetime:  getEnvDuration("DB_MAX_CONN_LIFETIME", time.Hour),
	}

	return cfg, nil
}
