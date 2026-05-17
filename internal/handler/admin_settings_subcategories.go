package handler

import (
	"github.com/gracchi-stdio/castogo/internal/view/settings_page"
	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"
)

func (h *AdminHandler) subcategoriesSSE(c echo.Context) error {
	category := c.QueryParam("category")
	options := settings_page.GetSubcategoryOptions(category)

	sse(c).MarshalAndPatchSignals(map[string]any{
		"subcategory_select": map[string]string{
			"value": "",
			"label": "",
		},
	})

	sse(c).PatchElementTempl(
		settings_page.SubcategoryFragment(options),
		datastar.WithSelectorID("subcategory_wrapper"),
		datastar.WithModeInner(),
	)

	return nil
}
