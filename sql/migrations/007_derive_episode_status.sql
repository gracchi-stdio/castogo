-- Migration: Derive episode status from publish_at instead of explicit status column
-- publish_at IS NULL           → draft
-- publish_at > NOW()           → scheduled
-- publish_at <= NOW()          → published
-- archived_at IS NOT NULL      → archived (overrides all above)

-- Step 1: Add archived_at column
ALTER TABLE episodes ADD COLUMN archived_at TIMESTAMPTZ;

-- Step 2: Migrate existing archived episodes — set archived_at from updated_at
UPDATE episodes SET archived_at = updated_at WHERE status = 'archived';

-- Step 3: Drop old indexes
DROP INDEX IF EXISTS idx_episodes_status;
DROP INDEX IF EXISTS idx_episodes_published_at;

-- Step 4: Drop status column and enum type
ALTER TABLE episodes DROP COLUMN status;
DROP TYPE episode_status;

-- Step 5: New indexes for derived status queries
-- Published episodes (publish_at in the past, not archived) — used by RSS feed and analytics
CREATE INDEX idx_episodes_published ON episodes (published_at DESC)
    WHERE published_at IS NOT NULL AND archived_at IS NULL;

-- Drafts (no publish_at, not archived) — used by admin episode list
CREATE INDEX idx_episodes_drafts ON episodes (created_at DESC)
    WHERE published_at IS NULL AND archived_at IS NULL;

-- Archived — used by admin episode list
CREATE INDEX idx_episodes_archived ON episodes (archived_at DESC)
    WHERE archived_at IS NOT NULL;
