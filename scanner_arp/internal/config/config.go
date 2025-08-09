package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	RabbitMQURL string
	ScannerName string
	Timeout     time.Duration
	MaxRetries  int
	RetryDelay  time.Duration
}

func Load() Config {
	err := godotenv.Load()
	if err != nil {
		log.Printf("No .env file found or error loading it: %v", err)
	}
	return Config{
		RabbitMQURL: getEnv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/"),
		ScannerName: getEnv("SCANNER_NAME", "arp_scanner"),
		Timeout:     getDurationEnv("SCANNER_TIMEOUT", 2*time.Second),
		MaxRetries:  getIntEnv("SCANNER_MAX_RETRIES", 2),
		RetryDelay:  getDurationEnv("SCANNER_RETRY_DELAY", 500*time.Millisecond),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getIntEnv(key string, defaultValue int) int {
	value := getEnv(key, "")
	if value == "" {
		return defaultValue
	}
	result, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return result
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	value := getEnv(key, "")
	if value == "" {
		return defaultValue
	}
	result, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}
	return result
}
