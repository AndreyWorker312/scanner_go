package scanner

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"
)

// mockLogger реализация интерфейса Logger
type mockLogger struct{}

func (m *mockLogger) Infof(format string, args ...interface{})  {}
func (m *mockLogger) Errorf(format string, args ...interface{}) {}

// TestPortScanner - тесты для реального сканера
func TestPortScanner(t *testing.T) {
	tests := []struct {
		name    string
		ip      string
		ports   string
		want    []int
		wantErr bool
	}{
		{
			name:    "single port",
			ip:      "127.0.0.1",
			ports:   "8080", // Используем высокий порт для тестов
			want:    []int{8080},
			wantErr: false,
		},
		{
			name:    "port range",
			ip:      "127.0.0.1",
			ports:   "8080-8082",
			want:    []int{8080, 8081, 8082},
			wantErr: false,
		},
		{
			name:    "invalid port",
			ip:      "127.0.0.1",
			ports:   "99999",
			want:    nil,
			wantErr: true,
		},
	}

	scanner := NewPortScanner(&mockLogger{}, 1*time.Second, 3, 100*time.Millisecond)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Запускаем тестовый сервер для портов, которые должны быть открыты
			if !tt.wantErr {
				for _, port := range tt.want {
					stop, err := startTestServer(port)
					if err != nil {
						t.Fatalf("Failed to start test server on port %d: %v", port, err)
					}
					defer stop()
				}
			}

			ctx := context.Background()
			got, err := scanner.ScanPorts(ctx, tt.ip, tt.ports)
			if (err != nil) != tt.wantErr {
				t.Errorf("ScanPorts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !comparePorts(got, tt.want) {
				t.Errorf("ScanPorts() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestPortScannerWithMock - тесты с mock-реализацией сканера
func TestPortScannerWithMock(t *testing.T) {
	tests := []struct {
		name        string
		openPorts   map[int]bool
		portsToScan string
		expected    []int
		wantErr     bool
	}{
		{
			name: "all ports open",
			openPorts: map[int]bool{
				80: true,
				81: true,
				82: true,
			},
			portsToScan: "80-82",
			expected:    []int{80, 81, 82},
			wantErr:     false,
		},
		{
			name: "some ports closed",
			openPorts: map[int]bool{
				80: true,
				81: false,
				82: true,
			},
			portsToScan: "80-82",
			expected:    []int{80, 82},
			wantErr:     false,
		},
		{
			name: "invalid port range",
			openPorts: map[int]bool{
				80: true,
			},
			portsToScan: "80-",
			expected:    nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockScanner{
				openPorts: tt.openPorts,
			}

			ports, err := mock.ScanPorts(context.Background(), "127.0.0.1", tt.portsToScan)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ScanPorts() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !comparePorts(ports, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, ports)
			}
		})
	}
}

// mockScanner - mock-реализация интерфейса PortScanner
type mockScanner struct {
	openPorts map[int]bool
}

func (m *mockScanner) ScanPorts(ctx context.Context, ip string, ports string) ([]int, error) {
	portList, err := parsePorts(ports)
	if err != nil {
		return nil, err
	}

	var result []int
	for _, port := range portList {
		if m.openPorts[port] {
			result = append(result, port)
		}
	}
	return result, nil
}

// Вспомогательные функции

// startTestServer запускает тестовый TCP-сервер на указанном порту
func startTestServer(port int) (func(), error) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()

	return func() { ln.Close() }, nil
}

// comparePorts сравнивает два списка портов
func comparePorts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
