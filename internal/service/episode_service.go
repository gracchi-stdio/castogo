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

	eps, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	if err := s.applyEpisodeNumbers(ctx, eps); err != nil {
		return nil, err
	}
	return eps, nil
}

// episodeRanks returns a map of episode ID -> 1-based chronological rank, built
// from ListForRanking (all non-archived episodes with a publish_at, in order).
// Episode numbers are derived, not stored — see episode_numbering.go.
func (s *EpisodeService) episodeRanks(ctx context.Context) (map[int64]int, error) {
	return episodeNumberRanks(ctx, s.repo)
}

// applyEpisodeNumbers sets each episode's EpisodeNumber from the global rank
// map. Episodes absent from the map (drafts, archived) keep number 0.
func (s *EpisodeService) applyEpisodeNumbers(ctx context.Context, eps []*domain.Episode) error {
	ranks, err := s.episodeRanks(ctx)
	if err != nil {
		return err
	}
	applyEpisodeNumbers(eps, ranks)
	return nil
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

	eps, err := s.repo.ListPublished(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	if err := s.applyEpisodeNumbers(ctx, eps); err != nil {
		return nil, err
	}
	return eps, nil
}

func (s *EpisodeService) ListPublishedWithPagePath(ctx context.Context, limit, offset int) ([]*domain.EpisodeWithPagePath, error) {
	if limit <= 0 {
		limit = 20
	}

	ewps, err := s.repo.ListPublishedWithPagePath(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	ranks, err := s.episodeRanks(ctx)
	if err != nil {
		return nil, err
	}
	for _, ewp := range ewps {
		ewp.EpisodeNumber = ranks[ewp.ID]
	}
	return ewps, nil
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

func (s *EpisodeService) SearchPublishedWithPagePath(ctx context.Context, query string, limit, offset int) ([]*domain.EpisodeWithPagePath, error) {
	if limit <= 0 {
		limit = 20
	}
	ewps, err := s.repo.SearchPublishedWithPagePath(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	ranks, err := s.episodeRanks(ctx)
	if err != nil {
		return nil, err
	}
	for _, ewp := range ewps {
		ewp.EpisodeNumber = ranks[ewp.ID]
	}
	return ewps, nil
}

func (s *EpisodeService) LinkPage(ctx context.Context, episodeID, pageID int64) error {
	return s.repo.UpdateLinkedPageID(ctx, episodeID, &pageID)
}

func (s *EpisodeService) UnlinkPage(ctx context.Context, episodeID int64) error {
	return s.repo.UpdateLinkedPageID(ctx, episodeID, nil)
}

func (s *EpisodeService) GetByLinkedPageID(ctx context.Context, pageID int64) (*domain.Episode, error) {
	return s.repo.GetByLinkedPageID(ctx, pageID)
}

// ListPublishedByLinkedPageID returns the published episodes linked to a page
// (the reverse of linked_page_id), with derived episode numbers applied — it
// feeds the public page's "listen" section.
func (s *EpisodeService) ListPublishedByLinkedPageID(ctx context.Context, pageID int64) ([]*domain.Episode, error) {
	eps, err := s.repo.ListPublishedByLinkedPageID(ctx, pageID)
	if err != nil {
		return nil, err
	}
	if err := s.applyEpisodeNumbers(ctx, eps); err != nil {
		return nil, err
	}
	return eps, nil
}
