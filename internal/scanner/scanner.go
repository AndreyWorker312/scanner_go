package scanner

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

type PortScanner interface {
	ScanPorts(ctx context.Context, ip string, ports string) ([]int, error)
}

type portScanner struct {
	timeout time.Duration
	logger  Logger
}

type Logger interface {
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

func NewPortScanner(logger Logger, timeout time.Duration) PortScanner {
	return &portScanner{
		timeout: timeout,
		logger:  logger,
	}
}

func (s *portScanner) ScanPorts(ctx context.Context, ip string, ports string) ([]int, error) {
	portList, err := parsePorts(ports)
	if err != nil {
		return nil, fmt.Errorf("invalid ports specification: %v", err)
	}

	var openPorts []int

	for _, port := range portList {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			address := fmt.Sprintf("%s:%d", ip, port)
			conn, err := net.DialTimeout("tcp", address, s.timeout)
			if err == nil {
				openPorts = append(openPorts, port)
				conn.Close()
			}
		}
	}

	return openPorts, nil
}

func parsePorts(ports string) ([]int, error) {
	if ports == "" {
		return nil, fmt.Errorf("ports string is empty")
	}

	// Обработка диапазона портов (например, "1-1024")
	if strings.Contains(ports, "-") {
		parts := strings.Split(ports, "-")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid port range format")
		}

		start, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, fmt.Errorf("invalid start port: %v", err)
		}

		end, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid end port: %v", err)
		}

		if start > end {
			return nil, fmt.Errorf("start port cannot be greater than end port")
		}

		if start < 1 || end > 65535 {
			return nil, fmt.Errorf("port range must be between 1 and 65535")
		}

		var result []int
		for i := start; i <= end; i++ {
			result = append(result, i)
		}
		return result, nil
	}

	// Обработка списка портов через запятую (например, "80,443,8080")
	if strings.Contains(ports, ",") {
		var result []int
		for _, p := range strings.Split(ports, ",") {
			port, err := strconv.Atoi(strings.TrimSpace(p))
			if err != nil {
				return nil, fmt.Errorf("invalid port number: %v", err)
			}
			if port < 1 || port > 65535 {
				return nil, fmt.Errorf("port %d is out of range (1-65535)", port)
			}
			result = append(result, port)
		}
		return result, nil
	}

	// Одиночный порт
	port, err := strconv.Atoi(ports)
	if err != nil {
		return nil, fmt.Errorf("invalid port number: %v", err)
	}
	if port < 1 || port > 65535 {
		return nil, fmt.Errorf("port %d is out of range (1-65535)", port)
	}

	return []int{port}, nil
}
