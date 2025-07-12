# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /network-scanner ./cmd/server/main.go

FROM alpine:3.18

WORKDIR /app
COPY --from=builder /network-scanner /app/network-scanner
COPY .env /app/.env

EXPOSE 8080
CMD ["/app/network-scanner"]

