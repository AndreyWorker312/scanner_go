package handler

import (
	"encoding/json"
	"net/http"
	"network-scanner/internal/models"
	"network-scanner/internal/repository"
	"network-scanner/internal/scanner"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type Handler struct {
	logger      Logger
	repo        repository.Repository
	portScanner scanner.PortScanner
}

type Logger interface {
	Info(msg string)
	Infof(format string, args ...interface{})
	Error(msg string)
	Errorf(format string, args ...interface{})
	Fatal(msg string)
	Fatalf(format string, args ...interface{})
}

func NewHandler(logger Logger, repo repository.Repository, scanner scanner.PortScanner) *Handler {
	return &Handler{
		logger:      logger,
		repo:        repo,
		portScanner: scanner,
	}
}

func (h *Handler) InitRoutes() *mux.Router {
	r := mux.NewRouter()

	// API routes
	api := r.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/scan", h.scan).Methods("POST")
	api.HandleFunc("/history", h.history).Methods("GET")
	api.HandleFunc("/results/{id}", h.results).Methods("GET")

	// Middleware
	api.Use(h.loggingMiddleware)
	api.Use(h.contentTypeMiddleware)

	return r
}

// scan godoc
// @Summary Сканировать порты
// @Description Сканирует указанные порты на IP и сохраняет результат
// @Tags scan
// @Accept  json
// @Produce  json
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
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.IP == "" {
		http.Error(w, "IP address is required", http.StatusBadRequest)
		return
	}

	if request.Ports == "" {
		request.Ports = "1-1024" // Default ports to scan
	}

	// Сохраняем запрос в базу данных
	scanReq := models.ScanRequest{
		IPAddress: request.IP,
		Ports:     request.Ports,
		CreatedAt: time.Now(),
	}

	requestID, err := h.repo.SaveScanRequest(r.Context(), &scanReq)
	if err != nil {
		h.logger.Errorf("Failed to save scan request: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Сканируем порты
	openPorts, err := h.portScanner.ScanPorts(r.Context(), request.IP, request.Ports)
	if err != nil {
		h.logger.Errorf("Failed to scan ports: %v", err)
		http.Error(w, "Failed to scan ports", http.StatusInternalServerError)
		return
	}

	// Сохраняем результаты сканирования
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
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Формируем ответ (теперь используем именованную структуру)
	response := models.ScanResponseSwagger{
		RequestID: requestID,
		IP:        request.IP,
		Ports:     request.Ports,
		OpenPorts: openPorts,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// history godoc
// @Summary История сканирований
// @Description Получить историю всех сканирований
// @Tags scan
// @Produce  json
// @Success 200 {array} models.ScanRequest
// @Failure 500 {string} string "internal error"
// @Router /history [get]
func (h *Handler) history(w http.ResponseWriter, r *http.Request) {
	history, err := h.repo.GetScanHistory(r.Context())
	if err != nil {
		h.logger.Errorf("Failed to get scan history: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

// results godoc
// @Summary Результаты сканирования
// @Description Получить результаты по request_id
// @Tags scan
// @Produce  json
// @Param id path int true "ID запроса"
// @Success 200 {array} models.ScanResult
// @Failure 400 {string} string "bad request"
// @Failure 500 {string} string "internal error"
// @Router /results/{id} [get]
func (h *Handler) results(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	requestID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid request ID", http.StatusBadRequest)
		return
	}

	results, err := h.repo.GetScanResults(r.Context(), requestID)
	if err != nil {
		h.logger.Errorf("Failed to get scan results: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
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
