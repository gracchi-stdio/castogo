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

# Run a specific migration by prefix (e.g. "just migrate 013")
[unix]
migrate file="":
	@if [ -z "{{file}}" ]; then echo "Usage: just migrate <prefix>\n\nAvailable migrations:"; ls sql/migrations/*.sql | sed 's/sql\/migrations\//  /'; exit 0; fi
	@matching=$$(ls sql/migrations/{{file}}*.sql 2>/dev/null); \
	if [ -z "$$matching" ]; then echo "No migration found matching '{{file}}'"; exit 1; fi; \
	echo "Running: $$matching"; \
	podman exec -i castogo_db_1 psql -U castogo -d castogo -f /dev/stdin < $$matching

[windows]
migrate file="":
	@powershell -NoProfile -Command "if ('{{file}}' -eq '') { Write-Host 'Usage: just migrate <prefix>'; Write-Host; Write-Host 'Available migrations:'; Get-ChildItem sql/migrations/*.sql | ForEach-Object { Write-Host ('  ' + $_.Name) } } else { $$m = Get-Item sql/migrations/{{file}}*.sql -ErrorAction SilentlyContinue; if (-not $$m) { Write-Host 'No migration found'; exit 1 }; Write-Host ('Running: ' + $$m.FullName); Get-Content $$m.FullName | docker exec -i castogo-db-1 psql -U castogo -d castogo }"

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
