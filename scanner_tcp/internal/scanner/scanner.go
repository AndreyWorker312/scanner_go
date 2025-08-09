package scanner

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type PortScanner interface {
	ScanPorts(ctx context.Context, ip string, ports string) ([]int, error)
}

type portScanner struct {
	timeout    time.Duration
	maxRetries int
	retryDelay time.Duration
}

func NewPortScanner(timeout time.Duration, maxRetries int, retryDelay time.Duration) PortScanner {
	return &portScanner{
		timeout:    timeout,
		maxRetries: maxRetries,
		retryDelay: retryDelay,
	}
}

func (s *portScanner) ScanPorts(ctx context.Context, ip, ports string) ([]int, error) {
	portList, err := parsePorts(ports)
	if err != nil {
		return nil, err
	}

	var (
		openPorts []int
		mu        sync.Mutex
		wg        sync.WaitGroup
	)

	for _, port := range portList {
		wg.Add(1)

		go func(p int) {
			defer wg.Done()

			if isPortOpen(ctx, ip, p, s.timeout, s.maxRetries, s.retryDelay) {
				mu.Lock()
				openPorts = append(openPorts, p)
				mu.Unlock()
			}
		}(port)
	}

	wg.Wait()
	return openPorts, nil
}

func isPortOpen(ctx context.Context, ip string, port int, timeout time.Duration, maxRetries int, retryDelay time.Duration) bool {
	for attempt := 0; attempt < maxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return false
		default:
			address := fmt.Sprintf("%s:%d", ip, port)
			conn, err := net.DialTimeout("tcp", address, timeout)
			if err == nil {
				conn.Close()
				return true
			}
			time.Sleep(retryDelay)
		}
	}
	return false
}

func parsePorts(ports string) ([]int, error) {
	if ports == "" {
		return nil, fmt.Errorf("ports string is empty")
	}

	if strings.Contains(ports, "-") {
		return parsePortRange(ports)
	}

	if strings.Contains(ports, ",") {
		return parsePortList(ports)
	}

	return parseSinglePort(ports)
}

func parsePortRange(ports string) ([]int, error) {
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

func parsePortList(ports string) ([]int, error) {
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

func parseSinglePort(ports string) ([]int, error) {
	port, err := strconv.Atoi(ports)
	if err != nil {
		return nil, fmt.Errorf("invalid port number: %v", err)
	}
	if port < 1 || port > 65535 {
		return nil, fmt.Errorf("port %d is out of range (1-65535)", port)
	}
	return []int{port}, nil
}
