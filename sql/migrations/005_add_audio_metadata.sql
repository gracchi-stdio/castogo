ALTER TABLE episodes ADD COLUMN audio_metadata JSONB NOT NULL DEFAULT '{}'; -- no nulls, default to empty object
