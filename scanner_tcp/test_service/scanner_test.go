package test_service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"scanner/internal/scanner"
)

func TestScanPorts(t *testing.T) {
	// Создаем тестовый сканер с параметрами:
	// timeout = 1s, maxRetries = 3, retryDelay = 100ms
	s := scanner.NewPortScanner(1*time.Second, 3, 100*time.Millisecond)

	tests := []struct {
		name      string
		ip        string
		ports     string
		wantPorts []int
		wantErr   bool
	}{
		{
			name:      "single open port",
			ip:        "127.0.0.1",
			ports:     "80",
			wantPorts: []int{80}, // предполагаем, что 80 порт открыт
		},
		{
			name:      "multiple ports",
			ip:        "127.0.0.1",
			ports:     "80,443,8080",
			wantPorts: []int{80, 443}, // предполагаем, что 80 и 443 открыты
		},
		{
			name:      "port range",
			ip:        "127.0.0.1",
			ports:     "80-82",
			wantPorts: []int{80}, // предполагаем, что только 80 открыт в этом диапазоне
		},
		{
			name:    "invalid port format",
			ip:      "127.0.0.1",
			ports:   "abc",
			wantErr: true,
		},
		{
			name:    "invalid port number",
			ip:      "127.0.0.1",
			ports:   "99999",
			wantErr: true,
		},
		{
			name:      "non-existent ip",
			ip:        "192.0.2.0", // TEST-NET-1, не должен существовать
			ports:     "80",
			wantPorts: []int{}, // не должно быть открытых портов
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			t.Logf("Starting scan for IP: %s, ports: %s", tt.ip, tt.ports)

			got, err := s.ScanPorts(ctx, tt.ip, tt.ports)
			if (err != nil) != tt.wantErr {
				t.Errorf("ScanPorts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Выводим информацию об открытых портах
				if len(got) > 0 {
					t.Logf("Found %d open port(s) for IP %s:", len(got), tt.ip)
					for _, port := range got {
						t.Logf(" - Port %d is open", port)
					}
				} else {
					t.Logf("No open ports found for IP %s", tt.ip)
				}

				if len(got) != len(tt.wantPorts) {
					t.Errorf("ScanPorts() got %v ports, want %v", got, tt.wantPorts)
				} else {
					for i, port := range got {
						if port != tt.wantPorts[i] {
							t.Errorf("ScanPorts() got %v, want %v", got, tt.wantPorts)
							break
						}
					}
				}
			} else {
				t.Logf("Expected error occurred: %v", err)
			}
		})
	}
}

// Дополнительная функция для более наглядного вывода при запуске тестов
func TestScanPortsWithDetailedOutput(t *testing.T) {
	s := scanner.NewPortScanner(1*time.Second, 3, 100*time.Millisecond)
	ip := "chat.deepseek.com"
	ports := "80,443,1802"

	fmt.Printf("\n=== Detailed port scan test ===\n")
	fmt.Printf("Scanning IP: %s\n", ip)
	fmt.Printf("Ports to scan: %s\n", ports)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	openPorts, err := s.ScanPorts(ctx, ip, ports)
	if err != nil {
		fmt.Printf("Scan failed: %v\n", err)
		t.Fatalf("Scan failed: %v", err)
	}

	fmt.Printf("\nScan results for %s:\n", ip)
	if len(openPorts) == 0 {
		fmt.Println("No open ports found")
	} else {
		fmt.Println("Open ports:")
		for _, port := range openPorts {
			fmt.Printf(" - %d\n", port)
		}
	}
	fmt.Println("============================")
}
