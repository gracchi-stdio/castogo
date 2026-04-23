CREATE TABLE IF NOT EXISTS http_sessions (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    key        TEXT NOT NULL UNIQUE,
    data       TEXT NOT NULL,
    created_on TIMESTAMPTZ DEFAULT NOW(),
    expires_on TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS http_sessions_expiry_idx ON http_sessions (expires_on);
CREATE INDEX IF NOT EXISTS http_sessions_key_idx ON http_sessions (key);
