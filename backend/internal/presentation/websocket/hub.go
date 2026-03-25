package websocket

import "sync"

// Hub maintains the set of all active WebSocket clients and allows
// broadcasting messages to every one of them.
//
// A singleton is used so that the RabbitMQ consumer goroutine started in
// main.go can broadcast change events without knowing individual clients.
type Hub struct {
	mu      sync.RWMutex
	clients map[*Client]struct{}
}

var globalHub = &Hub{
	clients: make(map[*Client]struct{}),
}

// GetHub returns the process-wide singleton Hub.
func GetHub() *Hub { return globalHub }

// Register adds a client to the hub.
func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()
}

// Unregister removes a client from the hub.
func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	delete(h.clients, c)
	h.mu.Unlock()
}

// Broadcast sends msg to every connected client.
// Slow / disconnected clients are skipped (non-blocking send).
func (h *Hub) Broadcast(msg Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.clients {
		select {
		case c.send <- msg:
		default:
			// client buffer is full — skip rather than block
		}
	}
}

