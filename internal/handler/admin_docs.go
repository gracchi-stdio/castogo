package handler

import (
	"errors"
	"net/http"

	"github.com/a-h/templ"
	"github.com/gracchi-stdio/castogo/internal/docs"
	"github.com/gracchi-stdio/castogo/internal/view/docsview"
	"github.com/labstack/echo/v5"
)

// Four thin route handlers funnel into one private helper. They're kept
// separate (rather than a /docs/:section param route) because Echo's radix
// tree rejects a parametrized segment that collides with the static
// /docs/developer and /docs/user segments.

func (h *AdminHandler) docsDeveloperIndex(c *echo.Context) error {
	return h.renderDocs(c, docs.SectionDeveloper, "")
}

func (h *AdminHandler) docsDeveloperShow(c *echo.Context) error {
	return h.renderDocs(c, docs.SectionDeveloper, c.Param("slug"))
}

func (h *AdminHandler) docsUserIndex(c *echo.Context) error {
	return h.renderDocs(c, docs.SectionUser, "")
}

func (h *AdminHandler) docsUserShow(c *echo.Context) error {
	return h.renderDocs(c, docs.SectionUser, c.Param("slug"))
}

// renderDocs serves both the section root (slug == "") and a specific doc
// (slug != ""). At the root the first listed document is rendered as the
// default view; if the section is empty, an empty-state placeholder is shown.
// Unknown section or slug maps to 404 (rendered by the custom error handler).
func (h *AdminHandler) renderDocs(c *echo.Context, section docs.Section, slug string) error {
	items, err := docs.List(section)
	if err != nil {
		if errors.Is(err, docs.ErrUnknownSection) {
			return echo.NewHTTPError(http.StatusNotFound, "Unknown docs section")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load docs listing")
	}

	var current *docs.Doc
	if slug == "" {
		// Section root: render the first doc (if any) as the default view.
		if len(items) > 0 {
			current, _ = docs.Load(section, items[0].Slug)
		}
	} else {
		current, err = docs.Load(section, slug)
		if err != nil {
			if errors.Is(err, docs.ErrNotFound) {
				return echo.NewHTTPError(http.StatusNotFound, "Doc not found")
			}
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load doc")
		}
	}

	title := sectionLabel(section)
	if current != nil {
		title = current.Title
	}

	args := docsview.DocsPageArgs{
		Section:      section,
		SectionLabel: sectionLabel(section),
		Items:        items,
		Current:      current,
	}

	return echo.WrapHandler(templ.Handler(docsview.DocsPage(getSharedData(c), title, args)))(c)
}

func sectionLabel(s docs.Section) string {
	switch s {
	case docs.SectionDeveloper:
		return "Developer Docs"
	case docs.SectionUser:
		return "User Docs"
	}
	return "Docs"
}
