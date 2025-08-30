package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Ullaakut/nmap/v3"
)

func main() {
	// Equivalent to `/usr/local/bin/nmap -p 80,443,843 google.com facebook.com youtube.com`,
	// with a 5-minute timeout.
	s, err := nmap.NewScanner(
		context.Background(),
		nmap.WithTargets("google.com", "facebook.com", "youtube.com"),
		nmap.WithPorts("80,443,843"),
	)
	if err != nil {
		log.Fatalf("unable to create nmap scanner: %v", err)
	}

	// Executes asynchronously, allowing results to be streamed in real time.
	done := make(chan error)
	result, warnings, err := s.Async(done).Run()
	if err != nil {
		log.Fatal(err)
	}

	// Blocks main until the scan has completed.
	if err := <-done; err != nil {
		if len(*warnings) > 0 {
			log.Printf("run finished with warnings: %s\n", *warnings) // Warnings are non-critical errors from nmap.
		}
		log.Fatal(err)
	}

	// Use the results to print an example output
	for _, host := range result.Hosts {
		if len(host.Ports) == 0 || len(host.Addresses) == 0 {
			continue
		}

		fmt.Printf("Host %q:\n", host.Addresses[0])

		for _, port := range host.Ports {
			fmt.Printf("\tPort %d/%s %s %s\n", port.ID, port.Protocol, port.State, port.Service.Name)
		}
	}
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
		nmap.WithTimingTemplate(5),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create UDP scanner: %w", err)
	}

	result, warnings, err := udpScanner.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to run UDP scanner: %w", err)
	}

	if len(*warnings) == 0 {
		log.Printf("Scan warning: %v\n", *warnings)
	}

	return result, nil
}
