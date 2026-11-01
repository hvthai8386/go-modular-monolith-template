SQLC := go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1

.PHONY: generate build test lint run

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
