package models

import "time"

// Базовые типы
type ScanType string

const (
	ScanTypeARP  ScanType = "arp"
	ScanTypeICMP ScanType = "icmp"
	ScanTypeNMAP ScanType = "nmap"
)

type ScanStatus string

const (
	ScanStatusPending   ScanStatus = "pending"
	ScanStatusRunning   ScanStatus = "running"
	ScanStatusCompleted ScanStatus = "completed"
	ScanStatusFailed    ScanStatus = "failed"
)

// Базовые структуры для всех сканирований
type ScanRequest struct {
	TaskID    string      `json:"task_id" bson:"task_id"`
	Type      ScanType    `json:"type" bson:"type"`
	CreatedAt time.Time   `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time   `json:"updated_at" bson:"updated_at"`
	Status    ScanStatus  `json:"status" bson:"status"`
	Request   interface{} `json:"request" bson:"request"` // Конкретный запрос сканера
}

type ScanResponse struct {
	TaskID    string      `json:"task_id" bson:"task_id"`
	Type      ScanType    `json:"type" bson:"type"`
	Status    string      `json:"status" bson:"status"`
	Response  interface{} `json:"response" bson:"response"` // Конкретный ответ сканера
	Error     string      `json:"error,omitempty" bson:"error,omitempty"`
	Duration  int64       `json:"duration_ms" bson:"duration_ms"`
	CreatedAt time.Time   `json:"created_at" bson:"created_at"`
}

// ==================== ARP SCANNER MODELS ====================
type ARPRequest struct {
	TaskID        string `json:"task_id"`
	InterfaceName string `json:"interface_name"`
	IPRange       string `json:"ip_range"`
}

type ARPResponse struct {
	TaskID  string       `json:"task_id"`
	Status  string       `json:"status"`
	Devices []ARPDevice  `json:"devices"`
	Error   string       `json:"error,omitempty"`
}

type ARPDevice struct {
	IP     string `json:"ip"`
	MAC    string `json:"mac"`
	Vendor string `json:"vendor,omitempty"`
	Status string `json:"status"`
}

// ==================== ICMP SCANNER MODELS ====================
type ICMPRequest struct {
	TaskID    string   `json:"task_id"`
	Targets   []string `json:"targets"`
	PingCount int      `json:"ping_count"`
}

type ICMPResponse struct {
	TaskID  string          `json:"task_id"`
	Status  string          `json:"status"`
	Results []ICMPResult    `json:"results"`
	Error   string          `json:"error,omitempty"`
}

type ICMPResult struct {
	Target            string  `json:"target"`
	Address           string  `json:"address"`
	PacketsSent       int     `json:"packets_sent"`
	PacketsReceived   int     `json:"packets_received"`
	PacketLossPercent float64 `json:"packet_loss_percent"`
	Error             string  `json:"error,omitempty"`
}

// ==================== NMAP SCANNER MODELS ====================

type NmapRequest struct {
	ScanMethod string `json:"scan_method"`
}

type NmapTcpUdpRequest struct {
	TaskID      string `json:"task_id"`
	IP          string `json:"ip"`
	ScannerType string `json:"scanner_type"` // "tcp" или "udp"
	Ports       string `json:"ports"`
}

type NmapTcpUdpResponse struct {
	TaskID   string              `json:"task_id"`
	Host     string              `json:"host"`
	PortInfo []NmapPortTcpUdpInfo `json:"port_info"`
	Status   string              `json:"status"`
	Error    string              `json:"error,omitempty"`
}

type NmapPortTcpUdpInfo struct {
	Status      string   `json:"status"`
	AllPorts    []uint16 `json:"close_ports"`
	Protocols   []string `json:"protocols"`
	State       []string `json:"state"`
	ServiceName []string `json:"service_name"`
}

// OS Detection сканирование
type NmapOsDetectionRequest struct {
	TaskID string `json:"task_id"`
	IP     string `json:"ip"`
}

type NmapOsDetectionResponse struct {
	TaskID   string `json:"task_id"`
	Host     string `json:"host"`
	Name     string `json:"name"`
	Accuracy int    `json:"accuracy"`
	Vendor   string `json:"vendor"`
	Family   string `json:"family"`
	Type     string `json:"type"`
	Status   string `json:"status"`
	Error    string `json:"error,omitempty"`
}

// Host Discovery сканирование
type NmapHostDiscoveryRequest struct {
	TaskID string `json:"task_id"`
	IP     string `json:"ip"`
}

type NmapHostDiscoveryResponse struct {
	TaskID    string `json:"task_id"`
	Host      string `json:"host"`
	HostUP    int    `json:"host_up"`
	HostTotal int    `json:"host_total"`
	Status    string `json:"status"`
	DNS       string `json:"dns"`
	Reason    string `json:"reason"`
	Error     string `json:"error,omitempty"`
}