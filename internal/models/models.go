package models

import "time"

type ScanRequest struct {
	ID        int64     `json:"id"`
	IPAddress string    `json:"ip_address"`
	Ports     string    `json:"ports"` // Может быть "80,443" или "1-1024"
	CreatedAt time.Time `json:"created_at"`
}

type ScanResult struct {
	ID        int64     `json:"id"`
	RequestID int64     `json:"request_id"`
	Port      int       `json:"port"`
	IsOpen    bool      `json:"is_open"`
	ScannedAt time.Time `json:"scanned_at"`
}

type ScanResponse struct {
	Request   *ScanRequest  `json:"request"`
	Results   []*ScanResult `json:"results"`
	OpenPorts []int         `json:"open_ports"`
}

// --- ДОБАВЬ ЭТО ДЛЯ SWAGGER ---

type ScanRequestSwagger struct {
	IP    string `json:"ip" example:"192.168.1.1"`
	Ports string `json:"ports" example:"22,80,443"`
}

type ScanResponseSwagger struct {
	RequestID int64  `json:"request_id" example:"1"`
	IP        string `json:"ip" example:"192.168.1.1"`
	Ports     string `json:"ports" example:"22,80,443"`
	OpenPorts []int  `json:"open_ports" example:"22,80"` // <--- так
}
