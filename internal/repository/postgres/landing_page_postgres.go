package postgres

import (
	"context"
	"encoding/json"

	"github.com/gracchi-stdio/castogo/internal/db"
	"github.com/gracchi-stdio/castogo/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LandingPagePostgres struct {
	q *db.Queries
}

func NewLandingPageRepository(pool *pgxpool.Pool) *LandingPagePostgres {
	return &LandingPagePostgres{
		q: db.New(pool),
	}
}

func (r *LandingPagePostgres) GetAll(ctx context.Context) ([]*domain.LandingPageSection, error) {
	rows, err := r.q.GetAllLandingSections(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*domain.LandingPageSection, len(rows))
	for i := range rows {
		result[i] = toDomainLandingSection(&rows[i])
	}
	return result, nil
}

func (r *LandingPagePostgres) GetVisible(ctx context.Context) ([]*domain.LandingPageSection, error) {
	rows, err := r.q.GetVisibleLandingSections(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*domain.LandingPageSection, len(rows))
	for i := range rows {
		result[i] = toDomainLandingSection(&rows[i])
	}
	return result, nil
}

func (r *LandingPagePostgres) GetByKey(ctx context.Context, key string) (*domain.LandingPageSection, error) {
	row, err := r.q.GetLandingSectionByKey(ctx, key)
	if err != nil {
		return nil, err
	}
	return toDomainLandingSection(&row), nil
}

func (r *LandingPagePostgres) Update(ctx context.Context, section *domain.UpdateLandingSection) (*domain.LandingPageSection, error) {
	params := db.UpdateLandingSectionParams{
		ID: section.ID,
	}
	if section.Content != nil {
		s := string(*section.Content)
		params.Content = &s
	}
	if section.IsVisible != nil {
		params.IsVisible = section.IsVisible
	}
	if section.SortOrder != nil {
		v := int32(*section.SortOrder)
		params.SortOrder = &v
	}
	row, err := r.q.UpdateLandingSection(ctx, params)
	if err != nil {
		return nil, err
	}
	return toDomainLandingSection(&row), nil
}

func toDomainLandingSection(s *db.LandingPageSection) *domain.LandingPageSection {
	return &domain.LandingPageSection{
		ID:         s.ID,
		SectionKey: s.SectionKey,
		Content:    json.RawMessage(s.Content),
		IsVisible:  s.IsVisible,
		SortOrder:  int(s.SortOrder),
		UpdatedAt:  s.UpdatedAt.Time,
	}
}
