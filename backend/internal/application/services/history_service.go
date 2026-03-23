package services

import (
	"backend/domain/models"
	"log"
	"sync"
)

type SearchRepository interface {
	RepositoryInterface
	GetICMPHistoryByTargets(targets []string, limit int) ([]models.ICMPHistoryRecord, error)
	GetICMPHistoryByID(id string) (*models.ICMPHistoryRecord, error)
	GetNmapTcpUdpHistoryByIP(ip string, limit int) ([]models.NmapTcpUdpHistoryRecord, error)
	GetNmapTcpUdpHistoryByID(id string) (*models.NmapTcpUdpHistoryRecord, error)
	GetNmapOsDetectionHistoryByIP(ip string, limit int) ([]models.NmapOsDetectionHistoryRecord, error)
	GetNmapOsDetectionHistoryByID(id string) (*models.NmapOsDetectionHistoryRecord, error)
	GetNmapHostDiscoveryHistoryByIP(ip string, limit int) ([]models.NmapHostDiscoveryHistoryRecord, error)
	GetNmapHostDiscoveryHistoryByID(id string) (*models.NmapHostDiscoveryHistoryRecord, error)
	GetARPHistoryByIPRange(ipRange string, limit int) ([]models.ARPHistoryRecord, error)
	GetARPHistoryByID(id string) (*models.ARPHistoryRecord, error)
	GetTCPHistoryByHostPort(host, port string, limit int) ([]models.TCPHistoryRecord, error)
	GetTCPHistoryByID(id string) (*models.TCPHistoryRecord, error)
}

type RepositoryInterface interface {
	SaveARPHistory(record *models.ARPHistoryRecord) error
	GetARPHistory(limit int) ([]models.ARPHistoryRecord, error)
	DeleteARPHistory() error

	SaveICMPHistory(record *models.ICMPHistoryRecord) error
	GetICMPHistory(limit int) ([]models.ICMPHistoryRecord, error)
	DeleteICMPHistory() error

	SaveNmapTcpUdpHistory(record *models.NmapTcpUdpHistoryRecord) error
	GetNmapTcpUdpHistory(limit int) ([]models.NmapTcpUdpHistoryRecord, error)
	DeleteNmapTcpUdpHistory() error

	SaveNmapOsDetectionHistory(record *models.NmapOsDetectionHistoryRecord) error
	GetNmapOsDetectionHistory(limit int) ([]models.NmapOsDetectionHistoryRecord, error)
	DeleteNmapOsDetectionHistory() error

	SaveNmapHostDiscoveryHistory(record *models.NmapHostDiscoveryHistoryRecord) error
	GetNmapHostDiscoveryHistory(limit int) ([]models.NmapHostDiscoveryHistoryRecord, error)
	DeleteNmapHostDiscoveryHistory() error

	SaveTCPHistory(record *models.TCPHistoryRecord) error
	GetTCPHistory(limit int) ([]models.TCPHistoryRecord, error)
	DeleteTCPHistory() error
}

type HistoryService struct {
	repo         RepositoryInterface
	requestCache map[string]interface{}
	mu           sync.RWMutex
}

func NewHistoryService(repo RepositoryInterface) *HistoryService {
	return &HistoryService{
		repo:         repo,
		requestCache: make(map[string]interface{}),
	}
}

func (hs *HistoryService) CacheRequest(taskID string, request interface{}) {
	hs.mu.Lock()
	defer hs.mu.Unlock()
	hs.requestCache[taskID] = request
}

func (hs *HistoryService) GetCachedRequest(taskID string) interface{} {
	hs.mu.RLock()
	defer hs.mu.RUnlock()
	return hs.requestCache[taskID]
}

func (hs *HistoryService) RemoveCachedRequest(taskID string) {
	hs.mu.Lock()
	defer hs.mu.Unlock()
	delete(hs.requestCache, taskID)
}

func (hs *HistoryService) GetRepo() RepositoryInterface {
	return hs.repo
}

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

	if ip == "" && result.Host != "" {
		ip = result.Host
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

	if ip == "" && result.Host != "" {
		ip = result.Host
	}
	host := result.Host
	if host == "" {
		host = ip
	}

	historyRecord := &models.NmapHostDiscoveryHistoryRecord{
		TaskID:    result.TaskID,
		IP:        ip,
		Host:      host,
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

func (hs *HistoryService) SaveTCPResponse(result models.TCPResponse) {
	if hs.repo == nil {
		return
	}

	var host, port string
	if cachedReq := hs.GetCachedRequest(result.TaskID); cachedReq != nil {
		if tcpReq, ok := cachedReq.(models.TCPRequest); ok {
			host = tcpReq.Host
			port = tcpReq.Port
		}
	}

	historyRecord := &models.TCPHistoryRecord{
		TaskID:       result.TaskID,
		Host:         host,
		Port:         port,
		HexObjectKey: result.HexObjectKey,
		DecodedText:  result.DecodedText,
		Status:       result.Status,
		Error:        result.Error,
	}

	if err := hs.repo.SaveTCPHistory(historyRecord); err != nil {
		log.Printf("Failed to save TCP history: %v", err)
	} else {
		log.Printf("Successfully saved TCP history for task %s", result.TaskID)
		hs.RemoveCachedRequest(result.TaskID)
	}
}
