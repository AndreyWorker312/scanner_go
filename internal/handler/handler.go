package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"network-scanner/internal/models"
	"network-scanner/internal/repository"
	"network-scanner/internal/scanner"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Handler структура для обработчиков WebSocket запросов
type Handler struct {
	logger      Logger
	repo        repository.Repository
	portScanner scanner.PortScanner
	wsManager   *WSManager
}

// WSManager управляет WebSocket соединениями
type WSManager struct {
	clients map[*websocket.Conn]bool
	mu      sync.Mutex
}

// NewWSManager создает новый менеджер WebSocket
func NewWSManager() *WSManager {
	return &WSManager{
		clients: make(map[*websocket.Conn]bool),
	}
}

// Logger интерфейс для логгирования
type Logger interface {
	Info(msg string)
	Infof(format string, args ...interface{})
	Error(msg string)
	Errorf(format string, args ...interface{})
	Fatal(msg string)
	Fatalf(format string, args ...interface{})
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// NewHandler создает новый экземпляр Handler
func NewHandler(logger Logger, repo repository.Repository, scanner scanner.PortScanner) *Handler {
	return &Handler{
		logger:      logger,
		repo:        repo,
		portScanner: scanner,
		wsManager:   NewWSManager(),
	}
}

// InitRoutes инициализирует маршруты API — только WebSocket
func (h *Handler) InitRoutes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", h.handleWebSocket)
	return mux
}

// AddClient добавляет нового WebSocket клиента
func (m *WSManager) AddClient(conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clients[conn] = true
}

// RemoveClient удаляет WebSocket клиента
func (m *WSManager) RemoveClient(conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.clients, conn)
}

// Broadcast отправляет сообщение всем подключенным клиентам
func (m *WSManager) Broadcast(message interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for conn := range m.clients {
		// Устанавливаем таймаут для записи
		conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

		if err := conn.WriteJSON(message); err != nil {
			log.Printf("WebSocket write error: %v", err)
			conn.Close()
			delete(m.clients, conn)
		}
	}
}

func (h *Handler) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Errorf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// Регистрируем соединение
	h.wsManager.AddClient(conn)
	defer h.wsManager.RemoveClient(conn)

	// Настройка обработчиков
	conn.SetPingHandler(func(appData string) error {
		h.logger.Info("Received ping")
		err := conn.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(time.Second))
		if err != nil {
			h.logger.Errorf("Failed to send pong: %v", err)
		}
		return err
	})

	conn.SetCloseHandler(func(code int, text string) error {
		h.logger.Infof("Connection closed: %d %s", code, text)
		return nil
	})

	// Приветственное сообщение
	if err := conn.WriteJSON(map[string]string{
		"type":    "welcome",
		"message": "Connected to WebSocket server",
	}); err != nil {
		h.logger.Errorf("Failed to send welcome message: %v", err)
		return
	}

	// Главный цикл обработки сообщений
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
		defer cancel()

		switch msg.Action {
		case "scan":
			h.handleScanWS(ctx, conn, msg.Data)
		case "history":
			h.handleHistoryWS(ctx, conn)
		case "get":
			h.handleGetScanByIDWS(ctx, conn, msg.Data)
		case "ping":
			conn.WriteJSON(map[string]string{"type": "pong"})
		default:
			h.sendErrorWS(conn, "Unknown action: "+msg.Action)
		}
	}
}

// sendErrorWS отправляет ошибку клиенту по WebSocket
func (h *Handler) sendErrorWS(conn *websocket.Conn, message string) {
	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	conn.WriteJSON(map[string]interface{}{
		"type":    "error",
		"message": message,
		"time":    time.Now().Format(time.RFC3339),
	})
}

// handleScanWS обрабатывает команду сканирования портов
func (h *Handler) handleScanWS(ctx context.Context, conn *websocket.Conn, data json.RawMessage) {
	var req struct {
		IP    string `json:"ip"`
		Ports string `json:"ports"`
	}
	if err := json.Unmarshal(data, &req); err != nil {
		h.sendErrorWS(conn, "Invalid scan request data")
		return
	}

	// Немедленно отвечаем клиенту
	conn.WriteJSON(map[string]interface{}{
		"type":  "scan_started",
		"ip":    req.IP,
		"ports": req.Ports,
	})
	if req.IP == "" {
		h.sendErrorWS(conn, "IP address is required")
		return
	}
	if req.Ports == "" {
		req.Ports = "1-1024"
	}

	// Отправляем подтверждение о начале сканирования
	conn.WriteJSON(map[string]interface{}{
		"type":  "scan_started",
		"ip":    req.IP,
		"ports": req.Ports,
	})

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
		"type":       "scan_result",
		"request_id": requestID,
		"ip":         req.IP,
		"ports":      req.Ports,
		"open_ports": openPorts,
		"time":       time.Now().Format(time.RFC3339),
	}

	// Отправляем результат клиенту
	conn.WriteJSON(response)

	// Рассылаем уведомление всем клиентам о завершении сканирования
	h.wsManager.Broadcast(response)
}

// handleHistoryWS обрабатывает команду получения истории сканирований
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

	conn.WriteJSON(map[string]interface{}{
		"type": "history_result",
		"data": response,
	})
}

// handleGetScanByIDWS обрабатывает команду получения результатов сканирования по ID
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

	conn.WriteJSON(response)
}
