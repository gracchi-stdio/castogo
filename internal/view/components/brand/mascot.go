package brand

import (
	"github.com/a-h/templ"

	"github.com/gracchi-stdio/castogo/internal/view/utils"
)

func Mascot(class ...string) templ.Component {
	cls := utils.TwMerge(append([]string{"size-12"}, class...)...)
	return templ.Raw(`<img src="/mascot.svg" alt="Castogo mascot" class="` + cls + `" />`)
}
