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
		return toast(c, "Invalid request", "error")
	}
	if err := validate.Struct(raw); err != nil {
		return patchFieldErrors(c, err, raw)
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

	// Settings only saves page metadata. Block content is saved per-block from the
	// Blocks tab (blockUpdateAction) — it must not be rebuilt here, since the Settings
	// tab no longer renders block signals and would otherwise blank every block.
	if _, err := h.pageService.UpdatePage(c.Request().Context(), id, update); err != nil {
		if slugErr := pageSlugError(err); slugErr != nil {
			return patchSignals(c, slugErr)
		}
		return toast(c, "Failed to update page", "error")
	}

	bustPagesCache(c, fmt.Sprintf("/admin/pages/%d/edit", id))
	return toast(c, "Page saved successfully", "success")
}

// page create action
func (h *AdminHandler) pageCreateAction(c echo.Context) error {
	var raw pageCreateInput
	if err := readSignals(c, &raw); err != nil {
		return toast(c, "Invalid request", "error")
	}
	if err := validate.Struct(raw); err != nil {
		return patchFieldErrors(c, err, raw)
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
		if slugErr := pageSlugError(err); slugErr != nil {
			return patchSignals(c, slugErr)
		}
		if errors.Is(err, domain.ErrMaxDepth) {
			return toast(c, "Maximum nesting depth exceeded (2 levels max)", "error")
		}
		return toast(c, "Failed to create page", "error")
	}

	bustPagesCache(c, "")
	return navigate(c, fmt.Sprintf("/admin/pages/%d/edit", page.ID), "", "")
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

	bustPagesCache(c, fmt.Sprintf("/admin/pages/%d/edit", id))
	return navigate(c, "/admin/pages", "", "")
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
