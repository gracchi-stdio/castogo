-- Normalize page paths to full URL paths (always starting with "/").
-- The homepage (empty slug) becomes "/"; every other stored path gets the
-- leading slash it was missing. Children of the homepage are detached first —
-- they were never valid (the homepage cannot be a parent), and detaching
-- before normalizing puts them at "/<slug>" instead of a double-slash path.
UPDATE pages SET parent_id = NULL
WHERE parent_id IN (SELECT id FROM pages WHERE slug = '');

UPDATE pages SET path = '/' || path WHERE path NOT LIKE '/%';

UPDATE pages SET path = '/' WHERE path = '';
