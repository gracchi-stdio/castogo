package postgres

import (
	"context"

	"github.com/gracchi-stdio/castogo/internal/db"
	"github.com/gracchi-stdio/castogo/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PageRepo struct {
	q *db.Queries
}

func NewPageRepo(pool *pgxpool.Pool) *PageRepo {
	return &PageRepo{
		q: db.New(pool),
	}
}

func (r *PageRepo) Create(ctx context.Context, page *domain.Page) (*domain.Page, error) {
	result, err := r.q.CreatePage(ctx, db.CreatePageParams{
		Title:       page.Title,
		Slug:        page.Slug,
		Layout:      page.Layout,
		ParentID:    page.ParentID,
		IsPublished: page.IsPublished,
		Metadata:    page.Metadata,
		Path:        page.Path,
		SortOrder:   int32(page.SortOrder),
	})
	if err != nil {
		return nil, err
	}
	return toDomainPage(&result), nil
}

func (r *PageRepo) GetByID(ctx context.Context, id int64) (*domain.Page, error) {
	page, err := r.q.GetPageByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toDomainPage(&page), nil
}

func (r *PageRepo) GetByPath(ctx context.Context, path string) (*domain.Page, error) {
	page, err := r.q.GetPageByPath(ctx, path)
	if err != nil {
		return nil, err
	}
	return toDomainPage(&page), nil
}

func (r *PageRepo) GetBySlug(ctx context.Context, slug string) (*domain.Page, error) {
	page, err := r.q.GetPageBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	return toDomainPage(&page), nil
}

func (r *PageRepo) List(ctx context.Context) ([]*domain.Page, error) {
	pages, err := r.q.ListPages(ctx, db.ListPagesParams{
		Search:     "",
		PageOffset: 0,
		PageLimit:  100,
	})
	if err != nil {
		return nil, err
	}
	result := make([]*domain.Page, len(pages))
	for i := range pages {
		result[i] = toDomainPage(&pages[i])
	}
	return result, nil
}

func (r *PageRepo) GetChildren(ctx context.Context, parentID int64) ([]*domain.Page, error) {
	pages, err := r.q.GetChildren(ctx, &parentID)
	if err != nil {
		return nil, err
	}
	result := make([]*domain.Page, len(pages))
	for i := range pages {
		result[i] = toDomainPage(&pages[i])
	}
	return result, nil
}

func (r *PageRepo) Update(ctx context.Context, page *domain.Page) (*domain.Page, error) {
	result, err := r.q.UpdatePage(ctx, db.UpdatePageParams{
		ID:          page.ID,
		Title:       &page.Title,
		Slug:        &page.Slug,
		Layout:      &page.Layout,
		ParentID:    page.ParentID,
		IsPublished: &page.IsPublished,
		Metadata:    page.Metadata,
		Path:        &page.Path,
		SortOrder:   ptrInt32(int32(page.SortOrder)),
	})
	if err != nil {
		return nil, err
	}
	return toDomainPage(&result), nil
}

func (r *PageRepo) Delete(ctx context.Context, id int64) error {
	return r.q.DeletePage(ctx, id)
}

func (r *PageRepo) UpdateDescendantPaths(ctx context.Context, oldPrefix, newPrefix string) error {
	return r.q.UpdateDescendantPaths(ctx, db.UpdateDescendantPathsParams{
		OldPrefix: &oldPrefix,
		NewPrefix: &newPrefix,
	})
}

// Block methods

func (r *PageRepo) CreateBlock(ctx context.Context, block *domain.PageBlock) (*domain.PageBlock, error) {
	result, err := r.q.CreateBlock(ctx, db.CreateBlockParams{
		Content:   block.Content,
		SortOrder: int32(block.SortOrder),
		BlockType: block.BlockType,
		PageID:    block.PageID,
	})
	if err != nil {
		return nil, err
	}
	return toDomainPageBlock(&result), nil
}

func (r *PageRepo) GetBlockByPageID(ctx context.Context, pageID int64) ([]*domain.PageBlock, error) {
	blocks, err := r.q.GetBlocksByPageID(ctx, pageID)
	if err != nil {
		return nil, err
	}
	result := make([]*domain.PageBlock, len(blocks))
	for i := range blocks {
		result[i] = toDomainPageBlock(&blocks[i])
	}
	return result, nil
}

func (r *PageRepo) UpdateBlock(ctx context.Context, block *domain.PageBlock) (*domain.PageBlock, error) {
	result, err := r.q.UpdateBlock(ctx, db.UpdateBlockParams{
		Content:   block.Content,
		SortOrder: ptrInt32(int32(block.SortOrder)),
		ID:        block.ID,
	})
	if err != nil {
		return nil, err
	}
	return toDomainPageBlock(&result), nil
}

func (r *PageRepo) DeleteBlock(ctx context.Context, id int64) error {
	return r.q.DeleteBlock(ctx, id)
}

func (r *PageRepo) GetPathWithChildrenCount(ctx context.Context, id int64) (string, int, error) {
	res, err := r.q.GetPagePathAndChildrenCountByID(ctx, id)
	if err != nil {
		return "", 0, err
	}
	return res.Path, int(res.ChildrenCount), nil
}

// Converters

func toDomainPage(p *db.Page) *domain.Page {
	return &domain.Page{
		ID:          p.ID,
		Title:       p.Title,
		Slug:        p.Slug,
		Layout:      p.Layout,
		ParentID:    p.ParentID,
		IsPublished: p.IsPublished,
		Metadata:    p.Metadata,
		SortOrder:   int(p.SortOrder),
		Path:        p.Path,
		CreatedAt:   p.CreatedAt.Time,
		UpdatedAt:   p.UpdatedAt.Time,
	}
}

func toDomainPageBlock(b *db.PageBlock) *domain.PageBlock {
	return &domain.PageBlock{
		ID:        b.ID,
		PageID:    b.PageID,
		BlockType: b.BlockType,
		Content:   b.Content,
		SortOrder: int(b.SortOrder),
		CreatedAt: b.CreatedAt.Time,
		UpdatedAt: b.UpdatedAt.Time,
	}
}

func ptrInt32(v int32) *int32 {
	return &v
}
