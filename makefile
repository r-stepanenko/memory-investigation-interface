.PHONY: install build run-server run-front serve demo test test-server test-front bench

install:
	cd server && go mod download
	cd front && npm install

build:
	cd server && go build ./...
	cd front && npm run build

run-server:
	cd server && go run ./cmd/event-memory-search-api

run-front:
	cd front && npm run dev

serve:
	powershell -Command "Start-Process powershell -ArgumentList '-NoExit','-Command','cd server; go run ./cmd/event-memory-search-api'"
	powershell -Command "Start-Process powershell -ArgumentList '-NoExit','-Command','cd front; npm run dev'"

test-server:
	cd server && go test ./...

test-front:
	cd front && npm run test:run

test:
	cd server && go test ./...
	cd front && npm run test:run

demo:
	cd server && go mod download
	cd front && npm install
	cd server && go build ./...
	cd front && npm run build
	cd server && go test ./...
	cd front && npm run test:run
	powershell -Command "Start-Process powershell -ArgumentList '-NoExit','-Command','cd server; go run ./cmd/event-memory-search-api'"
	powershell -Command "Start-Process powershell -ArgumentList '-NoExit','-Command','cd front; npm run dev'"

bench:
	cd server && go test -run='^$$' -bench=. -benchmem ./internal/search/...