package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gracchi-stdio/castogo/internal/domain"
	"github.com/gracchi-stdio/castogo/internal/repository"
)

type FeedService struct {
	configRepo  repository.PodcastConfigRepository
	episodeRepo repository.EpisodeRepository
}

func NewFeedService(configRepo repository.PodcastConfigRepository, episodeRepo repository.EpisodeRepository) *FeedService {
	return &FeedService{
		configRepo:  configRepo,
		episodeRepo: episodeRepo,
	}
}

func (s *FeedService) BuildFeed(ctx context.Context) (*domain.RSS, error) {
	config, err := s.configRepo.Get(ctx)
	if err != nil {
		return nil, err
	}

	// WithPagePath (not ListPublished) so each item's <link> can point at the
	// episode's real public page — there is no /episodes/{slug} route.
	ewps, err := s.episodeRepo.ListPublishedWithPagePath(ctx, 500, 0)

	if err != nil {
		return nil, err
	}

	// Episode numbers are derived from publish order; assign them before
	// building items so <itunes:episode> reflects the chronological rank.
	ranks, err := episodeNumberRanks(ctx, s.episodeRepo)
	if err != nil {
		return nil, fmt.Errorf("compute episode numbers: %w", err)
	}
	episodes := make([]*domain.Episode, len(ewps))
	for i, ewp := range ewps {
		episodes[i] = ewp.Episode
	}
	applyEpisodeNumbers(episodes, ranks)

	// Build category with optional subcategory
	var category *domain.ITunesCategory
	if config.Category != "" {
		category = &domain.ITunesCategory{
			Text: config.Category,
		}
		if config.Subcategory != "" {
			category.SubItems = &domain.ITunesCategory{
				Text: config.Subcategory,
			}
		}
	}

	// Build items from published episodes
	items := make([]domain.Item, 0, len(ewps))
	for _, ewp := range ewps {
		item := s.buildItem(ewp, config)
		items = append(items, item)
	}

	channel := domain.Channel{
		Title:          config.Title,
		Description:    config.Description,
		Link:           config.SiteURL,
		Language:       config.Language,
		Copyright:      config.Copyright,
		ITunesAuthor:   config.AuthorName,
		ITunesType:     "episodic",
		ITunesExplicit: "false",
		Generator:      "CASToGo",
		PodcastLocked:  "yes",
		Owner: &domain.ITunesOwner{
			Name:  config.OwnerName,
			Email: config.OwnerEmail,
		},
		Image:    &domain.ITunesImage{Href: config.CoverImageURL},
		Category: category,
		Items:    items,
	}

	return domain.NewRSSFeed(channel), nil
}

func (s *FeedService) buildItem(ep *domain.EpisodeWithPagePath, config *domain.PodcastConfig) domain.Item {
	// Determine the publish date: use PublishAt if set, otherwise CreatedAt
	pubTime := ep.CreatedAt
	if ep.PublishAt != nil {
		pubTime = *ep.PublishAt
	}

	item := domain.Item{
		Title:       ep.Title,
		Description: ep.Description,
		GUID: domain.GUID{
			IsPermaLink: "false",
			// Stable on the immutable episode id — NOT the number, which is now
			// derived from publish order and would change if dates change.
			Value: "podlog-ep-" + strconv.FormatInt(ep.ID, 10),
		},
		PubDate: &domain.PubDate{Time: pubTime},
		Enclosure: domain.Enclosure{
			URL:    ep.AudioSourceURL,
			Length: ep.AudioMetadata.FileSize,
			Type:   ep.AudioMetadata.MimeType,
		},
		ITunesDuration: domain.Duration{Duration: time.Duration(ep.Duration) * time.Second},
		ITunesExplicit: strconv.FormatBool(ep.Explicit),
		ITunesImage:    &domain.ITunesImage{Href: config.CoverImageURL}, // Default to podcast cover if episode-specific image is not set
	}

	// Optional: episode number
	if ep.EpisodeNumber > 0 {
		item.ITunesEpisode = &ep.EpisodeNumber
	}

	// Episode page link: the linked companion page when one exists (there is
	// no /episodes/{slug} route), otherwise the site root.
	if config.SiteURL != "" {
		item.Link = config.SiteURL
		if ep.PagePath != nil {
			item.Link += *ep.PagePath
		}
	}

	// Optional: per-episode cover image
	if ep.CoverImageURL != "" {
		item.ITunesImage = &domain.ITunesImage{Href: ep.CoverImageURL}
	}

	return item
}
