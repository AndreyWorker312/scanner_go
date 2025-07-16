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

// loggingMiddleware logs all incoming HTTP requests.
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

	h.logger.Infof("New WebSocket connection from %v", conn.RemoteAddr())
	h.wsManager.AddClient(conn)
	defer h.wsManager.RemoveClient(conn)

	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		h.logger.Infof("Received pong from %v", conn.RemoteAddr())
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	if err := h.sendWSMessage(conn, "welcome", map[string]string{
		"message": "Connected to Port Scanner",
	}); err != nil {
		h.logger.Errorf("Failed to send welcome message to %v: %v", conn.RemoteAddr(), err)
		return
	}

	for {
		mt, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.logger.Errorf("WebSocket unexpected close error from %v: %v", conn.RemoteAddr(), err)
			} else {
				h.logger.Infof("WebSocket connection closed by client %v: %v", conn.RemoteAddr(), err)
			}
			break
		}

		h.logger.Infof("Received message from %v: messageType=%d, len=%d", conn.RemoteAddr(), mt, len(msg))

		var request struct {
			Action string          `json:"action"`
			Data   json.RawMessage `json:"data"`
		}

		if err := json.Unmarshal(msg, &request); err != nil {
			h.logger.Errorf("Invalid message format from %v: %v", conn.RemoteAddr(), err)
			h.sendErrorWS(conn, "Invalid message format")
			continue
		}

		h.logger.Infof("Processing action '%s' from %v", request.Action, conn.RemoteAddr())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		switch request.Action {
		case "scan":
			h.handleScanRequest(ctx, conn, request.Data)
		case "ping":
			h.logger.Infof("Received ping from %v", conn.RemoteAddr())
			h.sendWSMessage(conn, "pong", nil)
		default:
			h.logger.Warnf("Unknown action '%s' from %v", request.Action, conn.RemoteAddr())
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
		h.logger.Errorf("Failed to unmarshal scan request: %v", err)
		h.sendErrorWS(conn, "Invalid scan request")
		return
	}

	h.logger.Infof("Received scan request from %v: IP=%s, Ports=%s", conn.RemoteAddr(), req.IP, req.Ports)

	if req.IP == "" {
		h.logger.Warnf("Scan request missing IP address from %v", conn.RemoteAddr())
		h.sendErrorWS(conn, "IP address is required")
		return
	}

	if req.Ports == "" {
		req.Ports = "1-1024"
		h.logger.Infof("Ports not provided, defaulting to %s", req.Ports)
	}

	taskID := generateTaskID()
	h.wsManager.RegisterTask(taskID, conn)

	if err := h.sendWSMessage(conn, "scan_queued", map[string]string{"task_id": taskID}); err != nil {
		h.logger.Errorf("Failed to send scan_queued message to %v: %v", conn.RemoteAddr(), err)
	}

	err := h.queue.PublishScanRequest(ctx, queue.ScanRequest{
		TaskID: taskID,
		IP:     req.IP,
		Ports:  req.Ports,
	})
	if err != nil {
		h.logger.Errorf("Failed to publish scan request to queue: %v", err)
		h.sendErrorWS(conn, "Failed to queue scan request")
		h.wsManager.UnregisterTask(taskID)
		return
	}

	h.logger.Infof("Published scan request %s for %s:%s to queue", taskID, req.IP, req.Ports)
}

func (h *Handler) sendWSMessage(conn *websocket.Conn, msgType string, data interface{}) error {
	conn.SetWriteDeadline(time.Now().Add(writeWait))
	err := conn.WriteJSON(map[string]interface{}{
		"type": msgType,
		"data": data,
	})
	if err != nil {
		h.logger.Errorf("Failed to send message of type '%s' to %v: %v", msgType, conn.RemoteAddr(), err)
	}
	return err
}

func (h *Handler) sendErrorWS(conn *websocket.Conn, message string) {
	h.logger.Infof("Sending error message to %v: %s", conn.RemoteAddr(), message)
	h.sendWSMessage(conn, "error", map[string]string{
		"message": message,
	})
}

func generateTaskID() string {
	return "task_" + time.Now().Format("20060102150405")
}
