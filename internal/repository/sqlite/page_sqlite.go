package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gracchi-stdio/castogo/internal/domain"
	"github.com/gracchi-stdio/castogo/internal/repository"
	"github.com/gracchi-stdio/castogo/internal/sqlcsqlite"
)

var _ repository.PageRepository = (*PageRepo)(nil)

type PageRepo struct {
	q *sqlcsqlite.Queries
}

func NewPageRepo(db *sql.DB) *PageRepo {
	return &PageRepo{q: sqlcsqlite.New(db)}
}

func navOrderToSQLite(v *int32) *int64 {
	if v == nil {
		return nil
	}
	return toPtr(int64(*v))
}

func navOrderFromSQLite(v *int64) *int32 {
	if v == nil {
		return nil
	}
	return toPtr(int32(*v))
}

func pageFromRow(p sqlcsqlite.Page) (*domain.Page, error) {
	createdAt, err := timeFromSQLite(p.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("parse page created_at: %w", err)
	}
	updatedAt, err := timeFromSQLite(p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("parse page updated_at: %w", err)
	}
	return &domain.Page{
		ID:          p.ID,
		Title:       p.Title,
		Slug:        p.Slug,
		Layout:      p.Layout,
		ParentID:    p.ParentID,
		IsPublished: p.IsPublished,
		NavOrder:    navOrderFromSQLite(p.NavOrder),
		Path:        p.Path,
		Metadata:    p.Metadata,
		SortOrder:   int(p.SortOrder),
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}, nil
}

func pagesFromList(list []sqlcsqlite.Page) ([]*domain.Page, error) {
	out := make([]*domain.Page, len(list))
	for i, row := range list {
		p, err := pageFromRow(row)
		if err != nil {
			return nil, err
		}
		out[i] = p
	}
	return out, nil
}

func blockFromRow(b sqlcsqlite.PageBlock) (*domain.PageBlock, error) {
	createdAt, err := timeFromSQLite(b.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("parse block created_at: %w", err)
	}
	updatedAt, err := timeFromSQLite(b.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("parse block updated_at: %w", err)
	}
	return &domain.PageBlock{
		ID:        b.ID,
		PageID:    b.PageID,
		BlockType: b.BlockType,
		Content:   json.RawMessage(b.Content),
		SortOrder: int(b.SortOrder),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

func blocksFromList(list []sqlcsqlite.PageBlock) ([]*domain.PageBlock, error) {
	out := make([]*domain.PageBlock, len(list))
	for i, row := range list {
		b, err := blockFromRow(row)
		if err != nil {
			return nil, err
		}
		out[i] = b
	}
	return out, nil
}

// Create implements [repository.PageRepository].
func (p *PageRepo) Create(ctx context.Context, page *domain.Page) (*domain.Page, error) {
	result, err := p.q.CreatePage(ctx, sqlcsqlite.CreatePageParams{
		Title:       page.Title,
		Slug:        page.Slug,
		Layout:      page.Layout,
		ParentID:    page.ParentID,
		IsPublished: page.IsPublished,
		NavOrder:    navOrderToSQLite(page.NavOrder),
		Metadata:    page.Metadata,
		Path:        page.Path,
		SortOrder:   int64(page.SortOrder),
	})
	if err != nil {
		return nil, fmt.Errorf("create page: %w", err)
	}
	return pageFromRow(result)
}

// CreateBlock implements [repository.PageRepository].
func (p *PageRepo) CreateBlock(ctx context.Context, block *domain.PageBlock) (*domain.PageBlock, error) {
	result, err := p.q.CreateBlock(ctx, sqlcsqlite.CreateBlockParams{
		Content:   domain.RawMessage(block.Content),
		SortOrder: int64(block.SortOrder),
		BlockType: block.BlockType,
		PageID:    block.PageID,
	})
	if err != nil {
		return nil, fmt.Errorf("create block: %w", err)
	}
	return blockFromRow(result)
}

// Delete implements [repository.PageRepository].
func (p *PageRepo) Delete(ctx context.Context, id int64) error {
	return p.q.DeletePage(ctx, id)
}

// DeleteBlock implements [repository.PageRepository].
func (p *PageRepo) DeleteBlock(ctx context.Context, id int64) error {
	return p.q.DeleteBlock(ctx, id)
}

// GetBlockByPageID implements [repository.PageRepository].
func (p *PageRepo) GetBlockByPageID(ctx context.Context, pageID int64) ([]*domain.PageBlock, error) {
	blocks, err := p.q.GetBlocksByPageID(ctx, pageID)
	if err != nil {
		return nil, fmt.Errorf("get blocks by page ID: %w", err)
	}
	return blocksFromList(blocks)
}

// GetByID implements [repository.PageRepository].
func (p *PageRepo) GetByID(ctx context.Context, id int64) (*domain.Page, error) {
	page, err := p.q.GetPageByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get page by ID: %w", err)
	}
	return pageFromRow(page)
}

// GetByPath implements [repository.PageRepository].
func (p *PageRepo) GetByPath(ctx context.Context, path string) (*domain.Page, error) {
	page, err := p.q.GetPageByPath(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("get page by path: %w", err)
	}
	return pageFromRow(page)
}

// GetBySlug implements [repository.PageRepository].
func (p *PageRepo) GetBySlug(ctx context.Context, slug string) (*domain.Page, error) {
	page, err := p.q.GetPageBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("get page by slug: %w", err)
	}
	return pageFromRow(page)
}

// GetChildren implements [repository.PageRepository].
func (p *PageRepo) GetChildren(ctx context.Context, parentID int64) ([]*domain.Page, error) {
	pages, err := p.q.GetChildren(ctx, &parentID)
	if err != nil {
		return nil, fmt.Errorf("get page children: %w", err)
	}
	return pagesFromList(pages)
}

// GetPublishedByPath implements [repository.PageRepository].
func (p *PageRepo) GetPublishedByPath(ctx context.Context, path string) (*domain.Page, error) {
	page, err := p.q.GetPublishedPageByPath(ctx, path)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get published page by path: %w", err)
	}
	return pageFromRow(page)
}

// GetTopLevelPublished implements [repository.PageRepository].
func (p *PageRepo) GetTopLevelPublished(ctx context.Context) ([]*domain.Page, error) {
	pages, err := p.q.GetPublishedTopLevelPages(ctx)
	if err != nil {
		return nil, fmt.Errorf("get top-level published pages: %w", err)
	}
	return pagesFromList(pages)
}

// List implements [repository.PageRepository].
func (p *PageRepo) List(ctx context.Context) ([]*domain.Page, error) {
	pages, err := p.q.ListPages(ctx, sqlcsqlite.ListPagesParams{
		Search:     "",
		PageOffset: 0,
		PageLimit:  100,
	})
	if err != nil {
		return nil, fmt.Errorf("list page: %w", err)
	}
	return pagesFromList(pages)
}

// SearchPublished implements [repository.PageRepository].
func (p *PageRepo) SearchPublished(ctx context.Context, query string, limit int, offset int) ([]*domain.Page, error) {
	pages, err := p.q.SearchPublishedPages(ctx, sqlcsqlite.SearchPublishedPagesParams{
		Search:     &query,
		PageOffset: int64(limit),
		PageLimit:  int64(offset),
	})
	if err != nil {
		return nil, fmt.Errorf("search published pages: %w", err)
	}
	return pagesFromList(pages)
}

// Update implements [repository.PageRepository].
func (p *PageRepo) Update(ctx context.Context, page *domain.Page) (*domain.Page, error) {
	result, err := p.q.UpdatePage(ctx, sqlcsqlite.UpdatePageParams{
		ID:          page.ID,
		Title:       &page.Title,
		Slug:        &page.Slug,
		Layout:      &page.Layout,
		ParentID:    page.ParentID,
		IsPublished: &page.IsPublished,
		NavOrder:    navOrderToSQLite(page.NavOrder),
		Metadata:    page.Metadata,
		Path:        &page.Path,
		SortOrder:   toPtr(int64(page.SortOrder)),
	})
	if err != nil {
		return nil, fmt.Errorf("update page: %w", err)
	}
	return pageFromRow(result)
}

// UpdateBlock implements [repository.PageRepository].
func (p *PageRepo) UpdateBlock(ctx context.Context, block *domain.PageBlock) (*domain.PageBlock, error) {
	result, err := p.q.UpdateBlock(ctx, sqlcsqlite.UpdateBlockParams{
		Content:   domain.RawMessage(block.Content),
		SortOrder: toPtr(int64(block.SortOrder)),
		ID:        block.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("update block: %w", err)
	}
	return blockFromRow(result)
}

// UpdateBlockOrder implements [repository.PageRepository].
func (p *PageRepo) UpdateBlockOrder(ctx context.Context, id int64, sortOrder int) error {
	return p.q.UpdateBlockOrder(ctx, sqlcsqlite.UpdateBlockOrderParams{
		ID:        id,
		SortOrder: int64(sortOrder),
	})
}

// UpdateDescendantPaths implements [repository.PageRepository].
func (p *PageRepo) UpdateDescendantPaths(ctx context.Context, oldPrefix string, newPrefix string) error {
	return p.q.UpdateDescendantPaths(ctx, sqlcsqlite.UpdateDescendantPathsParams{
		OldPrefix: &oldPrefix,
		NewPrefix: &newPrefix,
	})
}
