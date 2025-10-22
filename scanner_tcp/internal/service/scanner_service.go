package service

import (
	"context"
	"fmt"
	"scanner_tcp/internal/config"
	"scanner_tcp/internal/handler"
	"scanner_tcp/internal/scanner"
	"scanner_tcp/pkg/logger"
	"scanner_tcp/pkg/queue"
	"time"
)

func Run(ctx context.Context, cfg *config.Config, log logger.Logger) error {
	log.Infof("Starting TCP Scanner Service")
	log.Infof("Connecting to RabbitMQ at: %s", cfg.RabbitMQURL)

	// Подключаемся к RabbitMQ с повторными попытками
	var rabbitMQ *queue.RabbitMQ
	var err error
	maxRetries := 10
	
	for i := 0; i < maxRetries; i++ {
		rabbitMQ, err = queue.NewRabbitMQ(cfg.RabbitMQURL, cfg.ScannerName)
		if err == nil {
			break
		}
		log.Warnf("Failed to connect to RabbitMQ (attempt %d/%d): %v", i+1, maxRetries, err)
		time.Sleep(5 * time.Second)
	}
	
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ after %d attempts: %w", maxRetries, err)
	}
	defer rabbitMQ.Close()

	log.Infof("Successfully connected to RabbitMQ")

	// Создаем TCP сканер
	tcpScanner := scanner.NewTCPScanner(cfg.ConnTimeout, cfg.BannerTimeout, cfg.MaxBannerSize)
	log.Infof("TCP Scanner initialized with timeout: %v, banner timeout: %v", cfg.ConnTimeout, cfg.BannerTimeout)

	// Начинаем слушать очередь
	msgs, err := rabbitMQ.Consume(ctx)
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	log.Infof("Waiting for TCP scan requests on queue: %s", cfg.ScannerName)

	for {
		select {
		case <-ctx.Done():
			log.Info("Shutting down TCP scanner service...")
			return nil
		case msg, ok := <-msgs:
			if !ok {
				return fmt.Errorf("message channel closed")
			}
			go handler.HandleMessage(ctx, msg, rabbitMQ, tcpScanner, log)
		}
	}
}

