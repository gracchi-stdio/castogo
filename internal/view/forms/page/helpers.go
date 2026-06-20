package pageform

import (
	"fmt"

	"github.com/gracchi-stdio/castogo/internal/domain"
	selectcomponent "github.com/gracchi-stdio/castogo/internal/view/components/select"
)

func parentPageOptions(parentPages []*domain.Page) []selectcomponent.SelectOptionArgs {
	opts := make([]selectcomponent.SelectOptionArgs, 0, len(parentPages))

	for _, p := range parentPages {
		if p.ParentID != nil {
			continue
		}

		opts = append(opts, selectcomponent.SelectOptionArgs{
			Value: fmt.Sprintf("%d", p.ID),
			Label: p.Title + " (/" + p.Path + ")",
		})
	}
	return opts
}
