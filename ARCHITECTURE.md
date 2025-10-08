# Network Scanner - Архитектура системы

## Обзор системы

Система Network Scanner представляет собой микросервисную архитектуру для сканирования сети с тремя типами сканеров: ARP, ICMP и Nmap. Все результаты сканирования автоматически сохраняются в MongoDB для последующего просмотра через веб-интерфейс.

## Компоненты системы

### 1. Backend (Go)
- **main.go**: Точка входа, настройка серверов и маршрутов
- **WebSocket Handler**: Обработка real-time соединений с фронтендом
- **REST API**: Эндпоинты для работы с историей сканирования
- **Application Layer**: Бизнес-логика обработки запросов
- **Repository Layer**: Работа с базой данных MongoDB
- **Infrastructure**: Подключение к MongoDB и RabbitMQ

### 2. Сканеры (Go микросервисы)
- **ARP Scanner**: Обнаружение устройств в локальной сети
- **ICMP Scanner**: Ping сканирование хостов
- **Nmap Scanner**: Продвинутое сканирование портов и определение ОС

### 3. База данных
- **MongoDB**: Хранение истории сканирования
  - `arp_history`: Результаты ARP сканирования
  - `icmp_history`: Результаты ICMP ping
  - `nmap_tcp_udp_history`: Результаты TCP/UDP сканирования
  - `nmap_os_detection_history`: Результаты определения ОС
  - `nmap_host_discovery_history`: Результаты обнаружения хостов

### 4. Очереди сообщений
- **RabbitMQ**: Асинхронная обработка задач сканирования

### 5. Фронтенд
- **HTML/CSS/JavaScript**: Веб-интерфейс для управления сканированием
- **WebSocket клиент**: Real-time обновления результатов
- **REST API клиент**: Загрузка и управление историей

## Поток данных

```
1. Пользователь → Фронтенд → WebSocket → Backend
2. Backend → RabbitMQ → Сканер (ARP/ICMP/Nmap)
3. Сканер → RabbitMQ → Backend
4. Backend → MongoDB (автоматическое сохранение)
5. Backend → WebSocket → Фронтенд (real-time результаты)
6. Пользователь → Фронтенд → REST API → Backend → MongoDB (история)
```

## API Endpoints

### WebSocket
- `ws://localhost:8080/ws` - Real-time соединение для сканирования

### REST API
- `GET /api/history/arp` - Получить историю ARP сканирования
- `DELETE /api/history/arp/delete` - Очистить историю ARP
- `GET /api/history/icmp` - Получить историю ICMP ping
- `DELETE /api/history/icmp/delete` - Очистить историю ICMP
- `GET /api/history/nmap` - Получить историю Nmap сканирования
- `DELETE /api/history/nmap/delete` - Очистить историю Nmap

## Структура данных

### ARP History Record
```go
type ARPHistoryRecord struct {
    ID             string      `bson:"_id,omitempty" json:"id"`
    TaskID         string      `bson:"task_id" json:"task_id"`
    InterfaceName  string      `bson:"interface_name" json:"interface_name"`
    IPRange        string      `bson:"ip_range" json:"ip_range"`
    Status         string      `bson:"status" json:"status"`
    Devices        []ARPDevice `bson:"devices" json:"devices"`
    OnlineDevices  []ARPDevice `bson:"online_devices" json:"online_devices"`
    OfflineDevices []ARPDevice `bson:"offline_devices" json:"offline_devices"`
    TotalCount     int         `bson:"total_count" json:"total_count"`
    OnlineCount    int         `bson:"online_count" json:"online_count"`
    OfflineCount   int         `bson:"offline_count" json:"offline_count"`
    Error          string      `bson:"error,omitempty" json:"error,omitempty"`
    CreatedAt      time.Time   `bson:"created_at" json:"created_at"`
}
```

### ICMP History Record
```go
type ICMPHistoryRecord struct {
    ID        string       `bson:"_id,omitempty" json:"id"`
    TaskID    string       `bson:"task_id" json:"task_id"`
    Targets   []string     `bson:"targets" json:"targets"`
    PingCount int          `bson:"ping_count" json:"ping_count"`
    Status    string       `bson:"status" json:"status"`
    Results   []ICMPResult `bson:"results" json:"results"`
    Error     string       `bson:"error,omitempty" json:"error,omitempty"`
    CreatedAt time.Time    `bson:"created_at" json:"created_at"`
}
```

### Nmap History Records
- `NmapTcpUdpHistoryRecord` - TCP/UDP сканирование портов
- `NmapOsDetectionHistoryRecord` - Определение операционной системы
- `NmapHostDiscoveryHistoryRecord` - Обнаружение хостов

## Особенности реализации

1. **Автоматическое сохранение**: Все результаты сканирования автоматически сохраняются в MongoDB через `ProcessResponse` метод
2. **Real-time обновления**: WebSocket обеспечивает мгновенное отображение результатов
3. **История сканирования**: Полная история всех сканирований доступна через REST API
4. **Микросервисная архитектура**: Каждый сканер работает как отдельный сервис
5. **Масштабируемость**: Легко добавлять новые типы сканеров

## Развертывание

Система развертывается через Docker Compose с следующими сервисами:
- MongoDB (порт 27017)
- RabbitMQ (порт 5673, управление 15673)
- Backend (порт 8080)
- ARP Scanner
- ICMP Scanner  
- Nmap Scanner

Все сервисы настроены для работы в привилегированном режиме с необходимыми сетевыми правами для сканирования.
