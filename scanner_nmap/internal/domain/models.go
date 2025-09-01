package domain

type RawRequest struct {
	ScanMethod string `json:"scan_method"`
}
type ScanTcpUdpRequest struct {
	TaskID      string `json:"task_id"`
	IP          string `json:"ip"`
	ScannerType string `json:"scanner_type"`
	Ports       string `json:"ports"`
}

type ScanTcpUdpResponse struct {
	TaskID   string           `json:"task_id"`
	Host     string           `json:"host"`
	PortInfo []PortTcpUdpInfo `json:"port_info"`
	Status   string           `json:"status"`
}

type PortTcpUdpInfo struct {
	Status      string   `json:"status"`
	AllPorts    []uint16 `json:"close_ports"`
	Protocols   []string `json:"protocols"`
	State       []string `json:"state"`
	ServiceName []string `json:"service_name"`
}

type OsDetectionRequest struct {
	TaskID string `json:"task_id"`
	IP     string `json:"ip"`
}

type OsDetectionResponse struct {
	TaskID   string `json:"task_id"`
	Host     string `json:"host"`
	Name     string `json:"name"`
	Accuracy int    `json:"accuracy"`
	Vendor   string `json:"vendor"`
	Family   string `json:"family"`
	Type     string `json:"type"`
	Status   string `json:"status"`
}

type HostDiscoveryRequest struct {
	TaskID string `json:"task_id"`
	IP     string `json:"ip"`
}
type HostDiscoveryResponse struct {
	TaskID    string `json:"task_id"`
	Host      string `json:"host"`
	HostUP    int    `json:"host_up"`
	HostTotal int    `json:"host_total"`
	Status    string `json:"status"`
	DNS       string `json:"dns"`
	Reason    string `json:"reason"`
}
