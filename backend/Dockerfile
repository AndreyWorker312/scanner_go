FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /backend ./cmd/server/main.go

FROM alpine:3.18

WORKDIR /app

COPY --from=builder /backend /app/backend

# Копируем .env, если он нужен внутри контейнера (вариативно)
COPY .env ./

EXPOSE 8080

# Запускаем приложение
CMD ["/app/backend"]
