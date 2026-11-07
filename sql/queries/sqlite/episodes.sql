-- name: CreateEpisode :one
INSERT INTO episodes (
    title,
    slug,
    description,
    duration,
    explicit,
    cover_image_url,
    audio_source_url,
    audio_metadata,
    published_at
  )
VALUES (
    sqlc.narg('title'),
    sqlc.narg('slug'),
    sqlc.narg('description'),
    sqlc.narg('duration'),
    sqlc.narg('explicit'),
    sqlc.narg('cover_image_url'),
    sqlc.narg('audio_source_url'),
    sqlc.narg('audio_metadata'),
    sqlc.narg('published_at')
  )
RETURNING *;
-- name: GetEpisodeByID :one
SELECT *
FROM episodes
WHERE id = ?;
-- name: GetEpisodeBySlug :one
SELECT *
FROM episodes
WHERE slug = ?;
-- name: ListEpisodes :many
SELECT *
FROM episodes
WHERE (
    @search = ''
    OR title LIKE '%' || @search || '%'
  )
  AND (
    @status = ''
    OR CASE @status
      WHEN 'draft' THEN published_at IS NULL
        AND archived_at IS NULL
      WHEN 'scheduled' THEN published_at IS NOT NULL
        AND published_at > datetime('now')
        AND archived_at IS NULL
      WHEN 'published' THEN published_at IS NOT NULL
        AND published_at <= datetime('now')
        AND archived_at IS NULL
      WHEN 'archived' THEN archived_at IS NOT NULL
      ELSE false
    END
  )
ORDER BY published_at DESC,
  created_at DESC
LIMIT @page_limit OFFSET @page_offset;
-- name: CountEpisodesByStatus :one
SELECT COUNT(*) AS count
FROM episodes
WHERE CASE
    @status
    WHEN 'draft' THEN published_at IS NULL
    AND archived_at IS NULL
    WHEN 'scheduled' THEN published_at IS NOT NULL
    AND published_at > datetime('now')
    AND archived_at IS NULL
    WHEN 'published' THEN published_at IS NOT NULL
    AND published_at <= datetime('now')
    AND archived_at IS NULL
    WHEN 'archived' THEN archived_at IS NOT NULL
    ELSE false
  END;
-- name: ListPublishedEpisodes :many
SELECT *
FROM episodes
WHERE published_at IS NOT NULL
  AND published_at <=datetime('now')
  AND archived_at IS NULL
ORDER BY published_at DESC
LIMIT ? OFFSET ?;
-- name: ListPublishedEpisodesWithPagePath :many
SELECT sqlc.embed(e),
  p.path AS page_path
FROM episodes e
  LEFT JOIN pages p ON p.id = e.linked_page_id
    AND p.is_published = true
WHERE e.published_at IS NOT NULL
  AND e.published_at <= datetime('now')
  AND e.archived_at IS NULL
ORDER BY e.published_at DESC
LIMIT ? OFFSET ?;
-- name: UpdateEpisode :one
UPDATE episodes
SET title = COALESCE(sqlc.narg('title'), title),
  slug = COALESCE(sqlc.narg('slug'), slug),
  description = COALESCE(sqlc.narg('description'), description),
  duration = COALESCE(sqlc.narg('duration'), duration),
  explicit = COALESCE(sqlc.narg('explicit'), explicit),
  cover_image_url = COALESCE(sqlc.narg('cover_image_url'), cover_image_url),
  audio_source_url = COALESCE(sqlc.narg('audio_source_url'), audio_source_url),
  slides_md = COALESCE(sqlc.narg('slides_md'), slides_md),
  audio_metadata = COALESCE(sqlc.narg('audio_metadata'), audio_metadata),
  published_at = sqlc.narg('published_at'),
  archived_at = sqlc.narg('archived_at'),
  updated_at = datetime('now')
WHERE id = sqlc.arg('id')
RETURNING *;
-- name: UpdateEpisodeLinkedPageID :exec
UPDATE episodes
SET linked_page_id = ?,
  updated_at = datetime('now')
WHERE id = ?;
-- name: GetEpisodeByLinkedPageID :one
SELECT *
FROM episodes
WHERE linked_page_id = ?;
-- name: ListPublishedByLinkedPageID :many
-- Reverse direction of the episode->page link: the published episodes whose
-- companion page is this one. Feeds the public page's "listen" section.
SELECT *
FROM episodes
WHERE linked_page_id = ?
  AND published_at IS NOT NULL
  AND published_at <= datetime('now')
  AND archived_at IS NULL
ORDER BY published_at DESC;
-- name: DeleteEpisode :exec
DELETE FROM episodes
WHERE id = ?;
-- name: ListEpisodesForRanking :many
-- All non-archived episodes with a publish_at, in chronological order. The
-- service layer derives each episode's number (1-based) from row position here.
SELECT *
FROM episodes
WHERE published_at IS NOT NULL
  AND archived_at IS NULL
ORDER BY published_at ASC,
  id ASC;
-- name: SearchPublishedEpisodesWithPagePath :many
-- Like SearchPublishedEpisodes, but joins in the linked page's path so results
-- can link somewhere that actually exists (there is no /episodes/{slug} route).
SELECT sqlc.embed(e),
  p.path AS page_path
FROM episodes e
  LEFT JOIN pages p ON p.id = e.linked_page_id
    AND p.is_published = true
WHERE e.published_at IS NOT NULL
  AND e.published_at <= datetime('now')
  AND e.archived_at IS NULL
  AND (
    e.title LIKE '%' || @search || '%'
    OR e.description LIKE '%' || @search || '%'
  )
ORDER BY e.published_at DESC
LIMIT @page_limit OFFSET @page_offset;
