package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gracchi-stdio/castogo/internal/domain"
	"github.com/gracchi-stdio/castogo/internal/repository"
)

type LandingPageService struct {
	landingRepo repository.LandingPageRepository
	episodeRepo repository.EpisodeRepository
}

func NewLandingPageService(landingRepo repository.LandingPageRepository, episodeRepo repository.EpisodeRepository) *LandingPageService {
	return &LandingPageService{
		landingRepo: landingRepo,
		episodeRepo: episodeRepo,
	}
}

func (s *LandingPageService) GetAllSections(ctx context.Context) ([]*domain.LandingPageSection, error) {
	return s.landingRepo.GetAll(ctx)
}

func (s *LandingPageService) GetVisibleSections(ctx context.Context) ([]*domain.LandingPageSection, error) {
	return s.landingRepo.GetVisible(ctx)
}

func (s *LandingPageService) UpdateSection(ctx context.Context, section *domain.UpdateLandingSection) (*domain.LandingPageSection, error) {
	return s.landingRepo.Update(ctx, section)
}

func (s *LandingPageService) GetLandingPageData(ctx context.Context) (*domain.LandingPageData, error) {
	sections, err := s.landingRepo.GetVisible(ctx)
	if err != nil {
		return nil, fmt.Errorf("get visible sections: %w", err)
	}

	data := &domain.LandingPageData{}

	for _, sec := range sections {
		data.SectionOrder = append(data.SectionOrder, sec.SectionKey)
		switch sec.SectionKey {
		case "hero":
			var c domain.HeroContent
			if err := json.Unmarshal(sec.Content, &c); err == nil {
				data.Hero = &c
			}
		case "features":
			var c domain.FeaturesContent
			if err := json.Unmarshal(sec.Content, &c); err == nil {
				data.Features = &c
			}
		case "episodes_showcase":
			var c domain.EpisodesShowcaseContent
			if err := json.Unmarshal(sec.Content, &c); err == nil {
				data.EpisodesShowcase = &c
			}
		case "testimonials":
			var c domain.TestimonialsContent
			if err := json.Unmarshal(sec.Content, &c); err == nil {
				data.Testimonials = &c
			}
		case "cta":
			var c domain.CTAContent
			if err := json.Unmarshal(sec.Content, &c); err == nil {
				data.CTA = &c
			}
		case "footer":
			var c domain.FooterContent
			if err := json.Unmarshal(sec.Content, &c); err == nil {
				data.Footer = &c
			}
		}
	}

	if data.EpisodesShowcase != nil {
		max := data.EpisodesShowcase.MaxEpisodes
		if max <= 0 {
			max = 6
		}
		episodes, err := s.episodeRepo.ListPublished(ctx, max, 0)
		if err == nil {
			data.LatestEpisodes = episodes
		}
	}

	return data, nil
}
