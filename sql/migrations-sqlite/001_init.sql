-- Castogo baseline schema for SQLite / libSQL (Bunny DB).
-- Squash of sql/migrations/001-021 (postgres): dropped-only history is
-- omitted (landing_page_sections 009/010/011, homepage_id 013/015,
-- episode_number 019, episode status column 002/007), and the 020 path
-- normalization is baked into the seed data.
-- show_in_nav (018) is born as nav_order (022): NULL = not in nav, else
-- ascending sort position.
-- Translations: UUID -> TEXT + randomblob default, TIMESTAMPTZ/DATE -> TEXT
-- with datetime('now') defaults, BIGSERIAL/IDENTITY -> INTEGER PRIMARY KEY
-- (rowid alias), JSONB -> TEXT, enum -> TEXT (no enum columns survive 007).

CREATE TABLE users (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now'))
);

CREATE TABLE podcast_config (
    id              INTEGER PRIMARY KEY,
    title           TEXT NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    site_url        TEXT NOT NULL DEFAULT '',
    language        TEXT NOT NULL DEFAULT 'en-us',
    copyright       TEXT NOT NULL DEFAULT '',
    author_name     TEXT NOT NULL DEFAULT '',
    author_email    TEXT NOT NULL DEFAULT '',
    cover_image_url TEXT NOT NULL DEFAULT '',
    category        TEXT NOT NULL DEFAULT '',
    subcategory     TEXT NOT NULL DEFAULT '',
    owner_name      TEXT NOT NULL DEFAULT '',
    owner_email     TEXT NOT NULL DEFAULT ''
);

CREATE TABLE pages (
    id           INTEGER PRIMARY KEY,
    slug         TEXT NOT NULL,
    title        TEXT NOT NULL,
    parent_id    INTEGER REFERENCES pages(id) ON DELETE SET NULL,
    layout       TEXT NOT NULL DEFAULT 'default',
    is_published BOOLEAN NOT NULL DEFAULT false,
    metadata     TEXT NOT NULL DEFAULT '{}',
    path         TEXT NOT NULL UNIQUE,
    sort_order   INTEGER NOT NULL DEFAULT 0,
    nav_order    INTEGER,
    created_at   TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at   TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Prevents two siblings from having the same slug (expression index)
CREATE UNIQUE INDEX idx_pages_slug_parent ON pages (slug, COALESCE(parent_id, 0));

CREATE TABLE episodes (
    id              INTEGER PRIMARY KEY,
    title           TEXT NOT NULL,
    slug            TEXT UNIQUE NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    duration        INTEGER NOT NULL DEFAULT 0,
    explicit        BOOLEAN NOT NULL DEFAULT FALSE,
    cover_image_url TEXT,
    audio_source_url TEXT,
    published_at    TEXT,
    audio_metadata  TEXT NOT NULL DEFAULT '{}',
    archived_at     TEXT,
    linked_page_id  INTEGER REFERENCES pages(id) ON DELETE SET NULL,
    slides_md       TEXT NOT NULL DEFAULT '',
    created_at      TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at      TEXT DEFAULT (datetime('now'))
);

-- Derived status: published_at IS NULL -> draft; > now() -> scheduled;
-- <= now() -> published; archived_at IS NOT NULL -> archived (overrides).
CREATE INDEX idx_episodes_published ON episodes (published_at DESC)
    WHERE published_at IS NOT NULL AND archived_at IS NULL;
CREATE INDEX idx_episodes_drafts ON episodes (created_at DESC)
    WHERE published_at IS NULL AND archived_at IS NULL;
CREATE INDEX idx_episodes_archived ON episodes (archived_at DESC)
    WHERE archived_at IS NOT NULL;
CREATE INDEX idx_episodes_linked_page_id ON episodes(linked_page_id)
    WHERE linked_page_id IS NOT NULL;

CREATE TABLE page_blocks (
    id          INTEGER PRIMARY KEY,
    page_id     INTEGER NOT NULL REFERENCES pages(id) ON DELETE CASCADE,
    block_type  TEXT NOT NULL,
    content     TEXT NOT NULL DEFAULT '{}',
    sort_order  INTEGER NOT NULL DEFAULT 0,
    status      TEXT NOT NULL DEFAULT 'active',
    created_at  TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at  TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE http_sessions (
    id         INTEGER PRIMARY KEY,
    key        TEXT NOT NULL UNIQUE,
    data       TEXT NOT NULL,
    created_on TEXT DEFAULT (datetime('now')),
    expires_on TEXT NOT NULL
);

CREATE INDEX http_sessions_expiry_idx ON http_sessions (expires_on);
CREATE INDEX http_sessions_key_idx ON http_sessions (key);

-- Analytics cluster (006): daily rollups keyed on (podcast_id, date, dimension).

CREATE TABLE analytics_processed_files (
    id            INTEGER PRIMARY KEY,
    filename      TEXT NOT NULL UNIQUE,
    file_date     TEXT NOT NULL,
    entries_count INTEGER NOT NULL DEFAULT 0,
    processed_at  TEXT NOT NULL DEFAULT (datetime('now'))
);

-- IAB byte accumulator: hash = SHA1(date+ip+user_agent+episode_id);
-- once bytes_seen passes the 1-min threshold, counted = true
CREATE TABLE analytics_download_accumulator (
    hash       TEXT PRIMARY KEY,
    episode_id INTEGER NOT NULL,
    date       TEXT NOT NULL,
    bytes_seen INTEGER NOT NULL DEFAULT 0,
    counted    BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE analytics_podcasts (
    podcast_id      INTEGER NOT NULL,
    date            TEXT NOT NULL,
    hits            INTEGER NOT NULL DEFAULT 1,
    bandwidth       INTEGER NOT NULL DEFAULT 0,
    unique_listeners INTEGER NOT NULL DEFAULT 1,
    created_at      TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at      TEXT NOT NULL DEFAULT (datetime('now')),
    PRIMARY KEY (podcast_id, date)
);

CREATE TABLE analytics_podcasts_by_episode (
    podcast_id INTEGER NOT NULL,
    episode_id INTEGER NOT NULL,
    date       TEXT NOT NULL,
    age        INTEGER NOT NULL,
    hits       INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    PRIMARY KEY (podcast_id, episode_id, date)
);

CREATE TABLE analytics_podcasts_by_hour (
    podcast_id INTEGER NOT NULL,
    date       TEXT NOT NULL,
    hour       INTEGER NOT NULL,
    hits       INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    PRIMARY KEY (podcast_id, date, hour)
);

CREATE TABLE analytics_podcasts_by_player (
    podcast_id INTEGER NOT NULL,
    date       TEXT NOT NULL,
    service    TEXT NOT NULL DEFAULT '',
    app        TEXT NOT NULL DEFAULT '',
    device     TEXT NOT NULL DEFAULT '',
    os         TEXT NOT NULL DEFAULT '',
    is_bot     BOOLEAN NOT NULL DEFAULT FALSE,
    hits       INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    PRIMARY KEY (podcast_id, date, service, app, device, os, is_bot)
);

CREATE TABLE analytics_podcasts_by_country (
    podcast_id   INTEGER NOT NULL,
    date         TEXT NOT NULL,
    country_code TEXT NOT NULL,
    hits         INTEGER NOT NULL DEFAULT 1,
    created_at   TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at   TEXT NOT NULL DEFAULT (datetime('now')),
    PRIMARY KEY (podcast_id, date, country_code)
);

-- Seeds -------------------------------------------------------------------

INSERT INTO podcast_config (title)
VALUES ('')
ON CONFLICT DO NOTHING;

-- Sample public page for previewing theme + blocks (postgres 100, with
-- '{}'::jsonb casts dropped, paths already normalized per 020, and the
-- postgres-only data-modifying CTE restructured: sqlite CTEs are SELECT-only,
-- so the seed is two plain idempotent statements).
INSERT INTO pages (title, slug, parent_id, layout, is_published, metadata, path, sort_order)
VALUES ('Sample Page', 'sample', NULL, 'default', true, '{}', '/sample', 0)
ON CONFLICT (path) DO NOTHING;

-- Blocks only seed when the page exists and has none yet.
INSERT INTO page_blocks (page_id, block_type, content, sort_order)
SELECT p.id, 'hero', '{"headline":"Castogo: your podcast, your rules","subheadline":"Self-host your show with a clean workflow, beautiful public pages, and RSS that just works.","cta_text":"Explore episodes","cta_url":"/episodes","background_image":""}', 0
FROM pages p
WHERE p.path = '/sample' AND NOT EXISTS (SELECT 1 FROM page_blocks b WHERE b.page_id = p.id)
UNION ALL
SELECT p.id, 'features', '{"section_title":"Publish with confidence","section_description":"The building blocks you need for a polished podcast site.","items":[{"icon":"mic","title":"Episode workflow","description":"Draft, schedule, publish, and archive from one dashboard."},{"icon":"rss","title":"RSS done right","description":"Compliant feeds for Apple, Spotify, and Google."},{"icon":"layers","title":"Page builder","description":"Compose pages with reusable blocks and clean layouts."},{"icon":"shield","title":"Self-hosted","description":"Own your data and tune the experience to your brand."}]}', 1
FROM pages p
WHERE p.path = '/sample' AND NOT EXISTS (SELECT 1 FROM page_blocks b WHERE b.page_id = p.id)
UNION ALL
SELECT p.id, 'episodes_showcase', '{"section_title":"Latest episodes","section_description":"Fresh conversations and deep dives from the Castogo demo feed.","max_episodes":6,"display_mode":"grid"}', 2
FROM pages p
WHERE p.path = '/sample' AND NOT EXISTS (SELECT 1 FROM page_blocks b WHERE b.page_id = p.id)
UNION ALL
SELECT p.id, 'testimonials', '{"section_title":"Built for indie creators","section_description":"Simple tools that keep you focused on your show.","items":[{"quote":"We moved off a hosted platform and never looked back. Everything feels faster and clearer.","author":"Maya Liu","role":"Producer, The Signal","avatar_url":""},{"quote":"The blocks let us ship a landing page in an afternoon. It looks great on mobile.","author":"Jordan Patel","role":"Host, Night Shift","avatar_url":""},{"quote":"RSS just works. That alone saved us hours every release.","author":"Cam Torres","role":"Co-host, Field Notes","avatar_url":""}]}', 3
FROM pages p
WHERE p.path = '/sample' AND NOT EXISTS (SELECT 1 FROM page_blocks b WHERE b.page_id = p.id)
UNION ALL
SELECT p.id, 'prose', '{"body":"## About this sample page\n\nThis page is seeded from the database so you can preview public theming and blocks.\n\n- Hero, features, episodes, testimonials, and a CTA\n- Uses the **theme-public** token set\n- Fully editable from the admin Pages UI\n\n> Tip: create your own page in `/admin/pages` and publish it."}', 4
FROM pages p
WHERE p.path = '/sample' AND NOT EXISTS (SELECT 1 FROM page_blocks b WHERE b.page_id = p.id)
UNION ALL
SELECT p.id, 'cta', '{"headline":"Ready to build your public site?","description":"Jump into the admin and customize this page, or create a new one in minutes.","button_text":"Open admin","button_url":"/login"}', 5
FROM pages p
WHERE p.path = '/sample' AND NOT EXISTS (SELECT 1 FROM page_blocks b WHERE b.page_id = p.id);
