package config

import "time"

type Config struct {
	RabbitMQURL string
	ScannerName string
	Timeout     time.Duration
	MaxRetries  int
	RetryDelay  time.Duration
}

func Load() Config {
	return Config{
		RabbitMQURL: getEnv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/"),
		ScannerName: getEnv("SCANNER_NAME", "default_scanner"),
		Timeout:     getDurationEnv("SCANNER_TIMEOUT", 500*time.Millisecond),
		MaxRetries:  getIntEnv("SCANNER_MAX_RETRIES", 3),
		RetryDelay:  getDurationEnv("SCANNER_RETRY_DELAY", 1*time.Second),
	}
}
