package handler

import (
	"context"
	"encoding/json"
	"scanner_nmap/internal/domain"
	"scanner_nmap/internal/usecases"
	"scanner_nmap/pkg/logger"
	"scanner_nmap/pkg/queue"
)

func HandleMessage(ctx context.Context, msg queue.Delivery, rabbitMQ *queue.RabbitMQ, log logger.Logger) {
	var rawReq domain.RawRequest
	if err := json.Unmarshal(msg.Body, &rawReq); err != nil {
		log.Errorf("Failed to unmarshal scan request: %v", err)
		return
	}

	log.Infof("Received scan request")

	if msg.ReplyTo == "" {
		return
	}

	switch rawReq.ScanMethod {
	case "TCP/UDP":
		var udpRequest domain.ScanTcpUdpRequest
		if err := json.Unmarshal(msg.Body, &udpRequest); err != nil {
			log.Errorf("Failed to unmarshal scan request: %v", err)
			return
		}
		req, err := usecases.UdpTcpScanner(ctx, udpRequest)
		if err != nil {
			log.Errorf("Failed to scan udp response: %v", err)
		}
		sendResponse(rabbitMQ, msg, req, err, log)
	case "OC":
		var udpRequest domain.OsDetectionRequest
		if err := json.Unmarshal(msg.Body, &udpRequest); err != nil {
			log.Errorf("Failed to unmarshal scan request: %v", err)
			return
		}
		req, err := usecases.OSDetectionScanner(ctx, udpRequest)
		if err != nil {
			log.Errorf("Failed to scan udp response: %v", err)
		}
		sendResponse(rabbitMQ, msg, req, err, log)
	case "HOST":
		var udpRequest domain.HostDiscoveryRequest
		if err := json.Unmarshal(msg.Body, &udpRequest); err != nil {
			log.Errorf("Failed to unmarshal scan request: %v", err)
			return
		}
		req, err := usecases.HostDiscoveryScanner(ctx, udpRequest)
		if err != nil {
			log.Errorf("Failed to scan udp response: %v", err)
		}
		sendResponse(rabbitMQ, msg, req, err, log)
	default:
		log.Errorf("Invalid scan method: %s", rawReq.ScanMethod)
	}

	log.Infof("Scan completed")
}

func sendResponse[T domain.ScanTcpUdpResponse | domain.OsDetectionResponse | domain.HostDiscoveryResponse](
	rabbitMQ *queue.RabbitMQ,
	msg queue.Delivery,
	req T,
	err error,
	log logger.Logger,
) {
	switch r := any(req).(type) {
	case domain.ScanTcpUdpResponse:
		response := domain.ScanTcpUdpResponse{
			TaskID:   r.TaskID,
			Host:     r.Host,
			PortInfo: r.PortInfo,
			Status:   "completed",
		}
		if err != nil {
			response.Status = "failed"
			log.Errorf("TCP/UDP scan failed: %v", err)
		} else {
			log.Infof("TCP/UDP scan completed for task %s", r.TaskID)
		}

		body, _ := json.Marshal(response)

		if err := rabbitMQ.SendResponse(msg.ReplyTo, msg.CorrelationId, body); err != nil {
			log.Errorf("Failed to send TCP/UDP response: %v", err)
		}

	case domain.OsDetectionResponse:
		response := domain.OsDetectionResponse{
			TaskID:   r.TaskID,
			Host:     r.Host,
			Name:     r.Name,
			Accuracy: r.Accuracy,
			Vendor:   r.Vendor,
			Family:   r.Family,
			Type:     r.Type,
			Status:   "completed",
		}
		if err != nil {
			response.Status = "failed"
			log.Errorf("OS detection failed: %v", err)
		} else {
			log.Infof("OS detection completed for task %s", r.TaskID)
		}

		body, _ := json.Marshal(response)

		if err := rabbitMQ.SendResponse(msg.ReplyTo, msg.CorrelationId, body); err != nil {
			log.Errorf("Failed to send OS detection response: %v", err)
		}

	case domain.HostDiscoveryResponse:
		response := domain.HostDiscoveryResponse{
			TaskID:    r.TaskID,
			Host:      r.Host,
			Status:    r.Status,
			HostUP:    r.HostUP,
			HostTotal: r.HostTotal,
			DNS:       r.DNS,
			Reason:    r.Reason,
		}
		if err != nil {
			response.Status = "failed"
			log.Errorf("Host discovery failed: %v", err)
		} else {
			log.Infof("Host discovery completed for task %s", r.TaskID)
		}

		body, _ := json.Marshal(response)

		if err := rabbitMQ.SendResponse(msg.ReplyTo, msg.CorrelationId, body); err != nil {
			log.Errorf("Failed to send host discovery response: %v", err)
		}

	default:
		log.Errorf("Unknown response type: %T", req)
	}
}
