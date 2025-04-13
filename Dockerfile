# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Копируем файлы модулей
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o network-scanner ./cmd/server/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Копируем бинарник из builder
COPY --from=builder /app/network-scanner .
COPY --from=builder /app/migrations ./migrations

# Копируем .env файл (в продакшене лучше использовать секреты)
COPY .env .env

EXPOSE 8006

CMD ["./network-scanner"]