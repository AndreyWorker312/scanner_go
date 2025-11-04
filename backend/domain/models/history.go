package models

import (
	"time"
)

// ==================== ARP HISTORY MODELS ====================
type ARPHistoryRecord struct {
	ID             string      `bson:"_id,omitempty" json:"id"`
	TaskID         string      `bson:"task_id" json:"task_id"`
	InterfaceName  string      `bson:"interface_name" json:"interface_name"`
	IPRange        string      `bson:"ip_range" json:"ip_range"`
	Status         string      `bson:"status" json:"status"`
	Devices        []ARPDevice `bson:"devices" json:"devices"`
	OnlineDevices  []ARPDevice `bson:"online_devices" json:"online_devices"`
	OfflineDevices []ARPDevice `bson:"offline_devices" json:"offline_devices"`
	TotalCount     int         `bson:"total_count" json:"total_count"`
	OnlineCount    int         `bson:"online_count" json:"online_count"`
	OfflineCount   int         `bson:"offline_count" json:"offline_count"`
	Error          string      `bson:"error,omitempty" json:"error,omitempty"`
	CreatedAt      time.Time   `bson:"created_at" json:"created_at"`
}

// ==================== ICMP HISTORY MODELS ====================
type ICMPHistoryRecord struct {
	ID        string       `bson:"_id,omitempty" json:"id"`
	TaskID    string       `bson:"task_id" json:"task_id"`
	Targets   []string     `bson:"targets" json:"targets"`
	PingCount int          `bson:"ping_count" json:"ping_count"`
	Status    string       `bson:"status" json:"status"`
	Results   []ICMPResult `bson:"results" json:"results"`
	Error     string       `bson:"error,omitempty" json:"error,omitempty"`
	CreatedAt time.Time    `bson:"created_at" json:"created_at"`
}

// ==================== NMAP HISTORY MODELS ====================

// TCP/UDP Scan History
type NmapTcpUdpHistoryRecord struct {
	ID          string               `bson:"_id,omitempty" json:"id"`
	TaskID      string               `bson:"task_id" json:"task_id"`
	IP          string               `bson:"ip" json:"ip"`
	ScannerType string               `bson:"scanner_type" json:"scanner_type"`
	Ports       string               `bson:"ports" json:"ports"`
	Host        string               `bson:"host" json:"host"`
	PortInfo    []NmapPortTcpUdpInfo `bson:"port_info" json:"port_info"`
	Status      string               `bson:"status" json:"status"`
	Error       string               `bson:"error,omitempty" json:"error,omitempty"`
	CreatedAt   time.Time            `bson:"created_at" json:"created_at"`
}

// OS Detection History
type NmapOsDetectionHistoryRecord struct {
	ID        string    `bson:"_id,omitempty" json:"id"`
	TaskID    string    `bson:"task_id" json:"task_id"`
	IP        string    `bson:"ip" json:"ip"`
	Host      string    `bson:"host" json:"host"`
	Name      string    `bson:"name" json:"name"`
	Accuracy  int       `bson:"accuracy" json:"accuracy"`
	Vendor    string    `bson:"vendor" json:"vendor"`
	Family    string    `bson:"family" json:"family"`
	Type      string    `bson:"type" json:"type"`
	Status    string    `bson:"status" json:"status"`
	Error     string    `bson:"error,omitempty" json:"error,omitempty"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
}

// Host Discovery History
type NmapHostDiscoveryHistoryRecord struct {
	ID        string    `bson:"_id,omitempty" json:"id"`
	TaskID    string    `bson:"task_id" json:"task_id"`
	IP        string    `bson:"ip" json:"ip"`
	Host      string    `bson:"host" json:"host"`
	HostUP    int       `bson:"host_up" json:"host_up"`
	HostTotal int       `bson:"host_total" json:"host_total"`
	Status    string    `bson:"status" json:"status"`
	DNS       string    `bson:"dns" json:"dns"`
	Reason    string    `bson:"reason" json:"reason"`
	Error     string    `bson:"error,omitempty" json:"error,omitempty"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
}

// ==================== TCP HISTORY MODELS ====================
type TCPHistoryRecord struct {
	ID           string    `bson:"_id,omitempty" json:"id"`
	TaskID       string    `bson:"task_id" json:"task_id"`
	Host         string    `bson:"host" json:"host"`
	Port         string    `bson:"port" json:"port"`
	HexObjectKey string    `bson:"hex_object_key" json:"hex_object_key"`
	DecodedText  string    `bson:"decoded_text" json:"decoded_text"`
	Status       string    `bson:"status" json:"status"`
	Error        string    `bson:"error,omitempty" json:"error,omitempty"`
	CreatedAt    time.Time `bson:"created_at" json:"created_at"`
}

// ==================== API RESPONSE MODELS ====================
type HistoryResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Count   int         `json:"count,omitempty"`
}
