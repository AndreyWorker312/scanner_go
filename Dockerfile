# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o network-scanner ./cmd/server/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/network-scanner .
COPY --from=builder /app/migrations ./migrations
COPY .env .env

EXPOSE 8006

CMD ["./network-scanner"]
