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

type ARPRequest struct {
	TaskID        string `json:"task_id"`
	InterfaceName string `json:"interface_name"` // Теперь это обязательное поле
	IPRange       string `json:"ip_range"`
}

type ARPResponse struct {
	TaskID         string      `json:"task_id"`
	Status         string      `json:"status"`
	Devices        []ARPDevice `json:"devices"`
	OnlineDevices  []ARPDevice `json:"online_devices"`
	OfflineDevices []ARPDevice `json:"offline_devices"`
	TotalCount     int         `json:"total_count"`
	OnlineCount    int         `json:"online_count"`
	OfflineCount   int         `json:"offline_count"`
	Error          string      `json:"error,omitempty"`
}

type ARPDevice struct {
	IP     string `json:"ip"`
	MAC    string `json:"mac"`
	Vendor string `json:"vendor,omitempty"`
	Status string `json:"status"`
}

type Delivery struct {
	amqp.Delivery
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

func (r *RabbitMQ) SendResponse(replyTo string, correlationID string, response interface{}) error {
	body, err := json.Marshal(response)
	if err != nil {
		return err
	}

	return r.channel.Publish(
		"",
		replyTo,
		false,
		false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: correlationID,
			Body:          body,
		})
}

func (r *RabbitMQ) ConsumeScanRequests(ctx context.Context) (<-chan Delivery, error) {
	msgs, err := r.channel.Consume(
		r.queue.Name,
		"",
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	if err != nil {
		return nil, err
	}

	deliveries := make(chan Delivery)

	go func() {
		defer close(deliveries)
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgs:
				if !ok {
					return
				}

				var req ARPRequest
				if err := json.Unmarshal(msg.Body, &req); err != nil {
					_ = msg.Nack(false, false)
					continue
				}

				deliveries <- Delivery{msg}
				_ = msg.Ack(false)
			}
		}
	}()

	return deliveries, nil
}
