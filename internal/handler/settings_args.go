package handler

// settingsInput is the typed Datastar signal payload for POST /admin/settings.
// JSON tags match the editors/settings.Signals struct so readSignals decodes
// cleanly; the field names line up with the "<field>_error" signal keys
// produced by fieldValidationErrors.
type settingsInput struct {
	ID            int64  `json:"id" validate:"required"`
	Title         string `json:"title" validate:"required"`
	Description   string `json:"description"`
	SiteURL       string `json:"site_url" validate:"omitempty,url"`
	Language      string `json:"language"`
	Copyright     string `json:"copyright"`
	AuthorName    string `json:"author_name"`
	AuthorEmail   string `json:"author_email" validate:"omitempty,email"`
	CoverImageURL string `json:"cover_image_url" validate:"omitempty,url"`
	Category      string `json:"category"`
	Subcategory   string `json:"subcategory"`
	OwnerName     string `json:"owner_name"`
	OwnerEmail    string `json:"owner_email" validate:"omitempty,email"`
}
