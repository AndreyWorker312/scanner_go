package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
)

// Config - конфигурация приложения (экспортируемая структура)
type Config struct {
	Server  ServerConfig
	DB      DBConfig
	Scanner ScannerConfig
}

// ServerConfig - конфигурация сервера
type ServerConfig struct {
	Port string
}

// DBConfig - конфигурация базы данных
type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

// ScannerConfig - конфигурация сканера
type ScannerConfig struct {
	Timeout time.Duration
}

func Load() (*Config, error) {
	// Загрузка .env файла
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	var cfg Config

	// Настройки сервера
	cfg.Server.Port = os.Getenv("SERVER_PORT")
	if cfg.Server.Port == "" {
		cfg.Server.Port = "8006"
	}

	// Настройки базы данных
	cfg.DB.Host = os.Getenv("DB_HOST")
	cfg.DB.Port = os.Getenv("DB_PORT")
	cfg.DB.User = os.Getenv("DB_USER")
	cfg.DB.Password = os.Getenv("DB_PASSWORD")
	cfg.DB.Name = os.Getenv("DB_NAME")

	// Настройки сканера
	timeoutStr := os.Getenv("SCANNER_TIMEOUT")
	if timeoutStr == "" {
		timeoutStr = "500ms"
	}
	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		return nil, err
	}
	cfg.Scanner.Timeout = timeout

	return &cfg, nil
}
