package config

import (
	"os"
	"time"
)

type Config struct {
	RabbitMQURL   string
	ScannerName   string
	MongoURI      string
	MongoDB       string
	MongoColl     string
	MinIOEndpoint string
	MinIOAccess   string
	MinIOSecret   string
	MinIOBucket   string
	ConnTimeout   time.Duration
	ReadTimeout   time.Duration
}

func Load() *Config {
	return &Config{
		RabbitMQURL:   getEnv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/"),
		ScannerName:   getEnv("SCANNER_NAME", "tcp_service"),
		MongoURI:      getEnv("MONGODB_URI", "mongodb://mongodb:27017"),
		MongoDB:       getEnv("MONGODB_DATABASE", "network_scanner"),
		MongoColl:     getEnv("TCP_MONGO_COLLECTION", "tcp_banners"),
		MinIOEndpoint: getEnv("MINIO_ENDPOINT", "minio:9000"),
		MinIOAccess:   getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinIOSecret:   getEnv("MINIO_SECRET_KEY", "minioadmin"),
		MinIOBucket:   getEnv("MINIO_BUCKET", "tcp-raw"),
		ConnTimeout:   getDuration("TCP_CONN_TIMEOUT", 5*time.Second),
		ReadTimeout:   getDuration("TCP_READ_TIMEOUT", 10*time.Second),
	}
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func getDuration(k string, def time.Duration) time.Duration {
	if v := os.Getenv(k); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
