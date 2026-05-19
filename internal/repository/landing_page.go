package repository

import (
	"context"

	"github.com/gracchi-stdio/castogo/internal/domain"
)

type LandingPageRepository interface {
	GetAll(ctx context.Context) ([]*domain.LandingPageSection, error)
	GetVisible(ctx context.Context) ([]*domain.LandingPageSection, error)
	GetByKey(ctx context.Context, key string) (*domain.LandingPageSection, error)
	Update(ctx context.Context, section *domain.UpdateLandingSection) (*domain.LandingPageSection, error)
}
