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

type ProgressReporter interface {
	ReportProgress(progress float64, scanned, total int)
}

type PortScanner interface {
	ScanPorts(ctx context.Context, ip string, ports string, reporter ProgressReporter) ([]int, error)
}

type portScanner struct {
	timeout    time.Duration
	logger     Logger
	maxRetries int
	retryDelay time.Duration
}

type Logger interface {
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

func NewPortScanner(logger Logger, timeout time.Duration, maxRetries int, retryDelay time.Duration) PortScanner {
	return &portScanner{
		timeout:    timeout,
		logger:     logger,
		maxRetries: maxRetries,
		retryDelay: retryDelay,
	}
}

func (s *portScanner) ScanPorts(ctx context.Context, ip, ports string, reporter ProgressReporter) ([]int, error) {
	portList, err := parsePorts(ports)
	if err != nil {
		return nil, err
	}

	var (
		openPorts []int
		mu        sync.Mutex
		wg        sync.WaitGroup
		total     = len(portList)
	)

	for i, port := range portList {
		wg.Add(1)

		go func(port, i int) {
			defer wg.Done()

			if isPortOpen(ctx, ip, port, s.timeout, s.maxRetries, s.retryDelay) {
				mu.Lock()
				openPorts = append(openPorts, port)
				mu.Unlock()
			}

			// Отправляем прогресс каждые 5%
			if (i+1)%(total/20) == 0 || i == total-1 {
				progress := float64(i+1) / float64(total)
				if reporter != nil {
					reporter.ReportProgress(progress, i+1, total)
				}
			}
		}(port, i)
	}

	wg.Wait()
	return openPorts, nil
}

func isPortOpen(ctx context.Context, ip string, port int, timeout time.Duration, maxRetries int, retryDelay time.Duration) bool {
	var attempts int
	for {
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
			attempts++
			if attempts >= maxRetries {
				return false
			}
			time.Sleep(retryDelay)
		}
	}
}

func parsePorts(ports string) ([]int, error) {
	if ports == "" {
		return nil, fmt.Errorf("ports string is empty")
	}

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

	port, err := strconv.Atoi(ports)
	if err != nil {
		return nil, fmt.Errorf("invalid port number: %v", err)
	}
	if port < 1 || port > 65535 {
		return nil, fmt.Errorf("port %d is out of range (1-65535)", port)
	}

	return []int{port}, nil
}
