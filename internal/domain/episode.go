package domain

import "time"

// EpisodeStatus represents the derived state of an episode.
// It is NOT stored in the database — it is computed from publish_at and archived_at.
type EpisodeStatus string

const (
	EpisodeStatusDraft     EpisodeStatus = "draft"
	EpisodeStatusScheduled EpisodeStatus = "scheduled"
	EpisodeStatusPublished EpisodeStatus = "published"
	EpisodeStatusArchived  EpisodeStatus = "archived"
)

type AudioMetadata struct {
	Duration     int    `json:"duration"`    // in seconds
	SampleRate   int    `json:"sample_rate"` // in Hz (e.g., 44100)
	BitRate      int    `json:"bit_rate"`    // in kbps (e.g., 128)
	ChannelCount int    `json:"channels"`    // number of audio channels (e.g., 2 for stereo)
	FileSize     int64  `json:"file_size"`   // in bytes
	Format       string `json:"format"`      // e.g., "mp3", "wav"
	MimeType     string `json:"mimetype"`    // e.g., "audio/mpeg", "audio/wav"
}

type Episode struct {
	ID             int64         `json:"id"`
	Title          string        `json:"title"`
	Slug           string        `json:"slug"`
	Description    string        `json:"description"`
	Duration       int           `json:"duration"` // in seconds
	AudioMetadata  AudioMetadata `json:"audio_metadata"`
	Explicit       bool          `json:"explicit"`
	CoverImageURL  string        `json:"cover_image_url"`
	AudioSourceURL string        `json:"audio_source_url"`
	EpisodeNumber  int           `json:"episode_number"`
	LinkedPageID   *int64        `json:"linked_page_id"`
	PublishAt      *time.Time    `json:"publish_at,omitempty"`
	ArchivedAt     *time.Time    `json:"archived_at,omitempty"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
}

// Status derives the episode status from publish_at and archived_at.
// No database column needed — the state machine is:
//
//	publish_at IS NULL       → draft
//	publish_at > NOW()       → scheduled
//	publish_at <= NOW()      → published
//	archived_at IS NOT NULL  → archived (overrides all above)
func (ep *Episode) Status() EpisodeStatus {
	if ep.ArchivedAt != nil {
		return EpisodeStatusArchived
	}
	if ep.PublishAt == nil {
		return EpisodeStatusDraft
	}
	if ep.PublishAt.After(time.Now()) {
		return EpisodeStatusScheduled
	}
	return EpisodeStatusPublished
}

// UpdateEpisode represents a partial update to an episode.
// nil fields are ignored (COALESCE keeps existing DB value).
// Only ID is required.
type UpdateEpisode struct {
	ID             int64          `json:"id"`
	Title          *string        `json:"title,omitempty"`
	Slug           *string        `json:"slug,omitempty"`
	EpisodeNumber  *int           `json:"episode_number,omitempty"`
	Description    *string        `json:"description,omitempty"`
	Duration       *int           `json:"duration,omitempty"`
	Explicit       *bool          `json:"explicit,omitempty"`
	CoverImageURL  *string        `json:"cover_image_url,omitempty"`
	AudioSourceURL *string        `json:"audio_source_url,omitempty"`
	AudioMetadata  *AudioMetadata `json:"audio_metadata,omitempty"`
	PublishAt      *time.Time     `json:"publish_at,omitempty"`
	ArchivedAt     *time.Time     `json:"archived_at,omitempty"`
}
