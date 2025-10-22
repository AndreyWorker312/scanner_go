package rest

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"backend/domain/models"
)

type HistoryHandler struct {
	repo RepositoryInterface
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

	// TCP methods
	SaveTCPHistory(record *models.TCPHistoryRecord) error
	GetTCPHistory(limit int) ([]models.TCPHistoryRecord, error)
	DeleteTCPHistory() error
}

func NewHistoryHandler(repo RepositoryInterface) *HistoryHandler {
	return &HistoryHandler{repo: repo}
}

// ==================== ARP HISTORY ENDPOINTS ====================

func (h *HistoryHandler) GetARPHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse limit parameter
	limitStr := r.URL.Query().Get("limit")
	limit := 0
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	records, err := h.repo.GetARPHistory(limit)
	if err != nil {
		log.Printf("Error getting ARP history: %v", err)
		response := models.HistoryResponse{
			Success: false,
			Error:   "Failed to retrieve ARP history",
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := models.HistoryResponse{
		Success: true,
		Data:    records,
		Count:   len(records),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *HistoryHandler) DeleteARPHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := h.repo.DeleteARPHistory()
	if err != nil {
		log.Printf("Error deleting ARP history: %v", err)
		response := models.HistoryResponse{
			Success: false,
			Error:   "Failed to delete ARP history",
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := models.HistoryResponse{
		Success: true,
		Data:    "ARP history deleted successfully",
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ==================== ICMP HISTORY ENDPOINTS ====================

func (h *HistoryHandler) GetICMPHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 0
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	records, err := h.repo.GetICMPHistory(limit)
	if err != nil {
		log.Printf("Error getting ICMP history: %v", err)
		response := models.HistoryResponse{
			Success: false,
			Error:   "Failed to retrieve ICMP history",
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := models.HistoryResponse{
		Success: true,
		Data:    records,
		Count:   len(records),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *HistoryHandler) DeleteICMPHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := h.repo.DeleteICMPHistory()
	if err != nil {
		log.Printf("Error deleting ICMP history: %v", err)
		response := models.HistoryResponse{
			Success: false,
			Error:   "Failed to delete ICMP history",
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := models.HistoryResponse{
		Success: true,
		Data:    "ICMP history deleted successfully",
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ==================== NMAP HISTORY ENDPOINTS ====================

func (h *HistoryHandler) GetNmapHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	scanType := r.URL.Query().Get("type")
	if scanType == "" {
		scanType = "all"
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 0
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	result := make(map[string]interface{})

	if scanType == "all" || scanType == "tcp_udp" {
		tcpUdpRecords, err := h.repo.GetNmapTcpUdpHistory(limit)
		if err != nil {
			log.Printf("Error getting Nmap TCP/UDP history: %v", err)
		} else {
			result["tcp_udp"] = tcpUdpRecords
		}
	}

	if scanType == "all" || scanType == "os_detection" {
		osRecords, err := h.repo.GetNmapOsDetectionHistory(limit)
		if err != nil {
			log.Printf("Error getting Nmap OS Detection history: %v", err)
		} else {
			result["os_detection"] = osRecords
		}
	}

	if scanType == "all" || scanType == "host_discovery" {
		hostRecords, err := h.repo.GetNmapHostDiscoveryHistory(limit)
		if err != nil {
			log.Printf("Error getting Nmap Host Discovery history: %v", err)
		} else {
			result["host_discovery"] = hostRecords
		}
	}

	response := models.HistoryResponse{
		Success: true,
		Data:    result,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *HistoryHandler) DeleteNmapHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	scanType := r.URL.Query().Get("type")
	if scanType == "" {
		scanType = "all"
	}

	result := make(map[string]string)

	if scanType == "all" || scanType == "tcp_udp" {
		err := h.repo.DeleteNmapTcpUdpHistory()
		if err != nil {
			result["tcp_udp"] = "Failed to delete TCP/UDP history"
		} else {
			result["tcp_udp"] = "Deleted successfully"
		}
	}

	if scanType == "all" || scanType == "os_detection" {
		err := h.repo.DeleteNmapOsDetectionHistory()
		if err != nil {
			result["os_detection"] = "Failed to delete OS Detection history"
		} else {
			result["os_detection"] = "Deleted successfully"
		}
	}

	if scanType == "all" || scanType == "host_discovery" {
		err := h.repo.DeleteNmapHostDiscoveryHistory()
		if err != nil {
			result["host_discovery"] = "Failed to delete Host Discovery history"
		} else {
			result["host_discovery"] = "Deleted successfully"
		}
	}

	response := models.HistoryResponse{
		Success: true,
		Data:    result,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ==================== TCP HISTORY ENDPOINTS ====================

func (h *HistoryHandler) GetTCPHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 0
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	records, err := h.repo.GetTCPHistory(limit)
	if err != nil {
		log.Printf("Error getting TCP history: %v", err)
		response := models.HistoryResponse{
			Success: false,
			Error:   "Failed to retrieve TCP history",
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := models.HistoryResponse{
		Success: true,
		Data:    records,
		Count:   len(records),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *HistoryHandler) DeleteTCPHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := h.repo.DeleteTCPHistory()
	if err != nil {
		log.Printf("Error deleting TCP history: %v", err)
		response := models.HistoryResponse{
			Success: false,
			Error:   "Failed to delete TCP history",
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := models.HistoryResponse{
		Success: true,
		Data:    "TCP history deleted successfully",
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
