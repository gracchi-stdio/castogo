package repository

import (
	"context"

	"github.com/gracchi-stdio/castogo/internal/domain"
)

type PodcastConfigRepository interface {
	Get(ctx context.Context) (*domain.PodcastConfig, error)
	Update(ctx context.Context, config *domain.UpdatePodcastConfig) (*domain.PodcastConfig, error)
}
