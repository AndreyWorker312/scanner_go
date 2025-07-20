package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"scanner/internal/scanner"
	"scanner/pkg/logger"
	"scanner/pkg/queue"
)

func main() {
	log := logger.New()

	// Load configuration from environment
	rabbitMQURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/")
	scannerName := getEnv("SCANNER_NAME", "default_scanner")
	timeout := getDurationEnv("SCANNER_TIMEOUT", 500*time.Millisecond)
	maxRetries := getIntEnv("SCANNER_MAX_RETRIES", 3)
	retryDelay := getDurationEnv("SCANNER_RETRY_DELAY", 1*time.Second)

	// Initialize RabbitMQ
	rabbitMQ, err := queue.NewRabbitMQ(queue.RabbitMQConfig{
		URL:          rabbitMQURL,
		ScannerQueue: scannerName,
	})
	if err != nil {
		log.Errorf("Failed to connect to RabbitMQ: %v", err)
		os.Exit(1)
	}
	defer rabbitMQ.Close()

	// Initialize port scanner
	portScanner := scanner.NewPortScanner(timeout, maxRetries, retryDelay)

	// Start consuming scan requests
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	msgs, err := rabbitMQ.ConsumeScanRequests(ctx)
	if err != nil {
		log.Errorf("Failed to consume scan requests: %v", err)
		os.Exit(1)
	}

	log.Infof("Scanner service started (%s), waiting for tasks...", scannerName)

	// Process scan requests
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case delivery, ok := <-msgs:
			if !ok {
				log.Info("Channel closed, shutting down...")
				return
			}

			// Unmarshal the message body into ScanRequest
			var req queue.ScanRequest
			if err := json.Unmarshal(delivery.Body, &req); err != nil {
				log.Errorf("Failed to unmarshal scan request: %v", err)
				continue
			}

			log.Infof("Received scan request: %s %s", req.IP, req.Ports)

			// Process the scan
			openPorts, err := portScanner.ScanPorts(ctx, req.IP, req.Ports)

			// If this is an RPC request, send response
			if delivery.ReplyTo != "" {
				response := queue.ScanResponse{
					TaskID:    req.TaskID,
					Status:    "completed",
					OpenPorts: openPorts,
				}
				if err != nil {
					response.Error = err.Error()
					response.Status = "failed"
				}

				err = rabbitMQ.SendResponse(ctx, delivery.ReplyTo, delivery.CorrelationId, response)
				if err != nil {
					log.Errorf("Failed to send RPC response: %v", err)
				}
			}

			if err != nil {
				log.Errorf("Scan failed: %v", err)
				continue
			}

			log.Infof("Scan completed for %s, open ports: %v", req.IP, openPorts)

		case <-signalChan:
			log.Info("Shutting down scanner service...")
			return
		}
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getIntEnv(key string, defaultValue int) int {
	value := getEnv(key, "")
	if value == "" {
		return defaultValue
	}
	result, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return result
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	value := getEnv(key, "")
	if value == "" {
		return defaultValue
	}
	result, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}
	return result
}
