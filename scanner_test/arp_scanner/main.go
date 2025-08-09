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
	defaultScannerQueue = "scanner2"
	scanInterval        = 3 * time.Second
)

type ARPRequest struct {
	TaskID        string `json:"task_id"`
	InterfaceName string `json:"interface_name"`
	IPRange       string `json:"ip_range"`
}

type ARPResponse struct {
	TaskID  string       `json:"task_id"`
	Status  string       `json:"status"`
	Devices []DeviceInfo `json:"devices,omitempty"`
	Error   string       `json:"error,omitempty"`
}

type DeviceInfo struct {
	IP     string `json:"ip"`
	MAC    string `json:"mac"`
	Status string `json:"status"`
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Конфигурация
	rabbitMQURL := getEnv("RABBITMQ_URL", defaultRabbitMQURL)
	scannerQueue := getEnv("SCANNER_QUEUE", defaultScannerQueue)

	// Список сканирований для выполнения
	scanRequests := []ARPRequest{
		{
			TaskID:        "arp-scan-1",
			InterfaceName: "wlo1",
			IPRange:       "192.168.1.1-192.168.1.10",
		},
		{
			TaskID:        "arp-scan-2",
			InterfaceName: "br-94b2bf6e3bd5",
			IPRange:       "172.22.0.2",
		},
	}

	// Создаем контекст для graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Запускаем бесконечный цикл сканирования
	if err := runContinuousScans(ctx, rabbitMQURL, scannerQueue, scanRequests); err != nil {
		log.Fatalf("ARP scan failed: %v", err)
	}
}

func runContinuousScans(ctx context.Context, rabbitMQURL, scannerQueue string, requests []ARPRequest) error {
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

func runSingleScanBatch(ctx context.Context, rabbitMQURL, scannerQueue string, requests []ARPRequest) error {
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
				log.Printf("ARP scan %s failed: %v", req.TaskID, err)
			}
		}
	}

	return nil
}

func sendAndReceive(ch *amqp.Channel, scannerQueue, replyTo string, msgs <-chan amqp.Delivery, req ARPRequest) error {
	correlationID := fmt.Sprintf("%d", time.Now().UnixNano())

	// Добавляем поле ReplyTo в запрос
	fullReq := struct {
		ARPRequest
		ReplyTo string `json:"reply_to"`
	}{
		ARPRequest: req,
		ReplyTo:    replyTo,
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

	log.Printf("Sent ARP scan request: %s %s on %s", req.TaskID, req.IPRange, req.InterfaceName)

	// Ожидаем ответа с таймаутом
	select {
	case d := <-msgs:
		if d.CorrelationId == correlationID {
			var response ARPResponse
			if err := json.Unmarshal(d.Body, &response); err != nil {
				return fmt.Errorf("decode response: %w", err)
			}

			if response.Error != "" {
				log.Printf("[%s] ARP scan failed: %s", response.TaskID, response.Error)
			} else {
				log.Printf("[%s] ARP scan result: found %d devices", response.TaskID, len(response.Devices))
				for _, device := range response.Devices {
					log.Printf("  - %s (%s) status: %s", device.IP, device.MAC, device.Status)
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
