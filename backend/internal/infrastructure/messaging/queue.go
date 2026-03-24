package rabbitmq

import (
	"backend/domain/models"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

type RPCScannerPublisher struct {
	conn       *amqp.Connection
	channel    *amqp.Channel
	replies    map[string]chan *models.Response
	mu         sync.Mutex
	onResponse func(*models.Response)
}

var (
	rpcPublisherInstance *RPCScannerPublisher
	rpcPublisherOnce     sync.Once
	rpcPublisherErr      error
)

func GetRPCconnection(amqpURI string) (*RPCScannerPublisher, error) {
	rpcPublisherOnce.Do(func() {
		log.Printf("[RabbitMQ] Starting connection loop to %s", amqpURI)
		for i := 1; i <= 10; i++ {
			start := time.Now()
			log.Printf("[RabbitMQ] Attempt %d/10 at %s", i, start.Format(time.RFC3339))

			// ── DNS diagnostics ──────────────────────────────────────────
			host, port, parseErr := net.SplitHostPort(extractHostPort(amqpURI))
			if parseErr == nil {
				log.Printf("[RabbitMQ] Resolving host=%q port=%q", host, port)
				addrs, dnsErr := net.LookupHost(host)
				if dnsErr != nil {
					log.Printf("[RabbitMQ] DNS lookup FAILED for %q: %v", host, dnsErr)
				} else {
					log.Printf("[RabbitMQ] DNS resolved %q → %v", host, addrs)

					// ── TCP reachability check ───────────────────────────────
					addr := net.JoinHostPort(addrs[0], port)
					log.Printf("[RabbitMQ] TCP probe %s ...", addr)
					conn, tcpErr := net.DialTimeout("tcp", addr, 5*time.Second)
					if tcpErr != nil {
						log.Printf("[RabbitMQ] TCP probe FAILED: %v (elapsed %s)", tcpErr, time.Since(start))
					} else {
						conn.Close()
						log.Printf("[RabbitMQ] TCP probe OK in %s — port is reachable", time.Since(start))
					}
				}
			}

			// ── actual AMQP dial ─────────────────────────────────────────
			rpcPublisherInstance, rpcPublisherErr = newRPCScannerPublisher(amqpURI)
			elapsed := time.Since(start)
			if rpcPublisherErr == nil {
				log.Printf("[RabbitMQ] Connected successfully on attempt %d (took %s)", i, elapsed)
				return
			}
			log.Printf("[RabbitMQ] Attempt %d/10 FAILED after %s: %v", i, elapsed, rpcPublisherErr)

			delay := time.Duration(i) * 3 * time.Second
			log.Printf("[RabbitMQ] Waiting %s before next attempt...", delay)
			time.Sleep(delay)
		}
		log.Printf("[RabbitMQ] All 10 attempts exhausted — giving up")
	})
	return rpcPublisherInstance, rpcPublisherErr
}

// extractHostPort parses "amqp://user:pass@host:port/vhost" → "host:port"
func extractHostPort(uri string) string {
	// simple extraction without full URL parsing dependency
	const prefix = "://"
	s := uri
	if idx := indexOf(s, prefix); idx >= 0 {
		s = s[idx+len(prefix):]
	}
	// strip user:pass@
	if idx := indexOf(s, "@"); idx >= 0 {
		s = s[idx+1:]
	}
	// strip /vhost
	if idx := indexOf(s, "/"); idx >= 0 {
		s = s[:idx]
	}
	return s // "host:port"
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
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

func (p *RPCScannerPublisher) PublishTcp(req models.TCPRequest) (*models.Response, error) {
	return p.publishRPC("tcp_service", req)
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

	log.Printf("About to marshal task for %s: %+v", queueName, task)
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

func (p *RPCScannerPublisher) SetResponseCallback(callback func(*models.Response)) {
	p.onResponse = callback
}

func (p *RPCScannerPublisher) parseResponse(body []byte) (*models.Response, error) {
	log.Printf("Raw response body: %s", string(body))

	var icmpResp models.ICMPResponse
	if err := json.Unmarshal(body, &icmpResp); err == nil && icmpResp.TaskID != "" && len(icmpResp.Results) > 0 {
		log.Printf("Received ICMP response for task %s with %d results", icmpResp.TaskID, len(icmpResp.Results))
		response := &models.Response{
			TaskID: icmpResp.TaskID,
			Result: icmpResp,
		}

		if p.onResponse != nil {
			p.onResponse(response)
		}

		return response, nil
	}

	var arpResp models.ARPResponse
	if err := json.Unmarshal(body, &arpResp); err == nil && arpResp.TaskID != "" && (len(arpResp.OnlineDevices) > 0 || len(arpResp.OfflineDevices) > 0) {
		log.Printf("Received ARP response for task %s: Total=%d, Online=%d, Offline=%d",
			arpResp.TaskID, arpResp.TotalCount, arpResp.OnlineCount, arpResp.OfflineCount)
		log.Printf("ARP response details: Status=%s, Error=%s", arpResp.Status, arpResp.Error)
		response := &models.Response{
			TaskID: arpResp.TaskID,
			Result: arpResp,
		}

		if p.onResponse != nil {
			p.onResponse(response)
		}

		return response, nil
	}

	var nmapTcpUdpResp models.NmapTcpUdpResponse
	if err := json.Unmarshal(body, &nmapTcpUdpResp); err == nil && nmapTcpUdpResp.TaskID != "" && len(nmapTcpUdpResp.PortInfo) > 0 {
		log.Printf("Received Nmap TCP/UDP response for task %s", nmapTcpUdpResp.TaskID)
		response := &models.Response{
			TaskID: nmapTcpUdpResp.TaskID,
			Result: nmapTcpUdpResp,
		}

		if p.onResponse != nil {
			p.onResponse(response)
		}

		return response, nil
	}

	var nmapOsResp models.NmapOsDetectionResponse
	if err := json.Unmarshal(body, &nmapOsResp); err == nil && nmapOsResp.TaskID != "" && (nmapOsResp.Name != "" || nmapOsResp.Vendor != "" || nmapOsResp.Family != "") {
		log.Printf("Received Nmap OS detection response for task %s", nmapOsResp.TaskID)
		response := &models.Response{
			TaskID: nmapOsResp.TaskID,
			Result: nmapOsResp,
		}

		if p.onResponse != nil {
			p.onResponse(response)
		}

		return response, nil
	}

	var nmapHostResp models.NmapHostDiscoveryResponse
	if err := json.Unmarshal(body, &nmapHostResp); err == nil && nmapHostResp.TaskID != "" && (nmapHostResp.Status != "" || nmapHostResp.DNS != "" || nmapHostResp.Reason != "") {
		log.Printf("Received Nmap host discovery response for task %s", nmapHostResp.TaskID)
		response := &models.Response{
			TaskID: nmapHostResp.TaskID,
			Result: nmapHostResp,
		}

		if p.onResponse != nil {
			p.onResponse(response)
		}

		return response, nil
	}

	var response models.Response
	if err := json.Unmarshal(body, &response); err == nil && response.TaskID != "" {
		log.Printf("Received generic response for task %s", response.TaskID)

		if p.onResponse != nil {
			p.onResponse(&response)
		}

		return &response, nil
	}

	return nil, fmt.Errorf("unable to parse response as any known type")
}

func generateCorrelationID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
