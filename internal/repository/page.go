package repository

import (
	"context"

	"github.com/gracchi-stdio/castogo/internal/domain"
)

type PageRepository interface {
	Create(ctx context.Context, page *domain.Page) (*domain.Page, error)
	GetByID(ctx context.Context, id int64) (*domain.Page, error)
	GetByPath(ctx context.Context, path string) (*domain.Page, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Page, error)
	List(ctx context.Context) ([]*domain.Page, error)
	GetChildren(ctx context.Context, parentID int64) ([]*domain.Page, error)
	Update(ctx context.Context, page *domain.Page) (*domain.Page, error)
	Delete(ctx context.Context, id int64) error
	UpdateDescendantPaths(ctx context.Context, oldPrefix, newPrefix string) error

	// PageBlock methods
	CreateBlock(ctx context.Context, block *domain.PageBlock) (*domain.PageBlock, error)
	GetBlockByPageID(ctx context.Context, pageID int64) ([]*domain.PageBlock, error)
	UpdateBlock(ctx context.Context, block *domain.PageBlock) (*domain.PageBlock, error)
	DeleteBlock(ctx context.Context, id int64) error
	UpdateBlockOrder(ctx context.Context, id int64, sortOrder int) error

	// Navigation
	GetTopLevelPublished(ctx context.Context) ([]*domain.Page, error)
	SearchPublished(ctx context.Context, query string, limit, offset int) ([]*domain.Page, error)
}
