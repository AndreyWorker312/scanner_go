package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server  ServerConfig
	DB      DBConfig
	Scanner ScannerConfig
}

type ServerConfig struct {
	Port string
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type ScannerConfig struct {
	Timeout    time.Duration
	MaxRetries int
	RetryDelay time.Duration
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	var cfg Config

	cfg.Server.Port = os.Getenv("SERVER_PORT")
	if cfg.Server.Port == "" {
		cfg.Server.Port = "8080" // Default port
	}

	cfg.DB.Host = os.Getenv("DB_HOST")
	cfg.DB.Port = os.Getenv("DB_PORT")
	cfg.DB.User = os.Getenv("DB_USER")
	cfg.DB.Password = os.Getenv("DB_PASSWORD")
	cfg.DB.Name = os.Getenv("DB_NAME")

	timeoutStr := os.Getenv("SCANNER_TIMEOUT")
	if timeoutStr == "" {
		timeoutStr = "500ms"
	}
	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		return nil, err
	}
	cfg.Scanner.Timeout = timeout

	maxRetriesStr := os.Getenv("SCANNER_MAX_RETRIES")
	if maxRetriesStr == "" {
		maxRetriesStr = "3"
	}
	maxRetries, err := strconv.Atoi(maxRetriesStr)
	if err != nil {
		return nil, err
	}
	cfg.Scanner.MaxRetries = maxRetries

	retryDelayStr := os.Getenv("SCANNER_RETRY_DELAY")
	if retryDelayStr == "" {
		retryDelayStr = "1s"
	}
	retryDelay, err := time.ParseDuration(retryDelayStr)
	if err != nil {
		return nil, err
	}
	cfg.Scanner.RetryDelay = retryDelay

	return &cfg, nil
}
