package nmap_wrapper

import (
	"context"
	"fmt"
	"github.com/Ullaakut/nmap/v3"
	"log"
	"time"
)

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
