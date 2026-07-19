package handler

// episodeUpdateInput is the typed Datastar signal payload for POST /admin/episodes/:id.
// JSON tags match editors/episode.Signals so readSignals decodes cleanly.
type episodeUpdateInput struct {
	Title         string `json:"title" validate:"required"`
	Slug          string `json:"slug"`
	Description   string `json:"description"`
	EpisodeNumber int    `json:"episode_number"`
	Explicit      struct {
		Checked bool `json:"checked"`
	} `json:"explicit"`
	PublishAt string `json:"publish_at"` // "2006-01-02" or ""
}

// episodeLinkPageInput carries the chosen existing-page id when the user links
// an episode to an already-existing page.
type episodeLinkPageInput struct {
	ExistingPageID int64 `json:"existing_page_id"`
}

// episodeCreateInput is the validated text input for the multipart create form.
// Audio file + metadata are read separately (c.FormFile / server-side processing),
// so this only covers the two text fields. Single-word names map to the
// "<field>_error" signal keys via the lowercase fallback in formFieldName.
type episodeCreateInput struct {
	Title       string `validate:"required"`
	Description string
}

// episodeCreateCompanionInput is the dialog payload for creating a companion
// page. Title is required; Slug is optional (auto-generated when blank). The
// json tags carry the companion_* signal names so validation errors map to
// companion_title_error / companion_slug_error.
type episodeCreateCompanionInput struct {
	Title string `json:"companion_title" validate:"required"`
	Slug  string `json:"companion_slug"`
}
