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
	defaultScannerQueue = "scanner1"
)

type ScanRequest struct {
	TaskID string `json:"task_id"`
	IP     string `json:"ip"`
	Ports  string `json:"ports"`
}

type ScanResponse struct {
	TaskID    string `json:"task_id"`
	Status    string `json:"status"`
	OpenPorts []int  `json:"open_ports"`
	Error     string `json:"error,omitempty"`
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Конфигурация
	rabbitMQURL := getEnv("RABBITMQ_URL", defaultRabbitMQURL)
	scannerQueue := getEnv("SCANNER_QUEUE", defaultScannerQueue)

	// Список сканирований для выполнения (можете изменить под свои нужды)
	scanRequests := []ScanRequest{
		{
			TaskID: "scan-1",
			IP:     "127.0.0.1",   // Замените на нужный IP
			Ports:  "80,443,8080", // Замените на нужные порты
		},
		{
			TaskID: "scan-2",
			IP:     "chat.deepseek.com",
			Ports:  "80,443,8080,2301",
		},
	}

	// Создаем контекст для graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Выполняем сканирования
	if err := runScans(ctx, rabbitMQURL, scannerQueue, scanRequests); err != nil {
		log.Fatalf("Scan failed: %v", err)
	}
}

func runScans(ctx context.Context, rabbitMQURL, scannerQueue string, requests []ScanRequest) error {
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
				log.Printf("Scan %s failed: %v", req.TaskID, err)
			}
		}
	}

	return nil
}

func sendAndReceive(ch *amqp.Channel, scannerQueue, replyTo string, msgs <-chan amqp.Delivery, req ScanRequest) error {
	correlationID := fmt.Sprintf("%d", time.Now().UnixNano())

	// Сериализуем запрос
	body, err := json.Marshal(req)
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

	log.Printf("Sent scan request: %s %s", req.IP, req.Ports)

	// Ожидаем ответа с таймаутом
	select {
	case d := <-msgs:
		if d.CorrelationId == correlationID {
			var response ScanResponse
			if err := json.Unmarshal(d.Body, &response); err != nil {
				return fmt.Errorf("decode response: %w", err)
			}

			if response.Error != "" {
				log.Printf("[%s] Scan failed: %s", response.TaskID, response.Error)
			} else {
				log.Printf("[%s] Scan result for %s: open ports %v", response.TaskID, req.IP, response.OpenPorts)
			}
		}
	case <-time.After(30 * time.Second):
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
