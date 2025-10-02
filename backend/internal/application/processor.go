package application

import (
	"backend/domain/models"
	rabbitmq "backend/internal/infrastructure"
	"encoding/json"
	"log"
)

type App struct {
	publisher *rabbitmq.RPCScannerPublisher
}

func NewApp(publisher *rabbitmq.RPCScannerPublisher) *App {
	return &App{
		publisher: publisher,
	}
}

func (a *App) ProcessRequest(req *models.Request) *models.Response {
	switch req.ScannerService {
	case "nmap_service":
		return a.processNmapRequest(req.Options)
	case "arp_service":
		return a.processArpRequest(req.Options)
	case "icmp_service":
		return a.processIcmpRequest(req.Options)
	default:
		return &models.Response{
			TaskID: "unknown",
			Result: map[string]string{"error": "unknown scanner service"},
		}
	}
}

func (a *App) processNmapRequest(options any) *models.Response {
	// Определяем тип Nmap запроса на основе полей
	optionsJSON, err := json.Marshal(options)
	if err != nil {
		log.Printf("Failed to marshal nmap options: %v", err)
		return &models.Response{
			TaskID: "error",
			Result: map[string]string{"error": "invalid nmap options"},
		}
	}

	// Пробуем определить тип запроса по наличию полей
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

	// Обрабатываем в зависимости от типа сканирования
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
		return a.publishNmapRequest(tcpUdpReq)

	case scanType.ScanMethod == "os_detection":
		var osReq models.NmapOsDetectionRequest
		if err := json.Unmarshal(optionsJSON, &osReq); err != nil {
			log.Printf("Failed to unmarshal OS detection request: %v", err)
			return &models.Response{
				TaskID: "error",
				Result: map[string]string{"error": "invalid OS detection request"},
			}
		}
		return a.publishNmapRequest(osReq)

	case scanType.ScanMethod == "host_discovery":
		var hostReq models.NmapHostDiscoveryRequest
		if err := json.Unmarshal(optionsJSON, &hostReq); err != nil {
			log.Printf("Failed to unmarshal host discovery request: %v", err)
			return &models.Response{
				TaskID: "error",
				Result: map[string]string{"error": "invalid host discovery request"},
			}
		}
		return a.publishNmapRequest(hostReq)

	default:
		// Пробуем как базовый NmapRequest
		var nmapReq models.NmapRequest
		if err := json.Unmarshal(optionsJSON, &nmapReq); err != nil {
			log.Printf("Failed to unmarshal basic nmap request: %v", err)
			return &models.Response{
				TaskID: "error",
				Result: map[string]string{"error": "invalid nmap request"},
			}
		}
		return a.publishNmapRequest(nmapReq)
	}
}

func (a *App) processArpRequest(options any) *models.Response {
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

	resp, err := a.publisher.PublishArp(arpReq)
	if err != nil {
		log.Printf("Failed to publish ARP task: %v", err)
		return &models.Response{
			TaskID: arpReq.TaskID,
			Result: map[string]string{"error": err.Error()},
		}
	}
	return resp
}

func (a *App) processIcmpRequest(options any) *models.Response {
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

	resp, err := a.publisher.PublishIcmp(icmpReq)
	if err != nil {
		log.Printf("Failed to publish ICMP task: %v", err)
		return &models.Response{
			TaskID: icmpReq.TaskID,
			Result: map[string]string{"error": err.Error()},
		}
	}
	return resp
}

// Универсальный метод для публикации Nmap запросов
func (a *App) publishNmapRequest(req interface{}) *models.Response {
	resp, err := a.publisher.PublishNmap(req)
	if err != nil {
		log.Printf("Failed to publish Nmap task: %v", err)

		// Пытаемся извлечь TaskID из запроса
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
