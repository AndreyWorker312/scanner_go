package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	RabbitMQ RabbitMQConfig
}

type ServerConfig struct {
	Port string
}

type RabbitMQConfig struct {
	URL          string
	ScannerQueue string
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	cfg := &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
		RabbitMQ: RabbitMQConfig{
			URL:          getEnv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/"),
			ScannerQueue: getEnv("SCANNER_NAME", "scan_requests"),
		},
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
