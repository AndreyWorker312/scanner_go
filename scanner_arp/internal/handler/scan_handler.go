package handler

import (
	"arp_scanner/internal/scanner"
	"arp_scanner/pkg/logger"
	"arp_scanner/pkg/queue"
	"context"
	"encoding/json"
)

func HandleMessage(ctx context.Context, msg queue.Delivery, rabbitMQ *queue.RabbitMQ, log logger.Logger) {
	var req queue.ARPRequest
	if err := json.Unmarshal(msg.Body, &req); err != nil {
		log.Errorf("Failed to unmarshal ARP scan request: %v", err)
		return
	}

	log.Infof("Received ARP scan request for range: %s on interface: %s", req.IPRange, req.InterfaceName)

	arpScanner := scanner.NewARPScanner(
		req.InterfaceName,
		scanner.DefaultTimeout,
		scanner.DefaultMaxRetries,
		scanner.DefaultRetryDelay,
	)

	devices, err := arpScanner.Scan(ctx, req.IPRange)

	if msg.ReplyTo != "" {
		sendResponse(rabbitMQ, msg, req, devices, err, log)
	}

	if err != nil {
		log.Errorf("ARP scan failed: %v", err)
		return
	}

	log.Infof("ARP scan completed, found %d devices", len(devices))
}

func sendResponse(rabbitMQ *queue.RabbitMQ, msg queue.Delivery, req queue.ARPRequest, devices []scanner.DeviceInfo, err error, log logger.Logger) {
	response := queue.ARPResponse{
		TaskID:  req.TaskID,
		Status:  "completed",
		Devices: devices,
	}
	if err != nil {
		response.Error = err.Error()
		response.Status = "failed"
	}

	if err := rabbitMQ.SendResponse(msg.ReplyTo, msg.CorrelationId, response); err != nil {
		log.Errorf("Failed to send RPC response: %v", err)
	}
}
