package service

import (
	"context"
	"fmt"
	"scanner_icmp/internal/config"
	"scanner_icmp/internal/handler"
	"scanner_icmp/pkg/logger"
	"scanner_icmp/pkg/queue"
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

	msgs, err := rabbitMQ.ConsumeScanRequests(ctx)
	if err != nil {
		return fmt.Errorf("failed to consume scan requests: %w", err)
	}

	log.Infof("Ping Scanner service started (%s), waiting for tasks...", cfg.ScannerName)
	return processMessages(ctx, msgs, rabbitMQ, log, &cfg)
}

func processMessages(ctx context.Context, msgs <-chan queue.Delivery, rabbitMQ *queue.RabbitMQ, log logger.Logger, cfg *config.Config) error {
	for {
		select {
		case msg, ok := <-msgs:
			if !ok {
				return nil
			}
			handler.HandleMessage(ctx, msg, rabbitMQ, log, cfg)

		case <-ctx.Done():
			return nil
		}
	}
}
