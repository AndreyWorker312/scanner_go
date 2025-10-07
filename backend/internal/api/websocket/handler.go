package websocket

import (
	"encoding/json"
	"log"
	"net/http"

	"backend/domain/models"
	api "backend/internal/application"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type WSHandler struct {
	app *api.App
}

func NewWSHandler(app *api.App) *WSHandler {
	return &WSHandler{
		app: app,
	}
}

type Message struct {
	Type string           `json:"type"`
	Req  *models.Request  `json:"request,omitempty"`
	Resp *models.Response `json:"response,omitempty"`
}

type Client struct {
	conn *websocket.Conn
	send chan Message
	app  *api.App
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (h *WSHandler) WsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	client := &Client{
		conn: conn,
		send: make(chan Message, 256),
		app:  h.app,
	}

	go client.writePump()
	go client.readPump()
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

		scannerService := ""
		if msg.Req != nil {
			scannerService = msg.Req.ScannerService
		}
		log.Printf("Received message type=%s, scanner_service=%s", msg.Type, scannerService)

		if msg.Req != nil {
			// Обрабатываем запрос и преобразуем options в нужную структуру
			response := c.processRequest(msg.Req)

			c.send <- Message{
				Type: "response",
				Resp: response,
			}
		}
	}
}

func (c *Client) processRequest(req *models.Request) *models.Response {
	taskID := generateTaskID()

	switch req.ScannerService {
	case "arp_service":
		return c.processARPRequest(req.Options, taskID)
	case "icmp_service", "ping_service":
		return c.processICMPRequest(req.Options, taskID)
	case "nmap_service":
		return c.processNmapRequest(req.Options, taskID)
	default:
		return &models.Response{
			TaskID: taskID,
			Result: map[string]string{
				"error": "unsupported scanner_service: " + req.ScannerService,
			},
		}
	}
}

func (c *Client) processARPRequest(options any, taskID string) *models.Response {
	var arpOpts struct {
		InterfaceName string `json:"interface_name"`
		IPRange       string `json:"ip_range"`
	}

	if err := parseOptions(options, &arpOpts); err != nil {
		return &models.Response{
			TaskID: taskID,
			Result: map[string]string{"error": "invalid ARP options: " + err.Error()},
		}
	}

	// Валидация
	if arpOpts.InterfaceName == "" {
		return &models.Response{
			TaskID: taskID,
			Result: map[string]string{"error": "interface_name is required for ARP scan"},
		}
	}
	if arpOpts.IPRange == "" {
		return &models.Response{
			TaskID: taskID,
			Result: map[string]string{"error": "ip_range is required for ARP scan"},
		}
	}

	// Создаем ARPRequest с TaskID
	arpRequest := models.ARPRequest{
		TaskID:        taskID,
		InterfaceName: arpOpts.InterfaceName,
		IPRange:       arpOpts.IPRange,
	}

	// Отправляем в application слой
	return c.app.ProcessRequest(&models.Request{
		ScannerService: "arp_service",
		Options:        arpRequest,
	})
}

func (c *Client) processICMPRequest(options any, taskID string) *models.Response {
	var icmpOpts struct {
		Targets   []string `json:"targets"`
		PingCount int      `json:"ping_count"`
	}

	if err := parseOptions(options, &icmpOpts); err != nil {
		return &models.Response{
			TaskID: taskID,
			Result: map[string]string{"error": "invalid ICMP options: " + err.Error()},
		}
	}

	// Валидация
	if len(icmpOpts.Targets) == 0 {
		return &models.Response{
			TaskID: taskID,
			Result: map[string]string{"error": "targets are required for ICMP ping"},
		}
	}
	if icmpOpts.PingCount <= 0 {
		icmpOpts.PingCount = 4 // значение по умолчанию
	}

	// Создаем ICMPRequest с TaskID
	icmpRequest := models.ICMPRequest{
		TaskID:    taskID,
		Targets:   icmpOpts.Targets,
		PingCount: icmpOpts.PingCount,
	}

	// Отправляем в application слой
	return c.app.ProcessRequest(&models.Request{
		ScannerService: "icmp_service",
		Options:        icmpRequest,
	})
}

func (c *Client) processNmapRequest(options any, taskID string) *models.Response {
	var nmapOpts struct {
		ScanMethod  string `json:"scan_method"`
		IP          string `json:"ip"`
		Ports       string `json:"ports"`
		ScannerType string `json:"scanner_type"`
	}

	if err := parseOptions(options, &nmapOpts); err != nil {
		return &models.Response{
			TaskID: taskID,
			Result: map[string]string{"error": "invalid Nmap options: " + err.Error()},
		}
	}

	// В зависимости от типа сканирования создаем соответствующую структуру
	switch nmapOpts.ScanMethod {
	case "tcp_udp_scan":
		if nmapOpts.IP == "" {
			return &models.Response{
				TaskID: taskID,
				Result: map[string]string{"error": "IP is required for TCP/UDP scan"},
			}
		}

		nmapRequest := models.NmapTcpUdpRequest{
			TaskID:      taskID,
			IP:          nmapOpts.IP,
			ScannerType: nmapOpts.ScannerType,
			Ports:       nmapOpts.Ports,
		}

		return c.app.ProcessRequest(&models.Request{
			ScannerService: "nmap_service",
			Options:        nmapRequest,
		})

	case "os_detection":
		if nmapOpts.IP == "" {
			return &models.Response{
				TaskID: taskID,
				Result: map[string]string{"error": "IP is required for OS detection"},
			}
		}

		nmapRequest := models.NmapOsDetectionRequest{
			TaskID: taskID,
			IP:     nmapOpts.IP,
		}

		return c.app.ProcessRequest(&models.Request{
			ScannerService: "nmap_service",
			Options:        nmapRequest,
		})

	case "host_discovery":
		if nmapOpts.IP == "" {
			return &models.Response{
				TaskID: taskID,
				Result: map[string]string{"error": "IP is required for host discovery"},
			}
		}

		nmapRequest := models.NmapHostDiscoveryRequest{
			TaskID: taskID,
			IP:     nmapOpts.IP,
		}

		return c.app.ProcessRequest(&models.Request{
			ScannerService: "nmap_service",
			Options:        nmapRequest,
		})

	default:
		return &models.Response{
			TaskID: taskID,
			Result: map[string]string{"error": "unsupported nmap scan method: " + nmapOpts.ScanMethod},
		}
	}
}

// Вспомогательные функции
func generateTaskID() string {
	return uuid.New().String()
}

func parseOptions(options any, target interface{}) error {
	optionsJSON, err := json.Marshal(options)
	if err != nil {
		return err
	}
	return json.Unmarshal(optionsJSON, target)
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
			log.Printf("Sent response: %+v", message.Resp)
		}
	}
}
