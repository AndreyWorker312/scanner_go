package services

import (
	"backend/domain/models"
	rabbitmq "backend/internal/infrastructure/messaging"
	"log"
)

// PublisherService управляет публикацией сообщений
type PublisherService struct {
	publisher *rabbitmq.RPCScannerPublisher
}

// NewPublisherService создает новый сервис публикации
func NewPublisherService(publisher *rabbitmq.RPCScannerPublisher) *PublisherService {
	return &PublisherService{
		publisher: publisher,
	}
}

// PublishNmapRequest публикует Nmap запрос
func (ps *PublisherService) PublishNmapRequest(req interface{}) *models.Response {
	log.Printf("Publishing Nmap request: %+v", req)

	resp, err := ps.publisher.PublishNmap(req)
	if err != nil {
		log.Printf("Failed to publish Nmap task: %v", err)

		var taskID string
		switch r := req.(type) {
		case models.NmapTcpUdpRequest:
			taskID = r.TaskID
		case models.NmapOsDetectionRequest:
			taskID = r.TaskID
		case models.NmapHostDiscoveryRequest:
			taskID = r.TaskID
		case models.NmapRequest:
			taskID = "unknown"
		}

		return &models.Response{
			TaskID: taskID,
			Result: map[string]string{"error": err.Error()},
		}
	}
	return resp
}

// PublishARPRequest публикует ARP запрос
func (ps *PublisherService) PublishARPRequest(req models.ARPRequest) *models.Response {
	log.Printf("Publishing ARP request: %+v", req)

	resp, err := ps.publisher.PublishArp(req)
	if err != nil {
		log.Printf("Failed to publish ARP task: %v", err)
		return &models.Response{
			TaskID: req.TaskID,
			Result: map[string]string{"error": err.Error()},
		}
	}
	return resp
}

// PublishICMPRequest публикует ICMP запрос
func (ps *PublisherService) PublishICMPRequest(req models.ICMPRequest) *models.Response {
	log.Printf("Publishing ICMP request: %+v", req)

	resp, err := ps.publisher.PublishIcmp(req)
	if err != nil {
		log.Printf("Failed to publish ICMP task: %v", err)
		return &models.Response{
			TaskID: req.TaskID,
			Result: map[string]string{"error": err.Error()},
		}
	}
	return resp
}

// SetResponseCallback устанавливает callback для ответов
func (ps *PublisherService) SetResponseCallback(callback func(*models.Response)) {
	ps.publisher.SetResponseCallback(callback)
}
