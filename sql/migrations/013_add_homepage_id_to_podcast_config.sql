ALTER TABLE podcast_config
    ADD COLUMN homepage_id BIGINT REFERENCES pages(id) ON DELETE SET NULL;
