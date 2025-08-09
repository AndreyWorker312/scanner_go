package scanner

import (
	"context"
	"fmt"
	"github.com/go-ping/ping"
	"net"
	"time"
)

type PingResult struct {
	Target            string  `json:"target"`
	Address           string  `json:"address"`
	PacketsSent       int     `json:"packets_sent"`
	PacketsReceived   int     `json:"packets_received"`
	PacketLossPercent float64 `json:"packet_loss_percent"`
	Error             string  `json:"error,omitempty"`
}

type PingScanner interface {
	Ping(ctx context.Context, target string) PingResult
}

type pingScanner struct {
	count   int
	timeout time.Duration
}

func NewPingScanner(count int, timeout time.Duration) PingScanner {
	return &pingScanner{
		count:   count,
		timeout: timeout,
	}
}

func (s *pingScanner) Ping(ctx context.Context, target string) PingResult {
	var res PingResult
	res.Target = target

	// Разрешаем IP, если это домен
	ip := net.ParseIP(target)
	if ip == nil {
		ips, err := net.LookupIP(target)
		if err != nil || len(ips) == 0 {
			res.Error = fmt.Sprintf("DNS lookup error: %v", err)
			return res
		}
		ip = ips[0]
	}
	res.Address = ip.String()

	pinger, err := ping.NewPinger(res.Address)
	if err != nil {
		res.Error = fmt.Sprintf("Ping init error: %v", err)
		return res
	}
	pinger.Count = s.count
	pinger.Timeout = s.timeout
	pinger.SetPrivileged(true)

	err = pinger.Run()
	if err != nil {
		res.Error = fmt.Sprintf("Ping run error: %v", err)
		return res
	}
	stats := pinger.Statistics()
	res.PacketsSent = stats.PacketsSent
	res.PacketsReceived = stats.PacketsRecv
	res.PacketLossPercent = stats.PacketLoss

	return res
}
