package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"network-scanner/internal/models"
	"network-scanner/internal/repository"
	"network-scanner/internal/scanner"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
)

// Logger интерфейс для логирования
type Logger interface {
	Info(msg string)
	Infof(format string, args ...interface{})
	Error(msg string)
	Errorf(format string, args ...interface{})
	Fatal(msg string)
	Fatalf(format string, args ...interface{})
}

// WSManager управляет WebSocket клиентами
type WSManager struct {
	clients map[*websocket.Conn]bool
	mu      sync.Mutex
	logger  Logger
}

// NewWSManager создаёт WSManager с логгером
func NewWSManager(logger Logger) *WSManager {
	return &WSManager{
		clients: make(map[*websocket.Conn]bool),
		logger:  logger,
	}
}

func (m *WSManager) AddClient(conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clients[conn] = true
}

func (m *WSManager) RemoveClient(conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.clients, conn)
}

// Broadcast отправляет сообщение всем клиентам с логированием ошибок
func (m *WSManager) Broadcast(message interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for conn := range m.clients {
		conn.SetWriteDeadline(time.Now().Add(writeWait))
		if err := conn.WriteJSON(message); err != nil {
			m.logger.Errorf("WebSocket write error: %v", err)
			conn.Close()
			delete(m.clients, conn)
		}
	}
}

// Handler — основной обработчик WebSocket
type Handler struct {
	logger      Logger
	repo        repository.Repository
	portScanner scanner.PortScanner
	wsManager   *WSManager
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true // В продакшене лучше ограничить
	},
}

// NewHandler создаёт новый Handler
func NewHandler(logger Logger, repo repository.Repository, scanner scanner.PortScanner) *Handler {
	return &Handler{
		logger:      logger,
		repo:        repo,
		portScanner: scanner,
		wsManager:   NewWSManager(logger),
	}
}

func (h *Handler) InitRoutes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		h.logger.Info("HTTP request received on /ws")
		h.handleWebSocket(w, r)
	})
	return loggingMiddleware(mux, h.logger)
}
func loggingMiddleware(next http.Handler, logger Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Infof("HTTP %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// handleWebSocket обрабатывает WebSocket соединение
func (h *Handler) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Errorf("WebSocket upgrade failed: %v", err)
		return
	}
	h.logger.Info("WebSocket connection established")
	defer func() {
		h.logger.Info("WebSocket connection closed")
		conn.Close()
	}()

	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	h.wsManager.AddClient(conn)
	defer h.wsManager.RemoveClient(conn)

	// Отправляем приветственное сообщение
	if err := h.sendWSMessage(conn, "welcome", map[string]string{
		"message": "Connected to WebSocket server",
		"version": "1.0",
	}); err != nil {
		h.logger.Errorf("Failed to send welcome message: %v", err)
		return
	}

	// Запускаем пинг для поддержания соединения
	pingTicker := time.NewTicker(pingPeriod)
	defer pingTicker.Stop()

	go func() {
		for range pingTicker.C {
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}()

	// Основной цикл обработки сообщений
	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.logger.Errorf("WebSocket error: %v", err)
			}
			break
		}

		var msg struct {
			Action string          `json:"action"`
			Data   json.RawMessage `json:"data"`
		}

		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			h.sendErrorWS(conn, "Invalid message format")
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

		switch msg.Action {
		case "scan":
			h.handleScanWS(ctx, conn, msg.Data)
		case "history":
			h.handleHistoryWS(ctx, conn)
		case "get":
			h.handleGetScanByIDWS(ctx, conn, msg.Data)
		case "ping":
			h.sendWSMessage(conn, "pong", nil)
		default:
			h.sendErrorWS(conn, "Unknown action: "+msg.Action)
		}

		cancel()
	}
}

// sendWSMessage отправляет JSON сообщение
func (h *Handler) sendWSMessage(conn *websocket.Conn, msgType string, data interface{}) error {
	conn.SetWriteDeadline(time.Now().Add(writeWait))
	return conn.WriteJSON(map[string]interface{}{
		"type": msgType,
		"data": data,
	})
}

// sendErrorWS отправляет ошибку клиенту
func (h *Handler) sendErrorWS(conn *websocket.Conn, message string) {
	conn.SetWriteDeadline(time.Now().Add(writeWait))
	_ = conn.WriteJSON(map[string]interface{}{
		"type":    "error",
		"message": message,
		"time":    time.Now().Format(time.RFC3339),
	})
}

// handleScanWS обрабатывает команду сканирования
func (h *Handler) handleScanWS(ctx context.Context, conn *websocket.Conn, data json.RawMessage) {
	var req struct {
		IP    string `json:"ip"`
		Ports string `json:"ports"`
	}
	if err := json.Unmarshal(data, &req); err != nil {
		h.sendErrorWS(conn, "Invalid scan request data")
		return
	}

	if req.IP == "" {
		h.sendErrorWS(conn, "IP address is required")
		return
	}
	if req.Ports == "" {
		req.Ports = "1-1024"
	}

	if err := h.sendWSMessage(conn, "scan_started", map[string]string{
		"ip":    req.IP,
		"ports": req.Ports,
	}); err != nil {
		h.logger.Errorf("Failed to send scan_started: %v", err)
		return
	}

	scanReq := models.ScanRequest{
		IPAddress: req.IP,
		Ports:     req.Ports,
		CreatedAt: time.Now(),
	}

	requestID, err := h.repo.SaveScanRequest(ctx, &scanReq)
	if err != nil {
		h.logger.Errorf("Failed to save scan request: %v", err)
		h.sendErrorWS(conn, "Internal server error")
		return
	}

	openPorts, err := h.portScanner.ScanPorts(ctx, req.IP, req.Ports)
	if err != nil {
		h.logger.Errorf("Failed to scan ports: %v", err)
		h.sendErrorWS(conn, "Failed to scan ports")
		return
	}

	var scanResults []*models.ScanResult
	for _, port := range openPorts {
		scanResults = append(scanResults, &models.ScanResult{
			RequestID: requestID,
			Port:      port,
			IsOpen:    true,
			ScannedAt: time.Now(),
		})
	}

	if err := h.repo.SaveScanResults(ctx, scanResults); err != nil {
		h.logger.Errorf("Failed to save scan results: %v", err)
		h.sendErrorWS(conn, "Internal server error")
		return
	}

	response := map[string]interface{}{
		"request_id": requestID,
		"ip":         req.IP,
		"ports":      req.Ports,
		"open_ports": openPorts,
		"time":       time.Now().Format(time.RFC3339),
	}

	if err := h.sendWSMessage(conn, "scan_result", response); err != nil {
		h.logger.Errorf("Failed to send scan_result: %v", err)
	}

	h.wsManager.Broadcast(response)
}

// handleHistoryWS обрабатывает запрос истории
func (h *Handler) handleHistoryWS(ctx context.Context, conn *websocket.Conn) {
	history, err := h.repo.GetScanHistory(ctx)
	if err != nil {
		h.logger.Errorf("Failed to get scan history: %v", err)
		h.sendErrorWS(conn, "Internal server error")
		return
	}

	var response []map[string]interface{}
	for _, scan := range history {
		if scan.Request == nil {
			continue
		}
		scanData := map[string]interface{}{
			"id":         scan.Request.ID,
			"ip_address": scan.Request.IPAddress,
			"ports":      scan.Request.Ports,
			"created_at": scan.Request.CreatedAt.Format(time.RFC3339),
		}
		response = append(response, scanData)
	}

	if err := h.sendWSMessage(conn, "history_result", response); err != nil {
		h.logger.Errorf("Failed to send history_result: %v", err)
	}
}

// handleGetScanByIDWS обрабатывает запрос деталей сканирования по ID
func (h *Handler) handleGetScanByIDWS(ctx context.Context, conn *websocket.Conn, data json.RawMessage) {
	var req struct {
		ID interface{} `json:"id"`
	}
	if err := json.Unmarshal(data, &req); err != nil {
		h.sendErrorWS(conn, "Invalid request data")
		return
	}

	var requestID int64
	switch v := req.ID.(type) {
	case float64:
		requestID = int64(v)
	case string:
		idParsed, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			h.sendErrorWS(conn, "Invalid scan ID format")
			return
		}
		requestID = idParsed
	default:
		h.sendErrorWS(conn, "Invalid scan ID type")
		return
	}

	scanResponse, err := h.repo.GetScanResults(ctx, requestID)
	if err != nil {
		if err == repository.ErrNotFound {
			h.sendErrorWS(conn, "Scan not found")
		} else {
			h.logger.Errorf("Failed to get scan results: %v", err)
			h.sendErrorWS(conn, "Internal server error")
		}
		return
	}

	if scanResponse.Request == nil {
		h.sendErrorWS(conn, "Scan request data missing")
		return
	}

	response := map[string]interface{}{
		"type":       "scan_details",
		"id":         scanResponse.Request.ID,
		"ip_address": scanResponse.Request.IPAddress,
		"ports":      scanResponse.Request.Ports,
		"created_at": scanResponse.Request.CreatedAt.Format(time.RFC3339),
		"open_ports": scanResponse.OpenPorts,
	}

	if err := h.sendWSMessage(conn, "scan_details", response); err != nil {
		h.logger.Errorf("Failed to send scan_details: %v", err)
	}
}
