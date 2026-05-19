package handler

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/gracchi-stdio/castogo/internal/domain"
	"github.com/gracchi-stdio/castogo/internal/view/landingadminview"
	"github.com/labstack/echo/v4"
)

// ────────────────────────────────────────────────────────────
// GET /admin/landing — render the landing page editor
// ────────────────────────────────────────────────────────────

func (h *AdminHandler) landingEditor(c echo.Context) error {
	sections, err := h.landingService.GetAllSections(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(500, "Failed to load landing page sections")
	}
	return echo.WrapHandler(templ.Handler(
		landingadminview.LandingEditorPage(getSharedData(c), sections),
	))(c)
}

// ────────────────────────────────────────────────────────────
// POST /admin/landing/:sectionKey — save a single section
// ────────────────────────────────────────────────────────────

func (h *AdminHandler) landingSaveSection(c echo.Context) error {
	sectionKey := c.Param("sectionKey")

	// Look up the section to get its database ID
	sections, err := h.landingService.GetAllSections(c.Request().Context())
	if err != nil {
		sse(c).MarshalAndPatchSignals(map[string]any{
			"error":          "Failed to load section",
			"loading_status": false,
		})
		return nil
	}

	var sectionID int64
	for _, s := range sections {
		if s.SectionKey == sectionKey {
			sectionID = s.ID
			break
		}
	}
	if sectionID == 0 {
		sse(c).MarshalAndPatchSignals(map[string]any{
			"error":          fmt.Sprintf("Unknown section: %s", sectionKey),
			"loading_status": false,
		})
		return nil
	}

	// Parse submitted form data
	if err := c.Request().ParseForm(); err != nil {
		sse(c).MarshalAndPatchSignals(map[string]any{
			"error":          "Failed to parse form data",
			"loading_status": false,
		})
		return nil
	}
	form := c.Request().Form

	// Build the content JSON from form values
	content, parseErr := buildSectionContent(sectionKey, form)
	if parseErr != nil {
		sse(c).MarshalAndPatchSignals(map[string]any{
			"error":          parseErr.Error(),
			"loading_status": false,
		})
		return nil
	}

	raw, err := json.Marshal(content)
	if err != nil {
		sse(c).MarshalAndPatchSignals(map[string]any{
			"error":          "Failed to encode section data",
			"loading_status": false,
		})
		return nil
	}
	rawMsg := json.RawMessage(raw)

	isVisible := form.Get("is_visible") == "true"
	update := &domain.UpdateLandingSection{
		ID:        sectionID,
		Content:   &rawMsg,
		IsVisible: &isVisible,
	}

	if _, err := h.landingService.UpdateSection(c.Request().Context(), update); err != nil {
		sse(c).MarshalAndPatchSignals(map[string]any{
			"error":          "Failed to save section: " + err.Error(),
			"loading_status": false,
		})
		return nil
	}

	sse(c).MarshalAndPatchSignals(map[string]any{
		"success":        "Section saved!",
		"loading_status": false,
	})
	return nil
}

// ────────────────────────────────────────────────────────────
// Form-to-content builders (one per section key)
// ────────────────────────────────────────────────────────────

func buildSectionContent(sectionKey string, form url.Values) (any, error) {
	switch sectionKey {
	case "hero":
		return domain.HeroContent{
			Headline:        form.Get("headline"),
			Subheadline:     form.Get("subheadline"),
			CTAText:         form.Get("cta_text"),
			CTAURL:          form.Get("cta_url"),
			BackgroundImage: form.Get("background_image"),
		}, nil

	case "features":
		items := parseFeatureItems(form)
		return domain.FeaturesContent{
			SectionTitle:       form.Get("section_title"),
			SectionDescription: form.Get("section_description"),
			Items:              items,
		}, nil

	case "episodes_showcase":
		maxEp := 6
		if v := form.Get("max_episodes"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				maxEp = n
			}
		}
		return domain.EpisodesShowcaseContent{
			SectionTitle:       form.Get("section_title"),
			SectionDescription: form.Get("section_description"),
			MaxEpisodes:        maxEp,
		}, nil

	case "testimonials":
		items := parseTestimonialItems(form)
		return domain.TestimonialsContent{
			SectionTitle:       form.Get("section_title"),
			SectionDescription: form.Get("section_description"),
			Items:              items,
		}, nil

	case "cta":
		return domain.CTAContent{
			Headline:    form.Get("headline"),
			Description: form.Get("description"),
			ButtonText:  form.Get("button_text"),
			ButtonURL:   form.Get("button_url"),
		}, nil

	case "footer":
		links := parseFooterLinks(form)
		socials := parseSocialLinks(form)
		return domain.FooterContent{
			Copyright:   form.Get("copyright"),
			Links:       links,
			SocialLinks: socials,
		}, nil

	default:
		return nil, fmt.Errorf("unknown section key: %s", sectionKey)
	}
}

// ────────────────────────────────────────────────────────────
// Indexed form field parsers
// ────────────────────────────────────────────────────────────

// parseFeatureItems reads items.N.icon / items.N.title / items.N.description from form values.
func parseFeatureItems(form url.Values) []domain.FeatureItem {
	var items []domain.FeatureItem
	for i := 0; ; i++ {
		prefix := fmt.Sprintf("items.%d.", i)
		if !formHasPrefix(form, prefix) {
			break
		}
		items = append(items, domain.FeatureItem{
			Icon:        form.Get(prefix + "icon"),
			Title:       form.Get(prefix + "title"),
			Description: form.Get(prefix + "description"),
		})
	}
	return items
}

// parseTestimonialItems reads items.N.quote / items.N.author / items.N.role / items.N.avatar_url.
func parseTestimonialItems(form url.Values) []domain.TestimonialItem {
	var items []domain.TestimonialItem
	for i := 0; ; i++ {
		prefix := fmt.Sprintf("items.%d.", i)
		if !formHasPrefix(form, prefix) {
			break
		}
		items = append(items, domain.TestimonialItem{
			Quote:     form.Get(prefix + "quote"),
			Author:    form.Get(prefix + "author"),
			Role:      form.Get(prefix + "role"),
			AvatarURL: form.Get(prefix + "avatar_url"),
		})
	}
	return items
}

// parseFooterLinks reads links.N.label / links.N.url.
func parseFooterLinks(form url.Values) []domain.FooterLink {
	var links []domain.FooterLink
	for i := 0; ; i++ {
		prefix := fmt.Sprintf("links.%d.", i)
		if !formHasPrefix(form, prefix) {
			break
		}
		links = append(links, domain.FooterLink{
			Label: form.Get(prefix + "label"),
			URL:   form.Get(prefix + "url"),
		})
	}
	return links
}

// parseSocialLinks reads social_links.N.platform / social_links.N.url.
func parseSocialLinks(form url.Values) []domain.SocialLink {
	var socials []domain.SocialLink
	for i := 0; ; i++ {
		prefix := fmt.Sprintf("social_links.%d.", i)
		if !formHasPrefix(form, prefix) {
			break
		}
		socials = append(socials, domain.SocialLink{
			Platform: form.Get(prefix + "platform"),
			URL:      form.Get(prefix + "url"),
		})
	}
	return socials
}

// formHasPrefix returns true if any form key starts with the given prefix.
func formHasPrefix(form url.Values, prefix string) bool {
	for key := range form {
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}
	return false
}
