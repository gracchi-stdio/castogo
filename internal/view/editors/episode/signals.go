package episodeForm

import "github.com/gracchi-stdio/castogo/internal/domain"

type Signals struct {
	// ... existing fields ...

	PageLinkMode      string `json:"pageLinkMode"` // "none" | "existing" | "new"
	ExistingPageID    *int64 `json:"existingPageId"`
	ExistingPageError string `json:"existingPageError"`
}

func NewSignals(episode *domain.Episode) Signals {
	s := Signals{}

	// ... existing code ...
	s.PageLinkMode = "none"

	if episode != nil && episode.LinkedPageID != nil {
		s.PageLinkMode = "existing"
		s.ExistingPageID = episode.LinkedPageID
	}
	return s
}
