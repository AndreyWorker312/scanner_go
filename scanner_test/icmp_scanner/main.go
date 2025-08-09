package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/streadway/amqp"
)

const (
	defaultRabbitMQURL  = "amqp://guest:guest@localhost:5672/"
	defaultScannerQueue = "scanner3"
	scanInterval        = 3 * time.Second
)

type PingRequest struct {
	TaskID    string   `json:"task_id"`
	Targets   []string `json:"targets"`
	PingCount int      `json:"ping_count"`
}

type PingResponse struct {
	TaskID  string       `json:"task_id"`
	Status  string       `json:"status"`
	Results []PingResult `json:"results,omitempty"`
	Error   string       `json:"error,omitempty"`
}

type PingResult struct {
	Target            string  `json:"target"`
	Address           string  `json:"address"`
	PacketsSent       int     `json:"packets_sent"`
	PacketsReceived   int     `json:"packets_received"`
	PacketLossPercent float64 `json:"packet_loss_percent"`
	Error             string  `json:"error,omitempty"`
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Конфигурация
	rabbitMQURL := getEnv("RABBITMQ_URL", defaultRabbitMQURL)
	scannerQueue := getEnv("SCANNER_QUEUE", defaultScannerQueue)

	// Список сканирований для выполнения
	scanRequests := []PingRequest{
		{
			TaskID:    "ping-scan-1",
			Targets:   []string{"8.8.8.8", "1.1.1.1", "google.com"},
			PingCount: 4,
		},
		{
			TaskID:    "ping-scan-2",
			Targets:   []string{"192.168.1.1", "192.168.1.254"},
			PingCount: 2,
		},
	}

	// Создаем контекст для graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Запускаем бесконечный цикл сканирования
	if err := runContinuousScans(ctx, rabbitMQURL, scannerQueue, scanRequests); err != nil {
		log.Fatalf("Ping scan failed: %v", err)
	}
}

func runContinuousScans(ctx context.Context, rabbitMQURL, scannerQueue string, requests []PingRequest) error {
	ticker := time.NewTicker(scanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := runSingleScanBatch(ctx, rabbitMQURL, scannerQueue, requests); err != nil {
				log.Printf("Scan batch failed: %v", err)
			}
		}
	}
}

func runSingleScanBatch(ctx context.Context, rabbitMQURL, scannerQueue string, requests []PingRequest) error {
	// Подключаемся к RabbitMQ
	conn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	defer ch.Close()

	// Создаем временную очередь для ответов
	replyQueue, err := ch.QueueDeclare(
		"",    // имя генерируется автоматически
		false, // durable
		true,  // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare reply queue: %w", err)
	}

	// Подписываемся на очередь ответов
	msgs, err := ch.Consume(
		replyQueue.Name,
		"",
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	// Выполняем все запросы сканирования
	for _, req := range requests {
		select {
		case <-ctx.Done():
			return nil
		default:
			if err := sendAndReceive(ch, scannerQueue, replyQueue.Name, msgs, req); err != nil {
				log.Printf("Ping scan %s failed: %v", req.TaskID, err)
			}
		}
	}

	return nil
}

func sendAndReceive(ch *amqp.Channel, scannerQueue, replyTo string, msgs <-chan amqp.Delivery, req PingRequest) error {
	correlationID := fmt.Sprintf("%d", time.Now().UnixNano())

	// Добавляем поле ReplyTo в запрос
	fullReq := struct {
		PingRequest
		ReplyTo string `json:"reply_to"`
	}{
		PingRequest: req,
		ReplyTo:     replyTo,
	}

	// Сериализуем запрос
	body, err := json.Marshal(fullReq)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	// Публикуем запрос
	err = ch.Publish(
		"",
		scannerQueue,
		false,
		false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: correlationID,
			ReplyTo:       replyTo,
			Body:          body,
		})
	if err != nil {
		return fmt.Errorf("publish message: %w", err)
	}

	log.Printf("Sent Ping scan request: %s for targets %v", req.TaskID, req.Targets)

	// Ожидаем ответа с таймаутом
	select {
	case d := <-msgs:
		if d.CorrelationId == correlationID {
			var response PingResponse
			if err := json.Unmarshal(d.Body, &response); err != nil {
				return fmt.Errorf("decode response: %w", err)
			}

			if response.Error != "" {
				log.Printf("[%s] Ping scan failed: %s", response.TaskID, response.Error)
			} else {
				log.Printf("[%s] Ping scan result: scanned %d targets", response.TaskID, len(response.Results))
				for _, result := range response.Results {
					if result.Error != "" {
						log.Printf("  - %s (%s) error: %s", result.Target, result.Address, result.Error)
					} else {
						log.Printf("  - %s (%s) sent: %d, received: %d, loss: %.1f%%",
							result.Target, result.Address,
							result.PacketsSent, result.PacketsReceived,
							result.PacketLossPercent)
					}
				}
			}
		}
	case <-time.After(60 * time.Second):
		return fmt.Errorf("timeout waiting for response")
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
