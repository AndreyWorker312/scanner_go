# Примеры использования TCP Scanner

## Быстрый старт

### 1. Запуск через Docker Compose

```bash
# Из корневой директории проекта
docker-compose up scanner_tcp
```

### 2. Локальный запуск

```bash
cd scanner_tcp
export RABBITMQ_URL="amqp://guest:guest@localhost:5672/"
export SCANNER_NAME="tcp_scanner"
go run ./cmd/scanner/main.go
```

## Примеры сканирования

### Пример 1: Сканирование веб-сервера

**Запрос через WebSocket:**

```json
{
  "scanner_service": "tcp_service",
  "options": {
    "task_id": "web-scan-001",
    "host": "example.com",
    "ports": ["80", "443"]
  }
}
```

**Ожидаемый ответ:**

```json
{
  "task_id": "web-scan-001",
  "host": "example.com",
  "status": "completed",
  "results": [
    {
      "port": "80",
      "state": "open",
      "service": "http",
      "banner": "HTTP/1.1 200 OK\r\nServer: nginx/1.18.0\r\n...",
      "version": "nginx/1.18.0",
      "response_time": 45
    },
    {
      "port": "443",
      "state": "open",
      "service": "https",
      "banner": "...",
      "version": "",
      "response_time": 52
    }
  ]
}
```

### Пример 2: Сканирование SSH сервера

**Запрос:**

```json
{
  "scanner_service": "tcp_service",
  "options": {
    "task_id": "ssh-scan-001",
    "host": "192.168.1.100",
    "ports": ["22"]
  }
}
```

**Ожидаемый ответ:**

```json
{
  "task_id": "ssh-scan-001",
  "host": "192.168.1.100",
  "status": "completed",
  "results": [
    {
      "port": "22",
      "state": "open",
      "service": "ssh",
      "banner": "SSH-2.0-OpenSSH_8.2p1 Ubuntu-4ubuntu0.5",
      "version": "SSH-2.0-OpenSSH_8.2p1",
      "response_time": 23
    }
  ]
}
```

### Пример 3: Сканирование баз данных

**Запрос:**

```json
{
  "scanner_service": "tcp_service",
  "options": {
    "task_id": "db-scan-001",
    "host": "localhost",
    "ports": ["3306", "5432", "6379", "27017"]
  }
}
```

**Ожидаемый ответ:**

```json
{
  "task_id": "db-scan-001",
  "host": "localhost",
  "status": "completed",
  "results": [
    {
      "port": "3306",
      "state": "open",
      "service": "mysql",
      "banner": "...",
      "version": "",
      "response_time": 12
    },
    {
      "port": "5432",
      "state": "open",
      "service": "postgresql",
      "banner": "...",
      "version": "",
      "response_time": 15
    },
    {
      "port": "6379",
      "state": "open",
      "service": "redis",
      "banner": "-ERR unknown command",
      "version": "",
      "response_time": 8
    },
    {
      "port": "27017",
      "state": "open",
      "service": "mongodb",
      "banner": "",
      "version": "",
      "response_time": 10
    }
  ]
}
```

### Пример 4: Массовое сканирование портов

**Запрос:**

```json
{
  "scanner_service": "tcp_service",
  "options": {
    "task_id": "mass-scan-001",
    "host": "scanme.nmap.org",
    "ports": ["21", "22", "23", "25", "53", "80", "110", "143", "443", "3389", "8080"]
  }
}
```

### Пример 5: Сканирование с закрытыми портами

**Запрос:**

```json
{
  "scanner_service": "tcp_service",
  "options": {
    "task_id": "closed-scan-001",
    "host": "example.com",
    "ports": ["21", "23", "25", "3389"]
  }
}
```

**Ожидаемый ответ:**

```json
{
  "task_id": "closed-scan-001",
  "host": "example.com",
  "status": "completed",
  "results": [
    {
      "port": "21",
      "state": "closed",
      "service": "ftp",
      "banner": "",
      "error": "Connection refused",
      "response_time": 2
    },
    {
      "port": "23",
      "state": "closed",
      "service": "telnet",
      "banner": "",
      "error": "Connection refused",
      "response_time": 1
    }
  ]
}
```

## Интеграция с REST API

### Получение истории сканирований

```bash
# Получить все записи
curl http://localhost:8080/api/history/tcp

# Получить последние 10 записей
curl http://localhost:8080/api/history/tcp?limit=10
```

### Удаление истории

```bash
curl -X DELETE http://localhost:8080/api/history/tcp/delete
```

## JavaScript примеры для фронтенда

### WebSocket клиент

```javascript
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onopen = () => {
  console.log('WebSocket connected');
  
  // Отправка запроса на TCP сканирование
  const request = {
    scanner_service: 'tcp_service',
    options: {
      task_id: 'tcp-' + Date.now(),
      host: 'example.com',
      ports: ['80', '443', '22']
    }
  };
  
  ws.send(JSON.stringify(request));
};

ws.onmessage = (event) => {
  const response = JSON.parse(event.data);
  console.log('Scan results:', response);
  
  // Обработка результатов
  if (response.results) {
    response.results.forEach(result => {
      console.log(`Port ${result.port}: ${result.state}`);
      if (result.service) {
        console.log(`  Service: ${result.service}`);
      }
      if (result.banner) {
        console.log(`  Banner: ${result.banner.substring(0, 50)}...`);
      }
    });
  }
};

ws.onerror = (error) => {
  console.error('WebSocket error:', error);
};
```

### Получение истории через fetch

```javascript
// Получить историю TCP сканирований
fetch('http://localhost:8080/api/history/tcp?limit=10')
  .then(response => response.json())
  .then(data => {
    if (data.success) {
      console.log('History records:', data.data);
      console.log('Total records:', data.count);
    }
  })
  .catch(error => console.error('Error:', error));
```

## Рекомендации по использованию

### Для безопасного тестирования

1. **Используйте тестовые хосты:**
   - scanme.nmap.org (официальный тестовый хост)
   - testphp.vulnweb.com
   - demo.testfire.net

2. **Собственные системы:**
   - localhost
   - 127.0.0.1
   - Системы в вашей локальной сети

### Оптимизация производительности

1. **Не сканируйте слишком много портов одновременно**
   - Рекомендуется: до 20 портов за раз
   - Для большего количества разбивайте на несколько запросов

2. **Настройте таймауты:**
   ```bash
   export CONN_TIMEOUT=3s
   export BANNER_TIMEOUT=2s
   ```

3. **Сканируйте только нужные порты**
   - Используйте список известных портов вашего сервиса

### Обработка ошибок

Возможные статусы портов:
- `open` - порт открыт и принимает соединения
- `closed` - порт закрыт (Connection refused)
- `filtered` - порт фильтруется (Timeout)

Коды ошибок в поле `error`:
- "Connection refused" - порт закрыт
- "Connection timeout" - порт фильтруется или хост недоступен
- "DNS lookup error" - не удалось разрешить имя хоста

## Troubleshooting

### Проблема: Все порты показываются как filtered

**Решение:**
- Проверьте доступность хоста (ping)
- Увеличьте CONN_TIMEOUT
- Проверьте настройки файрволла

### Проблема: Баннер не получен

**Причины:**
- Некоторые сервисы не отправляют баннер автоматически
- Таймаут слишком короткий
- Сервис требует инициализации (SSL handshake для HTTPS)

**Решение:**
- Увеличьте BANNER_TIMEOUT
- Для некоторых сервисов баннер может быть пустым - это нормально

### Проблема: Медленное сканирование

**Решение:**
- Уменьшите количество портов в одном запросе
- Уменьшите таймауты для локальной сети
- Проверьте нагрузку на систему

