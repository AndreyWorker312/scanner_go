package service

import (
	"context"
	"fmt"
	"scanner/internal/config"
	"scanner/internal/handler"
	"scanner/internal/scanner"
	"scanner/pkg/logger"
	"scanner/pkg/queue"
)

func Run(ctx context.Context, cfg config.Config, log logger.Logger) error {
	rabbitMQ, err := queue.NewRabbitMQ(queue.RabbitMQConfig{
		URL:          cfg.RabbitMQURL,
		ScannerQueue: cfg.ScannerName,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	defer rabbitMQ.Close()

	portScanner := scanner.NewPortScanner(cfg.Timeout, cfg.MaxRetries, cfg.RetryDelay)
	msgs, err := rabbitMQ.ConsumeScanRequests(ctx)
	if err != nil {
		return fmt.Errorf("failed to consume scan requests: %w", err)
	}

	log.Infof("Scanner service started (%s), waiting for tasks...", cfg.ScannerName)
	return processMessages(ctx, msgs, portScanner, rabbitMQ, log)
}

func processMessages(ctx context.Context, msgs <-chan queue.Delivery, scanner scanner.PortScanner, rabbitMQ *queue.RabbitMQ, log logger.Logger) error {
	for {
		select {
		case msg, ok := <-msgs:
			if !ok {
				return nil
			}
			handler.HandleMessage(ctx, msg, scanner, rabbitMQ, log)

		case <-ctx.Done():
			return nil
		}
	}
}
