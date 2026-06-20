// Package repository interface
package repository

import (
	"context"

	"github.com/gracchi-stdio/castogo/internal/domain"
)

type EpisodeFilter struct {
	Search string
	Limit  int
	Offset int
}

type EpisodeRepository interface {
	Create(ctx context.Context, ep *domain.Episode) (*domain.Episode, error)
	GetByID(ctx context.Context, id int64) (*domain.Episode, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Episode, error)
	List(ctx context.Context, filter EpisodeFilter) ([]*domain.Episode, error)
	Update(ctx context.Context, ep *domain.UpdateEpisode) (*domain.Episode, error)
	UpdateLinkedPageID(ctx context.Context, episodeID int64, pageID *int64) error
	GetByLinkedPageID(ctx context.Context, pageID int64) (*domain.Episode, error)
	Delete(ctx context.Context, id int64) error
	CountByStatus(ctx context.Context, status domain.EpisodeStatus) (int, error)
	ListPublished(ctx context.Context, limit, offset int) ([]*domain.Episode, error)
	GetMaxEpisodeNumber(ctx context.Context) (int, error)
	SearchPublished(ctx context.Context, query string, limit, offset int) ([]*domain.Episode, error)
}
