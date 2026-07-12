-- name: CreateEpisode :one
INSERT INTO episodes (
    title,
    slug,
    episode_number,
    description,
    duration,
    explicit,
    cover_image_url,
    audio_source_url,
    audio_metadata,
    published_at
  )
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    sqlc.narg('audio_metadata'),
    $9
  )
RETURNING *;
-- name: GetEpisodeByID :one
SELECT *
FROM episodes
WHERE id = $1;
-- name: GetEpisodeBySlug :one
SELECT *
FROM episodes
WHERE slug = $1;
-- name: ListEpisodes :many
SELECT *
FROM episodes
WHERE (
    @search = ''
    OR title ILIKE '%' || @search || '%'
  )
ORDER BY episode_number DESC
LIMIT @page_limit OFFSET @page_offset;
-- name: CountEpisodesByStatus :one
SELECT COUNT(*) AS count
FROM episodes
WHERE CASE
    @status::text
    WHEN 'draft' THEN published_at IS NULL
    AND archived_at IS NULL
    WHEN 'scheduled' THEN published_at IS NOT NULL
    AND published_at > NOW()
    AND archived_at IS NULL
    WHEN 'published' THEN published_at IS NOT NULL
    AND published_at <= NOW()
    AND archived_at IS NULL
    WHEN 'archived' THEN archived_at IS NOT NULL
    ELSE false
  END;
-- name: ListPublishedEpisodes :many
SELECT sqlc.embed(e),
  p.path AS page_path
FROM episodes e
  LEFT JOIN pages p ON p.id = e.linked_page_id
WHERE e.published_at IS NOT NULL
  AND e.published_at <= NOW()
  AND e.archived_at IS NULL
ORDER BY e.published_at DESC
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
  published_at = sqlc.narg('published_at'),
  archived_at = sqlc.narg('archived_at'),
  updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING *;
-- name: UpdateEpisodeLinkedPageID :exec
UPDATE episodes
SET linked_page_id = $2,
  updated_at = NOW()
WHERE id = $1;
-- name: GetEpisodeByLinkedPageID :one
SELECT *
FROM episodes
WHERE linked_page_id = $1;
-- name: DeleteEpisode :exec
DELETE FROM episodes
WHERE id = $1;
-- name: GetMaxEpisodeNumber :one
SELECT COALESCE(MAX(episode_number), 0)::bigint
FROM episodes;
-- name: SearchPublishedEpisodes :many
SELECT *
FROM episodes
WHERE published_at IS NOT NULL
  AND published_at <= NOW()
  AND archived_at IS NULL
  AND (
    title ILIKE '%' || @search || '%'
    OR description ILIKE '%' || @search || '%'
  )
ORDER BY published_at DESC
LIMIT @page_limit OFFSET @page_offset;