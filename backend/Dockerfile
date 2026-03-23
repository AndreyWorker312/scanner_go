FROM golang:1.24-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags='-s -w -extldflags "-static"' \
    -o /server ./cmd/main.go

FROM builder AS security-check

RUN go install github.com/securego/gosec/v2/cmd/gosec@v2.21.4

RUN gosec ./...

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /server /server

EXPOSE 8080

ENTRYPOINT ["/server"]
