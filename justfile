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
dev: check-deps
	bash -lc "echo 'Starting dev environment...'; just dev-vite & just dev-templ & just dev-server & wait"

# Check that required dev dependencies are installed
[unix]
check-deps:
	@# ffmpeg
	@if ! command -v ffmpeg > /dev/null 2>&1; then echo "Missing dependency: ffmpeg\nInstall with: sudo apt install ffmpeg"; exit 1; fi

# Vite dev server (HMR on localhost:3000)
dev-vite:
	bun run dev -- --port 3000

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
	bun run build

# Build the Go binary (runs generate + frontend-build first)
build: generate frontend-build
	go build -o bin/castogo ./cmd/server

# Build the castogo-cli utility binary to bin/ (gitignored)
build-cli:
	go build -o bin/castogo-cli ./cmd/cli

# Run the built binary
[windows]
run: build
	.\bin\castogo

[unix]
run: build
	./bin/castogo

# Build assets and run server in production mode (no Vite dev server)
[unix]
prod: build
	ENV=production ./bin/castogo

# Run tests
test:
	go test -v ./... -count=1

# Apply pending database migrations (same code path as a production deploy)
migrate:
	go run ./cmd/cli migrate --dir sql/migrations

# List pending migrations without applying them
migrate-status:
	go run ./cmd/cli migrate --dir sql/migrations --dry-run

# Record all migration files as applied WITHOUT running them — for dev DBs
# seeded by docker-entrypoint-initdb.d or databases migrated by hand (psql)
migrate-adopt:
	go run ./cmd/cli migrate-adopt --dir sql/migrations --yes

# Run any castogo-cli command against the local dev database (from .env).
# e.g.  just cli migrate-status   |   just cli register --email me@x.io --generate
cli *args:
	go run ./cmd/cli {{args}}

# Run any castogo-cli command against the remote/production database.
# Set REMOTE_DATABASE_URL (e.g. the Aiven service URI) in .env — never commit it.
# e.g.  just cli-prod register --email me@x.io --generate
cli-prod *args:
	go run ./cmd/cli --database-url "{{env_var("REMOTE_DATABASE_URL")}}" {{args}}

# Docker
docker-up:
	docker compose up -d --build

docker-migrate:
	docker compose run --rm app /app/castogo-cli migrate --dir /app/sql/migrations

docker-build version="dev" commit="local" build_time="local":
	docker build --build-arg VERSION={{version}} --build-arg COMMIT={{commit}} --build-arg BUILD_TIME={{build_time}} -t castogo:{{version}} .

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
install:
	bun install
