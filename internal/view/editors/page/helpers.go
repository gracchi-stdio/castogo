package pageform

import (
	"fmt"

	"github.com/gracchi-stdio/castogo/internal/domain"
	selectcomponent "github.com/gracchi-stdio/castogo/internal/view/components/select"
)

func parentPageOptions(parentPages []*domain.Page) []selectcomponent.SelectOptionArgs {
	opts := make([]selectcomponent.SelectOptionArgs, 0, len(parentPages))

	for _, p := range parentPages {
		// Only root pages with a slug can parent children — the homepage
		// (empty slug) is the site root "/" itself and stays parent-less.
		if p.ParentID != nil || p.Slug == "" {
			continue
		}

		opts = append(opts, selectcomponent.SelectOptionArgs{
			Value: fmt.Sprintf("%d", p.ID),
			Label: p.Title + " (" + p.Path + ")",
		})
	}
	return opts
}
