-- name: InsertProcessedFile :one
INSERT INTO analytics_processed_files (filename, file_date, entries_count)
VALUES (?, ?, ?)
RETURNING *;

-- name: GetProcessedFileByName :one
SELECT * FROM analytics_processed_files
WHERE filename = ?;

-- name: UpsertAccumulator :one
INSERT INTO analytics_download_accumulator (hash, episode_id, date, bytes_seen)
VALUES (?, ?, ?, ?)
ON CONFLICT (hash)
DO UPDATE SET bytes_seen = analytics_download_accumulator.bytes_seen + EXCLUDED.bytes_seen
RETURNING *;

-- name: MarkAccumulatorCounted :exec
UPDATE analytics_download_accumulator
SET counted = true
WHERE hash = ?;

-- name: UpsertPodcastDaily :one
INSERT INTO analytics_podcasts (podcast_id, date, bandwidth)
VALUES (?, ?, ?)
ON CONFLICT (podcast_id, date)
DO UPDATE SET
    hits = analytics_podcasts.hits + 1,
    bandwidth = analytics_podcasts.bandwidth + EXCLUDED.bandwidth,
    updated_at = datetime('now')
RETURNING *;

-- name: UpsertPodcastByEpisode :one
INSERT INTO analytics_podcasts_by_episode (podcast_id, episode_id, date, age)
VALUES (?, ?, ?, ?)
ON CONFLICT (podcast_id, episode_id, date)
DO UPDATE SET
    hits = analytics_podcasts_by_episode.hits + 1,
    updated_at = datetime('now')
RETURNING *;

-- name: UpsertPodcastByHour :one
INSERT INTO analytics_podcasts_by_hour (podcast_id, date, hour)
VALUES (?, ?, ?)
ON CONFLICT (podcast_id, date, hour)
DO UPDATE SET
    hits = analytics_podcasts_by_hour.hits + 1,
    updated_at = datetime('now')
RETURNING *;

-- name: UpsertPodcastByPlayer :one
INSERT INTO analytics_podcasts_by_player (podcast_id, date, service, app, device, os, is_bot)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT (podcast_id, date, service, app, device, os, is_bot)
DO UPDATE SET
    hits = analytics_podcasts_by_player.hits + 1,
    updated_at = datetime('now')
RETURNING *;

-- name: UpsertPodcastByCountry :one
INSERT INTO analytics_podcasts_by_country (podcast_id, date, country_code)
VALUES (?, ?, ?)
ON CONFLICT (podcast_id, date, country_code)
DO UPDATE SET
    hits = analytics_podcasts_by_country.hits + 1,
    updated_at = datetime('now')
RETURNING *;
