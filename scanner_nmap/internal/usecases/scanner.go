package usecases

import (
	"context"
	"fmt"
	"github.com/Ullaakut/nmap/v3"
	"scanner_nmap/internal/domain"
	"scanner_nmap/internal/usecases/nmap_wrapper"
)

/*
	func main() {
		ctx := context.Background()

		scanResponse, err := UdpTcpScanner(ctx, "TCP", "scanme.nmap.org", "22,80,443")

		// Вывод результатов
		fmt.Println("=== РЕЗУЛЬТАТЫ TCP/UDP СКАНИРОВАНИЯ ===")
		fmt.Printf("Ошибка: %v\n", err)

		fmt.Println("\n=== ИНФОРМАЦИЯ О СКАНИРОВАНИИ ПОРТОВ ===")
		fmt.Printf("Host: %s (тип: %T)\n", scanResponse.Host, scanResponse.Host)
		fmt.Printf("Error: %s (тип: %T)\n", scanResponse.Error, scanResponse.Error)

		if len(scanResponse.PortInfo) > 0 {
			info := scanResponse.PortInfo[0]
			fmt.Printf("Status: %v (тип: %T)\n", info.Status, info.Status)
			fmt.Printf("AllPorts: %v (тип: %T)\n", info.AllPorts, info.AllPorts)
			fmt.Printf("Protocols: %v (тип: %T)\n", info.Protocols, info.Protocols)
			fmt.Printf("State: %v (тип: %T)\n", info.State, info.State)
			fmt.Printf("ServiceName: %v (тип: %T)\n", info.ServiceName, info.ServiceName)
		}

		fmt.Println("\n=== ПОЛНЫЙ ОБЪЕКТ ОТВЕТА ===")
		fmt.Printf("ScanTcpUdpResponse: %+v\n", scanResponse)

		fmt.Println("\n=== СВОДНАЯ ИНФОРМАЦИЯ ===")
		fmt.Printf("Целевой хост: %s\n", scanResponse.Host)
		if scanResponse.Error != "" {
			fmt.Printf("Ошибка сканирования: %s\n", scanResponse.Error)
		}

		fmt.Println("\n\n\n\n\n----------------------------\n\n\n\n")

		// Пример вызова функции определения ОС
		osResponse, err := OSDetectionScanner(ctx, "scanme.nmap.org")

		// Вывод результатов
		fmt.Println("=== РЕЗУЛЬТАТЫ ОПРЕДЕЛЕНИЯ ОС ===")
		fmt.Printf("Ошибка: %v\n", err)

		fmt.Println("\n=== ИНФОРМАЦИЯ ОБ ОПЕРАЦИОННОЙ СИСТЕМЕ ===")
		fmt.Printf("Host: %s (тип: %T)\n", osResponse.Host, osResponse.Host)
		fmt.Printf("Name: %s (тип: %T)\n", osResponse.Name, osResponse.Name)
		fmt.Printf("Accuracy: %d (тип: %T)\n", osResponse.Accuracy, osResponse.Accuracy)
		fmt.Printf("Vendor: %s (тип: %T)\n", osResponse.Vendor, osResponse.Vendor)
		fmt.Printf("Family: %s (тип: %T)\n", osResponse.Family, osResponse.Family)
		fmt.Printf("Type: %s (тип: %T)\n", osResponse.Type, osResponse.Type)

		fmt.Println("\n=== ПОЛНЫЙ ОБЪЕКТ ОТВЕТА ===")
		fmt.Printf("OsDetectionResponse: %+v\n", osResponse)

		fmt.Println("\n=== СВОДНАЯ ИНФОРМАЦИЯ ===")
		fmt.Printf("Хост: %s\n", osResponse.Host)
		fmt.Printf("Операционная система: %s\n", osResponse.Name)
		fmt.Printf("Точность определения: %d%%\n", osResponse.Accuracy)
		fmt.Printf("Производитель: %s\n", osResponse.Vendor)
		fmt.Printf("Семейство ОС: %s\n", osResponse.Family)
		fmt.Printf("Тип ОС: %s\n", osResponse.Type)

		if err != nil {
			fmt.Printf("Ошибка сканирования: %v\n", err)
		}

		fmt.Println("\n\n\n\n\n----------------------------\n\n\n\n")

		discoveryInfo, err := HostDiscoveryScanner(ctx, "scanme.nmap.org")

		// Вывод результатов
		fmt.Println("=== РЕЗУЛЬТАТЫ ОБНАРУЖЕНИЯ ХОСТОВ ===")
		fmt.Printf("Ошибка: %v\n", err)

		fmt.Println("\n=== ИНФОРМАЦИЯ ОБ ОБНАРУЖЕНИИ ХОСТОВ ===")
		fmt.Printf("Host: %s (тип: %T)\n", discoveryInfo.Host, discoveryInfo.Host)
		fmt.Printf("HostUP: %d (тип: %T)\n", discoveryInfo.HostUP, discoveryInfo.HostUP)
		fmt.Printf("HostTotal: %d (тип: %T)\n", discoveryInfo.HostTotal, discoveryInfo.HostTotal)
		fmt.Printf("Status: %s (тип: %T)\n", discoveryInfo.Status, discoveryInfo.Status)
		fmt.Printf("DNS: %s (тип: %T)\n", discoveryInfo.DNS, discoveryInfo.DNS)
		fmt.Printf("Reason: %s (тип: %T)\n", discoveryInfo.Reason, discoveryInfo.Reason)

		fmt.Println("\n=== ПОЛНЫЙ ОБЪЕКТ ОТВЕТА ===")
		fmt.Printf("HostDiscoveryResponse: %+v\n", discoveryInfo)

		fmt.Println("\n=== СВОДНАЯ ИНФОРМАЦИЯ ===")
		fmt.Printf("Основной хост: %s\n", discoveryInfo.Host)
		fmt.Printf("Обнаружено хостов: %d/%d\n", discoveryInfo.HostUP, discoveryInfo.HostTotal)
		fmt.Printf("Статус основного хоста: %s\n", discoveryInfo.Status)
		if discoveryInfo.DNS != "unknown" {
			fmt.Printf("DNS имя: %s\n", discoveryInfo.DNS)
		}
		fmt.Printf("Причина статуса: %s\n", discoveryInfo.Reason)

}
*/
func UdpTcpScanner(ctx context.Context, scannerType string, target string, ports string) (response domain.ScanTcpUdpResponse, err error) {
	var scanResult *nmap.Run

	if scannerType == "UDP" {
		scanResult, err = nmap_wrapper.UDPScan(ctx, target, ports)
	} else {
		scanResult, err = nmap_wrapper.TCPScan(ctx, target, ports)
	}

	if scanResult == nil {
		fmt.Println("Scanner doesn't have any results")
		return domain.ScanTcpUdpResponse{
			Error: "No scan results",
		}, err
	}

	var hostResult string
	for _, host := range scanResult.Hosts {
		if len(host.Addresses) > 0 {
			hostResult = host.Addresses[0].String()
			break
		}
	}

	portInfo := domain.PortTcpUdpInfo{}

	for _, host := range scanResult.Hosts {
		if len(host.Addresses) == 0 {
			continue
		}

		portInfo.Status = host.Status.State

		for _, port := range host.Ports {
			portInfo.AllPorts = append(portInfo.AllPorts, uint16(port.ID))
			portInfo.Protocols = append(portInfo.Protocols, port.Protocol)
			portInfo.State = append(portInfo.State, port.State.State)

			serviceName := port.Service.Name
			if serviceName == "" {
				serviceName = "unknown"
			}
			portInfo.ServiceName = append(portInfo.ServiceName, serviceName)
		}
	}

	responseResult := domain.ScanTcpUdpResponse{
		Host:     hostResult,
		PortInfo: []domain.PortTcpUdpInfo{portInfo},
	}

	if err != nil {
		responseResult.Error = err.Error()
	}

	return responseResult, err
}

func OSDetectionScanner(ctx context.Context, target string) (response domain.OsDetectionResponse, err error) {
	scanResult, err := nmap_wrapper.OSDetectionScan(ctx, target)

	if scanResult == nil {
		fmt.Println("OS detection scanner doesn't have any results")
		return domain.OsDetectionResponse{}, err
	}

	// Находим первый хост с адресом
	var hostResult string
	for _, hostItem := range scanResult.Hosts {
		if len(hostItem.Addresses) > 0 {
			hostResult = hostItem.Addresses[0].String()
			break
		}
	}

	// Создаем ответ по умолчанию
	responseResult := domain.OsDetectionResponse{
		Host:     hostResult,
		Name:     "unknown",
		Accuracy: 0,
		Vendor:   "unknown",
		Family:   "unknown",
		Type:     "unknown",
	}

	// Заполняем информацию об ОС из результатов сканирования
	for _, hostItem := range scanResult.Hosts {
		if len(hostItem.Addresses) == 0 {
			continue
		}

		if len(hostItem.OS.Matches) > 0 {
			osMatch := hostItem.OS.Matches[0]
			responseResult.Name = osMatch.Name
			responseResult.Accuracy = osMatch.Accuracy

			if len(osMatch.Classes) > 0 {
				osClass := osMatch.Classes[0]
				responseResult.Vendor = osClass.Vendor
				responseResult.Family = osClass.Family
				responseResult.Type = osClass.Type
			}
			break
		}
	}

	return responseResult, err
}

func HostDiscoveryScanner(ctx context.Context, target string) (response domain.HostDiscoveryResponse, err error) {
	scanResult, err := nmap_wrapper.HostDiscovery(ctx, target)

	if scanResult == nil {
		fmt.Println("Host discovery scanner doesn't have any results")
		return domain.HostDiscoveryResponse{}, err
	}

	var hostResult string
	for _, host := range scanResult.Hosts {
		if len(host.Addresses) > 0 {
			hostResult = host.Addresses[0].String()
			break
		}
	}

	responseResult := domain.HostDiscoveryResponse{
		Host:      hostResult,
		HostUP:    0,
		HostTotal: len(scanResult.Hosts),
		Status:    "unknown",
		DNS:       "unknown",
		Reason:    "unknown",
	}

	for _, host := range scanResult.Hosts {
		if len(host.Addresses) == 0 {
			continue
		}

		if host.Status.State == "up" {
			responseResult.HostUP++
		}

		if hostResult == host.Addresses[0].String() {
			responseResult.Status = host.Status.State
			responseResult.Reason = host.Status.Reason

			if len(host.Hostnames) > 0 {
				responseResult.DNS = host.Hostnames[0].Name
			}
		}
	}

	return responseResult, err
}
