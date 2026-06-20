-- name: CreatePage :one
INSERT INTO pages (
    title,
    slug,
    parent_id,
    layout,
    is_published,
    show_in_nav,
    metadata,
    path,
    sort_order
  )
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;
-- name: GetPageByID :one
SELECT *
FROM pages
WHERE id = $1;
-- name: GetPageByPath :one
SELECT *
FROM pages
WHERE path = $1;
-- name: GetPageBySlug :one
SELECT *
FROM pages
WHERE slug = $1;
-- name: ListPages :many
SELECT *
FROM pages
WHERE (
    @search = ''
    OR title ILIKE '%' || @search || '%'
  )
ORDER BY path ASC
LIMIT @page_limit OFFSET @page_offset;
-- name: GetChildren :many
SELECT *
FROM pages
WHERE parent_id = $1
ORDER BY sort_order ASC;
-- name: GetDescendants :many
SELECT *
FROM pages
WHERE path LIKE sqlc.arg('prefix') || '/%'
ORDER BY path ASC;
-- name: UpdatePage :one
UPDATE pages
SET title = COALESCE(sqlc.narg('title'), title),
  slug = COALESCE(sqlc.narg('slug'), slug),
  layout = COALESCE(sqlc.narg('layout'), layout),
  parent_id = COALESCE(sqlc.narg('parent_id'), parent_id),
  is_published = COALESCE(sqlc.narg('is_published'), is_published),
  show_in_nav = COALESCE(sqlc.narg('show_in_nav'), show_in_nav),
  metadata = COALESCE(sqlc.narg('metadata'), metadata),
  path = COALESCE(sqlc.narg('path'), path),
  sort_order = COALESCE(sqlc.narg('sort_order'), sort_order),
  updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING *;
-- name: DeletePage :exec
DELETE FROM pages
WHERE id = $1;
-- name: UpdateDescendantPaths :exec
UPDATE pages
SET path = REPLACE(
    path,
    sqlc.arg('old_prefix') || '/',
    sqlc.arg('new_prefix') || '/'
  ),
  updated_at = NOW()
WHERE path LIKE sqlc.arg('old_prefix') || '/%';
-- name: GetPageSiblings :many
SELECT *
FROM pages
WHERE parent_id = $1
  AND id != $2
ORDER BY sort_order ASC;
-- name: GetChildrenCount :one
SELECT COUNT(*)
FROM pages
WHERE parent_id = $1;
-- name: GetPagePathAndChildrenCountByID :one
SELECT p.path,
  (
    SELECT COUNT(*)
    FROM pages
    WHERE parent_id = p.id
  ) AS children_count
FROM pages p
WHERE p.id = $1;
-- name: CountPageWithoutParent :one
SELECT COUNT(*)
FROM pages
WHERE parent_id IS NULL;
-- name: GetPublishedTopLevelPages :many
SELECT *
FROM pages
WHERE parent_id IS NULL
  AND is_published = true
  AND show_in_nav = true
ORDER BY sort_order ASC;
-- name: SearchPublishedPages :many
SELECT *
FROM pages
WHERE is_published = true
  AND title ILIKE '%' || @search || '%'
ORDER BY sort_order ASC
LIMIT @page_limit OFFSET @page_offset;