package episodeForm

import "github.com/gracchi-stdio/castogo/internal/domain"

// Args is the view model for the episode edit form.
type Args struct {
	Episode    *domain.Episode
	Pages      []*domain.Page // candidate pages for "link existing"
	LinkedPage *domain.Page   // resolved linked page (nil unless Episode.LinkedPageID is set)
}

// ListArgs is the view model for the admin episode list. Status and Search carry
// the active filter so the filter bar can highlight the current pill and prefill
// the search box.
type ListArgs struct {
	Episodes []*domain.Episode
	Status   string // active status filter ("" = all)
	Search   string // active title search
}
