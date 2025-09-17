package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	wb "backend/api/websocket"
	"backend/domain/models"
	"backend/usecases"
	"github.com/gorilla/websocket"
)

type ScanHandler struct {
	scanUseCases *usecases.ScanUseCases
	wsHub        *wb.Hub
	upgrader     websocket.Upgrader
}

func NewScanHandler(scanUseCases *usecases.ScanUseCases, wsHub *wb.Hub) *ScanHandler {
	return &ScanHandler{
		scanUseCases: scanUseCases,
		wsHub:        wsHub,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // В продакшене нужно ограничить домены
			},
		},
	}
}

func (h *ScanHandler) StartARPScan(w http.ResponseWriter, r *http.Request) {
	var request models.ARPRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid ARP request: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Валидация
	if request.InterfaceName == "" {
		http.Error(w, "InterfaceName is required", http.StatusBadRequest)
		return
	}
	if request.IPRange == "" {
		http.Error(w, "IPRange is required", http.StatusBadRequest)
		return
	}
	if request.TaskID == "" {
		request.TaskID = generateTaskID("arp")
	}

	scan, err := h.scanUseCases.StartScan(r.Context(), models.ScanTypeARP, &request)
	if err != nil {
		http.Error(w, "Failed to start ARP scan: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"task_id":    scan.TaskID,
		"status":     scan.Status,
		"type":       scan.Type,
		"created_at": scan.CreatedAt,
		"ws_url":     fmt.Sprintf("ws://%s/ws?task_id=%s", r.Host, scan.TaskID),
	})
}

func (h *ScanHandler) StartICMPScan(w http.ResponseWriter, r *http.Request) {
	var request models.ICMPRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid ICMP request: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Валидация
	if len(request.Targets) == 0 {
		http.Error(w, "Targets are required", http.StatusBadRequest)
		return
	}
	if request.PingCount <= 0 {
		request.PingCount = 4 // default value
	}
	if request.TaskID == "" {
		request.TaskID = generateTaskID("icmp")
	}

	scan, err := h.scanUseCases.StartScan(r.Context(), models.ScanTypeICMP, &request)
	if err != nil {
		http.Error(w, "Failed to start ICMP scan: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"task_id":    scan.TaskID,
		"status":     scan.Status,
		"type":       scan.Type,
		"created_at": scan.CreatedAt,
		"ws_url":     fmt.Sprintf("ws://%s/ws?task_id=%s", r.Host, scan.TaskID),
	})
}

func (h *ScanHandler) StartNmapTcpUdpScan(w http.ResponseWriter, r *http.Request) {
	var request models.NmapTcpUdpRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid Nmap TCP/UDP request: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Валидация
	if request.IP == "" {
		http.Error(w, "IP is required", http.StatusBadRequest)
		return
	}
	if request.ScannerType != "TCP" && request.ScannerType != "UDP" {
		http.Error(w, "ScannerType must be 'TCP' or 'UDP'", http.StatusBadRequest)
		return
	}
	if request.Ports == "" {
		request.Ports = "1-1000" // default ports range
	}
	if request.TaskID == "" {
		request.TaskID = generateTaskID("nmap-tcpudp")
	}

	scan, err := h.scanUseCases.StartScan(r.Context(), models.ScanTypeNMAP, &request)
	if err != nil {
		http.Error(w, "Failed to start Nmap TCP/UDP scan: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"task_id":    scan.TaskID,
		"status":     scan.Status,
		"type":       scan.Type,
		"created_at": scan.CreatedAt,
		"ws_url":     fmt.Sprintf("ws://%s/ws?task_id=%s", r.Host, scan.TaskID),
	})
}

func (h *ScanHandler) StartNmapOsDetection(w http.ResponseWriter, r *http.Request) {
	var request models.NmapOsDetectionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid Nmap OS Detection request: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Валидация
	if request.IP == "" {
		http.Error(w, "IP is required", http.StatusBadRequest)
		return
	}
	if request.TaskID == "" {
		request.TaskID = generateTaskID("nmap-os")
	}

	scan, err := h.scanUseCases.StartScan(r.Context(), models.ScanTypeNMAP, &request)
	if err != nil {
		http.Error(w, "Failed to start Nmap OS Detection: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"task_id":    scan.TaskID,
		"status":     scan.Status,
		"type":       scan.Type,
		"created_at": scan.CreatedAt,
		"ws_url":     fmt.Sprintf("ws://%s/ws?task_id=%s", r.Host, scan.TaskID),
	})
}

func (h *ScanHandler) StartNmapHostDiscovery(w http.ResponseWriter, r *http.Request) {
	var request models.NmapHostDiscoveryRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid Nmap Host Discovery request: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Валидация
	if request.IP == "" {
		http.Error(w, "IP is required", http.StatusBadRequest)
		return
	}
	if request.TaskID == "" {
		request.TaskID = generateTaskID("nmap-host")
	}

	scan, err := h.scanUseCases.StartScan(r.Context(), models.ScanTypeNMAP, &request)
	if err != nil {
		http.Error(w, "Failed to start Nmap Host Discovery: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"task_id":    scan.TaskID,
		"status":     scan.Status,
		"type":       scan.Type,
		"created_at": scan.CreatedAt,
		"ws_url":     fmt.Sprintf("ws://%s/ws?task_id=%s", r.Host, scan.TaskID),
	})
}

func (h *ScanHandler) GetScanStatus(w http.ResponseWriter, r *http.Request) {
	taskID := r.URL.Path[len("/api/scans/"):]
	if taskID == "" {
		http.Error(w, "Task ID is required", http.StatusBadRequest)
		return
	}

	scan, responses, err := h.scanUseCases.GetScanStatus(r.Context(), taskID)
	if err != nil {
		http.Error(w, "Failed to get scan status: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"scan":      scan,
		"responses": responses,
	})
}

func (h *ScanHandler) ListScans(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	scans, err := h.scanUseCases.ListScans(r.Context(), limit, offset)
	if err != nil {
		http.Error(w, "Failed to list scans: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"scans":  scans,
		"limit":  limit,
		"offset": offset,
		"count":  len(scans),
	})
}

func (h *ScanHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	taskID := r.URL.Query().Get("task_id")
	if taskID == "" {
		http.Error(w, "task_id parameter is required", http.StatusBadRequest)
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "WebSocket upgrade failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	client := &wb.Client{
		Hub:     h.wsHub,
		Conn:    conn,
		Send:    make(chan []byte, 256),
		TaskIDs: map[string]bool{taskID: true},
	}

	// Регистрируем клиента
	client.Hub.Register <- client

	// Запускаем goroutine для чтения сообщений
	go client.ReadPump()

	// Запускаем goroutine для записи сообщений
	go client.WritePump()

	// Отправляем текущий статус сканирования при подключении
	scan, responses, err := h.scanUseCases.GetScanStatus(r.Context(), taskID)
	if err == nil {
		initialMessage := map[string]interface{}{
			"type":      "initial_status",
			"task_id":   taskID,
			"scan":      scan,
			"responses": responses,
			"timestamp": time.Now(),
		}

		messageBytes, _ := json.Marshal(initialMessage)
		conn.WriteMessage(websocket.TextMessage, messageBytes)
	}
}

// Вспомогательная функция для генерации task_id
func generateTaskID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}
