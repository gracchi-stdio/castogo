package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/gosimple/slug"
	"github.com/gracchi-stdio/castogo/internal/domain"
	"github.com/gracchi-stdio/castogo/internal/repository"
)

type EpisodeService struct {
	repo repository.EpisodeRepository
}

func NewEpisodeService(repo repository.EpisodeRepository) *EpisodeService {
	return &EpisodeService{repo: repo}
}

func (s *EpisodeService) Create(ctx context.Context, ep *domain.Episode) (*domain.Episode, error) {
	if ep.Slug == "" {
		ep.Slug = slug.Make(ep.Title)
	}

	if ep.EpisodeNumber == 0 {
		max, err := s.repo.GetMaxEpisodeNumber(ctx)
		if err != nil {
			return nil, fmt.Errorf("get next episode number: %w", err)
		}
		ep.EpisodeNumber = max + 1
	}

	created, err := s.repo.Create(ctx, ep)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique") {
			return nil, domain.ErrDuplicateSlug
		}
		return nil, fmt.Errorf("create episode: %w", err)
	}

	return created, nil
}

func (s *EpisodeService) GetByID(ctx context.Context, id int64) (*domain.Episode, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *EpisodeService) GetBySlug(ctx context.Context, slug string) (*domain.Episode, error) {
	return s.repo.GetBySlug(ctx, slug)
}

func (s *EpisodeService) List(ctx context.Context, filter repository.EpisodeFilter) ([]*domain.Episode, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	return s.repo.List(ctx, filter)
}

func (s *EpisodeService) Update(ctx context.Context, ep *domain.UpdateEpisode) (*domain.Episode, error) {
	return s.repo.Update(ctx, ep)
}

func (s *EpisodeService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s *EpisodeService) CountByStatus(ctx context.Context, status domain.EpisodeStatus) (int, error) {
	return s.repo.CountByStatus(ctx, status)
}

func (s *EpisodeService) ListPublished(ctx context.Context, limit, offset int) ([]*domain.Episode, error) {
	if limit <= 0 {
		limit = 20
	}

	return s.repo.ListPublished(ctx, limit, offset)
}

func (s *EpisodeService) GetDashboardStats(ctx context.Context) (*domain.DashboardStats, error) {
	published, err := s.repo.CountByStatus(ctx, domain.EpisodeStatusPublished)
	if err != nil {
		return nil, fmt.Errorf("count published: %w", err)
	}

	drafts, err := s.repo.CountByStatus(ctx, domain.EpisodeStatusDraft)
	if err != nil {
		return nil, fmt.Errorf("count drafts: %w", err)
	}

	scheduled, err := s.repo.CountByStatus(ctx, domain.EpisodeStatusScheduled)
	if err != nil {
		return nil, fmt.Errorf("count scheduled: %w", err)
	}

	archived, err := s.repo.CountByStatus(ctx, domain.EpisodeStatusArchived)
	if err != nil {
		return nil, fmt.Errorf("count archived: %w", err)
	}

	return &domain.DashboardStats{
		Total:     published + drafts + scheduled + archived,
		Published: published,
		Drafts:    drafts,
		Scheduled: scheduled,
	}, nil
}
