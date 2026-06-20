package service

import (
	"context"
	"errors"
	"strconv"

	"github.com/gosimple/slug"
	"github.com/gracchi-stdio/castogo/internal/domain"
	"github.com/gracchi-stdio/castogo/internal/repository"
	"github.com/jackc/pgx/v5/pgconn"
)

type PageService struct {
	page    repository.PageRepository
	episode repository.EpisodeRepository
}

func NewPageService(pageRepo repository.PageRepository, episodeRepo repository.EpisodeRepository) *PageService {
	return &PageService{
		page:    pageRepo,
		episode: episodeRepo,
	}
}

var reservedSlugs = map[string]bool{
	"admin":       true,
	"api":         true,
	"login":       true,
	"signup":      true,
	"register":    true,
	"healthcheck": true,
	"logout":      true,
	"feed":        true,
	"assets":      true,
}

type CreatePageInput struct {
	Title    string
	Slug     string
	Layout   string
	ParentID *int64
	Metadata domain.PageMetadata
}

type UpdatePageInput struct {
	Title       *string
	Slug        *string
	Layout      *string
	ParentID    **int64
	IsPublished *bool
	Metadata    *domain.PageMetadata
	SortOrder   *int
}

type PageWithBlocks struct {
	Page   *domain.Page
	Blocks []*domain.PageBlock
}

func (s *PageService) CreatePage(ctx context.Context, input CreatePageInput) (*domain.Page, error) {
	// Empty slug is only allowed for root-level pages (homepage)
	if input.Slug == "" && input.ParentID != nil {
		return nil, domain.ErrInvalidInput
	}

	if input.Slug != "" && reservedSlugs[input.Slug] {
		return nil, domain.ErrReservedSlug
	}

	// If creating a root page with empty slug, check one doesn't already exist
	if input.Slug == "" {
		existing, _ := s.page.GetBySlug(ctx, "")
		if existing != nil {
			return nil, domain.ErrHomepageExists
		}
	}

	path := input.Slug
	if input.ParentID != nil {
		parent, err := s.page.GetByID(ctx, *input.ParentID)
		if err != nil {
			return nil, domain.ErrInvalidParent
		}
		if parent.ParentID != nil {
			return nil, domain.ErrMaxDepth
		}
		path = parent.Path + "/" + input.Slug
	}

	page := &domain.Page{
		Title:    input.Title,
		Slug:     input.Slug,
		Layout:   input.Layout,
		ParentID: input.ParentID,
		Path:     path,
		Metadata: input.Metadata,
	}

	created, err := s.page.Create(ctx, page)
	if err != nil {
		if isDuplicateKeyError(err) {
			return nil, domain.ErrDuplicatePath
		}
		return nil, err
	}
	return created, nil
}

func (s *PageService) GetPage(ctx context.Context, id int64) (*domain.Page, error) {
	return s.page.GetByID(ctx, id)
}

func (s *PageService) GetPageByPath(ctx context.Context, path string) (*domain.Page, error) {
	return s.page.GetByPath(ctx, path)
}

func (s *PageService) GetPageBySlug(ctx context.Context, slug string) (*domain.Page, error) {
	return s.page.GetBySlug(ctx, slug)
}

func (s *PageService) ListPages(ctx context.Context) ([]*domain.Page, error) {
	return s.page.List(ctx)
}

func (s *PageService) GetChildren(ctx context.Context, parentID int64) ([]*domain.Page, error) {
	return s.page.GetChildren(ctx, parentID)
}

func (s *PageService) UpdatePage(ctx context.Context, id int64, input UpdatePageInput) (*domain.Page, error) {
	existing, err := s.page.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Slug != nil {
		newSlug := *input.Slug
		// Empty slug only allowed for root pages
		if newSlug == "" {
			if existing.ParentID != nil || (input.ParentID != nil && *input.ParentID != nil) {
				return nil, domain.ErrInvalidInput
			}
			// Check no other root page already has empty slug
			if existing.Slug != "" {
				existingHomepage, _ := s.page.GetBySlug(ctx, "")
				if existingHomepage != nil && existingHomepage.ID != existing.ID {
					return nil, domain.ErrHomepageExists
				}
			}
		} else if reservedSlugs[newSlug] && newSlug != existing.Slug {
			return nil, domain.ErrReservedSlug
		}
	}

	updated := applyUpdates(existing, input)

	needsPathUpdate := (input.Slug != nil && *input.Slug != existing.Slug) ||
		(input.ParentID != nil && derefInt64(*input.ParentID) != derefInt64(existing.ParentID))

	if needsPathUpdate {
		if updated.ParentID != nil {
			parent, err := s.page.GetByID(ctx, *updated.ParentID)
			if err != nil {
				return nil, domain.ErrInvalidParent
			}
			if parent.ParentID != nil {
				return nil, domain.ErrMaxDepth
			}
			updated.Path = parent.Path + "/" + updated.Slug
		} else {
			updated.Path = updated.Slug
		}
	}

	result, err := s.page.Update(ctx, updated)
	if err != nil {
		if isDuplicateKeyError(err) {
			return nil, domain.ErrDuplicatePath
		}
		return nil, err
	}

	if existing.Path != result.Path {
		if err := s.page.UpdateDescendantPaths(ctx, existing.Path, result.Path); err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (s *PageService) DeletePage(ctx context.Context, id int64) error {
	return s.page.Delete(ctx, id)
}

func (s *PageService) GetPageWithBlocks(ctx context.Context, id int64) (*PageWithBlocks, error) {
	page, err := s.page.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	blocks, err := s.page.GetBlockByPageID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &PageWithBlocks{Page: page, Blocks: blocks}, nil
}

func (s *PageService) SaveBlock(ctx context.Context, block *domain.PageBlock) (*domain.PageBlock, error) {
	if block.ID == 0 {
		return s.page.CreateBlock(ctx, block)
	}
	return s.page.UpdateBlock(ctx, block)
}

func (s *PageService) DeleteBlock(ctx context.Context, id int64) error {
	return s.page.DeleteBlock(ctx, id)
}

func (s *PageService) ReorderBlocks(ctx context.Context, pageID int64, blockIDs []int64) error {
	for i, id := range blockIDs {
		if err := s.page.UpdateBlockOrder(ctx, id, i); err != nil {
			return err
		}
	}
	return nil
}

func (s *PageService) GetTopLevelPublished(ctx context.Context) ([]*domain.Page, error) {
	return s.page.GetTopLevelPublished(ctx)
}

func (s *PageService) SearchPublished(ctx context.Context, query string, limit, offset int) ([]*domain.Page, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.page.SearchPublished(ctx, query, limit, offset)
}

func applyUpdates(existing *domain.Page, input UpdatePageInput) *domain.Page {
	p := &domain.Page{
		ID:          existing.ID,
		Title:       existing.Title,
		Slug:        existing.Slug,
		Layout:      existing.Layout,
		ParentID:    existing.ParentID,
		IsPublished: existing.IsPublished,
		Path:        existing.Path,
		Metadata:    existing.Metadata,
		SortOrder:   existing.SortOrder,
	}
	if input.Title != nil {
		p.Title = *input.Title
	}
	if input.Slug != nil {
		p.Slug = *input.Slug
	}
	if input.Layout != nil {
		p.Layout = *input.Layout
	}
	if input.ParentID != nil {
		p.ParentID = *input.ParentID
	}
	if input.IsPublished != nil {
		p.IsPublished = *input.IsPublished
	}
	if input.Metadata != nil {
		p.Metadata = *input.Metadata
	}
	if input.SortOrder != nil {
		p.SortOrder = *input.SortOrder
	}
	return p
}

func derefInt64(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}

func isDuplicateKeyError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

func (s *PageService) CreateCompanionPage(ctx context.Context, episode *domain.Episode) (*domain.Page, error) {
	slug := slug.Make(episode.Title)
	i := 1
	// Ensure slug is unique
	for {
		existing, err := s.page.GetBySlug(ctx, slug)
		if err != nil {
			return nil, err
		}
		if existing == nil {
			break
		}
		slug = slug + "-" + strconv.Itoa(i)
		i++
	}

	// Create page
	page, err := s.page.Create(ctx, &domain.Page{
		Title:     episode.Title,
		Slug:      slug,
		ShowInNav: false,
	})
	if err != nil {
		return nil, err
	}

	// Link episode to page
	if err := s.episode.UpdateLinkedPageID(ctx, episode.ID, &page.ID); err != nil {
		return nil, err
	}

	return page, nil
}
