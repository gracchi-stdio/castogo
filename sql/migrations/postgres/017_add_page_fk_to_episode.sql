ALTER TABLE episodes
ADD COLUMN linked_page_id BIGINT REFERENCES pages(id) ON DELETE
SET NULL;
CREATE INDEX idx_episodes_linked_page_id ON episodes(linked_page_id)
WHERE linked_page_id IS NOT NULL;
