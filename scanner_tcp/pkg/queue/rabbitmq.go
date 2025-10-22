package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   amqp.Queue
}

type Delivery = amqp.Delivery

// TCPRequest запрос на TCP сканирование
type TCPRequest struct {
	TaskID  string   `json:"task_id"`
	Host    string   `json:"host"`
	Ports   []string `json:"ports"`
	Timeout int      `json:"timeout,omitempty"`
}

// TCPResponse ответ от TCP сканера
type TCPResponse struct {
	TaskID  string           `json:"task_id"`
	Host    string           `json:"host"`
	Results []PortScanResult `json:"results"`
	Status  string           `json:"status"`
	Error   string           `json:"error,omitempty"`
}

// PortScanResult результат сканирования порта
type PortScanResult struct {
	Port         string `json:"port"`
	State        string `json:"state"`
	Service      string `json:"service"`
	Banner       string `json:"banner"`
	Version      string `json:"version"`
	Error        string `json:"error,omitempty"`
	ResponseTime int64  `json:"response_time"`
}

func NewRabbitMQ(url, queueName string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	q, err := ch.QueueDeclare(
		queueName,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	return &RabbitMQ{
		conn:    conn,
		channel: ch,
		queue:   q,
	}, nil
}

func (r *RabbitMQ) Consume(ctx context.Context) (<-chan Delivery, error) {
	msgs, err := r.channel.Consume(
		r.queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register consumer: %w", err)
	}

	return msgs, nil
}

func (r *RabbitMQ) SendResponse(replyTo, correlationID string, response interface{}) error {
	body, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = r.channel.PublishWithContext(
		ctx,
		"",
		replyTo,
		false,
		false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: correlationID,
			Body:          body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish response: %w", err)
	}

	return nil
}

func (r *RabbitMQ) Close() error {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}

