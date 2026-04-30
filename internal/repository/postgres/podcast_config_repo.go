package postgres

import (
	"context"

	"github.com/gracchi-stdio/castogo/internal/db"
	"github.com/gracchi-stdio/castogo/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PodcastConfigRepository struct {
	q *db.Queries
}

func NewPodcastConfigRepository(pool *pgxpool.Pool) *PodcastConfigRepository {
	return &PodcastConfigRepository{
		q: db.New(pool),
	}
}

func (r *PodcastConfigRepository) Get(ctx context.Context) (*domain.PodcastConfig, error) {
	res, err := r.q.GetPodcastConfig(ctx)
	if err != nil {
		return nil, err
	}
	return toDomainPodcastConfig(&res), nil
}

func (r *PodcastConfigRepository) Update(ctx context.Context, config *domain.PodcastConfig) (*domain.PodcastConfig, error) {
	params := db.UpdatePodcastConfigParams{
		ID:            config.ID,
		Title:         &config.Title,
		Description:   &config.Description,
		SiteUrl:       &config.SiteURL,
		Language:      &config.Language,
		Copyright:     &config.Copyright,
		AuthorName:    &config.AuthorName,
		AuthorEmail:   &config.AuthorEmail,
		CoverImageUrl: &config.CoverImageURL,
		Category:      &config.Category,
		Subcategory:   &config.Subcategory,
		OwnerName:     &config.OwnerName,
		OwnerEmail:    &config.OwnerEmail,
	}
	res, err := r.q.UpdatePodcastConfig(ctx, params)
	if err != nil {
		return nil, err
	}
	return toDomainPodcastConfig(&res), nil

}
