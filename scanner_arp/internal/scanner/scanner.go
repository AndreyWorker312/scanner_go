package scanner

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/netip"
	"os"
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

	ips, err := parseIPRange(ipRange)
	if err != nil {
		return nil, fmt.Errorf("parse IP range failed: %w", err)
	}
	log.Printf("Scanning %d IP addresses...", len(ips))

	var (
		results   = make(map[string]DeviceInfo)
		resultsMu sync.Mutex
	)

	var wg sync.WaitGroup
	const maxConcurrency = 200
	semaphore := make(chan struct{}, maxConcurrency)

	requestTimeout := 200 * time.Millisecond

	for _, ip := range ips {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			wg.Add(1)
			semaphore <- struct{}{}

			go func(targetIP netip.Addr) {
				defer wg.Done()
				defer func() { <-semaphore }()

				client, err := arp.Dial(iface)
				if err != nil {
					return
				}
				defer client.Close()

				client.SetReadDeadline(time.Now().Add(requestTimeout))
				mac, err := client.Resolve(targetIP)

				if err == nil && mac != nil {
					ipStr := targetIP.String()
					macStr := mac.String()

					resultsMu.Lock()
					results[ipStr] = DeviceInfo{
						IP:     ipStr,
						MAC:    macStr,
						Vendor: LookupVendor(macStr),
						Status: "online",
					}
					resultsMu.Unlock()
				}
			}(ip)
		}
	}

	wg.Wait()

	requestedIPs := make(map[string]bool, len(ips))
	for _, ip := range ips {
		requestedIPs[ip.String()] = true
	}

	systemDevices := readSystemARPTable(iface)
	resultsMu.Lock()
	for ip, mac := range systemDevices {
		if requestedIPs[ip] {
			if _, exists := results[ip]; !exists {
				results[ip] = DeviceInfo{
					IP:     ip,
					MAC:    mac,
					Vendor: LookupVendor(mac),
					Status: "online",
				}
			}
		}
	}
	resultsMu.Unlock()

	var devices []DeviceInfo
	resultsMu.Lock()
	for _, ip := range ips {
		ipStr := ip.String()
		if device, exists := results[ipStr]; exists {
			devices = append(devices, device)
		} else {
			devices = append(devices, DeviceInfo{
				IP:     ipStr,
				MAC:    "",
				Status: "offline",
			})
		}
	}
	resultsMu.Unlock()

	onlineCount := 0
	for _, d := range devices {
		if d.Status == "online" {
			onlineCount++
		}
	}
	log.Printf("Scan completed. Found %d devices (%d online, %d offline)", len(devices), onlineCount, len(devices)-onlineCount)
	return devices, nil
}

func readSystemARPTable(iface *net.Interface) map[string]string {
	devices := make(map[string]string)

	file, err := os.Open("/proc/net/arp")
	if err != nil {
		return devices
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		scanner.Text()
	}

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) >= 6 {
			ip := fields[0]
			mac := fields[3]
			device := fields[5]

			if device == iface.Name && mac != "00:00:00:00:00:00" && strings.Contains(mac, ":") {
				devices[ip] = mac
			}
		}
	}

	return devices
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
