package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"network-scanner/internal/models"
	"network-scanner/internal/repository"
	"network-scanner/internal/scanner"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// Handler структура для обработчиков HTTP запросов
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
		return true // Для разработки. В продакшене замените на проверку origin.
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

// InitRoutes инициализирует маршруты API
// @title Network Scanner API
// @version 1.0
// @description API для сканирования сетевых портов
// @host localhost:8080
// @BasePath /api/v1
func (h *Handler) InitRoutes() *mux.Router {
	r := mux.NewRouter()
	
	api := r.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/scan", h.scan).Methods("POST")
	api.HandleFunc("/scan/history", h.getHistory).Methods("GET")
	api.HandleFunc("/scan/{id}", h.getScanByID).Methods("GET")
	api.HandleFunc("/ws", h.handleWebSocket).Methods("GET")

	api.Use(h.loggingMiddleware)
	api.Use(h.contentTypeMiddleware)

	return r
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
		if err := conn.WriteJSON(message); err != nil {
			log.Printf("WebSocket write error: %v", err)
			conn.Close()
			delete(m.clients, conn)
		}
	}
}

// handleWebSocket обрабатывает WebSocket соединения
// @Summary WebSocket соединение
// @Description Устанавливает WebSocket соединение для получения уведомлений
// @Tags websocket
// @Produce json
// @Success 101 {string} string "Switching Protocols"
// @Router /ws [get]
func (h *Handler) handleWebSocket(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        h.logger.Errorf("WebSocket upgrade failed: %v", err)
        return
    }
    defer conn.Close()

    h.wsManager.AddClient(conn)
    defer h.wsManager.RemoveClient(conn)

    // Отправляем приветственное сообщение
    if err := conn.WriteJSON(map[string]string{
    "type":    "welcome",
    "message": "connected",
}); err != nil {
    h.logger.Errorf("WebSocket write error: %v", err)
    return
}


    // Читаем сообщения от клиента
    for {
        if _, _, err := conn.ReadMessage(); err != nil {
            break
        }
    }
}

// scan обрабатывает запрос на сканирование портов
// @Summary Сканировать порты
// @Description Сканирует указанные порты на IP и сохраняет результат
// @Tags scan
// @Accept json
// @Produce json
// @Param request body models.ScanRequestSwagger true "Запрос на сканирование"
// @Success 200 {object} models.ScanResponseSwagger
// @Failure 400 {string} string "bad request"
// @Failure 500 {string} string "internal error"
// @Router /scan [post]
func (h *Handler) scan(w http.ResponseWriter, r *http.Request) {
	var request struct {
		IP    string `json:"ip"`
		Ports string `json:"ports"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if request.IP == "" {
		h.respondWithError(w, http.StatusBadRequest, "IP address is required")
		return
	}

	if request.Ports == "" {
		request.Ports = "1-1024" // Default ports to scan
	}

	scanReq := models.ScanRequest{
		IPAddress: request.IP,
		Ports:     request.Ports,
		CreatedAt: time.Now(),
	}

	requestID, err := h.repo.SaveScanRequest(r.Context(), &scanReq)
	if err != nil {
		h.logger.Errorf("Failed to save scan request: %v", err)
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	openPorts, err := h.portScanner.ScanPorts(r.Context(), request.IP, request.Ports)
	if err != nil {
		h.logger.Errorf("Failed to scan ports: %v", err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to scan ports")
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

	if err := h.repo.SaveScanResults(r.Context(), scanResults); err != nil {
		h.logger.Errorf("Failed to save scan results: %v", err)
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	response := map[string]interface{}{
		"request_id": requestID,
		"ip":         request.IP,
		"ports":      request.Ports,
		"open_ports": openPorts,
		"created_at": time.Now().Format(time.RFC3339),
	}

	h.respondWithJSON(w, http.StatusOK, response)
	h.wsManager.Broadcast(map[string]interface{}{
		"event": "scan_completed",
		"data":  response,
	})
}


/// getHistory возвращает историю сканирований
// @Summary История сканирований
// @Description Получить историю всех сканирований
// @Tags scan
// @Produce json
// @Success 200 {array} models.ScanResponse
// @Failure 500 {string} string "internal error"
// @Router /scan/history [get]
func (h *Handler) getHistory(w http.ResponseWriter, r *http.Request) {
    history, err := h.repo.GetScanHistory(r.Context())
    if err != nil {
        h.logger.Errorf("Failed to get scan history: %v", err)
        h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
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

    h.respondWithJSON(w, http.StatusOK, response)
}



// getScanByID возвращает результаты сканирования по ID
// @Summary Получить сканирование по ID
// @Description Получить детали сканирования по его ID
// @Tags scan
// @Produce json
// @Param id path int true "ID сканирования"
// @Success 200 {object} models.ScanResponse
// @Failure 400 {string} string "bad request"
// @Failure 404 {string} string "not found"
// @Failure 500 {string} string "internal error"
// @Router /scan/{id} [get]
func (h *Handler) getScanByID(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    idStr := vars["id"]

    requestID, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        h.respondWithError(w, http.StatusBadRequest, "Invalid scan ID")
        return
    }

    scanResponse, err := h.repo.GetScanResults(r.Context(), requestID)
    if err != nil {
        if err == repository.ErrNotFound {
            h.respondWithError(w, http.StatusNotFound, "Scan not found")
        } else {
            h.logger.Errorf("Failed to get scan results: %v", err)
            h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
        }
        return
    }

    if scanResponse.Request == nil {
        h.respondWithError(w, http.StatusInternalServerError, "Scan request data missing")
        return
    }

    response := map[string]interface{}{
        "id":          scanResponse.Request.ID,
        "ip_address":  scanResponse.Request.IPAddress,
        "ports":       scanResponse.Request.Ports,
        "created_at":  scanResponse.Request.CreatedAt.Format(time.RFC3339),
        "open_ports":  scanResponse.OpenPorts,
    }

    h.respondWithJSON(w, http.StatusOK, response)
}
func (h *Handler) loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        h.logger.Infof("%s %s", r.Method, r.URL.Path)
        next.ServeHTTP(w, r)
    })
}

func (h *Handler) contentTypeMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        next.ServeHTTP(w, r)
    })
}

func (h *Handler) respondWithError(w http.ResponseWriter, code int, message string) {
    h.respondWithJSON(w, code, map[string]string{"error": message})
}

func (h *Handler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    json.NewEncoder(w).Encode(payload)
}
