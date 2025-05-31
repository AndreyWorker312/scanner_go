.PHONY: build run test migrate-up migrate-down docker-up docker-down docker-logs

build:
	go build -o bin/network-scanner ./cmd/server/main.go

run: build
	./bin/network-scanner

test:
	go test -v ./...

migrate-up:
	migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/network_scanner?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/network_scanner?sslmode=disable" down

docker-up:
	docker-compose up -d --build

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

