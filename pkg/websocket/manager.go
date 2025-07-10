package websocket

import (
    "github.com/gorilla/websocket"
    "log"
    "sync"
)

type Manager struct {
    clients map[*websocket.Conn]bool
    mutex   sync.Mutex
}

func NewManager() *Manager {
    return &Manager{
        clients: make(map[*websocket.Conn]bool),
    }
}

func (m *Manager) AddClient(conn *websocket.Conn) {
    m.mutex.Lock()
    defer m.mutex.Unlock()
    m.clients[conn] = true
}

func (m *Manager) RemoveClient(conn *websocket.Conn) {
    m.mutex.Lock()
    defer m.mutex.Unlock()
    delete(m.clients, conn)
}

func (m *Manager) Broadcast(msg Message) {
    m.mutex.Lock()
    defer m.mutex.Unlock()
    for client := range m.clients {
        if err := client.WriteJSON(msg); err != nil {
            log.Printf("WebSocket error: %v", err)
            client.Close()
            delete(m.clients, client)
        }
    }
}