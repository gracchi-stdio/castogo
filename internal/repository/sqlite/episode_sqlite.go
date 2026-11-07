package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/gracchi-stdio/castogo/internal/domain"
	"github.com/gracchi-stdio/castogo/internal/repository"
	"github.com/gracchi-stdio/castogo/internal/sqlcsqlite"
)

// Ensure that EpisodeRepo implements the EpisodeRepository interface
var _ repository.EpisodeRepository = (*EpisodeRepo)(nil)

type EpisodeRepo struct {
	q *sqlcsqlite.Queries
}

func NewEpisodeRepo(db *sql.DB) *EpisodeRepo {
	return &EpisodeRepo{q: sqlcsqlite.New(db)}
}
func episodeFromRow(e sqlcsqlite.Episode) (*domain.Episode, error) {
	createdAt, err := timeFromSQLite(e.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("parse episode created_at: %w", err)
	}

	publishedAt, err := timePtrFromSQLite(e.PublishedAt)
	if err != nil {
		return nil, fmt.Errorf("parse episode published_at: %w", err)
	}
	archivedAt, err := timePtrFromSQLite(e.ArchivedAt)
	if err != nil {
		return nil, fmt.Errorf("parse episode archived_at: %w", err)
	}

	return &domain.Episode{
		ID:             e.ID,
		Title:          e.Title,
		Slug:           e.Slug,
		Description:    e.Description,
		Duration:       int(e.Duration),
		Explicit:       e.Explicit,
		CoverImageURL:  orZero(e.CoverImageUrl),
		AudioSourceURL: orZero(e.AudioSourceUrl),
		SlidesMD:       e.SlidesMd,
		AudioMetadata:  e.AudioMetadata,
		LinkedPageID:   e.LinkedPageID,
		PublishAt:      publishedAt,
		ArchivedAt:     archivedAt,
		CreatedAt:      createdAt,
		UpdatedAt:      sqliteTime(e.UpdatedAt),
	}, nil
}

func episodesFromList(list []sqlcsqlite.Episode) ([]*domain.Episode, error) {
	out := make([]*domain.Episode, len(list))
	for i, result := range list {
		ep, err := episodeFromRow(result)
		if err != nil {
			return nil, err
		}
		out[i] = ep
	}
	return out, nil
}

// CountByStatus implements [repository.EpisodeRepository].
func (e *EpisodeRepo) CountByStatus(ctx context.Context, status domain.EpisodeStatus) (int, error) {
	result, err := e.q.CountEpisodesByStatus(ctx, string(status))
	if err != nil {
		return 0, fmt.Errorf("count episodes by status: %w", err)
	}
	return int(result), nil
}

// Create implements [repository.EpisodeRepository].
func (e *EpisodeRepo) Create(ctx context.Context, ep *domain.Episode) (*domain.Episode, error) {
	result, err := e.q.CreateEpisode(ctx, sqlcsqlite.CreateEpisodeParams{
		Title:          &ep.Title,
		Slug:           &ep.Slug,
		Description:    &ep.Description,
		Duration:       toPtr(int64(ep.Duration)),
		Explicit:       &ep.Explicit,
		CoverImageUrl:  &ep.CoverImageURL,
		AudioSourceUrl: &ep.AudioSourceURL,
		AudioMetadata:  ep.AudioMetadata,
		PublishedAt:    timePtrToSQLite(ep.PublishAt),
	})
	if err != nil {
		return nil, fmt.Errorf("create episode: %w", err)
	}
	return episodeFromRow(result)
}

// Delete implements [repository.EpisodeRepository].
func (e *EpisodeRepo) Delete(ctx context.Context, id int64) error {
	return e.q.DeleteEpisode(ctx, id)
}

// GetByID implements [repository.EpisodeRepository].
func (e *EpisodeRepo) GetByID(ctx context.Context, id int64) (*domain.Episode, error) {
	result, err := e.q.GetEpisodeByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get episode by ID: %w", err)
	}
	return episodeFromRow(result)
}

// GetByLinkedPageID implements [repository.EpisodeRepository].
func (e *EpisodeRepo) GetByLinkedPageID(ctx context.Context, pageID int64) (*domain.Episode, error) {
	result, err := e.q.GetEpisodeByLinkedPageID(ctx, &pageID)
	if err != nil {
		return nil, fmt.Errorf("get episode by linked page ID: %w", err)
	}
	return episodeFromRow(result)
}

// GetBySlug implements [repository.EpisodeRepository].
func (e *EpisodeRepo) GetBySlug(ctx context.Context, slug string) (*domain.Episode, error) {
	result, err := e.q.GetEpisodeBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("get episode by slug: %w", err)
	}
	return episodeFromRow(result)
}

// List implements [repository.EpisodeRepository].
func (e *EpisodeRepo) List(ctx context.Context, filter repository.EpisodeFilter) ([]*domain.Episode, error) {
	results, err := e.q.ListEpisodes(ctx, sqlcsqlite.ListEpisodesParams{
		Search:     filter.Search,
		Status:     filter.Status,
		PageOffset: int64(filter.Offset),
		PageLimit:  int64(filter.Limit),
	})
	if err != nil {
		return nil, fmt.Errorf("list episodes: %w", err)
	}
	return episodesFromList(results)
}

// ListForRanking implements [repository.EpisodeRepository].
func (e *EpisodeRepo) ListForRanking(ctx context.Context) ([]*domain.Episode, error) {
	results, err := e.q.ListEpisodesForRanking(ctx)
	if err != nil {
		return nil, fmt.Errorf("list episodes for ranking: %w", err)
	}

	return episodesFromList(results)
}

// ListPublished implements [repository.EpisodeRepository].
func (e *EpisodeRepo) ListPublished(ctx context.Context, limit int, offset int) ([]*domain.Episode, error) {
	results, err := e.q.ListPublishedEpisodes(ctx, sqlcsqlite.ListPublishedEpisodesParams{
		Limit:  int64(limit),
		Offset: int64(offset),
	})
	if err != nil {
		return nil, fmt.Errorf("list published episodes: %w", err)
	}
	return episodesFromList(results)
}

// ListPublishedByLinkedPageID implements [repository.EpisodeRepository].
func (e *EpisodeRepo) ListPublishedByLinkedPageID(ctx context.Context, pageID int64) ([]*domain.Episode, error) {
	results, err := e.q.ListPublishedByLinkedPageID(ctx, &pageID)
	if err != nil {
		return nil, fmt.Errorf("list published episodes by linked page ID: %w", err)
	}
	return episodesFromList(results)
}

// ListPublishedWithPagePath implements [repository.EpisodeRepository].
func (e *EpisodeRepo) ListPublishedWithPagePath(ctx context.Context, limit int, offset int) ([]*domain.EpisodeWithPagePath, error) {
	results, err := e.q.ListPublishedEpisodesWithPagePath(ctx, sqlcsqlite.ListPublishedEpisodesWithPagePathParams{
		Limit:  int64(limit),
		Offset: int64(offset),
	})
	if err != nil {
		return nil, fmt.Errorf("list published episodes: %w", err)
	}

	out := make([]*domain.EpisodeWithPagePath, len(results))
	for i, row := range results {
		ep, err := episodeFromRow(row.Episode)
		if err != nil {
			return nil, err
		}
		out[i] = &domain.EpisodeWithPagePath{
			Episode:  ep,
			PagePath: row.PagePath,
		}
	}
	return out, nil
}

// SearchPublishedWithPagePath implements [repository.EpisodeRepository].
func (e *EpisodeRepo) SearchPublishedWithPagePath(ctx context.Context, query string, limit int, offset int) ([]*domain.EpisodeWithPagePath, error) {
	results, err := e.q.SearchPublishedEpisodesWithPagePath(ctx, sqlcsqlite.SearchPublishedEpisodesWithPagePathParams{
		Search:     &query,
		PageLimit:  int64(limit),
		PageOffset: int64(offset),
	})
	if err != nil {
		return nil, fmt.Errorf("search published episodes: %w", err)
	}
	out := make([]*domain.EpisodeWithPagePath, len(results))
	for i, row := range results {
		ep, err := episodeFromRow(row.Episode)
		if err != nil {
			return nil, err
		}
		out[i] = &domain.EpisodeWithPagePath{
			Episode:  ep,
			PagePath: row.PagePath,
		}
	}
	return out, nil
}

// Update implements [repository.EpisodeRepository].
func (e *EpisodeRepo) Update(ctx context.Context, ep *domain.UpdateEpisode) (*domain.Episode, error) {
	var duration *int64
	if ep.Duration != nil {
		duration = toPtr(int64(*ep.Duration))
	}

	// nil → zero value, mirroring the postgres wrapper: the narg on an
	// overridden column is non-pointer, so an Update without metadata
	// overwrites it.
	result, err := e.q.UpdateEpisode(ctx, sqlcsqlite.UpdateEpisodeParams{
		ID:             ep.ID,
		Title:          ep.Title,
		Slug:           ep.Slug,
		Description:    ep.Description,
		Duration:       duration,
		Explicit:       ep.Explicit,
		CoverImageUrl:  ep.CoverImageURL,
		AudioSourceUrl: ep.AudioSourceURL,
		SlidesMd:       ep.SlidesMD,
		AudioMetadata:  orZero(ep.AudioMetadata),
		PublishedAt:    timePtrToSQLite(ep.PublishAt),
		ArchivedAt:     timePtrToSQLite(ep.ArchivedAt),
	})
	if err != nil {
		return nil, fmt.Errorf("update episode: %w", err)
	}
	return episodeFromRow(result)
}

// UpdateLinkedPageID implements [repository.EpisodeRepository].
func (e *EpisodeRepo) UpdateLinkedPageID(ctx context.Context, episodeID int64, pageID *int64) error {
	return e.q.UpdateEpisodeLinkedPageID(ctx, sqlcsqlite.UpdateEpisodeLinkedPageIDParams{
		ID:           episodeID,
		LinkedPageID: pageID,
	})
}
