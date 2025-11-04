package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	conn  *amqp.Connection
	ch    *amqp.Channel
	queue amqp.Queue
}

type Delivery = amqp.Delivery

type TCPRequest struct {
	TaskID string `json:"task_id"`
	Host   string `json:"host"`
	Port   string `json:"port"`
}

type TCPResponse struct {
	TaskID       string `json:"task_id"`
	Host         string `json:"host"`
	Port         string `json:"port"`
	HexObjectKey string `json:"hex_object_key"`
	DecodedText  string `json:"decoded_text"`
	Status       string `json:"status"`
	Error        string `json:"error,omitempty"`
}

func New(url, queueName string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}
	if err := ch.Qos(1, 0, false); err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}
	q, err := ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}
	return &RabbitMQ{conn: conn, ch: ch, queue: q}, nil
}

func (r *RabbitMQ) Consume(ctx context.Context) (<-chan Delivery, error) {
	msgs, err := r.ch.Consume(r.queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return nil, err
	}
	return msgs, nil
}

func (r *RabbitMQ) Ack(d Delivery)  { _ = d.Ack(false) }
func (r *RabbitMQ) Nack(d Delivery) { _ = d.Nack(false, false) }

func (r *RabbitMQ) Reply(replyTo, corrID string, v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return r.ch.PublishWithContext(ctx, "", replyTo, false, false, amqp.Publishing{
		ContentType:   "application/json",
		CorrelationId: corrID,
		Body:          b,
	})
}

func (r *RabbitMQ) Close() {
	if r.ch != nil {
		_ = r.ch.Close()
	}
	if r.conn != nil {
		_ = r.conn.Close()
	}
}

func ParseTCPRequest(body []byte) (TCPRequest, error) {
	var req TCPRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return req, fmt.Errorf("bad request: %w", err)
	}
	return req, nil
}
