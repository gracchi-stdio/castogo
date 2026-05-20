package service

import (
	"context"

	"github.com/gracchi-stdio/castogo/internal/domain"
	"github.com/gracchi-stdio/castogo/internal/repository"
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
	"admin": true,
	"api":   true,
	"login": true,
	"signup": true,
	"register": true,
	"healthcheck": true,
	"logout": true,
	"feed": true,
	// Add more reserved slugs as needed
}

type CreatePageInput struct {
	Title string
	Slug string
	Layout string
	ParentID *int64
}

func (s *PageService) CreatePage(ctx context.Context, input CreatePageInput) (*domain.Page, error) {
	if reservedSlugs[input.Slug] {
		return nil, domain.ErrReservedSlug
	}

	path := input.Slug
	order := 1
	if input.ParentID != nil {
		parent, err := s.page.GetPathWithChildrenCount(ctx, input.ParentID)
		if err != nil {
			return nil, err
		}
		path = parent.Path + "/" + path
		order = parent.ChildrenCount + 1
	}

	page := &domain.Page{
		Title:       input.Title,
		Slug:        input.Slug,
		Layout:      input.Layout,
		ParentID:    input.ParentID,
		Path:        path,
		SortOrder:        order,
	}

	page, err := s.page.Create(ctx, page)
	if err != nil {
		if isDuplicateKeyError(err) {
			return nil, domain.ErrSlugAlreadyExists
		}
		return nil, err
	}

	return page, nil
}
func (s *PageService) ListPages(ctx context.Context) ([]*domain.Page, error) {

}

func (s *PageService) GetChildren(ctx context.Context, parentID int64) ([]*domain.Page, error)
}

func (s *PageService) UpdatePage(ctx context.Context, page *domain.Page) (*domain.Page, error) {
}

func (s *PageService) DeletePage(ctx context.Context, id int64) error {
}

func (s *PageService) GetPageWithBlocks(ctx context.Context, id int64) (*domain.Page, error) {
}

func (s *PageService) SavePageBlock(ctx context.Context, pageID int64, blocks []*domain.PageBlock) ([]*domain.PageBlock, error) {
}


func isDuplicateKeyError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
