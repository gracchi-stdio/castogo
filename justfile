set dotenv-load

# Default recipe — list available commands
default:
    @just --list

# Start dev: Vite dev server (with HMR), templ watcher, and Go server via air
dev:
    @echo "Starting dev environment..."
    @just dev-vite & just dev-templ & just dev-server; kill 0

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

# Build frontend for production
frontend-build:
    yarn build

# Build the Go binary (runs generate + frontend-build first)
build: generate frontend-build
    go build -o bin/castogo ./cmd/server

# Run the built binary
run: build
    ./bin/castogo

# Run tests
test:
    go test -v ./... -count=1

# Docker
docker-up:
    docker compose up -d --build

docker-down:
    docker compose down

# Clean build artifacts
clean:
    rm -rf bin/ internal/db/ tmp/ public/assets/ public/.vite/

# Install frontend dependencies
yarn:
    yarn install
