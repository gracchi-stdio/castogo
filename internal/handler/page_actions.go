package handler

import (
	"encoding/json"
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

	var rawSignals map[string]any
	if err := readSignals(c, &rawSignals); err != nil {
		out := sse(c)
		out.ExecuteScript(toastScript("Invalid request", "error"))
		return nil
	}

	var raw pageUpdateInput
	rawBytes, _ := json.Marshal(rawSignals)
	json.Unmarshal(rawBytes, &raw)

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

	update := service.UpdatePageInput{
		Title:       &title,
		Slug:        &slug,
		Layout:      &layout,
		ParentID:    &parentVal,
		IsPublished: &isPublished,
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

	// Save all blocks
	pwb, err := h.pageService.GetPageWithBlocks(c.Request().Context(), id)
	if err == nil {
		for _, block := range pwb.Blocks {
			content := buildBlockContent(block.ID, block.BlockType, rawSignals)
			contentJSON, err := json.Marshal(content)
			if err != nil {
				continue
			}
			block.Content = contentJSON
			h.pageService.SaveBlock(c.Request().Context(), block)
		}
	}

	// Signal save success — leave block edit modes alone so the user's open
	// blocks stay open. Toast feedback confirms the save happened.
	sse(c).ExecuteScript(toastScript("Page saved successfully", "success"))
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

	sse(c).ExecuteScript(fmt.Sprintf("window.navigateAdmin(%q)", fmt.Sprintf("/admin/pages/%d/edit", page.ID)))
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

	sse(c).ExecuteScript(fmt.Sprintf("window.navigateAdmin(%q)", "/admin/pages"))
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
