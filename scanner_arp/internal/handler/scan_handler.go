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
	// Преобразуем scanner.DeviceInfo в queue.ARPDevice
	arpDevices := make([]queue.ARPDevice, len(devices))
	var onlineDevices []queue.ARPDevice
	var offlineDevices []queue.ARPDevice

	for i, device := range devices {
		arpDevice := queue.ARPDevice{
			IP:     device.IP,
			MAC:    device.MAC,
			Vendor: device.Vendor,
			Status: device.Status,
		}
		arpDevices[i] = arpDevice

		// Разделяем на онлайн и офлайн
		if device.Status == "online" {
			onlineDevices = append(onlineDevices, arpDevice)
		} else {
			offlineDevices = append(offlineDevices, arpDevice)
		}
	}

	response := queue.ARPResponse{
		TaskID:         req.TaskID,
		Status:         "completed",
		Devices:        arpDevices,
		OnlineDevices:  onlineDevices,
		OfflineDevices: offlineDevices,
		TotalCount:     len(arpDevices),
		OnlineCount:    len(onlineDevices),
		OfflineCount:   len(offlineDevices),
	}
	if err != nil {
		response.Error = err.Error()
		response.Status = "failed"
	}

	log.Infof("Sending ARP response: TaskID=%s, Status=%s, Total=%d, Online=%d, Offline=%d, Error=%s",
		response.TaskID, response.Status, response.TotalCount, response.OnlineCount, response.OfflineCount, response.Error)

	if err := rabbitMQ.SendResponse(msg.ReplyTo, msg.CorrelationId, response); err != nil {
		log.Errorf("Failed to send RPC response: %v", err)
	} else {
		log.Infof("Successfully sent ARP response for task %s", response.TaskID)
	}
}
