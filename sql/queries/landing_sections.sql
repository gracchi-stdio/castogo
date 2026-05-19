-- name: GetAllLandingSections :many
SELECT * FROM landing_page_sections ORDER BY sort_order ASC;

-- name: GetVisibleLandingSections :many
SELECT * FROM landing_page_sections WHERE is_visible = true ORDER BY sort_order ASC;

-- name: GetLandingSectionByKey :one
SELECT * FROM landing_page_sections WHERE section_key = $1;

-- name: UpdateLandingSection :one
UPDATE landing_page_sections
SET content = COALESCE(sqlc.narg('content'), content),
    is_visible = COALESCE(sqlc.narg('is_visible'), is_visible),
    sort_order = COALESCE(sqlc.narg('sort_order'), sort_order),
    updated_at = now()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: UpdateLandingSectionOrder :exec
UPDATE landing_page_sections
SET sort_order = $2,
    updated_at = now()
WHERE id = $1;
