-- name: CreateBlock :one
INSERT INTO page_blocks (content, sort_order, block_type, page_id)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetBlocksByPageID :many
SELECT * FROM page_blocks WHERE page_id = $1 ORDER BY sort_order ASC;

-- name: UpdateBlock :one
UPDATE page_blocks
SET content = COALESCE(sqlc.narg('content'), content),
    sort_order = COALESCE(sqlc.narg('sort_order'), sort_order),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: DeleteBlock :exec
DELETE FROM page_blocks WHERE id = $1;
