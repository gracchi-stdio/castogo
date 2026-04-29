package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gracchi-stdio/castogo/internal/db"
	"github.com/gracchi-stdio/castogo/internal/domain"
	"github.com/gracchi-stdio/castogo/internal/repository"
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
	metadataJSON, _ := json.Marshal(ep.AudioMetadata)

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
		AudioMetadata:  metadataJSON,
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

func (r *EpisodeRepo) GetByID(ctx context.Context, id int64) (*domain.Episode, error) {
	result, err := r.q.GetEpisodeByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get episode by ID: %w", err)
	}
	return toDomainEpisode(&result), nil
}

func (r *EpisodeRepo) GetBySlug(ctx context.Context, slug string) (*domain.Episode, error) {
	result, err := r.q.GetEpisodeBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("get episode by slug: %w", err)
	}
	return toDomainEpisode(&result), nil
}

func (r *EpisodeRepo) List(ctx context.Context, filter repository.EpisodeFilter) ([]*domain.Episode, error) {
	params := db.ListEpisodesParams{
		Status:     filter.Status,
		Search:     filter.Search,
		PageOffset: int32(filter.Offset),
		PageLimit:  int32(filter.Limit),
	}
	results, err := r.q.ListEpisodes(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list episodes: %w", err)
	}

	episodes := make([]*domain.Episode, len(results))
	for i, result := range results {
		episodes[i] = toDomainEpisode(&result)
	}
	return episodes, nil
}

func (r *EpisodeRepo) Update(ctx context.Context, ep *domain.Episode) (*domain.Episode, error) {
	params := db.UpdateEpisodeParams{
		ID:             ep.ID,
		Title:          stringPtr(ep.Title),
		Slug:           stringPtr(ep.Slug),
		EpisodeNumber:  int32Ptr(ep.EpisodeNumber),
		Description:    stringPtr(ep.Description),
		Duration:       int32Ptr(ep.Duration),
		Explicit:       boolPtr(ep.Explicit),
		CoverImageUrl:  stringPtr(ep.CoverImageURL),
		AudioSourceUrl: stringPtr(ep.AudioSourceURL),
		Status:         episodeStatusPtr(db.EpisodeStatus(ep.Status)),
	}

	if ep.PublishAt != nil {
		params.PublishedAt = pgtype.Timestamptz{Time: *ep.PublishAt, Valid: true}
	}

	result, err := r.q.UpdateEpisode(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("update episode: %w", err)
	}

	return toDomainEpisode(&result), nil
}

func (r *EpisodeRepo) Delete(ctx context.Context, id int64) error {
	return r.q.DeleteEpisode(ctx, id)
}

func (r *EpisodeRepo) CountByStatus(ctx context.Context, status domain.EpisodeStatus) (int, error) {
	count, err := r.q.CountEpisodesByStatus(ctx, db.EpisodeStatus(status))
	if err != nil {
		return 0, fmt.Errorf("count episodes by status: %w", err)
	}
	return int(count), nil
}

func (r *EpisodeRepo) ListPublished(ctx context.Context, limit, offset int) ([]*domain.Episode, error) {
	params := db.ListPublishedEpisodesParams{
		Limit:  int32(limit),
		Offset: int32(offset)}
	results, err := r.q.ListPublishedEpisodes(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list published episodes: %w", err)
	}

	episodes := make([]*domain.Episode, len(results))
	for i, result := range results {
		episodes[i] = toDomainEpisode(&result)
	}
	return episodes, nil
}

func (r *EpisodeRepo) GetMaxEpisodeNumber(ctx context.Context) (int, error) {
	max, err := r.q.GetMaxEpisodeNumber(ctx)
	if err != nil {
		return 0, fmt.Errorf("get max episode number: %w", err)
	}
	switch v := max.(type) {
	case int64:
		return int(v), nil
	case int32:
		return int(v), nil
	default:
		return 0, nil
	}
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
		AudioMetadata:  e.AudioMetadata,
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

func boolPtr(b bool) *bool {
	return &b
}

func episodeStatusPtr(s db.EpisodeStatus) *db.EpisodeStatus {
	return &s
}
