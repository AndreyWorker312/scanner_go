
### ICMP Scanner JSON Format

#### Принимает (PingRequest):
```json
{
  "task_id": "string",      // Уникальный идентификатор задачи
  "targets": ["string"],    // Массив целей для сканирования (IP или домены)
  "ping_count": number      // Количество ICMP пакетов для отправки (опционально, default: 4)
}
```

Пример:
```json
{
  "task_id": "ping_scan_123",
  "targets": ["8.8.8.8", "google.com", "192.168.1.1"],
  "ping_count": 5
}
```

#### Возвращает (PingResponse):
```json
{
  "task_id": "string",      // Тот же ID задачи
  "status": "string",       // "completed" или "failed"
  "results": [
    {
      "target": "string",   // Целевой IP/домен
      "address": "string",  // Разрешенный IP адрес
      "packets_sent": number,
      "packets_received": number,
      "packet_loss_percent": number,
      "error": "string"     // Описание ошибки (если была)
    }
  ],
  "error": "string"         // Общая ошибка задачи (опционально)
}
```

Пример успешного ответа:
```json
{
  "task_id": "ping_scan_123",
  "status": "completed",
  "results": [
    {
      "target": "8.8.8.8",
      "address": "8.8.8.8",
      "packets_sent": 5,
      "packets_received": 5,
      "packet_loss_percent": 0.0
    },
    {
      "target": "google.com",
      "address": "142.250.185.206",
      "packets_sent": 5,
      "packets_received": 4,
      "packet_loss_percent": 20.0
    }
  ]
}
```

Пример ответа с ошибкой:
```json
{
  "task_id": "ping_scan_123",
  "status": "failed",
  "results": [
    {
      "target": "192.168.1.1",
      "address": "192.168.1.1",
      "error": "Request timed out"
    }
  ],
  "error": "Partial scan completion"
}
```
