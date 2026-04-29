package repository

import (
	"context"

	"github.com/gracchi-stdio/castogo/internal/domain"
)

type EpisodeFilter struct {
	Status domain.EpisodeStatus
	Search string
	Limit  int
	Offset int
}

type EpisodeRepository interface {
	Create(ctx context.Context, ep *domain.Episode) (*domain.Episode, error)
	GetByID(ctx context.Context, id int64) (*domain.Episode, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Episode, error)
	List(ctx context.Context, filter EpisodeFilter) ([]*domain.Episode, error)
	Update(ctx context.Context, ep *domain.Episode) (*domain.Episode, error)
	Delete(ctx context.Context, id int64) error
	CountByStatus(ctx context.Context, status domain.EpisodeStatus) (int, error)
	ListPublished(ctx context.Context, limit, offset int) ([]*domain.Episode, error)
	GetMaxEpisodeNumber(ctx context.Context) (int, error)
}
