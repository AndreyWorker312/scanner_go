package application

import (
	"backend/domain/models"
	rabbitmq "backend/internal/infrastructure"
	"encoding/json"
	"log"
	"sync"
)

type App struct {
	publisher *rabbitmq.RPCScannerPublisher
	repo      RepositoryInterface
	// Кэш для хранения параметров запросов
	requestCache map[string]interface{}
	mu           sync.RWMutex
}

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

func NewApp(publisher *rabbitmq.RPCScannerPublisher, repo RepositoryInterface) *App {
	return &App{
		publisher:    publisher,
		repo:         repo,
		requestCache: make(map[string]interface{}),
	}
}

// Методы для работы с кэшем запросов
func (a *App) cacheRequest(taskID string, request interface{}) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.requestCache[taskID] = request
}

func (a *App) getCachedRequest(taskID string) interface{} {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.requestCache[taskID]
}

func (a *App) removeCachedRequest(taskID string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.requestCache, taskID)
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
		return a.PublishNmapRequest(tcpUdpReq)

	case scanType.ScanMethod == "os_detection":
		var osReq models.NmapOsDetectionRequest
		if err := json.Unmarshal(optionsJSON, &osReq); err != nil {
			log.Printf("Failed to unmarshal OS detection request: %v", err)
			return &models.Response{
				TaskID: "error",
				Result: map[string]string{"error": "invalid OS detection request"},
			}
		}
		return a.PublishNmapRequest(osReq)

	case scanType.ScanMethod == "host_discovery":
		var hostReq models.NmapHostDiscoveryRequest
		if err := json.Unmarshal(optionsJSON, &hostReq); err != nil {
			log.Printf("Failed to unmarshal host discovery request: %v", err)
			return &models.Response{
				TaskID: "error",
				Result: map[string]string{"error": "invalid host discovery request"},
			}
		}
		return a.PublishNmapRequest(hostReq)

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
		return a.PublishNmapRequest(nmapReq)
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

	// Кэшируем запрос для последующего использования
	a.cacheRequest(arpReq.TaskID, arpReq)

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

	// Кэшируем запрос для последующего использования
	a.cacheRequest(icmpReq.TaskID, icmpReq)

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
func (a *App) PublishNmapRequest(req interface{}) *models.Response {
	log.Printf("Publishing Nmap request: %+v", req)

	// Кэшируем запрос для последующего использования
	switch r := req.(type) {
	case models.NmapTcpUdpRequest:
		a.cacheRequest(r.TaskID, r)
	case models.NmapOsDetectionRequest:
		a.cacheRequest(r.TaskID, r)
	case models.NmapHostDiscoveryRequest:
		a.cacheRequest(r.TaskID, r)
	}

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

// ==================== HISTORY SAVING METHODS ====================

func (a *App) SaveARPResult(response *models.ARPResponse, request *models.ARPRequest) {
	if a.repo == nil {
		return
	}

	historyRecord := &models.ARPHistoryRecord{
		TaskID:         response.TaskID,
		InterfaceName:  request.InterfaceName,
		IPRange:        request.IPRange,
		Status:         response.Status,
		Devices:        response.Devices,
		OnlineDevices:  response.OnlineDevices,
		OfflineDevices: response.OfflineDevices,
		TotalCount:     response.TotalCount,
		OnlineCount:    response.OnlineCount,
		OfflineCount:   response.OfflineCount,
		Error:          response.Error,
	}

	if err := a.repo.SaveARPHistory(historyRecord); err != nil {
		log.Printf("Failed to save ARP history: %v", err)
	}
}

func (a *App) SaveICMPResult(response *models.ICMPResponse, request *models.ICMPRequest) {
	if a.repo == nil {
		return
	}

	historyRecord := &models.ICMPHistoryRecord{
		TaskID:    response.TaskID,
		Targets:   request.Targets,
		PingCount: request.PingCount,
		Status:    response.Status,
		Results:   response.Results,
		Error:     response.Error,
	}

	if err := a.repo.SaveICMPHistory(historyRecord); err != nil {
		log.Printf("Failed to save ICMP history: %v", err)
	}
}

func (a *App) SaveNmapTcpUdpResult(response *models.NmapTcpUdpResponse, request *models.NmapTcpUdpRequest) {
	if a.repo == nil {
		return
	}

	historyRecord := &models.NmapTcpUdpHistoryRecord{
		TaskID:      response.TaskID,
		IP:          request.IP,
		ScannerType: request.ScannerType,
		Ports:       request.Ports,
		Host:        response.Host,
		PortInfo:    response.PortInfo,
		Status:      response.Status,
		Error:       response.Error,
	}

	if err := a.repo.SaveNmapTcpUdpHistory(historyRecord); err != nil {
		log.Printf("Failed to save Nmap TCP/UDP history: %v", err)
	}
}

func (a *App) SaveNmapOsDetectionResult(response *models.NmapOsDetectionResponse, request *models.NmapOsDetectionRequest) {
	if a.repo == nil {
		return
	}

	historyRecord := &models.NmapOsDetectionHistoryRecord{
		TaskID:   response.TaskID,
		IP:       request.IP,
		Host:     response.Host,
		Name:     response.Name,
		Accuracy: response.Accuracy,
		Vendor:   response.Vendor,
		Family:   response.Family,
		Type:     response.Type,
		Status:   response.Status,
		Error:    response.Error,
	}

	if err := a.repo.SaveNmapOsDetectionHistory(historyRecord); err != nil {
		log.Printf("Failed to save Nmap OS Detection history: %v", err)
	}
}

func (a *App) SaveNmapHostDiscoveryResult(response *models.NmapHostDiscoveryResponse, request *models.NmapHostDiscoveryRequest) {
	if a.repo == nil {
		return
	}

	historyRecord := &models.NmapHostDiscoveryHistoryRecord{
		TaskID:    response.TaskID,
		IP:        request.IP,
		Host:      response.Host,
		HostUP:    response.HostUP,
		HostTotal: response.HostTotal,
		Status:    response.Status,
		DNS:       response.DNS,
		Reason:    response.Reason,
		Error:     response.Error,
	}

	if err := a.repo.SaveNmapHostDiscoveryHistory(historyRecord); err != nil {
		log.Printf("Failed to save Nmap Host Discovery history: %v", err)
	}
}

// ProcessResponse обрабатывает ответ от сканера и сохраняет в базу данных
func (a *App) ProcessResponse(response *models.Response) {
	if a.repo == nil || response == nil {
		log.Printf("ProcessResponse: repo or response is nil")
		return
	}

	log.Printf("ProcessResponse: processing response for task %s", response.TaskID)
	log.Printf("ProcessResponse: response result type: %T", response.Result)

	// Определяем тип ответа по типу структуры
	switch result := response.Result.(type) {
	case models.ARPResponse:
		log.Printf("Processing ARP response")
		a.saveARPResponseFromStruct(result)
	case models.ICMPResponse:
		log.Printf("Processing ICMP response")
		a.saveICMPResponseFromStruct(result)
	case models.NmapTcpUdpResponse:
		log.Printf("Processing Nmap TCP/UDP response")
		a.saveNmapTcpUdpResponseFromStruct(result)
	case models.NmapOsDetectionResponse:
		log.Printf("Processing Nmap OS Detection response")
		a.saveNmapOsDetectionResponseFromStruct(result)
	case models.NmapHostDiscoveryResponse:
		log.Printf("Processing Nmap Host Discovery response")
		a.saveNmapHostDiscoveryResponseFromStruct(result)
	default:
		log.Printf("Unknown response type: %T", result)
	}
}

// Вспомогательные методы для сохранения каждого типа ответа
func (a *App) saveARPResponse(resultData map[string]interface{}) {
	// Извлекаем данные из map
	taskID, _ := resultData["task_id"].(string)
	status, _ := resultData["status"].(string)
	errorMsg, _ := resultData["error"].(string)

	// Извлекаем массивы устройств
	var devices []models.ARPDevice
	if devicesData, ok := resultData["devices"].([]interface{}); ok {
		for _, deviceData := range devicesData {
			if deviceMap, ok := deviceData.(map[string]interface{}); ok {
				device := models.ARPDevice{
					IP:     getString(deviceMap, "ip"),
					MAC:    getString(deviceMap, "mac"),
					Vendor: getString(deviceMap, "vendor"),
					Status: getString(deviceMap, "status"),
				}
				devices = append(devices, device)
			}
		}
	}

	// Разделяем на онлайн и оффлайн устройства
	var onlineDevices, offlineDevices []models.ARPDevice
	for _, device := range devices {
		if device.Status == "online" {
			onlineDevices = append(onlineDevices, device)
		} else {
			offlineDevices = append(offlineDevices, device)
		}
	}

	// Извлекаем счетчики
	totalCount := getInt(resultData, "total_count")
	onlineCount := getInt(resultData, "online_count")
	offlineCount := getInt(resultData, "offline_count")

	historyRecord := &models.ARPHistoryRecord{
		TaskID:         taskID,
		InterfaceName:  "unknown", // TODO: получить из кэша запросов
		IPRange:        "unknown", // TODO: получить из кэша запросов
		Status:         status,
		Devices:        devices,
		OnlineDevices:  onlineDevices,
		OfflineDevices: offlineDevices,
		TotalCount:     totalCount,
		OnlineCount:    onlineCount,
		OfflineCount:   offlineCount,
		Error:          errorMsg,
	}

	if err := a.repo.SaveARPHistory(historyRecord); err != nil {
		log.Printf("Failed to save ARP history: %v", err)
	}
}

func (a *App) saveICMPResponse(resultData map[string]interface{}) {
	taskID, _ := resultData["task_id"].(string)
	status, _ := resultData["status"].(string)
	errorMsg, _ := resultData["error"].(string)

	// Извлекаем результаты
	var results []models.ICMPResult
	if resultsData, ok := resultData["results"].([]interface{}); ok {
		for _, resultData := range resultsData {
			if resultMap, ok := resultData.(map[string]interface{}); ok {
				result := models.ICMPResult{
					Target:            getString(resultMap, "target"),
					Address:           getString(resultMap, "address"),
					PacketsSent:       getInt(resultMap, "packets_sent"),
					PacketsReceived:   getInt(resultMap, "packets_received"),
					PacketLossPercent: getFloat64(resultMap, "packet_loss_percent"),
					Error:             getString(resultMap, "error"),
				}
				results = append(results, result)
			}
		}
	}

	historyRecord := &models.ICMPHistoryRecord{
		TaskID:    taskID,
		Targets:   []string{}, // TODO: получить из кэша запросов
		PingCount: 0,          // TODO: получить из кэша запросов
		Status:    status,
		Results:   results,
		Error:     errorMsg,
	}

	if err := a.repo.SaveICMPHistory(historyRecord); err != nil {
		log.Printf("Failed to save ICMP history: %v", err)
	}
}

func (a *App) saveNmapTcpUdpResponse(resultData map[string]interface{}) {
	taskID, _ := resultData["task_id"].(string)
	host, _ := resultData["host"].(string)
	status, _ := resultData["status"].(string)
	errorMsg, _ := resultData["error"].(string)

	// Извлекаем информацию о портах
	var portInfo []models.NmapPortTcpUdpInfo
	if portInfoData, ok := resultData["port_info"].([]interface{}); ok {
		for _, infoData := range portInfoData {
			if infoMap, ok := infoData.(map[string]interface{}); ok {
				info := models.NmapPortTcpUdpInfo{
					Status:      getString(infoMap, "status"),
					AllPorts:    getUint16Slice(infoMap, "all_ports"),
					Protocols:   getStringSlice(infoMap, "protocols"),
					State:       getStringSlice(infoMap, "state"),
					ServiceName: getStringSlice(infoMap, "service_name"),
				}
				portInfo = append(portInfo, info)
			}
		}
	}

	historyRecord := &models.NmapTcpUdpHistoryRecord{
		TaskID:      taskID,
		IP:          "unknown", // TODO: получить из кэша запросов
		ScannerType: "unknown", // TODO: получить из кэша запросов
		Ports:       "unknown", // TODO: получить из кэша запросов
		Host:        host,
		PortInfo:    portInfo,
		Status:      status,
		Error:       errorMsg,
	}

	if err := a.repo.SaveNmapTcpUdpHistory(historyRecord); err != nil {
		log.Printf("Failed to save Nmap TCP/UDP history: %v", err)
	}
}

func (a *App) saveNmapOsDetectionResponse(resultData map[string]interface{}) {
	taskID, _ := resultData["task_id"].(string)
	host, _ := resultData["host"].(string)
	name, _ := resultData["name"].(string)
	accuracy := getInt(resultData, "accuracy")
	vendor, _ := resultData["vendor"].(string)
	family, _ := resultData["family"].(string)
	typeStr, _ := resultData["type"].(string)
	status, _ := resultData["status"].(string)
	errorMsg, _ := resultData["error"].(string)

	historyRecord := &models.NmapOsDetectionHistoryRecord{
		TaskID:   taskID,
		IP:       "unknown", // TODO: получить из кэша запросов
		Host:     host,
		Name:     name,
		Accuracy: accuracy,
		Vendor:   vendor,
		Family:   family,
		Type:     typeStr,
		Status:   status,
		Error:    errorMsg,
	}

	if err := a.repo.SaveNmapOsDetectionHistory(historyRecord); err != nil {
		log.Printf("Failed to save Nmap OS Detection history: %v", err)
	}
}

func (a *App) saveNmapHostDiscoveryResponse(resultData map[string]interface{}) {
	taskID, _ := resultData["task_id"].(string)
	host, _ := resultData["host"].(string)
	hostUp := getInt(resultData, "host_up")
	hostTotal := getInt(resultData, "host_total")
	status, _ := resultData["status"].(string)
	dns, _ := resultData["dns"].(string)
	reason, _ := resultData["reason"].(string)
	errorMsg, _ := resultData["error"].(string)

	historyRecord := &models.NmapHostDiscoveryHistoryRecord{
		TaskID:    taskID,
		IP:        "unknown", // TODO: получить из кэша запросов
		Host:      host,
		HostUP:    hostUp,
		HostTotal: hostTotal,
		Status:    status,
		DNS:       dns,
		Reason:    reason,
		Error:     errorMsg,
	}

	if err := a.repo.SaveNmapHostDiscoveryHistory(historyRecord); err != nil {
		log.Printf("Failed to save Nmap Host Discovery history: %v", err)
	}
}

// Вспомогательные функции для извлечения данных из map
func getString(data map[string]interface{}, key string) string {
	if value, ok := data[key].(string); ok {
		return value
	}
	return ""
}

func getInt(data map[string]interface{}, key string) int {
	if value, ok := data[key].(int); ok {
		return value
	}
	if value, ok := data[key].(float64); ok {
		return int(value)
	}
	return 0
}

func getFloat64(data map[string]interface{}, key string) float64 {
	if value, ok := data[key].(float64); ok {
		return value
	}
	if value, ok := data[key].(int); ok {
		return float64(value)
	}
	return 0.0
}

func getStringSlice(data map[string]interface{}, key string) []string {
	if value, ok := data[key].([]interface{}); ok {
		var result []string
		for _, v := range value {
			if str, ok := v.(string); ok {
				result = append(result, str)
			}
		}
		return result
	}
	return []string{}
}

func getUint16Slice(data map[string]interface{}, key string) []uint16 {
	if value, ok := data[key].([]interface{}); ok {
		var result []uint16
		for _, v := range value {
			if num, ok := v.(float64); ok {
				result = append(result, uint16(num))
			}
		}
		return result
	}
	return []uint16{}
}

func getMapKeys(data map[string]interface{}) []string {
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	return keys
}

// Новые методы для работы со структурами Go
func (a *App) saveARPResponseFromStruct(result models.ARPResponse) {
	// Получаем кэшированный запрос
	var interfaceName, ipRange string
	if cachedReq := a.getCachedRequest(result.TaskID); cachedReq != nil {
		if arpReq, ok := cachedReq.(models.ARPRequest); ok {
			interfaceName = arpReq.InterfaceName
			ipRange = arpReq.IPRange
		}
	}

	// Разделяем на онлайн и оффлайн устройства
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

	if err := a.repo.SaveARPHistory(historyRecord); err != nil {
		log.Printf("Failed to save ARP history: %v", err)
	} else {
		log.Printf("Successfully saved ARP history for task %s", result.TaskID)
		// Удаляем из кэша после успешного сохранения
		a.removeCachedRequest(result.TaskID)
	}
}

func (a *App) saveICMPResponseFromStruct(result models.ICMPResponse) {
	// Получаем кэшированный запрос
	var targets []string
	var pingCount int
	if cachedReq := a.getCachedRequest(result.TaskID); cachedReq != nil {
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

	if err := a.repo.SaveICMPHistory(historyRecord); err != nil {
		log.Printf("Failed to save ICMP history: %v", err)
	} else {
		log.Printf("Successfully saved ICMP history for task %s", result.TaskID)
		// Удаляем из кэша после успешного сохранения
		a.removeCachedRequest(result.TaskID)
	}
}

func (a *App) saveNmapTcpUdpResponseFromStruct(result models.NmapTcpUdpResponse) {
	// Получаем кэшированный запрос
	var ip, scannerType, ports string
	if cachedReq := a.getCachedRequest(result.TaskID); cachedReq != nil {
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

	if err := a.repo.SaveNmapTcpUdpHistory(historyRecord); err != nil {
		log.Printf("Failed to save Nmap TCP/UDP history: %v", err)
	} else {
		log.Printf("Successfully saved Nmap TCP/UDP history for task %s", result.TaskID)
		// Удаляем из кэша после успешного сохранения
		a.removeCachedRequest(result.TaskID)
	}
}

func (a *App) saveNmapOsDetectionResponseFromStruct(result models.NmapOsDetectionResponse) {
	// Получаем кэшированный запрос
	var ip string
	if cachedReq := a.getCachedRequest(result.TaskID); cachedReq != nil {
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

	if err := a.repo.SaveNmapOsDetectionHistory(historyRecord); err != nil {
		log.Printf("Failed to save Nmap OS Detection history: %v", err)
	} else {
		log.Printf("Successfully saved Nmap OS Detection history for task %s", result.TaskID)
		// Удаляем из кэша после успешного сохранения
		a.removeCachedRequest(result.TaskID)
	}
}

func (a *App) saveNmapHostDiscoveryResponseFromStruct(result models.NmapHostDiscoveryResponse) {
	// Получаем кэшированный запрос
	var ip string
	if cachedReq := a.getCachedRequest(result.TaskID); cachedReq != nil {
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

	if err := a.repo.SaveNmapHostDiscoveryHistory(historyRecord); err != nil {
		log.Printf("Failed to save Nmap Host Discovery history: %v", err)
	} else {
		log.Printf("Successfully saved Nmap Host Discovery history for task %s", result.TaskID)
		// Удаляем из кэша после успешного сохранения
		a.removeCachedRequest(result.TaskID)
	}
}
