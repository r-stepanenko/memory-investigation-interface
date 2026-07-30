.PHONY: install build run serve test bench clean demo

install:
	go mod download

build:
	go build ./...

run:
	go run ./cmd/event-memory-search-api

serve:
	go run ./cmd/event-memory-search-api

test:
	go test ./...

bench:
	go test -run='^$' -bench=. -benchmem ./internal/search/...

clean:
	go clean

demo:
	go mod download
	go build ./...
	go test ./...