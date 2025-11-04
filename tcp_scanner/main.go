package main

import (
	"fmt"
	"net"
	"time"
)

// --- USER CONFIG ---
var (
	host = "telehack.com" // любой TCP хост
	port = "23"           // любой TCP порт
)

// -------------------

func main() {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), 5*time.Second)
	if err != nil {
		fmt.Printf("Ошибка TCP: %v\n", err)
		return
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(10 * time.Second))

	var buf []byte
	tmp := make([]byte, 4096)
	for {
		n, err := conn.Read(tmp)
		if n > 0 {
			buf = append(buf, tmp[:n]...)
		}
		if err != nil {
			break
		}
	}

	fmt.Println("================ КОДИРОВАННЫЙ ВИД (HEX) ================")
	for i, b := range buf {
		if i > 0 {
			fmt.Print(" ")
		}
		fmt.Printf("%02X", b)
	}
	fmt.Println("\n========================================================\n")

	fmt.Println("================ ДЕКОДИРОВАННЫЙ ВИД ====================")
	fmt.Println(humanString(buf))
	fmt.Println("========================================================")
}

// humanString – только печатные + \r, \n, \t
func humanString(buf []byte) string {
	out := make([]byte, 0, len(buf))
	for _, b := range buf {
		switch b {
		case '\r':
			out = append(out, '\\', 'r')
		case '\n':
			out = append(out, '\\', 'n')
		case '\t':
			out = append(out, '\\', 't')
		default:
			if b >= 32 && b <= 126 {
				out = append(out, b)
			}
		}
	}
	return string(out)
}
