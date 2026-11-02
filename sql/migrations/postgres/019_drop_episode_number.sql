-- Episode numbers are now DERIVED from publication order (publish_at), not
-- stored. Drop the column; ranking is computed in the service layer via
-- ListEpisodesForRanking. This also makes feed GUIDs switch to the immutable
-- episode id (see feed_service.buildItem), so existing podcast clients will
-- re-fetch episodes once — acceptable for a project with no subscribers yet.
ALTER TABLE episodes DROP COLUMN IF EXISTS episode_number;
