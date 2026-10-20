.PHONY: run build test coverage lint fmt docker-up docker-down migrate-up migrate-down migrate-version

run:
	go run ./cmd/server

build:
	mkdir -p bin
	go build -o bin/server ./cmd/server

test:
	go test ./...

coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out | grep total
lint:
	go vet ./...

fmt:
	gofmt -w .

docker-up:
	docker compose up -d

docker-down:
	docker compose down

MIGRATE := $(shell go env GOPATH)/bin/migrate
DATABASE_URL ?= postgres://villain@127.0.0.1:5432/villain?sslmode=disable

migrate-up:
	$(MIGRATE) -path db/migrations -database "$(DATABASE_URL)" up

migrate-down:
	$(MIGRATE) -path db/migrations -database "$(DATABASE_URL)" down 1

migrate-version:
	$(MIGRATE) -path db/migrations -database "$(DATABASE_URL)" version
