package episodeForm

import "github.com/gracchi-stdio/castogo/internal/domain"

type Args struct {
	Episode *domain.Episode
	Page    []*domain.Page

	// ... existing fields ...
}
