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
	api.HandleFunc("/scan/history", h.getHistory).Methods("GET") // Изменено на /scan/history
	api.HandleFunc("/scan/{id}", h.getScanByID).Methods("GET")   // Изменено на /scan/{id}

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

	// Save scan request
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

	// Scan ports
	openPorts, err := h.portScanner.ScanPorts(r.Context(), request.IP, request.Ports)
	if err != nil {
		h.logger.Errorf("Failed to scan ports: %v", err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to scan ports")
		return
	}

	// Save scan results
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

	// Prepare response
	response := models.ScanResponseSwagger{
		RequestID: requestID,
		IP:        request.IP,
		Ports:     request.Ports,
		OpenPorts: openPorts,
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// getHistory godoc
// @Summary История сканирований
// @Description Получить историю всех сканирований
// @Tags scan
// @Produce  json
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

	h.respondWithJSON(w, http.StatusOK, history)
}

// getScanByID godoc
// @Summary Получить сканирование по ID
// @Description Получить детали сканирования по его ID
// @Tags scan
// @Produce  json
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

	h.respondWithJSON(w, http.StatusOK, scanResponse)
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
