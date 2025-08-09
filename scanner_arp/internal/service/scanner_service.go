package service

import (
	"arp_scanner/internal/config"
	"arp_scanner/internal/handler"
	"arp_scanner/pkg/logger"
	"arp_scanner/pkg/queue"
	"context"
	"fmt"
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

	log.Infof("ARP Scanner service started (%s), waiting for tasks...", cfg.ScannerName)
	return processMessages(ctx, msgs, rabbitMQ, log)
}

func processMessages(ctx context.Context, msgs <-chan queue.Delivery, rabbitMQ *queue.RabbitMQ, log logger.Logger) error {
	for {
		select {
		case msg, ok := <-msgs:
			if !ok {
				return nil
			}
			handler.HandleMessage(ctx, msg, rabbitMQ, log)

		case <-ctx.Done():
			return nil
		}
	}
}
