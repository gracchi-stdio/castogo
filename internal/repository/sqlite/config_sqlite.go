package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/gracchi-stdio/castogo/internal/domain"
	"github.com/gracchi-stdio/castogo/internal/repository"
	"github.com/gracchi-stdio/castogo/internal/sqlcsqlite"
)

var _ repository.PodcastConfigRepository = (*PodcastConfigRepo)(nil)

type PodcastConfigRepo struct {
	q *sqlcsqlite.Queries
}

func podcastConfigFromRow(p *sqlcsqlite.PodcastConfig) *domain.PodcastConfig {
	return &domain.PodcastConfig{
		ID:            p.ID,
		Title:         p.Title,
		Description:   p.Description,
		SiteURL:       p.SiteUrl,
		Language:      p.Language,
		Copyright:     p.Copyright,
		AuthorName:    p.AuthorName,
		AuthorEmail:   p.AuthorEmail,
		CoverImageURL: p.CoverImageUrl,
		Category:      p.Category,
		Subcategory:   p.Subcategory,
		OwnerName:     p.OwnerName,
		OwnerEmail:    p.OwnerEmail,
	}
}

// Get implements [repository.PodcastConfigRepository].
func (p *PodcastConfigRepo) Get(ctx context.Context) (*domain.PodcastConfig, error) {
	res, err := p.q.GetPodcastConfig(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get podcast config: %w", err)
	}
	return podcastConfigFromRow(&res), nil
}

// Update implements [repository.PodcastConfigRepository].
func (p *PodcastConfigRepo) Update(ctx context.Context, config *domain.UpdatePodcastConfig) (*domain.PodcastConfig, error) {
	res, err := p.q.UpdatePodcastConfig(ctx, sqlcsqlite.UpdatePodcastConfigParams{
		ID:            config.ID,
		Title:         config.Title,
		Description:   config.Description,
		SiteUrl:       config.SiteURL,
		Language:      config.Language,
		Copyright:     config.Copyright,
		AuthorName:    config.AuthorName,
		AuthorEmail:   config.AuthorEmail,
		CoverImageUrl: config.CoverImageURL,
		Category:      config.Category,
		Subcategory:   config.Subcategory,
		OwnerName:     config.OwnerName,
		OwnerEmail:    config.OwnerEmail,
	})
	if err != nil {
		return nil, fmt.Errorf("update podcast config: %w", err)
	}
	return podcastConfigFromRow(&res), nil
}

func NewPodcastConfigRepo(db *sql.DB) *PodcastConfigRepo {
	return &PodcastConfigRepo{q: sqlcsqlite.New(db)}
}
