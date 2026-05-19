CREATE TABLE landing_page_sections (
    id          BIGSERIAL PRIMARY KEY,
    section_key TEXT NOT NULL UNIQUE,
    content     TEXT NOT NULL DEFAULT '{}',
    is_visible  BOOLEAN NOT NULL DEFAULT true,
    sort_order  INTEGER NOT NULL DEFAULT 0,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
