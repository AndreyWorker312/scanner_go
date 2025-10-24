# TCP & Telnet Тестирование

Набор утилит для работы с TCP подключениями и анализа протокола Telnet.

## 📦 Файлы

- **`main.go`** - Telnet клиент с подробным выводом данных
- **`decoder.go`** - Декодер HEX строк Telnet
- **`TELNET_REFERENCE.md`** - Полный справочник по протоколу Telnet

## 🚀 Быстрый старт

### 1. Подключение к Telnet серверу

```bash
# Редактируем main.go (строки 14-15):
host := "telehack.com"  # или ваш сервер
port := "23"

# Запускаем:
go run main.go
```

**Вывод покажет:**
- 📦 Каждую порцию данных с HEX дампом
- 🔧 Все IAC команды Telnet
- 📨 Приветственное сообщение
- 📋 Полный HEX дамп

### 2. Декодирование HEX строк

```bash
# Копируете HEX строку из вывода main.go и декодируете:
go run decoder.go "FF FB 01 FF FD 18 FF FD 1F"
```

**Результат:**
```
  1 │ FF FB 01 │ IAC WILL Echo
  2 │ FF FD 18 │ IAC DO Terminal Type
  3 │ FF FD 1F │ IAC DO Window Size
```

## 📖 Расшифровка вашего примера

```
FF FB 01  →  IAC WILL Echo
             Сервер: "Я буду отображать то, что ты печатаешь"

FF FD 18  →  IAC DO Terminal-Type
             Сервер: "Пожалуйста, скажи, какой у тебя терминал"

FF FD 1F  →  IAC DO Window-Size
             Сервер: "Пожалуйста, скажи размер окна"

FF FD 24  →  IAC DO Environment-Variables
             Сервер: "Передай переменные окружения"

FF FD 27  →  IAC DO New-Environment
             Сервер: "Передай расширенные переменные"

FF FD 00  →  IAC DO Binary-Transmission
             Сервер: "Давай использовать бинарный режим"

0D 0A     →  \r\n (новая строка)

43 6F 6E 6E 65 63 74 65 64 20 74 6F...  →  "Connected to TELEHACK port 157\r\n"
```

## 🎯 Основные Telnet команды

| Hex | Команда | Значение |
|-----|---------|----------|
| `FF` | IAC | Начало команды |
| `FB` | WILL | "Я буду использовать" |
| `FC` | WONT | "Я НЕ буду" |
| `FD` | DO | "Пожалуйста, используй" |
| `FE` | DONT | "Пожалуйста, НЕ используй" |

## 🎛️ Популярные опции

| Hex | Опция | Описание |
|-----|-------|----------|
| `00` | Binary | Бинарная передача |
| `01` | Echo | Эхо символов |
| `03` | Suppress GA | Полный дуплекс |
| `18` (24) | Terminal Type | Тип терминала |
| `1F` (31) | Window Size | Размер окна |
| `24` (36) | Environment | Переменные |

## 🔍 Как читать TCP данные в Go

### Простое чтение
```go
conn, _ := net.Dial("tcp", "example.com:23")
buffer := make([]byte, 1024)
n, _ := conn.Read(buffer)
data := buffer[:n]
```

### С таймаутом
```go
conn.SetReadDeadline(time.Now().Add(5 * time.Second))
reader := bufio.NewReader(conn)
data, _ := reader.Read(buffer)
```

### Построчно
```go
scanner := bufio.NewScanner(conn)
for scanner.Scan() {
    line := scanner.Text()
}
```

## 📚 Документация

См. **`TELNET_REFERENCE.md`** для:
- Полного списка команд и опций
- Примеров согласования (negotiation)
- Подробных объяснений протокола
- Информации о безопасности

## 🧪 Тестовые серверы

```go
// Интерактивная игра в стиле 80-х
host := "telehack.com"
port := "23"

// Погода (США)
host := "rainmaker.wunderground.com"
port := "23"

// Star Wars ASCII анимация
host := "towel.blinkenlights.nl"
port := "23"
```

## ⚠️ Безопасность

- ❌ Telnet **НЕ шифрует** данные
- ❌ Пароли передаются открытым текстом
- ✅ Используйте только для тестирования
- ✅ Для продакшена используйте **SSH**

## 🛠️ Полезные команды

```bash
# Компиляция
go build main.go
go build decoder.go

# Запуск
./main
./decoder "FF FB 01"

# Подключение через telnet (для сравнения)
telnet telehack.com 23

# Захват трафика
tcpdump -i any -X port 23

# Просмотр в Wireshark
wireshark -i eth0 -f "port 23"
```

## 💡 Советы

1. **Для отладки:** Используйте `main.go` - он показывает все данные
2. **Для расшифровки:** Копируйте HEX и используйте `decoder.go`
3. **Для понимания:** Читайте `TELNET_REFERENCE.md`
4. **Для своих проектов:** Код в `main.go` можно адаптировать

## 🎓 Что вы узнаете

- ✅ Как установить TCP соединение в Go
- ✅ Как читать данные из TCP канала
- ✅ Как работает протокол Telnet
- ✅ Как декодировать HEX данные
- ✅ Что такое IAC команды
- ✅ Как анализировать сетевой трафик

## 📝 Примеры использования

### Проверка доступности Telnet на роутере
```go
host := "192.168.1.1"
port := "23"
```

### Анализ баннера сервера
```go
// Запустите main.go и посмотрите на баннер
// Декодируйте HEX для деталей
```

### Тестирование своего Telnet сервера
```go
// На сервере:
// nc -l -p 23

// В main.go:
host := "localhost"
port := "23"
```

---

**Автор:** Создано для изучения TCP и Telnet протоколов  
**Лицензия:** Используйте свободно для обучения и тестирования

