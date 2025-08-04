package handler

import (
	"context"
	"encoding/json"
	"scanner/internal/scanner"
	"scanner/pkg/logger"
	"scanner/pkg/queue"
)

func HandleMessage(ctx context.Context, msg queue.Delivery, scanner scanner.PortScanner, rabbitMQ *queue.RabbitMQ, log logger.Logger) {
	var req queue.ScanRequest
	if err := json.Unmarshal(msg.Body, &req); err != nil {
		log.Errorf("Failed to unmarshal scan request: %v", err)
		return
	}

	log.Infof("Received scan request: %s %s", req.IP, req.Ports)
	openPorts, err := scanner.ScanPorts(ctx, req.IP, req.Ports)

	if msg.ReplyTo != "" {
		sendResponse(rabbitMQ, msg, req, openPorts, err, log)
	}

	if err != nil {
		log.Errorf("Scan failed: %v", err)
		return
	}

	log.Infof("Scan completed for %s, open ports: %v", req.IP, openPorts)
}

func sendResponse(rabbitMQ *queue.RabbitMQ, msg queue.Delivery, req queue.ScanRequest, openPorts []int, err error, log logger.Logger) {
	response := queue.ScanResponse{
		TaskID:    req.TaskID,
		Status:    "completed",
		OpenPorts: openPorts,
	}
	if err != nil {
		response.Error = err.Error()
		response.Status = "failed"
	}

	if err := rabbitMQ.SendResponse(msg.ReplyTo, msg.CorrelationId, response); err != nil {
		log.Errorf("Failed to send RPC response: %v", err)
	}
}
