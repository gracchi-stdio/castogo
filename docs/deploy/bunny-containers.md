# Bunny Containers

Castogo is packaged as one immutable Docker image. Bunny Container should run
the image from GitHub Container Registry and expose the container's `PORT`
(default `8080`). The application does not run migrations during normal boot;
run the migration command once before switching traffic to a new image.

## Required environment

Set these as Bunny secrets or environment variables:

| Variable | Purpose |
| --- | --- |
| `DATABASE_URL` | PostgreSQL connection string |
| `SESSION_SECRET` | Long random cookie encryption secret |
| `ENV` | Set to `production` |
| `SITE_URL` | Public HTTPS URL |
| `AUDIO_PROCESSING_ENABLED` | Set to `false` to upload original audio without `ffmpeg` conversion |
| `BUNNY_STORAGE_ENDPOINT` | Bunny Storage upload endpoint |
| `BUNNY_STORAGE_PASSWORD` | Bunny Storage access key |
| `STORAGE_CDN` | Public CDN URL for uploaded audio |
| `BUNNY_API_KEY` | Bunny logging API key |
| `BUNNY_PULL_ZONE_ID` | Pull zone used for analytics |

Use `PORT=8080` unless the Bunny service is configured for another internal
port. The image includes `ffmpeg` and the CLI at `/app/castogo-cli`.

Set `AUDIO_PROCESSING_ENABLED=false` while audio conversion is disabled. The
application then uploads the original file format and skips the CPU-heavy
`ffmpeg` step. Duration, bitrate, sample rate, and channel metadata are not
derived in this mode, so the episode's technical metadata remains incomplete
until a metadata reader is added. The default is `true` to preserve the
normalization pipeline for existing deployments.

## Health checks

- `/healthz` is liveness-only and does not require PostgreSQL.
- `/readyz` and `/healthcheck` verify PostgreSQL connectivity.
- `/version` reports the image version, commit, and build time.

Configure Bunny's container health check to call `/healthz` on the internal
HTTP port. This lets Bunny replace an unhealthy process without treating a
short database outage as an application crash.

## Release and smooth update

The GitHub Actions workflow publishes:

- an immutable `sha-<commit>` image tag;
- `main` for the latest main-branch build;
- semantic-version tags when a `vMAJOR.MINOR.PATCH` tag is pushed.

Point Bunny at the immutable `sha-<commit>` tag for each release. Do not use
`latest`; immutable tags make rollback a tag change rather than a rebuild.

For each release:

1. Push the commit or semantic-version tag and wait for the image workflow.
2. Run `/app/castogo-cli migrate --dir /app/sql/migrations` once against the
   production database using the new image. Migrations must remain compatible
   with the currently running version until traffic has moved.
3. Update Bunny to the new `sha-<commit>` image and let its health check pass.
4. Verify `/version`, then monitor logs and `/readyz`.
5. Roll back by selecting the previous immutable image tag if needed.

The process handles `SIGTERM`, allowing Bunny's replacement process to drain
requests and stop the analytics worker cleanly. Keep schema changes
backward-compatible across one release so old and new containers can overlap.