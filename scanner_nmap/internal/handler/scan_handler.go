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
	log.Infof("Received scan request: %s", string(msg.Body))

	if msg.ReplyTo == "" {
		return
	}

	// Пробуем определить тип запроса по наличию полей
	var scanType struct {
		ScanMethod  string `json:"scan_method"`
		ScannerType string `json:"scanner_type"`
	}

	if err := json.Unmarshal(msg.Body, &scanType); err != nil {
		log.Errorf("Failed to unmarshal scan type: %v", err)
		return
	}

	log.Infof("Scan method: %s, Scanner type: %s", scanType.ScanMethod, scanType.ScannerType)

	// Определяем тип сканирования и обрабатываем
	switch {
	case scanType.ScanMethod == "tcp_udp_scan" || scanType.ScannerType == "tcp_scan" || scanType.ScannerType == "udp_scan":
		var tcpUdpRequest domain.ScanTcpUdpRequest
		if err := json.Unmarshal(msg.Body, &tcpUdpRequest); err != nil {
			log.Errorf("Failed to unmarshal TCP/UDP request: %v", err)
			return
		}
		log.Infof("Processing TCP/UDP scan for %s on ports %s", tcpUdpRequest.IP, tcpUdpRequest.Ports)
		req, err := usecases.UdpTcpScanner(ctx, tcpUdpRequest)
		if err != nil {
			log.Errorf("Failed to scan TCP/UDP: %v", err)
		}
		sendResponse(rabbitMQ, msg, req, err, log)

	case scanType.ScanMethod == "os_detection":
		var osRequest domain.OsDetectionRequest
		if err := json.Unmarshal(msg.Body, &osRequest); err != nil {
			log.Errorf("Failed to unmarshal OS detection request: %v", err)
			return
		}
		log.Infof("Processing OS detection for %s", osRequest.IP)
		req, err := usecases.OSDetectionScanner(ctx, osRequest)
		if err != nil {
			log.Errorf("Failed to scan OS detection: %v", err)
		}
		sendResponse(rabbitMQ, msg, req, err, log)

	case scanType.ScanMethod == "host_discovery":
		var hostRequest domain.HostDiscoveryRequest
		if err := json.Unmarshal(msg.Body, &hostRequest); err != nil {
			log.Errorf("Failed to unmarshal host discovery request: %v", err)
			return
		}
		log.Infof("Processing host discovery for %s", hostRequest.IP)
		req, err := usecases.HostDiscoveryScanner(ctx, hostRequest)
		if err != nil {
			log.Errorf("Failed to scan host discovery: %v", err)
		}
		sendResponse(rabbitMQ, msg, req, err, log)

	default:
		log.Errorf("Invalid scan method: %s", scanType.ScanMethod)
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
