package main

import (
	"bufio"
	"fmt"
	"net"
	"time"
)

func main() {
	fmt.Println("=== Telnet подключение и чтение данных ===\n")

	// Укажите хост и порт Telnet сервера
	host := "telehack.com" // Публичный telnet сервер для тестов
	port := "23"

	// Можно также попробовать другие адреса:
	// host := "192.168.1.1"  // Ваш роутер
	// host := "localhost"    // Локальный сервер

	connectAndReadTelnet(host, port)
}

// connectAndReadTelnet устанавливает Telnet соединение и читает все данные
func connectAndReadTelnet(host, port string) {
	fmt.Printf("🔌 Подключение к %s:%s...\n", host, port)

	// Шаг 1: Установка TCP соединения
	address := net.JoinHostPort(host, port)
	dialer := &net.Dialer{
		Timeout: 5 * time.Second,
	}

	conn, err := dialer.Dial("tcp", address)
	if err != nil {
		fmt.Printf("❌ Ошибка подключения: %v\n", err)
		fmt.Println("\n💡 Попробуйте другой адрес или проверьте, что порт 23 открыт")
		return
	}
	defer conn.Close()

	fmt.Printf("✅ TCP соединение установлено!\n")
	fmt.Printf("   Локальный адрес: %s\n", conn.LocalAddr())
	fmt.Printf("   Удаленный адрес: %s\n\n", conn.RemoteAddr())

	// Шаг 2: Чтение данных из Telnet
	readTelnetData(conn)
}

// readTelnetData читает и анализирует данные из Telnet соединения
func readTelnetData(conn net.Conn) {
	fmt.Println("📖 Начинаю чтение данных из TCP канала...\n")

	// Устанавливаем таймаут на чтение
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	// Создаем буферизованный reader для эффективного чтения
	reader := bufio.NewReader(conn)

	// Хранилища для данных
	rawData := make([]byte, 0, 8192)       // Сырые данные
	cleanData := make([]byte, 0, 8192)     // Очищенные данные (без IAC)
	iacCommands := []IACCommand{}          // Telnet команды

	buffer := make([]byte, 1024)
	iteration := 0

	// Читаем данные порциями
	for {
		iteration++
		n, err := reader.Read(buffer)

		if n > 0 {
			fmt.Printf("📦 Порция #%d: получено %d байт\n", iteration, n)

			// Сохраняем сырые данные
			rawData = append(rawData, buffer[:n]...)

			// Показываем hex дамп (первые 64 байта)
			fmt.Print("   Hex: ")
			displayLimit := min(n, 64)
			for i := 0; i < displayLimit; i++ {
				fmt.Printf("%02X ", buffer[i])
				if (i+1)%16 == 0 {
					fmt.Print("\n        ")
				}
			}
			if n > 64 {
				fmt.Printf("... (+%d байт)", n-64)
			}
			fmt.Println()

			// Обрабатываем данные
			processedData, commands := processTelnetBytes(buffer[:n])
			cleanData = append(cleanData, processedData...)
			iacCommands = append(iacCommands, commands...)

			// Показываем читаемый текст
			if len(processedData) > 0 {
				fmt.Printf("   Текст: %q\n", string(processedData))
			}

			fmt.Println()
		}

		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				fmt.Println("⏱️  Таймаут чтения (больше данных не поступает)")
			} else {
				fmt.Printf("ℹ️  Чтение завершено: %v\n", err)
			}
			break
		}

		// Продляем таймаут если данные продолжают поступать
		conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	}

	// Выводим итоговую статистику
	printSummary(rawData, cleanData, iacCommands)
}

// IACCommand представляет Telnet IAC команду
type IACCommand struct {
	Command string
	Option  string
	Bytes   []byte
}

// processTelnetBytes обрабатывает байты и извлекает IAC команды
func processTelnetBytes(data []byte) (cleanData []byte, commands []IACCommand) {
	cleanData = make([]byte, 0, len(data))
	commands = []IACCommand{}

	i := 0
	for i < len(data) {
		b := data[i]

		// Проверяем на IAC команду (0xFF)
		if b == 0xFF && i+2 < len(data) {
			cmd := data[i+1]
			opt := data[i+2]

			commands = append(commands, IACCommand{
				Command: getIACCommandName(cmd),
				Option:  getTelnetOptionName(opt),
				Bytes:   []byte{b, cmd, opt},
			})

			i += 3
			continue
		}

		// IAC без параметров (например, IAC IAC = экранированный 0xFF)
		if b == 0xFF && i+1 < len(data) && data[i+1] == 0xFF {
			cleanData = append(cleanData, 0xFF)
			i += 2
			continue
		}

		// Пропускаем непечатные управляющие символы (кроме \n, \r, \t)
		if b >= 32 && b <= 126 || b == '\n' || b == '\r' || b == '\t' {
			cleanData = append(cleanData, b)
		}

		i++
	}

	return cleanData, commands
}

// printSummary выводит итоговую информацию
func printSummary(rawData, cleanData []byte, commands []IACCommand) {
	fmt.Println("\n" + repeatString("═", 70))
	fmt.Println("📊 ИТОГОВАЯ СТАТИСТИКА")
	fmt.Println(repeatString("═", 70))

	fmt.Printf("\n📦 Всего получено: %d байт сырых данных\n", len(rawData))
	fmt.Printf("📝 Очищенных данных: %d байт\n", len(cleanData))
	fmt.Printf("🔧 Telnet IAC команд: %d\n", len(commands))

	// Показываем IAC команды
	if len(commands) > 0 {
		fmt.Println("\n🔧 Обнаруженные Telnet IAC команды:")
		for i, cmd := range commands {
			fmt.Printf("   %d. %s %s [%02X %02X %02X]\n",
				i+1, cmd.Command, cmd.Option,
				cmd.Bytes[0], cmd.Bytes[1], cmd.Bytes[2])
		}
	}

	// Показываем приветственное сообщение
	if len(cleanData) > 0 {
		fmt.Println("\n📨 Приветственное сообщение (баннер):")
		fmt.Println(repeatString("─", 70))
		fmt.Println(string(cleanData))
		fmt.Println(repeatString("─", 70))
	}

	// Анализируем тип сервера
	serverType := identifyTelnetServer(string(cleanData))
	if serverType != "" {
		fmt.Printf("\n🖥️  Определен тип сервера: %s\n", serverType)
	}

	// Показываем полный hex дамп сырых данных
	if len(rawData) > 0 {
		fmt.Println("\n📋 Полный HEX дамп сырых данных:")
		printHexDump(rawData)
	}
}

// Вспомогательные функции

// getIACCommandName возвращает имя IAC команды
func getIACCommandName(cmd byte) string {
	commands := map[byte]string{
		0xFF: "IAC",
		0xFE: "DONT",
		0xFD: "DO",
		0xFC: "WONT",
		0xFB: "WILL",
		0xFA: "SB",
		0xF0: "SE",
	}
	if name, ok := commands[cmd]; ok {
		return name
	}
	return fmt.Sprintf("Unknown(0x%02X)", cmd)
}

// getTelnetOptionName возвращает имя Telnet опции
func getTelnetOptionName(opt byte) string {
	options := map[byte]string{
		0:  "Binary Transmission",
		1:  "Echo",
		3:  "Suppress Go Ahead",
		5:  "Status",
		6:  "Timing Mark",
		24: "Terminal Type",
		31: "Window Size",
		32: "Terminal Speed",
		33: "Remote Flow Control",
		34: "Linemode",
		36: "Environment Variables",
	}
	if name, ok := options[opt]; ok {
		return fmt.Sprintf("%s(0x%02X)", name, opt)
	}
	return fmt.Sprintf("Unknown(0x%02X)", opt)
}

// identifyTelnetServer пытается определить тип Telnet сервера по баннеру
func identifyTelnetServer(banner string) string {
	bannerLower := ""
	for _, r := range banner {
		if r >= 'A' && r <= 'Z' {
			bannerLower += string(r + 32)
		} else {
			bannerLower += string(r)
		}
	}

	if contains(bannerLower, "ubuntu") {
		return "Ubuntu Linux Telnet"
	}
	if contains(bannerLower, "debian") {
		return "Debian Linux Telnet"
	}
	if contains(bannerLower, "redhat") || contains(bannerLower, "red hat") {
		return "Red Hat Linux Telnet"
	}
	if contains(bannerLower, "centos") {
		return "CentOS Linux Telnet"
	}
	if contains(bannerLower, "freebsd") {
		return "FreeBSD Telnet"
	}
	if contains(bannerLower, "cisco") {
		return "Cisco IOS Telnet"
	}
	if contains(bannerLower, "mikrotik") {
		return "MikroTik RouterOS Telnet"
	}
	if contains(bannerLower, "juniper") {
		return "Juniper JunOS Telnet"
	}
	if contains(bannerLower, "windows") {
		return "Windows Telnet Server"
	}
	if contains(bannerLower, "linux") {
		return "Linux Telnet"
	}

	return ""
}

// printHexDump выводит hex дамп данных
func printHexDump(data []byte) {
	const bytesPerLine = 16
	for i := 0; i < len(data); i += bytesPerLine {
		// Смещение
		fmt.Printf("   %08X: ", i)

		// Hex значения
		for j := 0; j < bytesPerLine; j++ {
			if i+j < len(data) {
				fmt.Printf("%02X ", data[i+j])
			} else {
				fmt.Print("   ")
			}
			if j == 7 {
				fmt.Print(" ")
			}
		}

		// ASCII представление
		fmt.Print(" |")
		for j := 0; j < bytesPerLine && i+j < len(data); j++ {
			b := data[i+j]
			if b >= 32 && b <= 126 {
				fmt.Printf("%c", b)
			} else {
				fmt.Print(".")
			}
		}
		fmt.Println("|")

		// Ограничиваем вывод первыми 256 байтами
		if i >= 240 && len(data) > 256 {
			fmt.Printf("   ... (показано первые 256 байт из %d)\n", len(data))
			break
		}
	}
}

// contains проверяет, содержит ли строка подстроку
func contains(s, substr string) bool {
	return len(s) >= len(substr) && indexOfString(s, substr) >= 0
}

func indexOfString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// repeatString повторяет строку n раз
func repeatString(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}

