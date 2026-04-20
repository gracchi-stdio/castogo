package domain

import "time"

type EpisodeStatus string

const (
	EpisodeStatusDraft     EpisodeStatus = "draft"
	EpisodeStatusScheduled EpisodeStatus = "scheduled"
	EpisodeStatusPublished EpisodeStatus = "published"
	EpisodeStatusArchived  EpisodeStatus = "archived"
)

type Episode struct {
	ID             int64         `json:"id"`
	Title          string        `json:"title"`
	Slug           string        `json:"slug"`
	Description    string        `json:"description"`
	Duration       int           `json:"duration"` // in seconds
	Explicit       bool          `json:"explicit"`
	CoverImageURL  string        `json:"cover_image_url"`
	AudioSourceURL string        `json:"audio_source_url"`
	EpisodeNumber  int           `json:"episode_number"`
	PublishAt      *time.Time    `json:"publish_at,omitempty"`
	Status         EpisodeStatus `json:"status"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
}
