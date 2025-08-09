package handler

import (
	"context"
	"encoding/json"
	"scanner_icmp/internal/config"
	"scanner_icmp/internal/scanner"
	"scanner_icmp/pkg/logger"
	"scanner_icmp/pkg/queue"
)

func HandleMessage(ctx context.Context, msg queue.Delivery, rabbitMQ *queue.RabbitMQ, log logger.Logger, cfg *config.Config) {
	var req queue.PingRequest
	if err := json.Unmarshal(msg.Body, &req); err != nil {
		log.Errorf("Failed to unmarshal Ping scan request: %v", err)
		return
	}

	log.Infof("Received Ping scan request for targets: %v", req.Targets)

	pingScanner := scanner.NewPingScanner(req.PingCount, cfg.PingTimeout)

	results := make([]scanner.PingResult, 0, len(req.Targets))
	for _, target := range req.Targets {
		select {
		case <-ctx.Done():
			return
		default:
			result := pingScanner.Ping(ctx, target)
			results = append(results, result)
		}
	}

	if msg.ReplyTo != "" {
		sendResponse(rabbitMQ, msg, req, results, log)
	}

	log.Infof("Ping scan completed, scanned %d targets", len(results))
}

func sendResponse(rabbitMQ *queue.RabbitMQ, msg queue.Delivery, req queue.PingRequest, results []scanner.PingResult, log logger.Logger) {
	response := queue.PingResponse{
		TaskID:  req.TaskID,
		Status:  "completed",
		Results: results,
	}

	if err := rabbitMQ.SendResponse(msg.ReplyTo, msg.CorrelationId, response); err != nil {
		log.Errorf("Failed to send RPC response: %v", err)
	}
}
