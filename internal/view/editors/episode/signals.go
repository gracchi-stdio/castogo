package episodeForm

import "github.com/gracchi-stdio/castogo/internal/domain"

// Signals is the typed Datastar signal state for the episode edit form.
// Error fields use snake_case + _error suffix to align with the form field names
// and the fieldValidationErrors helper in internal/handler/helpers.go.
type Signals struct {
	Title         string `json:"title"`
	Slug          string `json:"slug"`
	Description   string `json:"description"`
	EpisodeNumber int    `json:"episode_number"`
	Explicit      struct {
		Checked bool `json:"checked"`
	} `json:"explicit"`
	PublishAt string `json:"publish_at"` // "2006-01-02" or ""

	// Page linkage — only the chosen existing-page id needs to round-trip; the
	// linked/not-linked state itself is derived from the episode, so there is no
	// separate "mode" signal to keep in sync across SSE patches.
	ExistingPageID int64 `json:"existing_page_id"`

	TitleError       string `json:"title_error,omitempty"`
	SlugError        string `json:"slug_error,omitempty"`
	EpisodeNumberErr string `json:"episode_number_error,omitempty"`
}

// NewSignals builds the initial signal state from an existing episode.
func NewSignals(episode *domain.Episode) Signals {
	if episode == nil {
		return Signals{}
	}
	s := Signals{
		Title:         episode.Title,
		Slug:          episode.Slug,
		Description:   episode.Description,
		EpisodeNumber: episode.EpisodeNumber,
		Explicit: struct {
			Checked bool `json:"checked"`
		}{Checked: episode.Explicit},
	}
	if episode.PublishAt != nil {
		s.PublishAt = episode.PublishAt.Format("2006-01-02")
	}
	if episode.LinkedPageID != nil {
		s.ExistingPageID = *episode.LinkedPageID
	}
	return s
}

// CreateSignals is the signal state for the episode create form. The audio_*
// fields are display-only — they are populated client-side by
// window.extractAudioMetadata for a preview; the server recomputes the real
// metadata from the processed audio, so no numeric meta_* signals are needed.
type CreateSignals struct {
	Title             string `json:"title"`
	Description       string `json:"description"`
	Uploading         bool   `json:"uploading"`
	UploadingStatus   string `json:"uploading_status"`
	TitleError        string `json:"title_error,omitempty"`
	DescriptionError  string `json:"description_error,omitempty"`
	AudioFileError    string `json:"audio_file_error,omitempty"`
	MetadataExtracted bool   `json:"metadata_extracted"`
	AudioDuration     string `json:"audio_duration"`
	AudioSampleRate   string `json:"audio_sample_rate"`
	AudioChannelCount string `json:"audio_channel_count"`
	AudioBitrate      string `json:"audio_bitrate"`
	AudioFormat       string `json:"audio_format"`
	AudioMimeType     string `json:"audio_mime_type"`
	AudioFileSize     string `json:"audio_file_size"`
}
