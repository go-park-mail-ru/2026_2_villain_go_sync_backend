.PHONY: run build test lint fmt docker-up docker-down

run:
	go run ./cmd/server

build:
	mkdir -p bin
	go build -o bin/server ./cmd/server

test:
	go test ./...

lint:
	go vet ./...

fmt:
	gofmt -w .

docker-up:
	docker compose up -d

docker-down:
	docker compose down