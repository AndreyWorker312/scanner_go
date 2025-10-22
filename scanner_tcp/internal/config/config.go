package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	RabbitMQURL    string
	ScannerName    string
	ConnTimeout    time.Duration
	BannerTimeout  time.Duration
	MaxBannerSize  int
}

func Load() *Config {
	return &Config{
		RabbitMQURL:    getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		ScannerName:    getEnv("SCANNER_NAME", "tcp_scanner"),
		ConnTimeout:    getDurationEnv("CONN_TIMEOUT", 5*time.Second),
		BannerTimeout:  getDurationEnv("BANNER_TIMEOUT", 3*time.Second),
		MaxBannerSize:  getIntEnv("MAX_BANNER_SIZE", 4096),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

