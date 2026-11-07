-- name: CreateBlock :one
INSERT INTO page_blocks (content, sort_order, block_type, page_id)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: GetBlocksByPageID :many
SELECT * FROM page_blocks WHERE page_id = ? ORDER BY sort_order ASC;

-- name: UpdateBlock :one
UPDATE page_blocks
SET content = COALESCE(sqlc.narg('content'), content),
    sort_order = COALESCE(sqlc.narg('sort_order'), sort_order),
    updated_at = datetime('now')
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: DeleteBlock :exec
DELETE FROM page_blocks WHERE id = ?;

-- name: UpdateBlockOrder :exec
UPDATE page_blocks
SET sort_order = ?, updated_at = datetime('now')
WHERE id = ?;
