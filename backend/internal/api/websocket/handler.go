package websocket

import (
	"log"
	"net/http"

	"backend/internal/domain/models"
	api "backend/internal/application"
	"github.com/gorilla/websocket"
)

type Message struct {
	Type string           `json:"type"`
	Req  *models.Request  `json:"request,omitempty"`
	Resp *models.Response `json:"response,omitempty"`
}

type Client struct {
	conn *websocket.Conn
	send chan Message
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	client := &Client{
		conn: conn,
		send: make(chan Message, 256),
	}

	go client.writePump()
	client.readPump()
}

func (c *Client) readPump() {
	defer c.conn.Close()

	for {
		var msg Message

		if err := c.conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}

		log.Printf("Received message type=%s request=%+v", msg.Type, msg.Req)

		if msg.Req != nil {
			response := api.ProcessRequest(msg.Req)

			c.send <- Message{
				Type: "response",
				Resp: response,
			}
		}
	}
}

func (c *Client) writePump() {
	defer c.conn.Close()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteJSON(message); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}
		}
	}
}
