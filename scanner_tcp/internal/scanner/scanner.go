package scanner

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
	"time"
)

// PortScanResult содержит результаты сканирования TCP-порта
type PortScanResult struct {
	Host        string `json:"host"`
	Port        string `json:"port"`
	State       string `json:"state"`        // open, closed, filtered
	Service     string `json:"service"`      // http, ssh, ftp и т.д.
	Banner      string `json:"banner"`       // Баннер, прочитанный из сокета
	Version     string `json:"version"`      // Версия сервиса, если удалось определить
	Error       string `json:"error,omitempty"`
	ResponseTime int64 `json:"response_time"` // Время отклика в миллисекундах
}

// TCPScanner интерфейс для TCP сканирования
type TCPScanner interface {
	ScanPort(ctx context.Context, host string, port string) PortScanResult
	ScanPorts(ctx context.Context, host string, ports []string) []PortScanResult
}

type tcpScanner struct {
	timeout       time.Duration
	bannerTimeout time.Duration
	maxBannerSize int
}

// NewTCPScanner создает новый экземпляр TCP сканера
func NewTCPScanner(timeout, bannerTimeout time.Duration, maxBannerSize int) TCPScanner {
	return &tcpScanner{
		timeout:       timeout,
		bannerTimeout: bannerTimeout,
		maxBannerSize: maxBannerSize,
	}
}

// ScanPort сканирует один TCP порт и выполняет banner grabbing
func (s *tcpScanner) ScanPort(ctx context.Context, host string, port string) PortScanResult {
	startTime := time.Now()
	result := PortScanResult{
		Host:  host,
		Port:  port,
		State: "closed",
	}

	address := net.JoinHostPort(host, port)

	// Устанавливаем соединение
	dialer := &net.Dialer{
		Timeout: s.timeout,
	}

	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			result.State = "filtered"
			result.Error = "Connection timeout"
		} else {
			result.State = "closed"
			result.Error = err.Error()
		}
		result.ResponseTime = time.Since(startTime).Milliseconds()
		return result
	}
	defer conn.Close()

	result.State = "open"
	result.ResponseTime = time.Since(startTime).Milliseconds()

	// Пробуем прочитать баннер
	banner, err := s.grabBanner(conn)
	if err == nil && banner != "" {
		result.Banner = banner
		result.Service, result.Version = s.identifyService(port, banner)
	} else {
		// Если баннер не получен автоматически, пробуем отправить HTTP запрос
		if port == "80" || port == "8080" || port == "443" || port == "8443" {
			banner = s.sendHTTPRequest(conn)
			if banner != "" {
				result.Banner = banner
				result.Service = "http"
			}
		}
	}

	// Если сервис не определен, используем известные порты
	if result.Service == "" {
		result.Service = s.getServiceByPort(port)
	}

	return result
}

// ScanPorts сканирует несколько портов
func (s *tcpScanner) ScanPorts(ctx context.Context, host string, ports []string) []PortScanResult {
	results := make([]PortScanResult, 0, len(ports))
	
	for _, port := range ports {
		select {
		case <-ctx.Done():
			return results
		default:
			result := s.ScanPort(ctx, host, port)
			results = append(results, result)
		}
	}
	
	return results
}

// grabBanner читает баннер из TCP соединения
func (s *tcpScanner) grabBanner(conn net.Conn) (string, error) {
	conn.SetReadDeadline(time.Now().Add(s.bannerTimeout))
	
	reader := bufio.NewReader(conn)
	banner := make([]byte, s.maxBannerSize)
	
	n, err := reader.Read(banner)
	if err != nil && n == 0 {
		return "", err
	}
	
	// Очищаем баннер от непечатных символов
	return strings.TrimSpace(string(banner[:n])), nil
}

// sendHTTPRequest отправляет HTTP GET запрос для получения баннера
func (s *tcpScanner) sendHTTPRequest(conn net.Conn) string {
	conn.SetWriteDeadline(time.Now().Add(s.bannerTimeout))
	conn.SetReadDeadline(time.Now().Add(s.bannerTimeout))
	
	// Отправляем простой HTTP GET запрос
	request := "GET / HTTP/1.0\r\n\r\n"
	_, err := conn.Write([]byte(request))
	if err != nil {
		return ""
	}
	
	// Читаем ответ
	reader := bufio.NewReader(conn)
	response := make([]byte, s.maxBannerSize)
	n, _ := reader.Read(response)
	
	if n > 0 {
		return strings.TrimSpace(string(response[:n]))
	}
	
	return ""
}

// identifyService пытается идентифицировать сервис по баннеру
func (s *tcpScanner) identifyService(port string, banner string) (service string, version string) {
	bannerLower := strings.ToLower(banner)
	
	// SSH
	if strings.Contains(bannerLower, "ssh") {
		service = "ssh"
		if parts := strings.Fields(banner); len(parts) > 0 {
			version = parts[0]
		}
		return
	}
	
	// HTTP/HTTPS
	if strings.Contains(bannerLower, "http") {
		service = "http"
		if strings.Contains(bannerLower, "apache") {
			service = "apache"
			if idx := strings.Index(bannerLower, "apache"); idx != -1 {
				parts := strings.Fields(banner[idx:])
				if len(parts) > 1 {
					version = parts[1]
				}
			}
		} else if strings.Contains(bannerLower, "nginx") {
			service = "nginx"
			if idx := strings.Index(bannerLower, "nginx"); idx != -1 {
				parts := strings.Fields(banner[idx:])
				if len(parts) > 0 {
					version = strings.TrimPrefix(parts[0], "nginx/")
				}
			}
		}
		return
	}
	
	// FTP
	if strings.HasPrefix(bannerLower, "220") && strings.Contains(bannerLower, "ftp") {
		service = "ftp"
		if parts := strings.Fields(banner); len(parts) > 1 {
			version = strings.Join(parts[1:], " ")
		}
		return
	}
	
	// SMTP
	if strings.HasPrefix(bannerLower, "220") && strings.Contains(bannerLower, "smtp") {
		service = "smtp"
		return
	}
	
	// MySQL
	if strings.Contains(bannerLower, "mysql") {
		service = "mysql"
		return
	}
	
	// PostgreSQL
	if strings.Contains(bannerLower, "postgresql") {
		service = "postgresql"
		return
	}
	
	// Redis
	if strings.HasPrefix(bannerLower, "-err") || strings.HasPrefix(bannerLower, "+pong") {
		service = "redis"
		return
	}
	
	// Telnet
	if len(banner) > 0 && (banner[0] == 0xFF || strings.Contains(bannerLower, "telnet")) {
		service = "telnet"
		return
	}
	
	// По умолчанию - неизвестный сервис
	service = "unknown"
	return
}

// getServiceByPort возвращает известный сервис по номеру порта
func (s *tcpScanner) getServiceByPort(port string) string {
	services := map[string]string{
		"21":    "ftp",
		"22":    "ssh",
		"23":    "telnet",
		"25":    "smtp",
		"53":    "dns",
		"80":    "http",
		"110":   "pop3",
		"143":   "imap",
		"443":   "https",
		"445":   "smb",
		"3306":  "mysql",
		"3389":  "rdp",
		"5432":  "postgresql",
		"5900":  "vnc",
		"6379":  "redis",
		"8080":  "http-proxy",
		"8443":  "https-alt",
		"27017": "mongodb",
	}
	
	if service, ok := services[port]; ok {
		return service
	}
	
	return "unknown"
}

