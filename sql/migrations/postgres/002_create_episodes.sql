CREATE TYPE episode_status AS ENUM ('draft', 'scheduled', 'published', 'archived');

CREATE TABLE episodes (
    id BIGSERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    slug TEXT UNIQUE NOT NULL,
    episode_number INTEGER NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    duration INTEGER NOT NULL DEFAULT 0,
    explicit BOOLEAN NOT NULL DEFAULT FALSE,
    cover_image_url TEXT,
    audio_source_url TEXT,
    published_at TIMESTAMPTZ,
    status episode_status NOT NULL DEFAULT 'draft',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_episodes_status ON episodes (status);
CREATE INDEX idx_episodes_published_at ON episodes (published_at) WHERE status = 'scheduled';
