package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gracchi-stdio/castogo/internal/domain"
	"github.com/gracchi-stdio/castogo/internal/service"
	"github.com/labstack/echo/v4"
)

// update page action
func (h *AdminHandler) pageUpdateAction(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid page ID")
	}

	var raw pageUpdateInput
	if err := readSignals(c, &raw); err != nil {
		sse(c).ExecuteScript(toastScript("Invalid request", "error"))
		return nil
	}

	if err := validate.Struct(raw); err != nil {
		sse(c).MarshalAndPatchSignals(fieldValidationErrors(err, raw))
		return nil
	}

	var parentVal *int64
	if raw.ParentID > 0 {
		pid := int64(raw.ParentID)
		parentVal = &pid
	}

	title := raw.Title
	slug := raw.Slug
	layout := raw.Layout
	isPublished := raw.IsPublished.Checked
	showInNav := raw.ShowInNav.Checked

	update := service.UpdatePageInput{
		Title:       &title,
		Slug:        &slug,
		Layout:      &layout,
		ParentID:    &parentVal,
		IsPublished: &isPublished,
		ShowInNav:   &showInNav,
	}

	_, err = h.pageService.UpdatePage(c.Request().Context(), id, update)
	if err != nil {
		if errors.Is(err, domain.ErrReservedSlug) {
			return sse(c).MarshalAndPatchSignals(map[string]string{
				"slug_error": "This slug is reserved and cannot be used",
			})
		}
		if errors.Is(err, domain.ErrDuplicatePath) {
			return sse(c).MarshalAndPatchSignals(map[string]string{
				"slug_error": "A page with this slug already exists",
			})
		}
		if errors.Is(err, domain.ErrHomepageExists) {
			return sse(c).MarshalAndPatchSignals(map[string]string{
				"slug_error": "A homepage already exists — only one root page may have an empty slug",
			})
		}
		out := sse(c)
		out.ExecuteScript(toastScript("Failed to update page", "error"))
		return nil
	}

	// Settings only saves page metadata. Block content is saved per-block from the
	// Blocks tab (blockUpdateAction) — it must not be rebuilt here, since the Settings
	// tab no longer renders block signals and would otherwise blank every block.
	sse(c).ExecuteScript(fmt.Sprintf("window.bustPagesCache(%q); "+toastScript("Page saved successfully", "success"), fmt.Sprintf("/admin/pages/%d/edit", id)))
	return nil
}

// page create action
func (h *AdminHandler) pageCreateAction(c echo.Context) error {
	var raw pageCreateInput
	if err := readSignals(c, &raw); err != nil {
		out := sse(c)
		out.ExecuteScript(toastScript("Invalid request", "error"))
		return nil
	}

	if err := validate.Struct(raw); err != nil {
		sse(c).MarshalAndPatchSignals(fieldValidationErrors(err, raw))
		return nil
	}

	var parentID *int64
	if raw.ParentID > 0 {
		pid := int64(raw.ParentID)
		parentID = &pid
	}

	input := service.CreatePageInput{
		Title:    raw.Title,
		Slug:     raw.Slug,
		Layout:   raw.Layout,
		ParentID: parentID,
	}

	page, err := h.pageService.CreatePage(c.Request().Context(), input)
	if err != nil {

		slugError := pageSlugError(err)
		if slugError != nil {
			return sse(c).MarshalAndPatchSignals(slugError)
		}

		// Toast Error for Max depth and else
		out := sse(c)
		if errors.Is(err, domain.ErrMaxDepth) {
			out.ExecuteScript(toastScript("Maximum nesting depth exceeded (2 levels max)", "error"))
			return nil
		}
		out.ExecuteScript(toastScript("Failed to create page", "error"))
		return nil
	}

	sse(c).ExecuteScript(fmt.Sprintf("window.navigateAdmin(%q); window.bustPagesCache()", fmt.Sprintf("/admin/pages/%d/edit", page.ID)))
	return nil
}

// page delete action
func (h *AdminHandler) pageDeleteAction(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid page ID")
	}

	if err := h.pageService.DeletePage(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete page")
	}

	sse(c).ExecuteScript(fmt.Sprintf("window.navigateAdmin(%q); window.bustPagesCache(%q)", "/admin/pages", fmt.Sprintf("/admin/pages/%d/edit", id)))
	return nil
}

// helpers
func pageSlugError(err error) map[string]string {
	switch {
	case errors.Is(err, domain.ErrReservedSlug):
		return map[string]string{"slug_error": "This slug is reserved and cannot be used"}
	case errors.Is(err, domain.ErrDuplicatePath):
		return map[string]string{"slug_error": "A page with this slug already exists"}
	case errors.Is(err, domain.ErrHomepageExists):
		return map[string]string{"slug_error": "A homepage already exists — only one root page may have an empty slug"}
	}

	return nil
}
