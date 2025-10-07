package scanner

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/netip"
	"strings"
	"sync"
	"time"

	"github.com/mdlayher/arp"
)

const (
	DefaultTimeout    = 5 * time.Second
	DefaultMaxRetries = 3
	DefaultRetryDelay = 1 * time.Second
)

type DeviceInfo struct {
	IP     string `json:"ip"`
	MAC    string `json:"mac"`
	Vendor string `json:"vendor,omitempty"`
	Status string `json:"status"`
}

type ARPScanner interface {
	Scan(ctx context.Context, ipRange string) ([]DeviceInfo, error)
}

type arpScanner struct {
	ifaceName  string
	timeout    time.Duration
	maxRetries int
	retryDelay time.Duration
}

func NewARPScanner(ifaceName string, timeout time.Duration, maxRetries int, retryDelay time.Duration) ARPScanner {
	return &arpScanner{
		ifaceName:  ifaceName,
		timeout:    timeout,
		maxRetries: maxRetries,
		retryDelay: retryDelay,
	}
}

func (s *arpScanner) Scan(ctx context.Context, ipRange string) ([]DeviceInfo, error) {
	log.Printf("Starting ARP scan on interface %s for range %s", s.ifaceName, ipRange)

	iface, err := net.InterfaceByName(s.ifaceName)
	if err != nil {
		return nil, fmt.Errorf("interface not found: %w", err)
	}
	log.Printf("Found interface: %s, MAC: %s", iface.Name, iface.HardwareAddr)

	ips, err := parseIPRange(ipRange)
	if err != nil {
		return nil, fmt.Errorf("parse IP range failed: %w", err)
	}
	log.Printf("Parsed %d IP addresses to scan", len(ips))

	client, err := arp.Dial(iface)
	if err != nil {
		return nil, fmt.Errorf("failed to open ARP client: %w", err)
	}
	defer client.Close()
	log.Printf("ARP client opened successfully")

	var (
		results []DeviceInfo
		mu      sync.Mutex
		wg      sync.WaitGroup
	)

	for _, ip := range ips {
		wg.Add(1)

		go func(ip netip.Addr) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				return
			default:
				var mac net.HardwareAddr
				var success bool

				for attempt := 0; attempt < s.maxRetries; attempt++ {
					// Устанавливаем более короткий таймаут для чтения
					if err := client.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
						log.Printf("Failed to set deadline for %s: %v", ip, err)
						break
					}
					mac, err = client.Resolve(ip)
					if err == nil && mac != nil {
						log.Printf("Successfully resolved %s to %s", ip, mac)
						success = true
						break
					} else if err != nil {
						// Просто логируем ошибку, не пытаемся извлечь MAC из таймаута
						log.Printf("Failed to resolve %s (attempt %d): %v", ip, attempt+1, err)
					}
					time.Sleep(s.retryDelay)
				}

				status := "offline"
				macStr := ""
				if success && mac != nil {
					status = "online"
					macStr = mac.String()
					log.Printf("Successfully resolved %s to %s", ip, macStr)
				} else {
					// Не считаем устройство онлайн, если получили только таймаут с MAC роутера
					status = "offline"
					macStr = ""
					log.Printf("Device %s is offline (timeout or no response)", ip)
				}

				device := DeviceInfo{
					IP:     ip.String(),
					MAC:    macStr,
					Status: status,
				}

				mu.Lock()
				results = append(results, device)
				mu.Unlock()
			}
		}(ip)
	}

	wg.Wait()
	return results, nil
}

func parseIPRange(ipRange string) ([]netip.Addr, error) {
	if strings.Contains(ipRange, "/") {
		prefix, err := netip.ParsePrefix(ipRange)
		if err != nil {
			return nil, fmt.Errorf("invalid CIDR: %w", err)
		}
		var ips []netip.Addr
		for ip := prefix.Masked().Addr(); prefix.Contains(ip); ip = ip.Next() {
			ips = append(ips, ip)
		}
		if prefix.Addr().Is4() && len(ips) > 2 {
			return ips[1 : len(ips)-1], nil
		}
		return ips, nil
	} else if strings.Contains(ipRange, "-") {
		parts := strings.Split(ipRange, "-")
		if len(parts) != 2 {
			return nil, errors.New("invalid IP range format")
		}
		start, err := netip.ParseAddr(strings.TrimSpace(parts[0]))
		if err != nil {
			return nil, fmt.Errorf("invalid start IP: %w", err)
		}
		end, err := netip.ParseAddr(strings.TrimSpace(parts[1]))
		if err != nil {
			return nil, fmt.Errorf("invalid end IP: %w", err)
		}
		if start.Compare(end) > 0 {
			return nil, errors.New("start IP is after end IP")
		}
		var ips []netip.Addr
		for ip := start; ip.Compare(end) <= 0; ip = ip.Next() {
			ips = append(ips, ip)
		}
		return ips, nil
	} else {
		ip, err := netip.ParseAddr(strings.TrimSpace(ipRange))
		if err != nil {
			return nil, fmt.Errorf("invalid IP address: %w", err)
		}
		return []netip.Addr{ip}, nil
	}
}
