package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Утилита для декодирования HEX строк из Telnet
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Использование: go run decoder.go <hex-строка>")
		fmt.Println("\nПример:")
		fmt.Println("  go run decoder.go \"FF FB 01 0D 0A 48 65 6C 6C 6F\"")
		fmt.Println("\nИли вставьте hex прямо из вывода программы:")
		fmt.Println("  go run decoder.go \"FF FB 01 FF FD 18 FF FD 1F FF FD 24\"")
		return
	}

	// Объединяем все аргументы
	hexString := strings.Join(os.Args[1:], " ")
	
	// Парсим hex строку
	bytes := parseHexString(hexString)
	
	if len(bytes) == 0 {
		fmt.Println("❌ Не удалось распарсить hex строку")
		return
	}

	fmt.Printf("📦 Всего байт: %d\n\n", len(bytes))
	
	// Декодируем
	decodeBytes(bytes)
}

func parseHexString(s string) []byte {
	// Удаляем лишние символы
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\t", "")
	s = strings.ReplaceAll(s, "0x", "")
	s = strings.ToUpper(s)
	
	if len(s)%2 != 0 {
		return nil
	}
	
	result := make([]byte, 0, len(s)/2)
	for i := 0; i < len(s); i += 2 {
		b, err := strconv.ParseUint(s[i:i+2], 16, 8)
		if err != nil {
			return nil
		}
		result = append(result, byte(b))
	}
	
	return result
}

func decodeBytes(data []byte) {
	i := 0
	lineNum := 1
	
	for i < len(data) {
		b := data[i]
		
		// Проверяем на IAC команду
		if b == 0xFF && i+2 < len(data) {
			cmd := data[i+1]
			opt := data[i+2]
			
			fmt.Printf("%3d │ ", lineNum)
			fmt.Printf("FF %02X %02X │ ", cmd, opt)
			fmt.Printf("IAC %s %s\n", getIACName(cmd), getOptionName(opt))
			
			lineNum++
			i += 3
			continue
		}
		
		// Проверяем на IAC без опции
		if b == 0xFF && i+1 < len(data) {
			cmd := data[i+1]
			
			// Если это IAC IAC (экранированный 0xFF)
			if cmd == 0xFF {
				fmt.Printf("%3d │ FF FF    │ Escaped 0xFF (byte value 255)\n", lineNum)
				lineNum++
				i += 2
				continue
			}
			
			fmt.Printf("%3d │ FF %02X    │ IAC %s\n", lineNum, cmd, getIACName(cmd))
			lineNum++
			i += 2
			continue
		}
		
		// Обычные байты - группируем в текстовые строки
		textBytes := []byte{}
		
		for i < len(data) && data[i] != 0xFF {
			textBytes = append(textBytes, data[i])
			i++
		}
		
		// Выводим текстовые байты
		if len(textBytes) > 0 {
			decodeLine(lineNum, textBytes)
			lineNum++
		}
	}
}

func decodeLine(lineNum int, bytes []byte) {
	fmt.Printf("%3d │ ", lineNum)
	
	// Выводим hex
	for j, b := range bytes {
		fmt.Printf("%02X ", b)
		if j >= 15 {
			fmt.Print("...")
			break
		}
	}
	
	// Дополняем пробелами для выравнивания
	if len(bytes) <= 15 {
		for j := len(bytes); j < 16; j++ {
			fmt.Print("   ")
		}
	}
	
	fmt.Print("│ ")
	
	// Декодируем специальные символы
	hasSpecial := false
	for _, b := range bytes {
		if b == 0x0D || b == 0x0A || b == 0x09 || b < 0x20 {
			hasSpecial = true
			break
		}
	}
	
	// Выводим текст
	if hasSpecial {
		// Показываем с экранированием
		for _, b := range bytes {
			switch b {
			case 0x0D:
				fmt.Print("\\r")
			case 0x0A:
				fmt.Print("\\n")
			case 0x09:
				fmt.Print("\\t")
			case 0x00:
				fmt.Print("\\0")
			default:
				if b >= 32 && b <= 126 {
					fmt.Printf("%c", b)
				} else {
					fmt.Printf("\\x%02X", b)
				}
			}
		}
	} else {
		// Обычный текст
		fmt.Printf("\"%s\"", string(bytes))
	}
	
	fmt.Println()
}

func getIACName(cmd byte) string {
	names := map[byte]string{
		0xFF: "IAC",
		0xFE: "DONT",
		0xFD: "DO",
		0xFC: "WONT",
		0xFB: "WILL",
		0xFA: "SB (Subnegotiation Begin)",
		0xF0: "SE (Subnegotiation End)",
		0xF9: "GA (Go Ahead)",
		0xF8: "EL (Erase Line)",
		0xF7: "EC (Erase Character)",
		0xF6: "AYT (Are You There)",
		0xF5: "AO (Abort Output)",
		0xF4: "IP (Interrupt Process)",
		0xF3: "BRK (Break)",
		0xF2: "DM (Data Mark)",
		0xF1: "NOP (No Operation)",
	}
	
	if name, ok := names[cmd]; ok {
		return fmt.Sprintf("%-25s (0x%02X)", name, cmd)
	}
	return fmt.Sprintf("Unknown                   (0x%02X)", cmd)
}

func getOptionName(opt byte) string {
	options := map[byte]string{
		0:  "Binary Transmission",
		1:  "Echo",
		3:  "Suppress Go Ahead",
		5:  "Status",
		6:  "Timing Mark",
		24: "Terminal Type",
		31: "Window Size (NAWS)",
		32: "Terminal Speed",
		33: "Remote Flow Control",
		34: "Linemode",
		36: "Environment Variables",
		39: "New Environment",
	}
	
	if name, ok := options[opt]; ok {
		return fmt.Sprintf("%-30s (0x%02X)", name, opt)
	}
	return fmt.Sprintf("Unknown Option             (0x%02X)", opt)
}

