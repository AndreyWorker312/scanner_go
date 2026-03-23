FROM golang:1.24-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags='-s -w -extldflags "-static"' \
    -o /scanner ./cmd/scanner/main.go

FROM builder AS security-check

RUN go install github.com/securego/gosec/v2/cmd/gosec@v2.21.4

RUN gosec ./...

FROM alpine:3.20

RUN apk add --no-cache nmap nmap-scripts ca-certificates libcap \
    && setcap cap_net_raw,cap_net_admin,cap_net_bind_service+eip /usr/bin/nmap \
    && rm -rf /var/cache/apk/*

COPY --from=builder /scanner /scanner

ENTRYPOINT ["/scanner"]
