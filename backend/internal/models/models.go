package models

import "time"

type ScanRequest struct {
	TaskID    string    `json:"task_id"`
	IP        string    `json:"ip"`
	Ports     string    `json:"ports"`
	CreatedAt time.Time `json:"created_at"`
}

type ScanResponse struct {
	TaskID    string    `json:"task_id"`
	IP        string    `json:"ip"`
	Ports     string    `json:"ports"`
	OpenPorts []int     `json:"open_ports"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}
