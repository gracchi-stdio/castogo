package episodeForm

import "github.com/gracchi-stdio/castogo/internal/domain"

// Args is the view model for the episode edit form.
type Args struct {
	Episode    *domain.Episode
	Pages      []*domain.Page // candidate pages for "link existing"
	LinkedPage *domain.Page   // resolved linked page (nil unless Episode.LinkedPageID is set)
}
