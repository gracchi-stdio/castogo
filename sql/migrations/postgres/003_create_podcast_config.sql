 CREATE TABLE podcast_config (
      id              BIGSERIAL PRIMARY KEY,
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