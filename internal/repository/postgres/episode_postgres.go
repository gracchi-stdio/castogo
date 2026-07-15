package postgres

import (
	"context"
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
	params := db.CreateEpisodeParams{
		Title:          ep.Title,
		Slug:           ep.Slug,
		EpisodeNumber:  int32(ep.EpisodeNumber),
		Description:    ep.Description,
		Duration:       int32(ep.Duration),
		Explicit:       ep.Explicit,
		CoverImageUrl:  &ep.CoverImageURL,
		AudioSourceUrl: &ep.AudioSourceURL,
		AudioMetadata:  ep.AudioMetadata,
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

func (r *EpisodeRepo) Update(ctx context.Context, ep *domain.UpdateEpisode) (*domain.Episode, error) {
	var episodeNumber *int32
	if ep.EpisodeNumber != nil {
		v := int32(*ep.EpisodeNumber)
		episodeNumber = &v
	}

	var duration *int32
	if ep.Duration != nil {
		v := int32(*ep.Duration)
		duration = &v
	}

	var audioMetadata domain.AudioMetadata
	if ep.AudioMetadata != nil {
		audioMetadata = *ep.AudioMetadata
	}

	var publishedAt pgtype.Timestamptz
	if ep.PublishAt != nil {
		publishedAt = pgtype.Timestamptz{Time: *ep.PublishAt, Valid: true}
	}

	var archivedAt pgtype.Timestamptz
	if ep.ArchivedAt != nil {
		archivedAt = pgtype.Timestamptz{Time: *ep.ArchivedAt, Valid: true}
	}

	params := db.UpdateEpisodeParams{
		ID:             ep.ID,
		Title:          ep.Title,
		Slug:           ep.Slug,
		EpisodeNumber:  episodeNumber,
		Description:    ep.Description,
		Duration:       duration,
		Explicit:       ep.Explicit,
		CoverImageUrl:  ep.CoverImageURL,
		AudioSourceUrl: ep.AudioSourceURL,
		AudioMetadata:  audioMetadata,
		PublishedAt:    publishedAt,
		ArchivedAt:     archivedAt,
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
	count, err := r.q.CountEpisodesByStatus(ctx, string(status))
	if err != nil {
		return 0, fmt.Errorf("count episodes by status: %w", err)
	}
	return int(count), nil
}

func (r *EpisodeRepo) ListPublished(ctx context.Context, limit, offset int) ([]*domain.Episode, error) {
	params := db.ListPublishedEpisodesParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	}
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

func (r *EpisodeRepo) ListPublishedWithPagePath(ctx context.Context, limit, offset int) ([]*domain.EpisodeWithPagePath, error) {
	params := db.ListPublishedEpisodesWithPagePathParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	}
	results, err := r.q.ListPublishedEpisodesWithPagePath(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list published episodes: %w", err)
	}

	out := make([]*domain.EpisodeWithPagePath, len(results))
	for i, row := range results {
		// episodes[i] = toDomainEpisode(&result)
		out[i] = &domain.EpisodeWithPagePath{
			Episode:  toDomainEpisode(&row.Episode),
			PagePath: row.PagePath,
		}
	}
	return out, nil
}

func (r *EpisodeRepo) GetMaxEpisodeNumber(ctx context.Context) (int, error) {
	max, err := r.q.GetMaxEpisodeNumber(ctx)
	if err != nil {
		return 0, fmt.Errorf("get max episode number: %w", err)
	}
	return int(max), nil
}

func (r *EpisodeRepo) SearchPublished(ctx context.Context, query string, limit, offset int) ([]*domain.Episode, error) {
	results, err := r.q.SearchPublishedEpisodes(ctx, db.SearchPublishedEpisodesParams{
		Search:     &query,
		PageLimit:  int32(limit),
		PageOffset: int32(offset),
	})
	if err != nil {
		return nil, fmt.Errorf("search published episodes: %w", err)
	}
	episodes := make([]*domain.Episode, len(results))
	for i, result := range results {
		episodes[i] = toDomainEpisode(&result)
	}
	return episodes, nil
}

func (r *EpisodeRepo) UpdateLinkedPageID(ctx context.Context, episodeID int64, pageID *int64) error {
	return r.q.UpdateEpisodeLinkedPageID(ctx, db.UpdateEpisodeLinkedPageIDParams{
		ID:           episodeID,
		LinkedPageID: pageID,
	})
}

func (r *EpisodeRepo) GetByLinkedPageID(ctx context.Context, pageID int64) (*domain.Episode, error) {
	result, err := r.q.GetEpisodeByLinkedPageID(ctx, &pageID)
	if err != nil {
		return nil, fmt.Errorf("get episode by linked page ID: %w", err)
	}
	return toDomainEpisode(&result), nil
}
