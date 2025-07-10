package websocket

type Message struct {
    Type    string      `json:"type"`    // "progress", "open_port", "error"
    Data    interface{} `json:"data"`    // Может быть int, string, map и т.д.
    RequestID int64     `json:"request_id,omitempty"` // ID запроса (если нужно)
}