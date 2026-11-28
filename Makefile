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

.PHONY: onlineconf-db-init onlineconf-migrate-up onlineconf-migrate-down onlineconf-migrate-version onlineconf-admin-bootstrap

ONLINECONF_MIGRATE_BIN ?= $(MIGRATE)
export ONLINECONF_ADMIN_DATABASE_URL ONLINECONF_MIGRATE_DATABASE_URL ONLINECONF_PRODUCT_DATABASE_URL
export ONLINECONF_MIGRATE_BIN ONLINECONF_OWNER_USER_ID

onlineconf-db-init:
	@bash db/onlineconf/manage.sh init

onlineconf-migrate-up:
	@bash db/onlineconf/manage.sh up

onlineconf-migrate-down:
	@bash db/onlineconf/manage.sh down

onlineconf-migrate-version:
	@bash db/onlineconf/manage.sh version

onlineconf-admin-bootstrap:
	@bash db/onlineconf/manage.sh bootstrap
