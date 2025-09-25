package rabbitmq

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"backend/internal/domain/models"

	"github.com/streadway/amqp"
)

type RPCScannerPublisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	replies map[string]chan *models.Response
	mu      sync.Mutex
}

var (
	rpcPublisherInstance *RPCScannerPublisher
	rpcPublisherOnce     sync.Once
	rpcPublisherErr      error
)

func GetRPCconnection(amqpURI string) (*RPCScannerPublisher, error) {
	rpcPublisherOnce.Do(func() {
		rpcPublisherInstance, rpcPublisherErr = newRPCScannerPublisher(amqpURI)
	})
	return rpcPublisherInstance, rpcPublisherErr
}

func newRPCScannerPublisher(amqpURI string) (*RPCScannerPublisher, error) {
	conn, err := amqp.Dial(amqpURI)
	if err != nil {
		return nil, err
	}
	channel, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	publisher := &RPCScannerPublisher{
		conn:    conn,
		channel: channel,
		replies: make(map[string]chan *models.Response),
	}

	err = publisher.startReplyConsumer()
	if err != nil {
		return nil, err
	}

	return publisher, nil
}

// PublishNmap публикует задачу в очередь nmap_tasks и ожидает ответ
func (p *RPCScannerPublisher) PublishNmap(req models.NmapRequest) (*models.Response, error) {
	return p.publishRPC("nmap_service", req)
}

// PublishArp публикует задачу в arp_tasks
func (p *RPCScannerPublisher) PublishArp(req models.ARPRequest) (*models.Response, error) {
	return p.publishRPC("arp_service", req)
}

// PublishIcmp публикует задачу в icmp_tasks
func (p *RPCScannerPublisher) PublishIcmp(req models.ICMPRequest) (*models.Response, error) {
	return p.publishRPC("icmp_service", req)
}

func (p *RPCScannerPublisher) publishRPC(queueName string, task interface{}) (*models.Response, error) {
	correlationID := generateCorrelationID()
	replyChan := make(chan *models.Response, 1)

	p.mu.Lock()
	p.replies[correlationID] = replyChan
	p.mu.Unlock()

	defer func() {
		p.mu.Lock()
		delete(p.replies, correlationID)
		p.mu.Unlock()
	}()

	body, err := json.Marshal(task)
	if err != nil {
		return nil, err
	}

	err = p.channel.Publish(
		"",
		queueName,
		false,
		false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: correlationID,
			ReplyTo:       "amq.rabbitmq.reply-to",
			Body:          body,
		},
	)
	if err != nil {
		return nil, err
	}

	select {
	case response := <-replyChan:
		return response, nil
	case <-time.After(30 * time.Second):
		return nil, fmt.Errorf("RPC timeout")
	}
}

func (p *RPCScannerPublisher) startReplyConsumer() error {
	msgs, err := p.channel.Consume(
		"amq.rabbitmq.reply-to",
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	go func() {
		for msg := range msgs {
			p.mu.Lock()
			replyChan, exists := p.replies[msg.CorrelationId]
			p.mu.Unlock()

			if exists {
				var response models.Response
				err := json.Unmarshal(msg.Body, &response)
				if err != nil {
					log.Fatalf("Failed to unmarshal response: %v\n", err)
					continue
				}
				replyChan <- &response
			}
		}
	}()

	return nil
}

func generateCorrelationID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
