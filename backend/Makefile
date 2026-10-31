.PHONY: run test lint build

run:
	go run ./cmd/server

test:
	go test -race ./...

lint:
	go vet ./...
	@command -v golangci-lint >/dev/null && golangci-lint run || echo "golangci-lint not installed, skipped"

build:
	go build -o bin/server ./cmd/server
