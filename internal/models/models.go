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
	Request   *ScanRequest  `json:"request"` // Изменено на указатель
	Results   []*ScanResult `json:"results"` // Изменено на указатель
	OpenPorts []int         `json:"open_ports"`
}
