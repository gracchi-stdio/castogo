package settings

import "github.com/gracchi-stdio/castogo/internal/domain"

// Signals is the typed Datastar signal state for the settings form.
// Error fields use snake_case + _error suffix to align with the form field names
// and the fieldValidationErrors helper in internal/handler/helpers.go.
type Signals struct {
	Title         string `json:"title"`
	Description   string `json:"description"`
	SiteURL       string `json:"site_url"`
	Language      string `json:"language"`
	Copyright     string `json:"copyright"`
	AuthorName    string `json:"author_name"`
	AuthorEmail   string `json:"author_email"`
	CoverImageURL string `json:"cover_image_url"`
	OwnerName     string `json:"owner_name"`
	OwnerEmail    string `json:"owner_email"`
	Category      string `json:"category"`
	Subcategory   string `json:"subcategory"`

	// error signals — reactive, cleared on input
	TitleError         string `json:"title_error"`
	SiteURLError       string `json:"site_url_error"`
	AuthorEmailError   string `json:"author_email_error"`
	OwnerEmailError    string `json:"owner_email_error"`
	CoverImageURLError string `json:"cover_image_url_error"`
	SubcategoryError   string `json:"subcategory_error"`

	// cover upload sub-action state — separate from $fetching (which tracks the main save)
	// because cover upload is a multipart POST with multi-step progress.
	CoverUploading bool   `json:"cover_uploading"`
	CoverStatus    string `json:"cover_status"`
}

// NewSignals builds the initial signal state from an existing config.
// Nil-safe: returns zero-value Signals when config is nil (first-run case).
func NewSignals(config *domain.PodcastConfig) Signals {
	if config == nil {
		return Signals{
			Language: "en",
		}
	}
	return Signals{
		Title:         config.Title,
		Description:   config.Description,
		SiteURL:       config.SiteURL,
		Language:      config.Language,
		Copyright:     config.Copyright,
		AuthorName:    config.AuthorName,
		AuthorEmail:   config.AuthorEmail,
		CoverImageURL: config.CoverImageURL,
		OwnerName:     config.OwnerName,
		OwnerEmail:    config.OwnerEmail,
		Category:      config.Category,
		Subcategory:   config.Subcategory,
	}
}
