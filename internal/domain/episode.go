package domain

import "time"

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
	PublishAt      *time.Time    `json:"publish_at,omitempty"`
	Status         EpisodeStatus `json:"status"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
}
