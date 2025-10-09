package services

import (
	"backend/domain/models"
	"encoding/json"
	"log"
)

// RequestService обрабатывает входящие запросы
type RequestService struct{}

// NewRequestService создает новый сервис запросов
func NewRequestService() *RequestService {
	return &RequestService{}
}

// ProcessRequest обрабатывает входящие запросы
func (rs *RequestService) ProcessRequest(req *models.Request) *models.Response {
	switch req.ScannerService {
	case "nmap_service":
		return rs.processNmapRequest(req.Options)
	case "arp_service":
		return rs.processArpRequest(req.Options)
	case "icmp_service":
		return rs.processIcmpRequest(req.Options)
	default:
		return &models.Response{
			TaskID: "unknown",
			Result: map[string]string{"error": "unknown scanner service"},
		}
	}
}

// processNmapRequest обрабатывает Nmap запросы
func (rs *RequestService) processNmapRequest(options any) *models.Response {
	optionsJSON, err := json.Marshal(options)
	if err != nil {
		log.Printf("Failed to marshal nmap options: %v", err)
		return &models.Response{
			TaskID: "error",
			Result: map[string]string{"error": "invalid nmap options"},
		}
	}

	var scanType struct {
		ScannerType string `json:"scanner_type"`
		ScanMethod  string `json:"scan_method"`
	}

	if err := json.Unmarshal(optionsJSON, &scanType); err != nil {
		log.Printf("Failed to unmarshal nmap scan type: %v", err)
		return &models.Response{
			TaskID: "error",
			Result: map[string]string{"error": "invalid nmap request format"},
		}
	}

	switch {
	case scanType.ScannerType == "tcp_udp_scan" || scanType.ScanMethod == "tcp_udp_scan":
		var tcpUdpReq models.NmapTcpUdpRequest
		if err := json.Unmarshal(optionsJSON, &tcpUdpReq); err != nil {
			log.Printf("Failed to unmarshal TCP/UDP request: %v", err)
			return &models.Response{
				TaskID: "error",
				Result: map[string]string{"error": "invalid TCP/UDP request"},
			}
		}
		return &models.Response{TaskID: tcpUdpReq.TaskID, Result: tcpUdpReq}

	case scanType.ScanMethod == "os_detection":
		var osReq models.NmapOsDetectionRequest
		if err := json.Unmarshal(optionsJSON, &osReq); err != nil {
			log.Printf("Failed to unmarshal OS detection request: %v", err)
			return &models.Response{
				TaskID: "error",
				Result: map[string]string{"error": "invalid OS detection request"},
			}
		}
		return &models.Response{TaskID: osReq.TaskID, Result: osReq}

	case scanType.ScanMethod == "host_discovery":
		var hostReq models.NmapHostDiscoveryRequest
		if err := json.Unmarshal(optionsJSON, &hostReq); err != nil {
			log.Printf("Failed to unmarshal host discovery request: %v", err)
			return &models.Response{
				TaskID: "error",
				Result: map[string]string{"error": "invalid host discovery request"},
			}
		}
		return &models.Response{TaskID: hostReq.TaskID, Result: hostReq}

	default:
		var nmapReq models.NmapRequest
		if err := json.Unmarshal(optionsJSON, &nmapReq); err != nil {
			log.Printf("Failed to unmarshal basic nmap request: %v", err)
			return &models.Response{
				TaskID: "error",
				Result: map[string]string{"error": "invalid nmap request"},
			}
		}
		return &models.Response{TaskID: "unknown", Result: nmapReq}
	}
}

// processArpRequest обрабатывает ARP запросы
func (rs *RequestService) processArpRequest(options any) *models.Response {
	var arpReq models.ARPRequest
	optionsJSON, err := json.Marshal(options)
	if err != nil {
		log.Printf("Failed to marshal ARP options: %v", err)
		return &models.Response{
			TaskID: "error",
			Result: map[string]string{"error": "invalid ARP options"},
		}
	}

	if err := json.Unmarshal(optionsJSON, &arpReq); err != nil {
		log.Printf("Failed to unmarshal ARP request: %v", err)
		return &models.Response{
			TaskID: "error",
			Result: map[string]string{"error": "invalid ARP request"},
		}
	}

	return &models.Response{TaskID: arpReq.TaskID, Result: arpReq}
}

// processIcmpRequest обрабатывает ICMP запросы
func (rs *RequestService) processIcmpRequest(options any) *models.Response {
	var icmpReq models.ICMPRequest
	optionsJSON, err := json.Marshal(options)
	if err != nil {
		log.Printf("Failed to marshal ICMP options: %v", err)
		return &models.Response{
			TaskID: "error",
			Result: map[string]string{"error": "invalid ICMP options"},
		}
	}

	if err := json.Unmarshal(optionsJSON, &icmpReq); err != nil {
		log.Printf("Failed to unmarshal ICMP request: %v", err)
		return &models.Response{
			TaskID: "error",
			Result: map[string]string{"error": "invalid ICMP request"},
		}
	}

	return &models.Response{TaskID: icmpReq.TaskID, Result: icmpReq}
}
