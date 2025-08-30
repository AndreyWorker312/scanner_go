package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Config struct {
	RabbitMQURL string
	ScannerName string
}

func Load() Config {
	err := godotenv.Load()
	if err != nil {
		log.Printf("No .env file found or error loading it: %v", err)
	}
	return Config{
		RabbitMQURL: getEnv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/"),
		ScannerName: getEnv("SCANNER_NAME", "default_scanner"),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
