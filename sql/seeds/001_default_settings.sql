INSERT INTO podcast_config (title, description, site_url, author_name, owner_name, owner_email)
VALUES ('CASToGo', 'A simple podcasting app', 'http://localhost:8080', 'John Doe', 'Jane Smith', 'jane@example.com')
RETURNING *;