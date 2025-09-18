package websocket

import (
	"sync"
)

type Hub struct {
	clients    map[string]map[*Client]bool // scanID -> clients
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

type Message struct {
	ScanID  string      `json:"scan_id"`
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]map[*Client]bool),
		broadcast:  make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.scanID] == nil {
				h.clients[client.scanID] = make(map[*Client]bool)
			}
			h.clients[client.scanID][client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, exists := h.clients[client.scanID]; exists {
				delete(clients, client)
				if len(clients) == 0 {
					delete(h.clients, client.scanID)
				}
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			if clients, exists := h.clients[message.ScanID]; exists {
				for client := range clients {
					select {
					case client.send <- message:
					default:
						close(client.send)
						delete(clients, client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}
