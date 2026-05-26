-- name: GetPodcastConfig :one
SELECT * FROM podcast_config ORDER BY id ASC LIMIT 1;

-- name: UpdatePodcastConfig :one
UPDATE podcast_config
SET title = COALESCE(sqlc.narg('title'), title),
    description = COALESCE(sqlc.narg('description'), description),
    site_url = COALESCE(sqlc.narg('site_url'), site_url),
    language = COALESCE(sqlc.narg('language'), language),
    copyright = COALESCE(sqlc.narg('copyright'), copyright),
    author_name = COALESCE(sqlc.narg('author_name'), author_name),
    author_email = COALESCE(sqlc.narg('author_email'), author_email),
    cover_image_url = COALESCE(sqlc.narg('cover_image_url'), cover_image_url),
    category = COALESCE(sqlc.narg('category'), category),
    subcategory = COALESCE(sqlc.narg('subcategory'), subcategory),
    owner_name = COALESCE(sqlc.narg('owner_name'), owner_name),
    owner_email = COALESCE(sqlc.narg('owner_email'), owner_email),
    homepage_id = COALESCE(sqlc.narg('homepage_id'), homepage_id)
WHERE id = sqlc.arg('id')
RETURNING *;
