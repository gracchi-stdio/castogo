-- name: InsertProcessedFile :one
INSERT INTO analytics_processed_files (filename, file_date, entries_count)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetProcessedFileByName :one
SELECT * FROM analytics_processed_files
WHERE filename = $1;

-- name: UpsertAccumulator :one
INSERT INTO analytics_download_accumulator (hash, episode_id, date, bytes_seen)
VALUES ($1, $2, $3, $4)
ON CONFLICT (hash)
DO UPDATE SET bytes_seen = analytics_download_accumulator.bytes_seen + EXCLUDED.bytes_seen
RETURNING *;

-- name: MarkAccumulatorCounted :exec
UPDATE analytics_download_accumulator
SET counted = true
WHERE hash = $1;

-- name: UpsertPodcastDaily :one
INSERT INTO analytics_podcasts (podcast_id, date, bandwidth)
VALUES ($1, $2, $3)
ON CONFLICT (podcast_id, date)
DO UPDATE SET
    hits = analytics_podcasts.hits + 1,
    bandwidth = analytics_podcasts.bandwidth + EXCLUDED.bandwidth,
    updated_at = NOW()
RETURNING *;

-- name: UpsertPodcastByEpisode :one
INSERT INTO analytics_podcasts_by_episode (podcast_id, episode_id, date, age)
VALUES ($1, $2, $3, $4)
ON CONFLICT (podcast_id, episode_id, date)
DO UPDATE SET
    hits = analytics_podcasts_by_episode.hits + 1,
    updated_at = NOW()
RETURNING *;

-- name: UpsertPodcastByHour :one
INSERT INTO analytics_podcasts_by_hour (podcast_id, date, hour)
VALUES ($1, $2, $3)
ON CONFLICT (podcast_id, date, hour)
DO UPDATE SET
    hits = analytics_podcasts_by_hour.hits + 1,
    updated_at = NOW()
RETURNING *;

-- name: UpsertPodcastByPlayer :one
INSERT INTO analytics_podcasts_by_player (podcast_id, date, service, app, device, os, is_bot)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (podcast_id, date, service, app, device, os, is_bot)
DO UPDATE SET
    hits = analytics_podcasts_by_player.hits + 1,
    updated_at = NOW()
RETURNING *;

-- name: UpsertPodcastByCountry :one
INSERT INTO analytics_podcasts_by_country (podcast_id, date, country_code)
VALUES ($1, $2, $3)
ON CONFLICT (podcast_id, date, country_code)
DO UPDATE SET
    hits = analytics_podcasts_by_country.hits + 1,
    updated_at = NOW()
RETURNING *;
