-- Create page table
CREATE TABLE pages (
    id           BIGSERIAL PRIMARY KEY,
    slug         TEXT NOT NULL,
    title        TEXT NOT NULL,
    parent_id    BIGINT REFERENCES pages(id) ON DELETE SET NULL,
    layout       TEXT NOT NULL DEFAULT 'default',
    is_published BOOLEAN NOT NULL DEFAULT false,
    metadata     JSONB NOT NULL DEFAULT '{}',
    path         TEXT NOT NULL UNIQUE,
    sort_order   INTEGER NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Prevents two siblings from having the same slug
CREATE UNIQUE INDEX idx_pages_slug_parent ON pages (slug, COALESCE(parent_id, 0));

-- Create page_blocks table
CREATE TABLE page_blocks (
    id          BIGSERIAL PRIMARY KEY,
    page_id     BIGINT NOT NULL REFERENCES pages(id) ON DELETE CASCADE,
    block_type  TEXT NOT NULL,
    content     JSONB NOT NULL DEFAULT '{}',
    sort_order  INTEGER NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Prevents duplicate block types on the same page (e.g., two "hero" blocks)
CREATE UNIQUE INDEX idx_page_blocks_page_type ON page_blocks (page_id, block_type);
