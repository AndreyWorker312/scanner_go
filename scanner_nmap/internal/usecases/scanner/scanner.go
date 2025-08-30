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
	result, err := UDPScan(context.Background(), "google.com", "53,67,68,123,161,162,500,4500")
	if err != nil {
		fmt.Println("Это печально:", err)
	}

	PrintUDPScanResults(result)
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
			// Считаем open|filtered как потенциально открытые
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
