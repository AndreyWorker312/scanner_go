package handler

import (
	"context"
	"encoding/json"
	"scanner_tcp/internal/scanner"
	"scanner_tcp/pkg/logger"
	"scanner_tcp/pkg/queue"
)

func HandleMessage(ctx context.Context, msg queue.Delivery, rabbitMQ *queue.RabbitMQ, tcpScanner scanner.TCPScanner, log logger.Logger) {
	var req queue.TCPRequest
	if err := json.Unmarshal(msg.Body, &req); err != nil {
		log.Errorf("Failed to unmarshal TCP scan request: %v", err)
		return
	}

	log.Infof("Received TCP scan request for host: %s, ports: %v", req.Host, req.Ports)

	// Выполняем сканирование
	scanResults := tcpScanner.ScanPorts(ctx, req.Host, req.Ports)

	// Преобразуем результаты
	results := make([]queue.PortScanResult, len(scanResults))
	for i, sr := range scanResults {
		results[i] = queue.PortScanResult{
			Port:         sr.Port,
			State:        sr.State,
			Service:      sr.Service,
			Banner:       sr.Banner,
			Version:      sr.Version,
			Error:        sr.Error,
			ResponseTime: sr.ResponseTime,
		}
	}

	if msg.ReplyTo != "" {
		sendResponse(rabbitMQ, msg, req, results, log)
	}

	log.Infof("TCP scan completed for task %s", req.TaskID)
}

func sendResponse(rabbitMQ *queue.RabbitMQ, msg queue.Delivery, req queue.TCPRequest, results []queue.PortScanResult, log logger.Logger) {
	response := queue.TCPResponse{
		TaskID:  req.TaskID,
		Host:    req.Host,
		Results: results,
		Status:  "completed",
	}

	// Проверяем, есть ли ошибки в результатах
	hasErrors := false
	for _, result := range results {
		if result.Error != "" {
			hasErrors = true
			break
		}
	}

	if hasErrors {
		log.Warnf("TCP scan completed with some errors for task %s", req.TaskID)
	}

	log.Infof("Sending TCP response: TaskID=%s, Status=%s, Host=%s, Results=%d",
		response.TaskID, response.Status, response.Host, len(response.Results))

	if err := rabbitMQ.SendResponse(msg.ReplyTo, msg.CorrelationId, response); err != nil {
		log.Errorf("Failed to send RPC response: %v", err)
	} else {
		log.Infof("Successfully sent TCP response for task %s", response.TaskID)
	}
}

