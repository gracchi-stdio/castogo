package landingadminview

import (
	"encoding/json"
	"strings"

	"github.com/gracchi-stdio/castogo/internal/domain"
)

func buildSectionMap(sections []*domain.LandingPageSection) map[string]*domain.LandingPageSection {
	m := make(map[string]*domain.LandingPageSection)
	for _, s := range sections {
		m[s.SectionKey] = s
	}
	return m
}

func sectionID(m map[string]*domain.LandingPageSection, key string) int64 {
	if s, ok := m[key]; ok {
		return s.ID
	}
	return 0
}

func sectionVisible(m map[string]*domain.LandingPageSection, key string) bool {
	if s, ok := m[key]; ok {
		return s.IsVisible
	}
	return true
}

func heroContent(m map[string]*domain.LandingPageSection) domain.HeroContent {
	s, ok := m["hero"]
	if !ok {
		return domain.HeroContent{}
	}
	var c domain.HeroContent
	json.Unmarshal(s.Content, &c)
	return c
}

func featuresContent(m map[string]*domain.LandingPageSection) domain.FeaturesContent {
	s, ok := m["features"]
	if !ok {
		return domain.FeaturesContent{}
	}
	var c domain.FeaturesContent
	json.Unmarshal(s.Content, &c)
	return c
}

func episodesContent(m map[string]*domain.LandingPageSection) domain.EpisodesShowcaseContent {
	s, ok := m["episodes_showcase"]
	if !ok {
		return domain.EpisodesShowcaseContent{MaxEpisodes: 6}
	}
	var c domain.EpisodesShowcaseContent
	json.Unmarshal(s.Content, &c)
	return c
}

func testimonialsContent(m map[string]*domain.LandingPageSection) domain.TestimonialsContent {
	s, ok := m["testimonials"]
	if !ok {
		return domain.TestimonialsContent{}
	}
	var c domain.TestimonialsContent
	json.Unmarshal(s.Content, &c)
	return c
}

func ctaContent(m map[string]*domain.LandingPageSection) domain.CTAContent {
	s, ok := m["cta"]
	if !ok {
		return domain.CTAContent{}
	}
	var c domain.CTAContent
	json.Unmarshal(s.Content, &c)
	return c
}

func footerContent(m map[string]*domain.LandingPageSection) domain.FooterContent {
	s, ok := m["footer"]
	if !ok {
		return domain.FooterContent{}
	}
	var c domain.FooterContent
	json.Unmarshal(s.Content, &c)
	return c
}

func escapeJS(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, `\`, `\\`), `'`, `\'`)
}
