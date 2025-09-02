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
	defaultScannerQueue = "scanner4"
	scanInterval        = 30 * time.Second
	responseTimeout     = 120 * time.Second
)

type RawRequest struct {
	ScanMethod string `json:"scan_method"`
}

type ScanTcpUdpRequest struct {
	TaskID      string `json:"task_id"`
	IP          string `json:"ip"`
	ScannerType string `json:"scanner_type"`
	Ports       string `json:"ports"`
}

type OsDetectionRequest struct {
	TaskID string `json:"task_id"`
	IP     string `json:"ip"`
}

type HostDiscoveryRequest struct {
	TaskID string `json:"task_id"`
	IP     string `json:"ip"`
}

type ScanTcpUdpResponse struct {
	TaskID   string           `json:"task_id"`
	Host     string           `json:"host"`
	PortInfo []PortTcpUdpInfo `json:"port_info"`
	Status   string           `json:"status"`
	Error    string           `json:"error,omitempty"`
}

type PortTcpUdpInfo struct {
	Status      string   `json:"status"`
	AllPorts    []uint16 `json:"close_ports"`
	Protocols   []string `json:"protocols"`
	State       []string `json:"state"`
	ServiceName []string `json:"service_name"`
}

type OsDetectionResponse struct {
	TaskID   string `json:"task_id"`
	Host     string `json:"host"`
	Name     string `json:"name"`
	Accuracy int    `json:"accuracy"`
	Vendor   string `json:"vendor"`
	Family   string `json:"family"`
	Type     string `json:"type"`
	Status   string `json:"status"`
	Error    string `json:"error,omitempty"`
}

type HostDiscoveryResponse struct {
	TaskID    string `json:"task_id"`
	Host      string `json:"host"`
	HostUP    int    `json:"host_up"`
	HostTotal int    `json:"host_total"`
	Status    string `json:"status"`
	DNS       string `json:"dns"`
	Reason    string `json:"reason"`
	Error     string `json:"error,omitempty"`
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Конфигурация
	rabbitMQURL := getEnv("RABBITMQ_URL", defaultRabbitMQURL)
	scannerQueue := getEnv("SCANNER_QUEUE", defaultScannerQueue)

	// Список сканирований для выполнения
	scanRequests := []interface{}{
		struct {
			RawRequest
			ScanTcpUdpRequest
		}{
			RawRequest: RawRequest{ScanMethod: "TCP/UDP"},
			ScanTcpUdpRequest: ScanTcpUdpRequest{
				TaskID:      "tcp-scan-1",
				IP:          "google.com",
				ScannerType: "TCP",
				Ports:       "22,80,443,8080",
			},
		},
		struct {
			RawRequest
			ScanTcpUdpRequest
		}{
			RawRequest: RawRequest{ScanMethod: "TCP/UDP"},
			ScanTcpUdpRequest: ScanTcpUdpRequest{
				TaskID:      "udp-scan-1",
				IP:          "192.168.1.1",
				ScannerType: "UDP",
				Ports:       "53,67,68,123",
			},
		},
		struct {
			RawRequest
			OsDetectionRequest
		}{
			RawRequest: RawRequest{ScanMethod: "OC"},
			OsDetectionRequest: OsDetectionRequest{
				TaskID: "os-scan-1",
				IP:     "chat.deepseek.com",
			},
		},
		struct {
			RawRequest
			HostDiscoveryRequest
		}{
			RawRequest: RawRequest{ScanMethod: "HOST"},
			HostDiscoveryRequest: HostDiscoveryRequest{
				TaskID: "host-discovery-1",
				IP:     "chat.deepseek.com",
			},
		},
	}

	// Создаем контекст для graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Запускаем бесконечный цикл сканирования
	if err := runContinuousScans(ctx, rabbitMQURL, scannerQueue, scanRequests); err != nil {
		log.Fatalf("NMAP scan failed: %v", err)
	}
}

func runContinuousScans(ctx context.Context, rabbitMQURL, scannerQueue string, requests []interface{}) error {
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

func runSingleScanBatch(ctx context.Context, rabbitMQURL, scannerQueue string, requests []interface{}) error {
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
				log.Printf("Scan failed: %v", err)
			}
		}
	}

	return nil
}

func sendAndReceive(ch *amqp.Channel, scannerQueue, replyTo string, msgs <-chan amqp.Delivery, req interface{}) error {
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

	log.Printf("Sent scan request: %s", string(body))

	// Ожидаем ответа с таймаутом
	select {
	case d := <-msgs:
		if d.CorrelationId == correlationID {
			// Определяем тип ответа по структуре
			var raw map[string]interface{}
			if err := json.Unmarshal(d.Body, &raw); err != nil {
				return fmt.Errorf("decode response: %w", err)
			}

			// Проверяем наличие полей для определения типа ответа
			if _, hasPortInfo := raw["port_info"]; hasPortInfo {
				var response ScanTcpUdpResponse
				if err := json.Unmarshal(d.Body, &response); err != nil {
					return fmt.Errorf("decode TCP/UDP response: %w", err)
				}
				if response.Error != "" {
					log.Printf("[%s] TCP/UDP scan failed: %s", response.TaskID, response.Error)
				} else {
					log.Printf("[%s] TCP/UDP scan result: host %s, status: %s",
						response.TaskID, response.Host, response.Status)
					for i, portInfo := range response.PortInfo {
						log.Printf("  PortInfo[%d]: Status=%s, Ports=%v, Protocols=%v, State=%v, Services=%v",
							i, portInfo.Status, portInfo.AllPorts, portInfo.Protocols,
							portInfo.State, portInfo.ServiceName)
					}
				}
			} else if _, hasAccuracy := raw["accuracy"]; hasAccuracy {
				var response OsDetectionResponse
				if err := json.Unmarshal(d.Body, &response); err != nil {
					return fmt.Errorf("decode OS detection response: %w", err)
				}
				if response.Error != "" {
					log.Printf("[%s] OS detection failed: %s", response.TaskID, response.Error)
				} else {
					log.Printf("[%s] OS detection result: %s (accuracy: %d%%), Vendor: %s, Family: %s, Type: %s",
						response.TaskID, response.Name, response.Accuracy, response.Vendor,
						response.Family, response.Type)
				}
			} else if _, hasHostUP := raw["host_up"]; hasHostUP {
				var response HostDiscoveryResponse
				if err := json.Unmarshal(d.Body, &response); err != nil {
					return fmt.Errorf("decode host discovery response: %w", err)
				}
				if response.Error != "" {
					log.Printf("[%s] Host discovery failed: %s", response.TaskID, response.Error)
				} else {
					log.Printf("[%s] Host discovery result: %d/%d hosts up, Status: %s, DNS: %s, Reason: %s",
						response.TaskID, response.HostUP, response.HostTotal, response.Status,
						response.DNS, response.Reason)
				}
			} else {
				log.Printf("Unknown response type: %s", string(d.Body))
			}
		}
	case <-time.After(responseTimeout):
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
