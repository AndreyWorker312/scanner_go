package domain

type ScanTcpUdpRequest struct {
	TaskID string `json:"task_id"`
	IP     string `json:"ip"`
	Ports  string `json:"ports"`
}

type ScanTcpUdpResponse struct {
	TaskID   string           `json:"task_id"`
	Host     string           `json:"host"`
	PortInfo []PortTcpUdpInfo `json:"port_info"`
	Error    string           `json:"error,omitempty"`
}

type PortTcpUdpInfo struct {
	Status      []string `json:"status"`
	OpenPorts   []uint16 `json:"open_ports"`
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
}

type HostDiscoveryRequest struct {
	TaskID string `json:"task_id"`
	IP     string `json:"ip"`
}
type HostDiscoveryResponse struct {
	TaskID    string `json:"task_id"`
	HostUP    int    `json:"host_up"`
	HostTotal int    `json:"host_total"`
	Status    string `json:"status"`
	DNS       string `json:"dns"`
	Reason    string `json:"reason"`
}
