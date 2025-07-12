package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"network-scanner/internal/scanner"
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

type Logger interface {
	Info(msg string)
	Infof(format string, args ...interface{})
	Error(msg string)
	Errorf(format string, args ...interface{})
	Fatal(msg string)
	Fatalf(format string, args ...interface{})
}

type WSManager struct {
	clients map[*websocket.Conn]bool
	mu      sync.Mutex
	logger  Logger
}

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

func (m *WSManager) Broadcast(messageType string, data interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for conn := range m.clients {
		conn.SetWriteDeadline(time.Now().Add(writeWait))
		if err := conn.WriteJSON(map[string]interface{}{
			"type": messageType,
			"data": data,
		}); err != nil {
			m.logger.Errorf("WebSocket write error: %v", err)
			conn.Close()
			delete(m.clients, conn)
		}
	}
}

type Handler struct {
	logger      Logger
	portScanner scanner.PortScanner
	wsManager   *WSManager
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewHandler(logger Logger, scanner scanner.PortScanner) *Handler {
	return &Handler{
		logger:      logger,
		portScanner: scanner,
		wsManager:   NewWSManager(logger),
	}
}

func (h *Handler) InitRoutes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", h.handleWebSocket)
	return mux
}

func (h *Handler) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Errorf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	h.logger.Info("New WebSocket connection")
	h.wsManager.AddClient(conn)
	defer h.wsManager.RemoveClient(conn)

	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// Отправляем приветственное сообщение
	h.sendWSMessage(conn, "welcome", map[string]string{
		"message": "Connected to Port Scanner",
	})

	// Обработка сообщений
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.logger.Errorf("WebSocket error: %v", err)
			}
			break
		}

		var request struct {
			Action string          `json:"action"`
			Data   json.RawMessage `json:"data"`
		}

		if err := json.Unmarshal(msg, &request); err != nil {
			h.sendErrorWS(conn, "Invalid message format")
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		switch request.Action {
		case "scan":
			h.handleScanRequest(ctx, conn, request.Data)
		case "ping":
			h.sendWSMessage(conn, "pong", nil)
		default:
			h.sendErrorWS(conn, "Unknown action: "+request.Action)
		}
	}
}

func (h *Handler) handleScanRequest(ctx context.Context, conn *websocket.Conn, data json.RawMessage) {
	var req struct {
		IP    string `json:"ip"`
		Ports string `json:"ports"`
	}

	if err := json.Unmarshal(data, &req); err != nil {
		h.sendErrorWS(conn, "Invalid scan request")
		return
	}

	if req.IP == "" {
		h.sendErrorWS(conn, "IP address is required")
		return
	}

	if req.Ports == "" {
		req.Ports = "1-1024" // Default ports range
	}

	h.logger.Infof("Starting scan for %s on ports %s", req.IP, req.Ports)
	h.sendWSMessage(conn, "scan_started", map[string]string{
		"ip":    req.IP,
		"ports": req.Ports,
	})

	// Выполняем сканирование портов
	openPorts, err := h.portScanner.ScanPorts(ctx, req.IP, req.Ports)
	if err != nil {
		h.logger.Errorf("Scan failed: %v", err)
		h.sendErrorWS(conn, "Scan failed: "+err.Error())
		return
	}

	// Отправляем результаты
	h.sendWSMessage(conn, "scan_result", map[string]interface{}{
		"ip":         req.IP,
		"ports":      req.Ports,
		"open_ports": openPorts,
		"timestamp":  time.Now().Format(time.RFC3339),
	})

	h.logger.Infof("Scan completed for %s. Open ports: %v", req.IP, openPorts)
}

func (h *Handler) sendWSMessage(conn *websocket.Conn, msgType string, data interface{}) error {
	conn.SetWriteDeadline(time.Now().Add(writeWait))
	return conn.WriteJSON(map[string]interface{}{
		"type": msgType,
		"data": data,
	})
}

func (h *Handler) sendErrorWS(conn *websocket.Conn, message string) {
	h.sendWSMessage(conn, "error", map[string]string{
		"message": message,
	})
}
