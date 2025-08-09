package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"time"
)

type Config struct {
	RabbitMQURL string
	ScannerName string
	PingTimeout time.Duration
}

func Load() Config {
	err := godotenv.Load()
	if err != nil {
		log.Printf("No .env file found: %v", err)
	}

	return Config{
		RabbitMQURL: getEnv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/"),
		ScannerName: getEnv("SCANNER_NAME", "scanner_icmp"),
		PingTimeout: getDurationEnv("PING_TIMEOUT", 5*time.Second),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if dur, err := time.ParseDuration(value); err == nil {
			return dur
		}
	}
	return defaultValue
}
