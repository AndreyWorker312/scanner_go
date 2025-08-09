package scanner

import (
	"context"
	"errors"
	"fmt"
	"github.com/mdlayher/arp"
	"net"
	"net/netip"
	"strings"
	"sync"
	"time"
)

const (
	DefaultTimeout    = 2 * time.Second
	DefaultMaxRetries = 2
	DefaultRetryDelay = 500 * time.Millisecond
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
	iface, err := net.InterfaceByName(s.ifaceName)
	if err != nil {
		return nil, fmt.Errorf("interface not found: %w", err)
	}

	ips, err := parseIPRange(ipRange)
	if err != nil {
		return nil, fmt.Errorf("parse IP range failed: %w", err)
	}

	client, err := arp.Dial(iface)
	if err != nil {
		return nil, fmt.Errorf("failed to open ARP client: %w", err)
	}
	defer client.Close()

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
					if err := client.SetReadDeadline(time.Now().Add(s.timeout)); err != nil {
						break
					}
					mac, err = client.Resolve(ip)
					if err == nil && mac != nil {
						success = true
						break
					}
					time.Sleep(s.retryDelay)
				}

				status := "offline"
				macStr := ""
				if success && mac != nil {
					status = "online"
					macStr = mac.String()
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
