package docsview

import "github.com/gracchi-stdio/castogo/internal/docs"

// DocsPageArgs holds everything the docs page needs: the section being viewed,
// the full left-nav listing, and the currently-selected document (nil when the
// section is empty).
type DocsPageArgs struct {
	Section      docs.Section
	SectionLabel string
	Items        []docs.DocMeta
	Current      *docs.Doc
}

// ternary returns t when b is true, else f. Used to compose Tailwind class
// strings conditionally inside templ attributes.
func ternary(b bool, t, f string) string {
	if b {
		return t
	}
	return f
}
