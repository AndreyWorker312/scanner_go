package services

import (
	"backend/domain/models"
	rabbitmq "backend/internal/infrastructure/messaging"
	"log"
)

type PublisherService struct {
	publisher *rabbitmq.RPCScannerPublisher
}

func NewPublisherService(publisher *rabbitmq.RPCScannerPublisher) *PublisherService {
	return &PublisherService{
		publisher: publisher,
	}
}

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

func (ps *PublisherService) PublishTCPRequest(req models.TCPRequest) *models.Response {
	log.Printf("Publishing TCP request: %+v", req)

	resp, err := ps.publisher.PublishTcp(req)
	if err != nil {
		log.Printf("Failed to publish TCP task: %v", err)
		return &models.Response{
			TaskID: req.TaskID,
			Result: map[string]string{"error": err.Error()},
		}
	}
	return resp
}

func (ps *PublisherService) SetResponseCallback(callback func(*models.Response)) {
	ps.publisher.SetResponseCallback(callback)
}
