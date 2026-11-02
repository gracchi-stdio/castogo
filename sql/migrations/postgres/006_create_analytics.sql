-- Which log files have we already processed?
CREATE TABLE analytics_processed_files (
    id BIGSERIAL PRIMARY KEY,
    filename TEXT NOT NULL UNIQUE,
    file_date DATE NOT NULL,
    entries_count INTEGER NOT NULL DEFAULT 0,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- IAB byte accumulator for analytics
-- Hash = SHA1(date+ip+user_agent+episode_id)
-- Once bytes_seen passes the 1-min threshold, counted = true
CREATE TABLE analytics_download_accumulator (
    hash CHAR(40) PRIMARY KEY, -- SHA1 hex string
    episode_id BIGINT NOT NULL,
    date DATE NOT NULL,
    bytes_seen BIGINT NOT NULL DEFAULT 0,
    counted BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Daily podcast totals
CREATE TABLE analytics_podcasts (
    podcast_id BIGINT NOT NULL,
    date DATE NOT NULL,
    hits INTEGER NOT NULL DEFAULT 1,
    bandwidth BIGINT NOT NULL DEFAULT 0,
    unique_listeners INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (podcast_id, date)
);

-- Per episode per day
CREATE TABLE analytics_podcasts_by_episode (
    podcast_id BIGINT NOT NULL,
    episode_id BIGINT NOT NULL,
    date DATE NOT NULL,
    age INT NOT NULL, -- days since publication
    hits INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (podcast_id, episode_id, date)
);

-- Per hour of day (listening patterns)
CREATE TABLE analytics_podcasts_by_hour (
    podcast_id BIGINT NOT NULL,
    date DATE NOT NULL,
    hour INT NOT NULL, -- 0-23
    hits INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (podcast_id, date, hour)
);

-- Per player/app
CREATE TABLE analytics_podcasts_by_player (
    podcast_id BIGINT NOT NULL,
    date DATE NOT NULL,
    service VARCHAR(128) NOT NULL DEFAULT '',
    app VARCHAR(128) NOT NULL DEFAULT '',
    device VARCHAR(32) NOT NULL DEFAULT '',
    os VARCHAR(32) NOT NULL DEFAULT '',
    is_bot BOOLEAN NOT NULL DEFAULT FALSE,
    hits INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (podcast_id, date, service, app, device, os, is_bot)
);

-- Per country per day
CREATE TABLE analytics_podcasts_by_country (
    podcast_id BIGINT NOT NULL,
    date DATE NOT NULL,
    country_code CHAR(2) NOT NULL,
    hits INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (podcast_id, date, country_code)
);
