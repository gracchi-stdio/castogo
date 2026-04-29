-- name: CreateEpisode :one
INSERT INTO episodes (title, slug, episode_number, description, duration, explicit, cover_image_url, audio_source_url, audio_metadata, published_at, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, sqlc.narg('audio_metadata'), $9, $10)
RETURNING *;

-- name: GetEpisodeByID :one
SELECT * FROM episodes WHERE id = $1;

-- name: GetEpisodeBySlug :one
SELECT * FROM episodes WHERE slug = $1;

-- name: ListEpisodes :many
SELECT * FROM episodes
WHERE (@status = '' OR status::text = @status)
  AND (@search = '' OR title ILIKE '%' || @search || '%')
ORDER BY episode_number DESC
LIMIT @page_limit OFFSET @page_offset;

-- name: CountEpisodesByStatus :one
SELECT COUNT(*) AS count FROM episodes WHERE status = @status::episode_status;

-- name: ListPublishedEpisodes :many
SELECT * FROM episodes
WHERE status = 'published'
ORDER BY published_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateEpisode :one
UPDATE episodes
SET title = COALESCE(sqlc.narg('title'), title),
    slug = COALESCE(sqlc.narg('slug'), slug),
    episode_number = COALESCE(sqlc.narg('episode_number'), episode_number),
    description = COALESCE(sqlc.narg('description'), description),
    duration = COALESCE(sqlc.narg('duration'), duration),
    explicit = COALESCE(sqlc.narg('explicit'), explicit),
    cover_image_url = COALESCE(sqlc.narg('cover_image_url'), cover_image_url),
    audio_source_url = COALESCE(sqlc.narg('audio_source_url'), audio_source_url),
    audio_metadata = COALESCE(sqlc.narg('audio_metadata'), audio_metadata),
    published_at = COALESCE(sqlc.narg('published_at'), published_at),
    status = COALESCE(sqlc.narg('status'), status),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: DeleteEpisode :exec
DELETE FROM episodes WHERE id = $1;

-- name: GetMaxEpisodeNumber :one
SELECT COALESCE(MAX(episode_number), 1) FROM episodes;
