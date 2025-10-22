# Справочник TCP портов и сервисов

## Наиболее распространенные порты

### Файловые сервисы
| Порт | Сервис | Описание | Типичный баннер |
|------|--------|----------|----------------|
| 20 | FTP-DATA | FTP передача данных | - |
| 21 | FTP | File Transfer Protocol | `220 ProFTPD Server` |
| 69 | TFTP | Trivial FTP | - |
| 445 | SMB | Server Message Block | - |
| 2049 | NFS | Network File System | - |

### Удаленный доступ
| Порт | Сервис | Описание | Типичный баннер |
|------|--------|----------|----------------|
| 22 | SSH | Secure Shell | `SSH-2.0-OpenSSH_8.2p1` |
| 23 | Telnet | Telnet | `\xFF\xFD...` (IAC) |
| 3389 | RDP | Remote Desktop Protocol | - |
| 5900 | VNC | Virtual Network Computing | `RFB 003.008` |
| 5985 | WinRM | Windows Remote Management | - |

### Веб-сервисы
| Порт | Сервис | Описание | Типичный баннер |
|------|--------|----------|----------------|
| 80 | HTTP | Hypertext Transfer Protocol | `Server: nginx/1.18.0` |
| 443 | HTTPS | HTTP Secure | SSL/TLS handshake |
| 8000 | HTTP-ALT | Alternative HTTP | `Server: SimpleHTTP` |
| 8080 | HTTP-PROXY | HTTP Proxy | `Server: Apache` |
| 8443 | HTTPS-ALT | Alternative HTTPS | SSL/TLS |
| 8888 | HTTP-ALT | Alternative HTTP | - |

### Почтовые сервисы
| Порт | Сервис | Описание | Типичный баннер |
|------|--------|----------|----------------|
| 25 | SMTP | Simple Mail Transfer Protocol | `220 mail.example.com ESMTP` |
| 110 | POP3 | Post Office Protocol v3 | `+OK POP3 server ready` |
| 143 | IMAP | Internet Message Access Protocol | `* OK IMAP4rev1 Service Ready` |
| 465 | SMTPS | SMTP Secure | SSL/TLS |
| 587 | SMTP | SMTP Submission | `220 Ready to start TLS` |
| 993 | IMAPS | IMAP Secure | SSL/TLS |
| 995 | POP3S | POP3 Secure | SSL/TLS |

### Базы данных
| Порт | Сервис | Описание | Типичный баннер |
|------|--------|----------|----------------|
| 1433 | MSSQL | Microsoft SQL Server | - |
| 1521 | Oracle | Oracle Database | - |
| 3306 | MySQL | MySQL Database | `5.7.33-0ubuntu0.18.04.1` |
| 5432 | PostgreSQL | PostgreSQL Database | - |
| 6379 | Redis | Redis Key-Value Store | `-ERR unknown command` |
| 7000 | Cassandra | Apache Cassandra | - |
| 9042 | Cassandra | Cassandra CQL | - |
| 27017 | MongoDB | MongoDB Database | - |
| 27018 | MongoDB | MongoDB Shard | - |

### DNS и сетевые сервисы
| Порт | Сервис | Описание | Типичный баннер |
|------|--------|----------|----------------|
| 53 | DNS | Domain Name System | - |
| 67 | DHCP | DHCP Server | - |
| 68 | DHCP | DHCP Client | - |
| 123 | NTP | Network Time Protocol | - |
| 161 | SNMP | Simple Network Management | - |
| 389 | LDAP | Lightweight Directory Access | - |
| 636 | LDAPS | LDAP Secure | SSL/TLS |

### Приложения и фреймворки
| Порт | Сервис | Описание | Типичный баннер |
|------|--------|----------|----------------|
| 3000 | Node.js | Node.js/React Dev Server | - |
| 4200 | Angular | Angular Dev Server | - |
| 5000 | Flask | Flask Development Server | - |
| 5432 | Rails | Ruby on Rails | - |
| 8000 | Django | Django Development Server | - |
| 8080 | Tomcat | Apache Tomcat | `Server: Apache-Coyote` |
| 8081 | Nexus | Sonatype Nexus | - |
| 9000 | SonarQube | SonarQube | - |

### Messaging и очереди
| Порт | Сервис | Описание | Типичный баннер |
|------|--------|----------|----------------|
| 1883 | MQTT | Message Queue Telemetry | - |
| 4222 | NATS | NATS Messaging | `INFO {"version":"2.1.0"}` |
| 5671 | AMQPS | AMQP Secure | SSL/TLS |
| 5672 | AMQP | Advanced Message Queue | `AMQP` |
| 6379 | Redis | Redis Pub/Sub | - |
| 9092 | Kafka | Apache Kafka | - |
| 15672 | RabbitMQ | RabbitMQ Management | HTTP |

### Мониторинг и метрики
| Порт | Сервис | Описание | Типичный баннер |
|------|--------|----------|----------------|
| 3000 | Grafana | Grafana Dashboard | HTTP |
| 8086 | InfluxDB | InfluxDB Time Series | HTTP |
| 9090 | Prometheus | Prometheus Metrics | HTTP |
| 9200 | Elasticsearch | Elasticsearch | HTTP JSON |
| 9300 | Elasticsearch | Elasticsearch Transport | - |

### Игровые серверы
| Порт | Сервис | Описание | Типичный баннер |
|------|--------|----------|----------------|
| 25565 | Minecraft | Minecraft Java Edition | - |
| 27015 | Steam | Source Engine Games | - |
| 3074 | Xbox Live | Xbox Live | - |

### Другие важные порты
| Порт | Сервис | Описание | Типичный баннер |
|------|--------|----------|----------------|
| 179 | BGP | Border Gateway Protocol | - |
| 502 | Modbus | Modbus Protocol | - |
| 1080 | SOCKS | SOCKS Proxy | - |
| 1194 | OpenVPN | OpenVPN | - |
| 3128 | Squid | Squid Proxy | - |
| 4444 | Metasploit | Metasploit Framework | - |
| 5060 | SIP | Session Initiation Protocol | - |
| 6000-6063 | X11 | X Window System | - |
| 8888 | Jupyter | Jupyter Notebook | HTTP |

## Диапазоны портов

### Well-Known Ports (0-1023)
Зарезервированы для системных служб и требуют привилегий root/admin

### Registered Ports (1024-49151)
Используются приложениями, регистрируются IANA

### Dynamic/Private Ports (49152-65535)
Динамически выделяемые временные порты

## Примеры баннеров по сервисам

### SSH
```
SSH-2.0-OpenSSH_8.2p1 Ubuntu-4ubuntu0.5
SSH-2.0-OpenSSH_7.4
SSH-1.99-Cisco-1.25
```

### HTTP
```
HTTP/1.1 200 OK
Server: nginx/1.18.0
Server: Apache/2.4.41 (Ubuntu)
Server: Microsoft-IIS/10.0
```

### FTP
```
220 ProFTPD Server ready.
220 Microsoft FTP Service
220 vsftpd 3.0.3
```

### SMTP
```
220 mail.example.com ESMTP Postfix
220 smtp.gmail.com ESMTP
```

### MySQL
```
5.7.33-0ubuntu0.18.04.1
8.0.23 MySQL Community Server
```

### Redis
```
-ERR unknown command
+PONG
```

## Рекомендации по сканированию

### Базовый набор для веб-сервера
```json
{
  "ports": ["80", "443", "8080", "8443"]
}
```

### Базовый набор для сервера приложений
```json
{
  "ports": ["22", "80", "443", "3000", "3306", "5432", "6379"]
}
```

### Полное сканирование популярных портов
```json
{
  "ports": [
    "21", "22", "23", "25", "53", "80", "110", "143", "443",
    "445", "3306", "3389", "5432", "5900", "6379", "8080", "27017"
  ]
}
```

### Сканирование баз данных
```json
{
  "ports": ["1433", "1521", "3306", "5432", "6379", "7000", "9042", "27017"]
}
```

## Безопасность

⚠️ **Важные замечания:**

1. **Открытые порты ≠ Уязвимость**
   - Порт может быть открыт и безопасен
   - Важен контекст и конфигурация сервиса

2. **Баннеры содержат версии**
   - Полезно для инвентаризации
   - Может помочь злоумышленникам
   - Рекомендуется скрывать версии в продакшене

3. **Сканирование портов**
   - Законно для собственных систем
   - Требует разрешения для чужих систем
   - Может нарушать политики безопасности

4. **Защита**
   - Используйте файрволы
   - Закрывайте неиспользуемые порты
   - Скрывайте баннеры сервисов
   - Используйте fail2ban для защиты от сканирования

