package pageform

import "github.com/gracchi-stdio/castogo/internal/domain"

type Args struct {
	Page          *domain.Page
	ParentPages   []*domain.Page
	Blocks        []*domain.PageBlock
	LinkedEpisode *domain.Episode // read-only display in Settings;
}
