.PHONY: build test run clean

build:
	go build ./...

test:
	go test ./...

run:
	go run ./cmd/point-mcp

clean:
	go clean ./...
