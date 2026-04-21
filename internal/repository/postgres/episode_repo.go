package postgres

import (
	"context"
	"fmt"

	"github.com/gracchi-stdio/castogo/internal/db"
	"github.com/gracchi-stdio/castogo/internal/domain"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EpisodeRepo struct {
	q *db.Queries
}

func NewEpisodeRepo(pool *pgxpool.Pool) *EpisodeRepo {
	return &EpisodeRepo{
		q: db.New(pool),
	}
}

func (r *EpisodeRepo) Create(ctx context.Context, ep *domain.Episode) (*domain.Episode, error) {
	params := db.CreateEpisodeParams{
		Title:          ep.Title,
		Slug:           ep.Slug,
		EpisodeNumber:  int32(ep.EpisodeNumber),
		Description:    ep.Description,
		Duration:       int32(ep.Duration),
		Explicit:       ep.Explicit,
		CoverImageUrl:  (&ep.CoverImageURL),
		AudioSourceUrl: (&ep.AudioSourceURL),
		Status:         db.EpisodeStatus(ep.Status),
	}

	if ep.PublishAt != nil {
		params.PublishedAt = pgtype.Timestamptz{Time: *ep.PublishAt, Valid: true}
	}

	result, err := r.q.CreateEpisode(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("create episode: %w", err)
	}

	return toDomainEpisode(&result), nil
}

// --- Type mapping helper ---
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
		Status:         domain.EpisodeStatus(e.Status),
		CreatedAt:      e.CreatedAt.Time,
		UpdatedAt:      e.UpdatedAt.Time,
	}
	if e.PublishedAt.Valid {
		ep.PublishAt = &e.PublishedAt.Time
	}
	return ep
}

func stringPtr(s string) *string {
	return &s
}

func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func int32Ptr(i int) *int32 {
	v := int32(i)
	return &v
}
