package postgres

import (
	"github.com/gracchi-stdio/castogo/internal/db"
	"github.com/gracchi-stdio/castogo/internal/domain"
)

// --- Helper functions for type mapping between db and domain models ---

// --- Type mapping helper ---
func toDomainPodcastConfig(c *db.PodcastConfig) *domain.PodcastConfig {
	return &domain.PodcastConfig{
		ID:            c.ID,
		Title:         c.Title,
		Description:   c.Description,
		SiteURL:       c.SiteUrl,
		Copyright:     c.Copyright,
		AuthorName:    c.AuthorName,
		AuthorEmail:   c.AuthorEmail,
		CoverImageURL: c.CoverImageUrl,
		Language:      c.Language,
		OwnerName:     c.OwnerName,
		OwnerEmail:    c.OwnerEmail,
		Category:      c.Category,
		Subcategory:   c.Subcategory,
	}
}

func toDomainEpisode(e *db.Episode) *domain.Episode {
	ep := &domain.Episode{
		ID:             e.ID,
		Title:          e.Title,
		Slug:           e.Slug,
		Description:    e.Description,
		EpisodeNumber:  int(e.EpisodeNumber),
		Duration:       int(e.Duration),
		Explicit:       e.Explicit,
		CoverImageURL:  stringValue(e.CoverImageUrl),
		AudioSourceURL: stringValue(e.AudioSourceUrl),
		AudioMetadata:  e.AudioMetadata,
		LinkedPageID:   e.LinkedPageID,
		CreatedAt:      e.CreatedAt.Time,
		UpdatedAt:      e.UpdatedAt.Time,
	}
	if e.PublishedAt.Valid {
		ep.PublishAt = &e.PublishedAt.Time
	}
	if e.ArchivedAt.Valid {
		ep.ArchivedAt = &e.ArchivedAt.Time
	}
	return ep
}

func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
