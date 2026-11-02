-- Replace the boolean show_in_nav with an explicit nav position:
-- nav_order IS NULL  -> page is not shown in the public nav
-- nav_order = N      -> page is shown, ascending order (smallest first)
-- Ties break on path in the nav query.
ALTER TABLE pages ADD COLUMN nav_order INTEGER;

-- Pages that were opted in keep their visibility; all start at position 0
-- and are disambiguated by the tie-break (path) until reordered.
UPDATE pages SET nav_order = 0 WHERE show_in_nav = TRUE;

ALTER TABLE pages DROP COLUMN show_in_nav;
