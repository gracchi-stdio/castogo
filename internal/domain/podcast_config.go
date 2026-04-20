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
}
