# ============ СТАДИЯ 1: сборка ============
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Кэшируем зависимости отдельным слоем
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

# Статическая компиляция: -s -w убирают отладочную информацию,
# -trimpath убирает локальные пути из бинарника
# Статика (public/*) вшивается в бинарник через go:embed
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags='-s -w -extldflags "-static"' \
    -o /server ./cmd/main.go

# ============ СТАДИЯ 2: минимальный runtime ============
FROM scratch

# SSL-сертификаты — нужны для TLS-подключений к MongoDB / RabbitMQ
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Один статический бинарник (public/* уже внутри через embed.FS)
COPY --from=builder /server /server

EXPOSE 8080

ENTRYPOINT ["/server"]
