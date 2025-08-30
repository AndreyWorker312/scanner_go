package domain

type ScanRequest struct {
	TaskID string `json:"task_id"`
	IP     string `json:"ip"`
	Ports  string `json:"ports"`
}

type ScanResponse struct {
	TaskID    string `json:"task_id"`
	Status    string `json:"status"`
	OpenPorts []int  `json:"open_ports"`
	Error     string `json:"error,omitempty"`
}
