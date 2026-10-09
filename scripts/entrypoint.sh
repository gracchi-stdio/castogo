#!/bin/sh
# Container entrypoint: apply pending database migrations, then hand off to
# the server binary.
#
# - Runs BEFORE the server starts, so the app never serves against a schema
#   it doesn't expect.
# - `set -e`: a failed migration aborts the container (Bunny shows it as
#   crashing) instead of starting a half-migrated app.
# - `exec "$@"`: replaces this shell with the server, so SIGTERM from the
#   platform (STOPSIGNAL in the Dockerfile) reaches the Go binary directly
#   for graceful shutdown.
set -e

# Only auto-migrate for the default server command, so explicit command
# overrides keep working: `podman run image /app/castogo-cli migrate --dry-run`
# runs the CLI directly instead of migrating and then exec'ing it as PID 1.
if [ "$1" = "/app/castogo" ]; then
    echo "[entrypoint] applying database migrations"
    /app/castogo-cli migrate --dir /app/sql/migrations
fi

echo "[entrypoint] starting: $1"
exec "$@"
