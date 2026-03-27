FROM golang:1.24-alpine AS builder

WORKDIR /build

RUN apk add --no-cache wget

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

# Download the latest IEEE OUI database and embed it into the binary.
# Falls back silently to the placeholder (empty vendor lookups) if offline.
RUN wget -q -O internal/scanner/oui.csv \
    "https://standards-oui.ieee.org/oui/oui.csv" \
    || echo "OUI download failed – vendor lookup will return empty strings"

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags='-s -w -extldflags "-static"' \
    -o /scanner ./cmd/scanner/main.go

FROM builder AS security-check

RUN go install github.com/securego/gosec/v2/cmd/gosec@v2.21.4

RUN gosec ./...

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /scanner /scanner

ENTRYPOINT ["/scanner"]
