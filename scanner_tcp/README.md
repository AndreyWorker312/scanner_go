# TCP Scanner с Banner Grabbing

TCP сканер для проверки доступности портов и получения баннеров сервисов.

## Возможности

- **TCP Port Scanning**: Проверка доступности TCP портов
- **Banner Grabbing**: Чтение данных из сокета для идентификации сервиса
- **Service Detection**: Автоматическое определение типа сервиса по баннеру
- **Version Detection**: Извлечение информации о версии сервиса
- **HTTP Probing**: Специальная обработка HTTP/HTTPS портов
- **Response Time**: Измерение времени отклика

## Архитектура

Сканер работает как микросервис и взаимодействует с основным backend через RabbitMQ:

```
Backend → RabbitMQ → TCP Scanner → RabbitMQ → Backend → MongoDB
```

## Определение сервисов

Сканер автоматически определяет следующие сервисы по баннерам:

- **SSH**: Определение по "SSH" в баннере
- **HTTP/HTTPS**: Apache, Nginx, IIS и другие веб-серверы
- **FTP**: Определение по коду 220 и "FTP"
- **SMTP**: Определение по коду 220 и "SMTP"
- **MySQL/PostgreSQL**: Определение по ключевым словам
- **Redis**: Определение по специфичным ответам
- **Telnet**: Определение по IAC последовательности

Также используется база известных портов для идентификации стандартных сервисов:
- 21 (FTP), 22 (SSH), 23 (Telnet)
- 25 (SMTP), 53 (DNS)
- 80 (HTTP), 443 (HTTPS)
- 3306 (MySQL), 5432 (PostgreSQL)
- 6379 (Redis), 27017 (MongoDB)
- И другие...

## Формат запроса

```json
{
  "task_id": "unique-task-id",
  "host": "example.com",
  "ports": ["80", "443", "22", "3306"],
  "timeout": 5
}
```

### Параметры:
- `task_id` (string): Уникальный ID задачи
- `host` (string): IP адрес или доменное имя
- `ports` ([]string): Список портов для сканирования
- `timeout` (int, optional): Таймаут подключения в секундах

## Формат ответа

```json
{
  "task_id": "unique-task-id",
  "host": "example.com",
  "status": "completed",
  "results": [
    {
      "port": "80",
      "state": "open",
      "service": "http",
      "banner": "HTTP/1.1 200 OK\r\nServer: nginx/1.18.0",
      "version": "nginx/1.18.0",
      "response_time": 45
    },
    {
      "port": "22",
      "state": "open",
      "service": "ssh",
      "banner": "SSH-2.0-OpenSSH_8.2p1 Ubuntu-4ubuntu0.5",
      "version": "SSH-2.0-OpenSSH_8.2p1",
      "response_time": 23
    },
    {
      "port": "3306",
      "state": "closed",
      "service": "mysql",
      "error": "Connection refused",
      "response_time": 1
    }
  ]
}
```

### Поля результата:
- `port` (string): Номер порта
- `state` (string): Состояние порта (open, closed, filtered)
- `service` (string): Определенный сервис
- `banner` (string): Полученный баннер
- `version` (string): Версия сервиса (если определена)
- `error` (string): Описание ошибки (если есть)
- `response_time` (int64): Время отклика в миллисекундах

## Конфигурация

Переменные окружения:

- `RABBITMQ_URL`: URL для подключения к RabbitMQ (по умолчанию: `amqp://guest:guest@localhost:5672/`)
- `SCANNER_NAME`: Имя очереди сканера (по умолчанию: `tcp_scanner`)
- `CONN_TIMEOUT`: Таймаут TCP подключения (по умолчанию: `5s`)
- `BANNER_TIMEOUT`: Таймаут чтения баннера (по умолчанию: `3s`)
- `MAX_BANNER_SIZE`: Максимальный размер баннера в байтах (по умолчанию: `4096`)

## Сборка и запуск

### Локальная сборка

```bash
cd scanner_tcp
go mod download
go build -o tcp_scanner ./cmd/scanner
./tcp_scanner
```

### Docker

```bash
docker build -t scanner-tcp .
docker run --network host \
  -e RABBITMQ_URL=amqp://guest:guest@localhost:5672/ \
  -e SCANNER_NAME=tcp_scanner \
  scanner-tcp
```

### Docker Compose

```bash
# Из корневой директории проекта
docker-compose up scanner_tcp
```

## Примеры использования

### Через WebSocket (из фронтенда):

```javascript
{
  "scanner_service": "tcp_service",
  "options": {
    "task_id": "tcp-scan-1",
    "host": "scanme.nmap.org",
    "ports": ["22", "80", "443", "8080"]
  }
}
```

### Через RabbitMQ (напрямую):

```go
request := TCPRequest{
    TaskID: "tcp-scan-1",
    Host:   "example.com",
    Ports:  []string{"80", "443", "22"},
    Timeout: 5,
}

body, _ := json.Marshal(request)
// Публикация в очередь "tcp_scanner"
```

## История сканирования

Все результаты сохраняются в MongoDB в коллекции `tcp_history`.

### REST API для доступа к истории:

```bash
# Получить историю
GET /api/history/tcp?limit=10

# Удалить историю
DELETE /api/history/tcp/delete
```

## Banner Grabbing

Сканер использует несколько методов для получения баннеров:

1. **Passive Banner Grabbing**: Ожидание данных от сервера после подключения
2. **Active HTTP Probing**: Отправка HTTP GET запроса для веб-серверов
3. **Timeout-based**: Использование коротких таймаутов для быстрого сканирования

## Безопасность

⚠️ **Важно**: 
- Используйте сканер только на собственных системах или с явного разрешения
- Banner grabbing может быть расценен как попытка взлома
- Соблюдайте законы и политики безопасности
- В производственной среде ограничивайте доступ к сканеру

## Ограничения

- Не все сервисы отправляют баннеры автоматически
- Некоторые файрволы могут блокировать сканирование
- Таймауты могут привести к ложноположительным результатам "filtered"
- Идентификация сервисов не всегда точна на 100%

## Зависимости

```
github.com/rabbitmq/amqp091-go v1.10.0
github.com/sirupsen/logrus v1.9.3
```

## Лицензия

Часть проекта WebScanAPI

