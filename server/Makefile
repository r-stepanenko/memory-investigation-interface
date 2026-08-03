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

generate:
	go run ./cmd/generate --count 100000 -out internal/datasets/events100k.jsonl
	go run ./cmd/generate --count 1000000 -out internal/datasets/events1m.jsonl

clean:
	go clean

demo:
	go mod download
	go build ./...
	go test ./...