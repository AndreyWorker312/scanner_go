package usecases

import (
	"context"
	"fmt"
	"scanner_nmap/internal/domain"
	"scanner_nmap/internal/usecases/nmap_wrapper"

	"github.com/Ullaakut/nmap/v3"
)

func UdpTcpScanner(ctx context.Context, request domain.ScanTcpUdpRequest) (response domain.ScanTcpUdpResponse, err error) {
	var scanResult *nmap.Run

	// Проверяем разные варианты UDP сканирования
	if request.ScannerType == "UDP" || request.ScannerType == "udp_scan" {
		scanResult, err = nmap_wrapper.UDPScan(ctx, request.IP, request.Ports)
	} else {
		scanResult, err = nmap_wrapper.TCPScan(ctx, request.IP, request.Ports)
	}

	if scanResult == nil {
		fmt.Println("Scanner doesn't have any results")
		return domain.ScanTcpUdpResponse{}, err
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
			portInfo.AllPorts = append(portInfo.AllPorts, port.ID)
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
		TaskID:   request.TaskID,
		Host:     hostResult,
		PortInfo: []domain.PortTcpUdpInfo{portInfo},
	}

	return responseResult, err
}

func OSDetectionScanner(ctx context.Context, request domain.OsDetectionRequest) (response domain.OsDetectionResponse, err error) {
	scanResult, err := nmap_wrapper.OSDetectionScan(ctx, request.IP)

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
		TaskID:   request.TaskID,
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

func HostDiscoveryScanner(ctx context.Context, request domain.HostDiscoveryRequest) (response domain.HostDiscoveryResponse, err error) {
	fmt.Printf("Starting host discovery for %s\n", request.IP)
	scanResult, err := nmap_wrapper.HostDiscovery(ctx, request.IP)

	if scanResult == nil {
		fmt.Println("Host discovery scanner doesn't have any results")
		return domain.HostDiscoveryResponse{}, err
	}

	fmt.Printf("Host discovery found %d hosts\n", len(scanResult.Hosts))

	var hostResult string
	for _, host := range scanResult.Hosts {
		if len(host.Addresses) > 0 {
			hostResult = host.Addresses[0].String()
			break
		}
	}

	// Если не нашли хост в результатах, используем исходный IP из запроса
	if hostResult == "" {
		hostResult = request.IP
	}

	responseResult := domain.HostDiscoveryResponse{
		TaskID:    request.TaskID,
		Host:      hostResult,
		HostUP:    0,
		HostTotal: len(scanResult.Hosts),
		Status:    "unknown",
		DNS:       "unknown",
		Reason:    "unknown",
	}

	for i, host := range scanResult.Hosts {
		fmt.Printf("Host %d: Addresses=%d, Status=%s, Reason=%s\n", i, len(host.Addresses), host.Status.State, host.Status.Reason)

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

	fmt.Printf("Final result: Host=%s, UP=%d, Total=%d, Status=%s\n", responseResult.Host, responseResult.HostUP, responseResult.HostTotal, responseResult.Status)

	return responseResult, err
}
