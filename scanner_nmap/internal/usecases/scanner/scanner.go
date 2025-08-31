package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Ullaakut/nmap/v3"
)

func main() {
	result1, err := UDPScan(context.Background(), "google.com", "53,67,68,123,161,162,500,4500")
	result2, err := TCPScan(context.Background(), "google.com", "443,8080")
	if err != nil {
		fmt.Println("Это печально:", err)
	}

	PrintUDPScanResults(result1)
	PrintUDPScanResults(result2)

	result3, err := OSDetectionScan(context.Background(), "google.com")

	PrintOSDetectionResults(result3)

	result4, err := HostDiscovery(context.Background(), "google.com")
	PrintHostDiscoveryResults(result4)
}
func UDPScan(ctx context.Context, target string, ports string) (*nmap.Run, error) {
	scanCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	udpScanner, err := nmap.NewScanner(
		scanCtx,
		nmap.WithTargets(target),
		nmap.WithPorts(ports),
		nmap.WithUDPScan(),
		nmap.WithSkipHostDiscovery(),
		nmap.WithTimingTemplate(3),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create UDP scanner: %w", err)
	}

	result, warnings, err := udpScanner.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to run UDP scanner: %w", err)
	}

	if len(*warnings) > 0 {
		log.Printf("Scan warning: %v\n", *warnings)
	}

	return result, nil
}

func TCPScan(ctx context.Context, target string, ports string) (*nmap.Run, error) {
	scanCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	tcpScanner, err := nmap.NewScanner(
		scanCtx,
		nmap.WithTargets(target),
		nmap.WithPorts(ports),
		nmap.WithConnectScan(),
		nmap.WithSkipHostDiscovery(),
		nmap.WithTimingTemplate(3),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create TCP scanner: %w", err)
	}

	result, warnings, err := tcpScanner.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to run TCP scanner: %w", err)
	}

	if len(*warnings) > 0 {
		log.Printf("Scan warning: %v\n", *warnings)
	}

	return result, nil
}

func OSDetectionScan(ctx context.Context, target string) (*nmap.Run, error) {
	scanCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	scanner, err := nmap.NewScanner(
		scanCtx,
		nmap.WithTargets(target),
		nmap.WithOSDetection(),
		nmap.WithTimingTemplate(5),
		nmap.WithSkipHostDiscovery(),
		nmap.WithMaxRetries(0),
		nmap.WithOSScanGuess(),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create OS detection scanner: %w", err)
	}

	result, warnings, err := scanner.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to run OS detection scanner: %w", err)
	}

	if len(*warnings) > 0 {
		log.Printf("OS detection warnings: %v\n", *warnings)
	}

	return result, nil
}

func HostDiscovery(ctx context.Context, target string) (*nmap.Run, error) {
	scanCtx, cancel := context.WithTimeout(ctx, 10*time.Second) // Короткий таймаут
	defer cancel()

	scanner, err := nmap.NewScanner(
		scanCtx,
		nmap.WithTargets(target),
		nmap.WithPingScan(),                 // Только проверка доступности
		nmap.WithTimingTemplate(5),          // Максимальная скорость
		nmap.WithMaxRetries(1),              // Минимум попыток
		nmap.WithHostTimeout(5*time.Second), // Таймаут на хост
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create host discovery scanner: %w", err)
	}

	result, warnings, err := scanner.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to run host discovery scanner: %w", err)
	}

	if len(*warnings) > 0 {
		log.Printf("Host discovery warnings: %v\n", *warnings)
	}

	return result, nil
}

func PrintUDPScanResults(result *nmap.Run) {
	if result == nil {
		fmt.Println("Нет результатов сканирования")
		return
	}

	fmt.Printf("UDP сканирование завершено за %v\n", result.Stats.Finished.Elapsed)
	fmt.Printf("Найдено хостов: %d\n", len(result.Hosts))

	for i, host := range result.Hosts {
		if len(host.Addresses) == 0 {
			continue
		}

		fmt.Printf("\n=== Хост %d: %s ===\n", i+1, host.Addresses[0].String())
		fmt.Printf("Статус: %s\n", host.Status.State)

		if len(host.Ports) == 0 {
			fmt.Println("  Портов не найдено")
			continue
		}

		openPorts := 0
		openFilteredPorts := 0

		for _, port := range host.Ports {

			if strings.Contains(port.State.State, "open") {
				openFilteredPorts++
			}

			fmt.Printf("  Порт: %d/%s - %s", port.ID, port.Protocol, port.State.State)

			if port.Service.Name != "" {
				fmt.Printf(" - Сервис: %s", port.Service.Name)
				if port.Service.Version != "" {
					fmt.Printf(" (%s)", port.Service.Version)
				}
			}
			fmt.Println()

			if port.State.State == "open" {
				openPorts++
			}
		}

		fmt.Printf("  Всего портов: %d\n", len(host.Ports))
		fmt.Printf("  Точно открыто: %d\n", openPorts)
		fmt.Printf("  Возможно открыто: %d\n", openFilteredPorts)

		// Анализ конкретных портов
		fmt.Println("\n  Вероятно работающие сервисы:")
		for _, port := range host.Ports {
			if port.Service.Name != "" && strings.Contains(port.State.State, "open") {
				fmt.Printf("    • Порт %d: %s\n", port.ID, port.Service.Name)
			}
		}
	}
}

func PrintOSDetectionResults(result *nmap.Run) {
	if result == nil {
		fmt.Println("Нет результатов сканирования")
		return
	}

	for _, host := range result.Hosts {
		if len(host.Addresses) == 0 {
			continue
		}

		fmt.Printf("\n🎯 Хост: %s\n", host.Addresses[0].String())

		if len(host.OS.Matches) > 0 {
			bestMatch := host.OS.Matches[0]
			fmt.Printf("🖥️  ОС: %s (%d%%)\n", bestMatch.Name, bestMatch.Accuracy)

			if len(bestMatch.Classes) > 0 {
				class := bestMatch.Classes[0]
				fmt.Printf("📊 %s | %s | %s\n",
					class.Vendor,
					class.Family,
					class.Type)
			}
		} else {
			fmt.Println("❌ ОС не определена")
		}
	}

	fmt.Printf("\n⏱️ Время: %v\n", result.Stats.Finished.Elapsed)
}
func PrintHostDiscoveryResults(result *nmap.Run) {
	if result == nil {
		fmt.Println("Нет результатов сканирования")
		return
	}

	fmt.Printf("Проверка доступности завершена за %v\n", result.Stats.Finished.Elapsed)
	fmt.Printf("Обнаружено хостов: %d из %d\n",
		result.Stats.Hosts.Up,
		result.Stats.Hosts.Total)

	for i, host := range result.Hosts {
		if len(host.Addresses) == 0 {
			continue
		}

		status := "❌ Недоступен"
		if host.Status.State == "up" {
			status = "✅ Доступен"
		}

		fmt.Printf("\n%d. %s - %s\n", i+1, host.Addresses[0].String(), status)

		if len(host.Hostnames) > 0 {
			fmt.Printf("   DNS: %s\n", host.Hostnames[0].Name)
		}

		if host.Status.Reason != "" {
			fmt.Printf("   Причина: %s\n", host.Status.Reason)
		}
	}
}
