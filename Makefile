.PHONY: run build test coverage lint fmt docker-up docker-down migrate-up migrate-down migrate-version proto-tools proto

run:
	go run ./cmd/server

build:
	mkdir -p bin
	go build -o bin/server ./cmd/server

test:
	go test ./...

coverage:
	go test ./... -coverprofile=coverage.out
	awk 'NR==1 || $$1 !~ /\/mock_[^/]*\.go:/' coverage.out > coverage_no_mocks.out
	go tool cover -func=coverage_no_mocks.out | grep total

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

PROTO_DIR := api/proto
PROTO_OUT := gen/go
PROTO_FILES := $(shell find $(PROTO_DIR) -name '*.proto')
export PATH := $(shell go env GOPATH)/bin:$(PATH)

proto-tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.12
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.6.2

proto:
	mkdir -p $(PROTO_OUT)
	protoc \
 		-I $(PROTO_DIR) \
 		--go_out=$(PROTO_OUT) \
 		--go_opt=paths=source_relative \
 		--go-grpc_out=$(PROTO_OUT) \
 		--go-grpc_opt=paths=source_relative \
 		$(PROTO_FILES)
 		