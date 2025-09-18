package handlers

import (
	"encoding/json"
	_ "github.com/your-username/my-scanner-app/internal/application/usecases"
	"net/http"
)

type ScanHandler struct {
	startScanUseCase *usecases.StartScanUseCase
}

func NewScanHandler(startScanUseCase *usecases.StartScanUseCase) *ScanHandler {
	return &ScanHandler{startScanUseCase: startScanUseCase}
}

func (h *ScanHandler) StartScan(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Target   string                 `json:"target"`
		ScanType string                 `json:"scan_type"`
		Options  map[string]interface{} `json:"options"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	scanID, err := h.startScanUseCase.Execute(request.Target, request.ScanType, request.Options)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]string{"scan_id": scanID}
	json.NewEncoder(w).Encode(response)
}
