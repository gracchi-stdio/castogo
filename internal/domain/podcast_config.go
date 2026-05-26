package domain

type PodcastConfig struct {
	ID            int64  `db:"id"`
	Title         string `db:"title"`
	Description   string `db:"description"`
	SiteURL       string `db:"site_url"`
	Language      string `db:"language"`
	Copyright     string `db:"copyright"`
	AuthorName    string `db:"author_name"`
	AuthorEmail   string `db:"author_email"`
	CoverImageURL string `db:"cover_image_url"`
	Category      string `db:"category"`
	Subcategory   string `db:"subcategory"`
	OwnerName     string `db:"owner_name"`
	OwnerEmail    string `db:"owner_email"`
	HomepageID    *int64 `db:"homepage_id"`
}

// UpdatePodcastConfig represents a partial update to podcast_config.
// nil fields are ignored (COALESCE keeps existing DB value).
// Only ID is required.
type UpdatePodcastConfig struct {
	ID            int64   `db:"id"`
	Title         *string `db:"title"`
	Description   *string `db:"description"`
	SiteURL       *string `db:"site_url"`
	Language      *string `db:"language"`
	Copyright     *string `db:"copyright"`
	AuthorName    *string `db:"author_name"`
	AuthorEmail   *string `db:"author_email"`
	CoverImageURL *string `db:"cover_image_url"`
	Category      *string `db:"category"`
	Subcategory   *string `db:"subcategory"`
	OwnerName     *string `db:"owner_name"`
	OwnerEmail    *string `db:"owner_email"`
	HomepageID    *int64  `db:"homepage_id"`
}
