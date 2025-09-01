package usecases

import (
	"context"
	"fmt"
	"log"
	"scanner_nmap/internal/domain"
	"scanner_nmap/internal/usecases/nmap_wrapper"
	"time"

	"github.com/Ullaakut/nmap/v3"
)

func main() {
	ctx := context.Background()

	// Пример вызова функции
	host, err, info := UdpTcpScanner(ctx, "TCP", "scanme.nmap.org", "22,80,443")

	// Вывод результатов
	fmt.Println("=== РЕЗУЛЬТАТЫ СКАНИРОВАНИЯ ===")
	fmt.Printf("Хост: %s\n", host)
	fmt.Printf("Ошибка: %v\n", err)

	fmt.Println("\n=== ИНФОРМАЦИЯ О ПОРТАХ ===")
	fmt.Printf("Status: %v (тип: %T)\n", info.Status, info.Status)
	fmt.Printf("OpenPorts: %v (тип: %T)\n", info.OpenPorts, info.OpenPorts)
	fmt.Printf("Protocols: %v (тип: %T)\n", info.Protocols, info.Protocols)
	fmt.Printf("State: %v (тип: %T)\n", info.State, info.State)
	fmt.Printf("ServiceName: %v (тип: %T)\n", info.ServiceName, info.ServiceName)

	// Детальный вывод
	fmt.Println("\n=== ДЕТАЛЬНАЯ ИНФОРМАЦИЯ ===")
	for i := 0; i < len(info.OpenPorts); i++ {
		fmt.Printf("Порт %d: %s %s - %s\n",
			info.OpenPorts[i],
			info.Protocols[i],
			info.State[i],
			info.ServiceName[i])
	}

	fmt.Println("\n=== СТАТУС ХОСТА ===")
	for i, status := range info.Status {
		fmt.Printf("Хост %d: %s\n", i+1, status)
	}

	fmt.Println("\n\n\n\n\n----------------------------\n\n\n\n")
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
func UdpTcpScanner(ctx context.Context, scannerType string, target string, ports string) (host string, err error, info domain.PortTcpUdpInfo) {
	var scanResult *nmap.Run
	var hostResult string

	if scannerType == "UDP" {
		scanResult, err = nmap_wrapper.UDPScan(ctx, target, ports)
	} else {
		scanResult, err = nmap_wrapper.TCPScan(ctx, target, ports)
	}

	if scanResult == nil {
		fmt.Println("Scanner don't have any results")
		return "", err, domain.PortTcpUdpInfo{}
	}

	for _, hosts := range scanResult.Hosts {
		if len(hosts.Addresses) > 0 {
			hostResult = hosts.Addresses[0].String()
			break
		}
	}

	infoResults := domain.PortTcpUdpInfo{}

	for _, hosts := range scanResult.Hosts {
		if len(hosts.Addresses) == 0 {
			continue
		}

		infoResults.Status = append(infoResults.Status, hosts.Status.State)

		for _, port := range hosts.Ports {
			if port.State.State == "open" {
				infoResults.OpenPorts = append(infoResults.OpenPorts, port.ID)
			}

			infoResults.Protocols = append(infoResults.Protocols, port.Protocol)

			infoResults.State = append(infoResults.State, port.State.State)

			serviceName := port.Service.Name
			if serviceName == "" {
				serviceName = "unknown"
			}
			infoResults.ServiceName = append(infoResults.ServiceName, serviceName)
		}
	}

	return hostResult, err, infoResults
}
