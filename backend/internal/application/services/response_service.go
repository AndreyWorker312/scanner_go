package services

import (
	"backend/domain/models"
	"log"
)

// ResponseService обрабатывает ответы от сканеров
type ResponseService struct {
	historyService *HistoryService
}

// NewResponseService создает новый сервис ответов
func NewResponseService(historyService *HistoryService) *ResponseService {
	return &ResponseService{
		historyService: historyService,
	}
}

// ProcessResponse обрабатывает ответы от сканеров
func (rs *ResponseService) ProcessResponse(response *models.Response) {
	if response == nil {
		log.Printf("ProcessResponse: response is nil")
		return
	}

	log.Printf("ProcessResponse: processing response for task %s", response.TaskID)
	log.Printf("ProcessResponse: response result type: %T", response.Result)

	switch result := response.Result.(type) {
	case models.ARPResponse:
		log.Printf("Processing ARP response")
		rs.historyService.SaveARPResponse(result)
	case models.ICMPResponse:
		log.Printf("Processing ICMP response")
		rs.historyService.SaveICMPResponse(result)
	case models.NmapTcpUdpResponse:
		log.Printf("Processing Nmap TCP/UDP response")
		rs.historyService.SaveNmapTcpUdpResponse(result)
	case models.NmapOsDetectionResponse:
		log.Printf("Processing Nmap OS Detection response")
		rs.historyService.SaveNmapOsDetectionResponse(result)
	case models.NmapHostDiscoveryResponse:
		log.Printf("Processing Nmap Host Discovery response")
		rs.historyService.SaveNmapHostDiscoveryResponse(result)
	default:
		log.Printf("Unknown response type: %T", result)
	}
}
