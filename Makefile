.PHONY: build run test generate docker-up docker-down clean dev

dev: generate
	go run ./cmd/server

build: generate
	go build -o bin/castogo ./cmd/server

run: build
	./bin/castogo

test:
	go test -v ./... -count=1

generate:
	sqlc generate
	templ generate

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down

clean:
	rm -rf bin/ internal/db/ tmp/
