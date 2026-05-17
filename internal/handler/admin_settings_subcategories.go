package handler

import (
	"github.com/gracchi-stdio/castogo/internal/domain"
	"github.com/gracchi-stdio/castogo/internal/view/settings_page"
	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"
)

func (h *AdminHandler) subcategoriesSSE(c echo.Context) error {
	category := c.QueryParam("category")
	selected := c.QueryParam("subcategory")
	// Guard against "undefined" string literal from race conditions in JS signal init order
	if selected == "" || selected == "undefined" {
		selected = ""
	}
	options := settings_page.GetSubcategoryOptions(category)
	if selected != "" && !domain.IsValidCategory(category, selected) {
		selected = ""
	}

	sse(c).MarshalAndPatchSignals(map[string]any{
		"subcategory_select": map[string]string{
			"value": selected,
			"label": "",
		},
	})

	sse(c).PatchElementTempl(
		settings_page.SubcategoryFragment(options, selected),
		datastar.WithSelectorID("subcategory_wrapper"),
		datastar.WithModeInner(),
	)

	return nil
}
