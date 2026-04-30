package service

import (
	"context"

	"github.com/gracchi-stdio/castogo/internal/domain"
	"github.com/gracchi-stdio/castogo/internal/repository"
)

type SettingsService struct {
	configRepo repository.PodcastConfigRepository
}

func NewSettingsService(configRepo repository.PodcastConfigRepository) *SettingsService {
	return &SettingsService{configRepo: configRepo}
}

func (s *SettingsService) GetPodcastConfig(ctx context.Context) (*domain.PodcastConfig, error) {
	return s.configRepo.Get(ctx)
}

func (s *SettingsService) UpdatePodcastConfig(ctx context.Context, config *domain.PodcastConfig) (*domain.PodcastConfig, error) {
	return s.configRepo.Update(ctx, config)
}
