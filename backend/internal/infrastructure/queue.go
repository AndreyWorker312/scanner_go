package rabbitmq

import (
	"backend/domain/models"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

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

func (p *RPCScannerPublisher) PublishNmap(req interface{}) (*models.Response, error) {
	return p.publishRPC("nmap_service", req)
}

func (p *RPCScannerPublisher) PublishArp(req models.ARPRequest) (*models.Response, error) {
	return p.publishRPC("arp_service", req)
}

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

	log.Printf("Publishing to %s: %s", queueName, string(body))

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
		log.Printf("Received response for %s: %+v", correlationID, response)
		return response, nil
	case <-time.After(30 * time.Second):
		return nil, fmt.Errorf("RPC timeout for queue %s", queueName)
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
				// Пробуем определить тип ответа и преобразовать в универсальный Response
				response, err := p.parseResponse(msg.Body)
				if err != nil {
					log.Printf("Failed to parse response: %v", err)
					log.Printf("Response body: %s", string(msg.Body))
					continue
				}
				replyChan <- response
			}
		}
	}()

	return nil
}

// parseResponse пытается определить тип ответа и преобразовать его в универсальный Response
func (p *RPCScannerPublisher) parseResponse(body []byte) (*models.Response, error) {
	log.Printf("Raw response body: %s", string(body))

	// Сначала пробуем как ARPResponse (более специфичный)
	var arpResp models.ARPResponse
	if err := json.Unmarshal(body, &arpResp); err == nil && arpResp.TaskID != "" {
		log.Printf("Received ARP response for task %s: Total=%d, Online=%d, Offline=%d",
			arpResp.TaskID, arpResp.TotalCount, arpResp.OnlineCount, arpResp.OfflineCount)
		log.Printf("ARP response details: Status=%s, Error=%s", arpResp.Status, arpResp.Error)
		return &models.Response{
			TaskID: arpResp.TaskID,
			Result: arpResp,
		}, nil
	}

	// Потом пробуем как обычный Response
	var response models.Response
	if err := json.Unmarshal(body, &response); err == nil && response.TaskID != "" {
		log.Printf("Received generic response for task %s", response.TaskID)
		return &response, nil
	}

	// Пробуем как ICMPResponse
	var icmpResp models.ICMPResponse
	if err := json.Unmarshal(body, &icmpResp); err == nil && icmpResp.TaskID != "" {
		log.Printf("Received ICMP response for task %s with %d results", icmpResp.TaskID, len(icmpResp.Results))
		return &models.Response{
			TaskID: icmpResp.TaskID,
			Result: icmpResp,
		}, nil
	}

	// Пробуем как NmapTcpUdpResponse
	var nmapTcpUdpResp models.NmapTcpUdpResponse
	if err := json.Unmarshal(body, &nmapTcpUdpResp); err == nil && nmapTcpUdpResp.TaskID != "" {
		log.Printf("Received Nmap TCP/UDP response for task %s", nmapTcpUdpResp.TaskID)
		return &models.Response{
			TaskID: nmapTcpUdpResp.TaskID,
			Result: nmapTcpUdpResp,
		}, nil
	}

	// Пробуем как NmapOsDetectionResponse
	var nmapOsResp models.NmapOsDetectionResponse
	if err := json.Unmarshal(body, &nmapOsResp); err == nil && nmapOsResp.TaskID != "" {
		log.Printf("Received Nmap OS detection response for task %s", nmapOsResp.TaskID)
		return &models.Response{
			TaskID: nmapOsResp.TaskID,
			Result: nmapOsResp,
		}, nil
	}

	// Пробуем как NmapHostDiscoveryResponse
	var nmapHostResp models.NmapHostDiscoveryResponse
	if err := json.Unmarshal(body, &nmapHostResp); err == nil && nmapHostResp.TaskID != "" {
		log.Printf("Received Nmap host discovery response for task %s", nmapHostResp.TaskID)
		return &models.Response{
			TaskID: nmapHostResp.TaskID,
			Result: nmapHostResp,
		}, nil
	}

	return nil, fmt.Errorf("unable to parse response as any known type")
}

func generateCorrelationID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
