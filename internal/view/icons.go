package view

import (
	"github.com/a-h/templ"
	"github.com/kaugesaar/lucide-go"
)

// Icon renders a Lucide icon as inline SVG for use in templ templates.
// name: Lucide icon name (e.g. "headphones", "send", "gear")
// class: Tailwind CSS classes for sizing/styling (e.g. "size-4", "size-6 text-muted-foreground")
func Icon(name, class string) templ.Component {
	return templ.Raw(lucide.Icon(name, map[string]any{
		"class": class,
	}))
}
