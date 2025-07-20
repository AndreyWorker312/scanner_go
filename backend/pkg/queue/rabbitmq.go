package queue

import (
	"backend/pkg/logger"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"

	"github.com/streadway/amqp"
)

type RabbitMQConfig struct {
	URL          string
	ScannerQueue string
}

type ScanRequest struct {
	TaskID string `json:"task_id"`
	IP     string `json:"ip"`
	Ports  string `json:"ports"`
}

type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   amqp.Queue
	config  RabbitMQConfig
}

func NewRabbitMQ(config RabbitMQConfig) (*RabbitMQ, error) {
	conn, err := amqp.Dial(config.URL)
	if err != nil {
		return nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	queue, err := channel.QueueDeclare(
		config.ScannerQueue,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, err
	}

	return &RabbitMQ{
		conn:    conn,
		channel: channel,
		queue:   queue,
		config:  config,
	}, nil
}

func (r *RabbitMQ) Close() error {
	if r.channel != nil {
		if err := r.channel.Close(); err != nil {
			return err
		}
	}
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}

func (r *RabbitMQ) PublishScanRequest(ctx context.Context, req ScanRequest) error {
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	return r.channel.Publish(
		"",           // exchange
		r.queue.Name, // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		})
}

// Добавим новую структуру для ответа
type ScanResponse struct {
	TaskID    string `json:"task_id"`
	Status    string `json:"status"`
	OpenPorts []int  `json:"open_ports"`
	Error     string `json:"error,omitempty"`
}

// Добавим метод для RPC вызова с Direct Reply-To
func (r *RabbitMQ) RPCCall(ctx context.Context, req ScanRequest) (*ScanResponse, error) {
	// Сериализуем запрос
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	// Создаем канал для получения ответа
	replies, err := r.channel.Consume(
		"amq.rabbitmq.reply-to", // специальная псевдо-очередь
		"",                      // автоматически сгенерированный consumer tag
		true,                    // auto-ack
		false,                   // exclusive
		false,                   // no-local
		false,                   // no-wait
		nil,                     // args
	)
	if err != nil {
		return nil, err
	}

	// Генерируем уникальный correlation ID
	correlationID := generateCorrelationID()

	// Публикуем запрос
	err = r.channel.Publish(
		"",           // exchange
		r.queue.Name, // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: correlationID,
			ReplyTo:       "amq.rabbitmq.reply-to",
			Body:          body,
		})
	if err != nil {
		return nil, err
	}

	// Ждем ответа с таймаутом
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case msg := <-replies:
		// Проверяем correlation ID (на всякий случай)
		if msg.CorrelationId != correlationID {
			return nil, fmt.Errorf("mismatched correlation ID")
		}

		var response ScanResponse
		if err := json.Unmarshal(msg.Body, &response); err != nil {
			return nil, err
		}

		return &response, nil
	}
}

// Вспомогательная функция для генерации correlation ID
func generateCorrelationID() string {
	return uuid.New().String()
}
func (r *RabbitMQ) ConsumeScanRequests(ctx context.Context, log logger.LoggerInterface) (<-chan ScanRequest, error) {
	msgs, err := r.channel.Consume(
		r.queue.Name, // queue
		"",           // consumer tag
		false,        // auto-ack, ждём ручного подтверждения
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	if err != nil {
		log.Errorf("Failed to register consumer on queue %s: %v", r.queue.Name, err)
		return nil, err
	}

	reqChan := make(chan ScanRequest)

	go func() {
		defer close(reqChan)
		log.Infof("Started consuming scan requests from queue %s", r.queue.Name)
		for {
			select {
			case <-ctx.Done():
				log.Info("Stopping consuming scan requests (context cancelled)")
				return
			case msg, ok := <-msgs:
				if !ok {
					log.Warnf("Messages channel closed, stopping consuming")
					return
				}

				log.Infof("Received raw message: %s", string(msg.Body))

				var req ScanRequest
				if err := json.Unmarshal(msg.Body, &req); err != nil {
					log.Errorf("Failed to unmarshal scan request: %v", err)
					if err := msg.Nack(false, false); err != nil {
						log.Errorf("Failed to nack message: %v", err)
					}
					continue
				}

				log.Infof("Parsed scan request: task_id=%s, ip=%s, ports=%s", req.TaskID, req.IP, req.Ports)

				reqChan <- req

				if err := msg.Ack(false); err != nil {
					log.Errorf("Failed to ack message: %v", err)
				} else {
					log.Infof("Acknowledged message: task_id=%s", req.TaskID)
				}
			}
		}
	}()

	return reqChan, nil
}
