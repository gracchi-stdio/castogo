package icon

import (
	"github.com/a-h/templ"
	"github.com/kaugesaar/lucide-go"

	"github.com/gracchi-stdio/castogo/internal/view/utils"
)

func Lucide(name string, class ...string) templ.Component {
	cls := utils.TwMerge(append([]string{"size-4"}, class...)...)
	return templ.Raw(lucide.Icon(name, map[string]any{
		"class": cls,
	}))
}
