.PHONY: build run test docker-up docker-down docker-logs

build:
	go build -o bin/network-scanner ./cmd/server/main.go

run: build
	./bin/network-scanner

test:
	go test -v ./...

docker-up:
	docker-compose up -d --build

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f