package services

import (
	"backend/domain/models"
	"log"
	"sync"
)

// RepositoryInterface определяет интерфейс для работы с данными
type RepositoryInterface interface {
	// ARP methods
	SaveARPHistory(record *models.ARPHistoryRecord) error
	GetARPHistory(limit int) ([]models.ARPHistoryRecord, error)
	DeleteARPHistory() error

	// ICMP methods
	SaveICMPHistory(record *models.ICMPHistoryRecord) error
	GetICMPHistory(limit int) ([]models.ICMPHistoryRecord, error)
	DeleteICMPHistory() error

	// Nmap TCP/UDP methods
	SaveNmapTcpUdpHistory(record *models.NmapTcpUdpHistoryRecord) error
	GetNmapTcpUdpHistory(limit int) ([]models.NmapTcpUdpHistoryRecord, error)
	DeleteNmapTcpUdpHistory() error

	// Nmap OS Detection methods
	SaveNmapOsDetectionHistory(record *models.NmapOsDetectionHistoryRecord) error
	GetNmapOsDetectionHistory(limit int) ([]models.NmapOsDetectionHistoryRecord, error)
	DeleteNmapOsDetectionHistory() error

	// Nmap Host Discovery methods
	SaveNmapHostDiscoveryHistory(record *models.NmapHostDiscoveryHistoryRecord) error
	GetNmapHostDiscoveryHistory(limit int) ([]models.NmapHostDiscoveryHistoryRecord, error)
	DeleteNmapHostDiscoveryHistory() error
}

// HistoryService управляет историей сканирования
type HistoryService struct {
	repo         RepositoryInterface
	requestCache map[string]interface{}
	mu           sync.RWMutex
}

// NewHistoryService создает новый сервис истории
func NewHistoryService(repo RepositoryInterface) *HistoryService {
	return &HistoryService{
		repo:         repo,
		requestCache: make(map[string]interface{}),
	}
}

// CacheRequest кэширует запрос
func (hs *HistoryService) CacheRequest(taskID string, request interface{}) {
	hs.mu.Lock()
	defer hs.mu.Unlock()
	hs.requestCache[taskID] = request
}

// GetCachedRequest получает кэшированный запрос
func (hs *HistoryService) GetCachedRequest(taskID string) interface{} {
	hs.mu.RLock()
	defer hs.mu.RUnlock()
	return hs.requestCache[taskID]
}

// RemoveCachedRequest удаляет из кэша
func (hs *HistoryService) RemoveCachedRequest(taskID string) {
	hs.mu.Lock()
	defer hs.mu.Unlock()
	delete(hs.requestCache, taskID)
}

// GetRepo возвращает репозиторий
func (hs *HistoryService) GetRepo() RepositoryInterface {
	return hs.repo
}

// SaveARPResponse сохраняет ARP ответ
func (hs *HistoryService) SaveARPResponse(result models.ARPResponse) {
	if hs.repo == nil {
		return
	}

	var interfaceName, ipRange string
	if cachedReq := hs.GetCachedRequest(result.TaskID); cachedReq != nil {
		if arpReq, ok := cachedReq.(models.ARPRequest); ok {
			interfaceName = arpReq.InterfaceName
			ipRange = arpReq.IPRange
		}
	}

	var onlineDevices, offlineDevices []models.ARPDevice
	for _, device := range result.Devices {
		if device.Status == "online" {
			onlineDevices = append(onlineDevices, device)
		} else {
			offlineDevices = append(offlineDevices, device)
		}
	}

	historyRecord := &models.ARPHistoryRecord{
		TaskID:         result.TaskID,
		InterfaceName:  interfaceName,
		IPRange:        ipRange,
		Status:         result.Status,
		Devices:        result.Devices,
		OnlineDevices:  onlineDevices,
		OfflineDevices: offlineDevices,
		TotalCount:     result.TotalCount,
		OnlineCount:    result.OnlineCount,
		OfflineCount:   result.OfflineCount,
		Error:          result.Error,
	}

	if err := hs.repo.SaveARPHistory(historyRecord); err != nil {
		log.Printf("Failed to save ARP history: %v", err)
	} else {
		log.Printf("Successfully saved ARP history for task %s", result.TaskID)
		hs.RemoveCachedRequest(result.TaskID)
	}
}

// SaveICMPResponse сохраняет ICMP ответ
func (hs *HistoryService) SaveICMPResponse(result models.ICMPResponse) {
	if hs.repo == nil {
		return
	}

	var targets []string
	var pingCount int
	if cachedReq := hs.GetCachedRequest(result.TaskID); cachedReq != nil {
		if icmpReq, ok := cachedReq.(models.ICMPRequest); ok {
			targets = icmpReq.Targets
			pingCount = icmpReq.PingCount
		}
	}

	historyRecord := &models.ICMPHistoryRecord{
		TaskID:    result.TaskID,
		Targets:   targets,
		PingCount: pingCount,
		Status:    result.Status,
		Results:   result.Results,
		Error:     result.Error,
	}

	if err := hs.repo.SaveICMPHistory(historyRecord); err != nil {
		log.Printf("Failed to save ICMP history: %v", err)
	} else {
		log.Printf("Successfully saved ICMP history for task %s", result.TaskID)
		hs.RemoveCachedRequest(result.TaskID)
	}
}

// SaveNmapTcpUdpResponse сохраняет Nmap TCP/UDP ответ
func (hs *HistoryService) SaveNmapTcpUdpResponse(result models.NmapTcpUdpResponse) {
	if hs.repo == nil {
		return
	}

	var ip, scannerType, ports string
	if cachedReq := hs.GetCachedRequest(result.TaskID); cachedReq != nil {
		if tcpUdpReq, ok := cachedReq.(models.NmapTcpUdpRequest); ok {
			ip = tcpUdpReq.IP
			scannerType = tcpUdpReq.ScannerType
			ports = tcpUdpReq.Ports
		}
	}

	historyRecord := &models.NmapTcpUdpHistoryRecord{
		TaskID:      result.TaskID,
		IP:          ip,
		ScannerType: scannerType,
		Ports:       ports,
		Host:        result.Host,
		PortInfo:    result.PortInfo,
		Status:      result.Status,
		Error:       result.Error,
	}

	if err := hs.repo.SaveNmapTcpUdpHistory(historyRecord); err != nil {
		log.Printf("Failed to save Nmap TCP/UDP history: %v", err)
	} else {
		log.Printf("Successfully saved Nmap TCP/UDP history for task %s", result.TaskID)
		hs.RemoveCachedRequest(result.TaskID)
	}
}

// SaveNmapOsDetectionResponse сохраняет Nmap OS Detection ответ
func (hs *HistoryService) SaveNmapOsDetectionResponse(result models.NmapOsDetectionResponse) {
	if hs.repo == nil {
		return
	}

	var ip string
	if cachedReq := hs.GetCachedRequest(result.TaskID); cachedReq != nil {
		if osReq, ok := cachedReq.(models.NmapOsDetectionRequest); ok {
			ip = osReq.IP
		}
	}

	historyRecord := &models.NmapOsDetectionHistoryRecord{
		TaskID:   result.TaskID,
		IP:       ip,
		Host:     result.Host,
		Name:     result.Name,
		Accuracy: result.Accuracy,
		Vendor:   result.Vendor,
		Family:   result.Family,
		Type:     result.Type,
		Status:   result.Status,
		Error:    result.Error,
	}

	if err := hs.repo.SaveNmapOsDetectionHistory(historyRecord); err != nil {
		log.Printf("Failed to save Nmap OS Detection history: %v", err)
	} else {
		log.Printf("Successfully saved Nmap OS Detection history for task %s", result.TaskID)
		hs.RemoveCachedRequest(result.TaskID)
	}
}

// SaveNmapHostDiscoveryResponse сохраняет Nmap Host Discovery ответ
func (hs *HistoryService) SaveNmapHostDiscoveryResponse(result models.NmapHostDiscoveryResponse) {
	if hs.repo == nil {
		return
	}

	var ip string
	if cachedReq := hs.GetCachedRequest(result.TaskID); cachedReq != nil {
		if hostReq, ok := cachedReq.(models.NmapHostDiscoveryRequest); ok {
			ip = hostReq.IP
		}
	}

	historyRecord := &models.NmapHostDiscoveryHistoryRecord{
		TaskID:    result.TaskID,
		IP:        ip,
		Host:      result.Host,
		HostUP:    result.HostUP,
		HostTotal: result.HostTotal,
		Status:    result.Status,
		DNS:       result.DNS,
		Reason:    result.Reason,
		Error:     result.Error,
	}

	if err := hs.repo.SaveNmapHostDiscoveryHistory(historyRecord); err != nil {
		log.Printf("Failed to save Nmap Host Discovery history: %v", err)
	} else {
		log.Printf("Successfully saved Nmap Host Discovery history for task %s", result.TaskID)
		hs.RemoveCachedRequest(result.TaskID)
	}
}
