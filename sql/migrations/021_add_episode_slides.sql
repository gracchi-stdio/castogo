-- Per-episode slideshow as a single markdown source (Slidev-style).
-- Slides are separated by lines that are exactly "---"; a slide may open with
-- directive lines (e.g. "time: 03:12" for audio-synced reveal) before its
-- first blank line. Parsed at render time — see pageview/blocks/slideshow.templ.
ALTER TABLE episodes ADD COLUMN slides_md TEXT NOT NULL DEFAULT '';
