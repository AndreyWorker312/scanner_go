package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"backend/pkg/logger"
	"backend/pkg/queue"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
	rpcTimeout     = 30 * time.Second
)

type WSManager struct {
	clients   map[*websocket.Conn]bool
	taskConns map[string]*websocket.Conn
	mu        sync.Mutex
	logger    logger.LoggerInterface
}

func NewWSManager(logger logger.LoggerInterface) *WSManager {
	return &WSManager{
		clients:   make(map[*websocket.Conn]bool),
		taskConns: make(map[string]*websocket.Conn),
		logger:    logger,
	}
}

func (m *WSManager) AddClient(conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clients[conn] = true
	m.logger.Infof("Client added: %v, total clients: %d", conn.RemoteAddr(), len(m.clients))
}

func (m *WSManager) RemoveClient(conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.clients, conn)
	m.logger.Infof("Client removed: %v, total clients: %d", conn.RemoteAddr(), len(m.clients))
}

func (m *WSManager) RegisterTask(taskID string, conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.taskConns[taskID] = conn
	m.logger.Infof("Task registered: %s for client: %v", taskID, conn.RemoteAddr())
}

func (m *WSManager) UnregisterTask(taskID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.taskConns, taskID)
	m.logger.Infof("Task unregistered: %s", taskID)
}

func (m *WSManager) GetConnForTask(taskID string) (*websocket.Conn, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	conn, ok := m.taskConns[taskID]
	return conn, ok
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
			m.logger.Errorf("WebSocket write error to %v: %v", conn.RemoteAddr(), err)
			conn.Close()
			delete(m.clients, conn)
			m.logger.Infof("Client removed due to write error: %v, total clients: %d", conn.RemoteAddr(), len(m.clients))
		}
	}
}

type Handler struct {
	logger    logger.LoggerInterface
	queue     *queue.RabbitMQ
	wsManager *WSManager
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewHandler(logger logger.LoggerInterface, queue *queue.RabbitMQ) *Handler {
	return &Handler{
		logger:    logger,
		queue:     queue,
		wsManager: NewWSManager(logger),
	}
}

func (h *Handler) InitRoutes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", h.handleWebSocket)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "frontend/static/index.html")
	})
	return h.loggingMiddleware(mux)
}

func (h *Handler) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.logger.Infof("HTTP %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	h.logger.Infof("handleWebSocket called from %s", r.RemoteAddr)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Errorf("WebSocket upgrade failed: %v", err)
		return
	}
	defer func() {
		h.logger.Infof("WebSocket connection closed: %v", conn.RemoteAddr())
		conn.Close()
	}()

	h.wsManager.AddClient(conn)
	defer h.wsManager.RemoveClient(conn)

	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	if err := h.sendWSMessage(conn, "welcome", map[string]string{
		"message": "Connected to Port Scanner",
	}); err != nil {
		h.logger.Errorf("Failed to send welcome message: %v", err)
		return
	}

	// Запускаем обработчик ping/pong
	go h.handlePingPong(conn)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.logger.Errorf("WebSocket error: %v", err)
			}
			break
		}

		h.processMessage(conn, msg)
	}
}

func (h *Handler) handlePingPong(conn *websocket.Conn) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := h.sendWSMessage(conn, "ping", nil); err != nil {
				h.logger.Errorf("Failed to send ping: %v", err)
				return
			}
		}
	}
}

func (h *Handler) processMessage(conn *websocket.Conn, msg []byte) {
	var request struct {
		Action string          `json:"action"`
		Data   json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(msg, &request); err != nil {
		h.logger.Errorf("Invalid message format: %v", err)
		h.sendErrorWS(conn, "Invalid message format")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), rpcTimeout)
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
		req.Ports = "1-1024"
	}

	taskID := generateTaskID()
	h.wsManager.RegisterTask(taskID, conn)
	defer h.wsManager.UnregisterTask(taskID)

	// Используем новый RPC вызов
	resp, err := h.queue.RPCCall(ctx, queue.ScanRequest{
		TaskID: taskID,
		IP:     req.IP,
		Ports:  req.Ports,
	})

	if err != nil {
		h.logger.Errorf("RPC call failed: %v", err)
		h.sendErrorWS(conn, "Scan failed")
		return
	}
	if resp.Error != "" {
		h.logger.Infof(
			"Scan result: task_id=%s, ip=%s, ports=%s, status=%s, error=%s",
			resp.TaskID, req.IP, req.Ports, resp.Status, resp.Error,
		)
	} else {
		h.logger.Infof(
			"Scan result: task_id=%s, ip=%s, ports=%s, status=%s, open_ports=%v",
			resp.TaskID, req.IP, req.Ports, resp.Status, resp.OpenPorts,
		)
	}
	h.sendWSMessage(conn, "scan_result", map[string]interface{}{
		"task_id":    resp.TaskID,
		"status":     resp.Status,
		"open_ports": resp.OpenPorts,
		"error":      resp.Error,
	})
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

func generateTaskID() string {
	return "task_" + time.Now().Format("20060102150405")
}
