package queue

import (
	"context"
	"encoding/json"

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

func (r *RabbitMQ) ConsumeScanRequests(ctx context.Context) (<-chan ScanRequest, error) {
	msgs, err := r.channel.Consume(
		r.queue.Name, // queue
		"",           // consumer
		false,        // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	if err != nil {
		return nil, err
	}

	reqChan := make(chan ScanRequest)

	go func() {
		defer close(reqChan)
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgs:
				if !ok {
					return
				}

				var req ScanRequest
				if err := json.Unmarshal(msg.Body, &req); err != nil {
					msg.Nack(false, false)
					continue
				}

				reqChan <- req
				msg.Ack(false)
			}
		}
	}()

	return reqChan, nil
}
