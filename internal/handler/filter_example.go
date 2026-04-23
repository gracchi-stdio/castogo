package handler

import (
	"fmt"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/gracchi-stdio/castogo/internal/view"
	"github.com/starfederation/datastar-go/datastar"
)

// placeholder episodes — will come from database later
var episodes = []struct {
	Title  string
	Status string
}{
	{"Episode 1: Getting Started", "published"},
	{"Episode 2: Setting Up Tools", "published"},
	{"Episode 3: Draft Notes", "draft"},
	{"Episode 4: Old Recording", "archived"},
	{"Episode 5: Work in Progress", "draft"},
	{"Episode 6: Live Show", "published"},
	{"Episode 7: Retired Episode", "archived"},
}

type FilterHandler struct{}

func NewFilterHandler() *FilterHandler {
	return &FilterHandler{}
}

func (h *FilterHandler) RegisterRoutes(e *echo.Echo) {
	e.GET("/filter", echo.WrapHandler(templ.Handler(view.FilterPage())))
	e.GET("/filter/items", h.items)
}

func (h *FilterHandler) items(c echo.Context) error {
	signals := &struct {
		Status string `json:"status"`
	}{}
	if err := readSignals(c, signals); err != nil {
		return err
	}

	// filter episodes by selected status
	filtered := episodes
	if signals.Status != "all" {
		filtered = nil
		for _, ep := range episodes {
			if ep.Status == signals.Status {
				filtered = append(filtered, ep)
			}
		}
	}

	// build HTML for the filtered list
	html := `<div id="list"><ul>`
	if len(filtered) == 0 {
		html += `<li>No episodes found.</li>`
	}
	for _, ep := range filtered {
		html += fmt.Sprintf(`<li>%s <small>(%s)</small></li>`, ep.Title, ep.Status)
	}
	html += `</ul></div>`

	sse(c).PatchElements(html, datastar.WithUseViewTransitions(true))
	return nil
}
