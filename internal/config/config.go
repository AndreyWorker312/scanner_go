package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Scanner  ScannerConfig
	RabbitMQ RabbitMQConfig
}

type ServerConfig struct {
	Port string
}

type ScannerConfig struct {
	Timeout    time.Duration
	MaxRetries int
	RetryDelay time.Duration
}

type RabbitMQConfig struct {
	URL       string
	QueueName string
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	var cfg Config

	// Server configuration
	cfg.Server.Port = os.Getenv("SERVER_PORT")
	if cfg.Server.Port == "" {
		cfg.Server.Port = "8080"
	}

	// Scanner configuration
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

	// RabbitMQ configuration
	cfg.RabbitMQ.URL = os.Getenv("RABBITMQ_URL")
	if cfg.RabbitMQ.URL == "" {
		cfg.RabbitMQ.URL = "amqp://guest:guest@localhost:5672/"
	}

	cfg.RabbitMQ.QueueName = os.Getenv("RABBITMQ_QUEUE_NAME")
	if cfg.RabbitMQ.QueueName == "" {
		cfg.RabbitMQ.QueueName = "scan_requests"
	}

	return &cfg, nil
}
