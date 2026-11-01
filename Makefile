SQLC  := go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1
# Goose targets are silenced with @ so the DATABASE_URL in the command line never reaches logs.
GOOSE := go run github.com/pressly/goose/v3/cmd/goose@v3.29.0 -dir db/migrations postgres "$(DATABASE_URL)"

.PHONY: generate build test lint run migrate migrate-status migrate-create

generate:
	cd api && buf lint && buf generate
	$(SQLC) generate

build:
	go build ./...

test:
	go test -race ./...

lint:
	golangci-lint run ./...

run:
	go run ./cmd/app

migrate:
	@$(GOOSE) up

migrate-status:
	@$(GOOSE) status

# Usage: make migrate-create NAME=add_orders_table
migrate-create:
	@$(GOOSE) create $(NAME) sql
