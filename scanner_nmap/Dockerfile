# ============ СТАДИЯ 1: сборка ============
FROM golang:1.24-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags='-s -w -extldflags "-static"' \
    -o /scanner ./cmd/scanner/main.go

# ============ СТАДИЯ 2: runtime с nmap ============
# scratch невозможен — github.com/Ullaakut/nmap вызывает бинарник nmap
FROM alpine:3.20

RUN apk add --no-cache nmap nmap-scripts ca-certificates \
    && rm -rf /var/cache/apk/*

COPY --from=builder /scanner /scanner

ENTRYPOINT ["/scanner"]
