package brand

import (
	"github.com/a-h/templ"

	"github.com/gracchi-stdio/castogo/internal/view/utils"
)

func Mascot(class ...string) templ.Component {
	cls := utils.TwMerge(append([]string{"size-12"}, class...)...)
	return templ.Raw(`<img src="/mascot_dark.svg" alt="Castogo mascot" class="hidden dark:block ` + cls + `" /><img src="/mascot_light.svg" alt="Castogo mascot" class="dark:hidden ` + cls + `" />`)
}
