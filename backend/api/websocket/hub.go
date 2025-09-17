package websocket

import (
	"sync"

	"backend/domain/models"
)

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan *BroadcastMessage
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

// ... остальные методы Hub ...

func (h *Hub) RegisterClient(client *Client) {
	h.mu.Lock()
	h.clients[client] = true
	h.mu.Unlock()
}

func (h *Hub) UnregisterClient(client *Client) {
	h.mu.Lock()
	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.Send)
	}
	h.mu.Unlock()
}

func (h *Hub) BroadcastToTask(taskID string, message *BroadcastMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		if client.TaskIDs[taskID] {
			select {
			case client.Send <- h.marshalMessage(message):
			default:
				h.UnregisterClient(client)
			}
		}
	}
}
