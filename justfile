set dotenv-load
set windows-shell := ["powershell.exe", "-NoProfile", "-Command"]

# Default recipe — list available commands
default:
	@just --list

# Start dev: Vite dev server (with HMR), templ watcher, and Go server via air
[windows]
dev:
	powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-dev.ps1

[unix]
dev:
	bash -lc "echo 'Starting dev environment...'; just dev-vite & just dev-templ & just dev-server & wait"

# Vite dev server (HMR on localhost:3000)
dev-vite:
	yarn dev --port 3000

# templ file watcher
dev-templ:
	templ generate -watch

# Go server with air (auto-rebuild + restart on .go changes)
dev-server:
	air -c .air.toml

# One-shot generate (sqlc + templ)
generate:
	sqlc generate
	templ generate

# Build CSS
[windows]
css-build:
	powershell -ExecutionPolicy Bypass -File scripts/build-css.ps1

[unix]
css-build:
	bash scripts/build-css.sh

# Build frontend for production
frontend-build:
	yarn build

# Build the Go binary (runs generate + frontend-build first)
build: generate frontend-build
	go build -o bin/castogo ./cmd/server

# Run the built binary
[windows]
run: build
	.\bin\castogo

[unix]
run: build
	./bin/castogo

# Run tests
test:
	go test -v ./... -count=1

# Docker
docker-up:
	docker compose up -d --build

podman-db-up:
	podman compose up -d db

docker-down:
	docker compose down

# Clean build artifacts
[windows]
clean:
	powershell -NoProfile -ExecutionPolicy Bypass -Command "Remove-Item -Path 'bin','internal/db','tmp','public/assets','public/.vite' -Recurse -Force -ErrorAction SilentlyContinue"

[unix]
clean:
	rm -rf bin internal/db tmp public/assets public/.vite

# Install frontend dependencies
yarn:
	yarn install
