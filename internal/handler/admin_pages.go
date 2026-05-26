package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/a-h/templ"
	"github.com/gracchi-stdio/castogo/internal/domain"
	"github.com/gracchi-stdio/castogo/internal/service"
	"github.com/gracchi-stdio/castogo/internal/view/pageadminview"
	"github.com/labstack/echo/v4"
)

func (h *AdminHandler) pageList(c echo.Context) error {
	pages, err := h.pageService.ListPages(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load pages")
	}

	return echo.WrapHandler(templ.Handler(pageadminview.PageListPage(getSharedData(c), pages)))(c)
}

func (h *AdminHandler) pageCreatePage(c echo.Context) error {
	parentPages, err := h.pageService.ListPages(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load parent pages")
	}

	return echo.WrapHandler(templ.Handler(pageadminview.PageFormPage(getSharedData(c), nil, parentPages, nil, false)))(c)
}

type pageCreateInput struct {
	Title    string  `json:"title" validate:"required"`
	Slug     string  `json:"slug" validate:"required"`
	Layout   string  `json:"page_layout"`
	ParentID float64 `json:"parent_id"`
}

func (h *AdminHandler) pageCreateAction(c echo.Context) error {
	var raw pageCreateInput
	if err := readSignals(c, &raw); err != nil {
		return sse(c).MarshalAndPatchSignals(map[string]string{
			"error": "Invalid request",
		})
	}

	if err := validate.Struct(raw); err != nil {
		sse(c).MarshalAndPatchSignals(fieldValidationErrors(err))
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
		if errors.Is(err, domain.ErrMaxDepth) {
			return sse(c).MarshalAndPatchSignals(map[string]string{
				"error": "Maximum nesting depth exceeded (2 levels max)",
			})
		}
		return sse(c).MarshalAndPatchSignals(map[string]string{
			"error": "Failed to create page",
		})
	}

	sse(c).ExecuteScript(fmt.Sprintf("window.navigateAdmin(%q)", fmt.Sprintf("/admin/pages/%d/edit", page.ID)))
	return nil
}

func (h *AdminHandler) pageEditPage(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid page ID")
	}

	pageWithBlocks, err := h.pageService.GetPageWithBlocks(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Page not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load page")
	}

	parentPages, err := h.pageService.ListPages(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load parent pages")
	}

	return echo.WrapHandler(templ.Handler(pageadminview.PageFormPage(
		getSharedData(c),
		pageWithBlocks.Page,
		parentPages,
		pageWithBlocks.Blocks,
		true,
	)))(c)
}

func (h *AdminHandler) pageUpdateAction(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid page ID")
	}

	var raw map[string]any
	if err := readSignals(c, &raw); err != nil {
		return sse(c).MarshalAndPatchSignals(map[string]string{
			"error": "Invalid request",
		})
	}

	input := service.UpdatePageInput{}

	if v, ok := raw["title"].(string); ok {
		input.Title = &v
	}
	if v, ok := raw["slug"].(string); ok {
		input.Slug = &v
	}
	if v, ok := raw["page_layout"].(string); ok {
		input.Layout = &v
	}
	if v, ok := raw["parent_id"].(float64); ok {
		pid := int64(v)
		ppid := &pid
		input.ParentID = &ppid
	}
	if v, ok := raw["is_published"].(bool); ok {
		input.IsPublished = &v
	}

	_, err = h.pageService.UpdatePage(c.Request().Context(), id, input)
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
		return sse(c).MarshalAndPatchSignals(map[string]string{
			"error": "Failed to update page",
		})
	}

	sse(c).MarshalAndPatchSignals(map[string]string{
		"error": "",
	})
	return nil
}

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

func (h *AdminHandler) blockCreateAction(c echo.Context) error {
	pageID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid page ID")
	}

	var raw map[string]any
	if err := readSignals(c, &raw); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}

	blockType, _ := raw["new_block_type"].(string)

	block := &domain.PageBlock{
		PageID:    pageID,
		BlockType: blockType,
		Content:   json.RawMessage(`{}`),
	}

	_, err = h.pageService.SaveBlock(c.Request().Context(), block)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create block")
	}

	sse(c).ExecuteScript(fmt.Sprintf("window.navigateAdmin(%q)", fmt.Sprintf("/admin/pages/%d/edit", pageID)))
	return nil
}

func (h *AdminHandler) blockDeleteAction(c echo.Context) error {
	pageID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid page ID")
	}

	blockID, err := strconv.ParseInt(c.Param("blockId"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid block ID")
	}

	if err := h.pageService.DeleteBlock(c.Request().Context(), blockID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete block")
	}

	sse(c).ExecuteScript(fmt.Sprintf("window.navigateAdmin(%q)", fmt.Sprintf("/admin/pages/%d/edit", pageID)))
	return nil
}

func (h *AdminHandler) blockUpdateAction(c echo.Context) error {
	pageID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid page ID")
	}

	blockID, err := strconv.ParseInt(c.Param("blockId"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid block ID")
	}

	// Load current blocks first — block type comes from the database (source of truth)
	pageWithBlocks, err := h.pageService.GetPageWithBlocks(c.Request().Context(), pageID)
	if err != nil {
		return sse(c).MarshalAndPatchSignals(map[string]string{
			"error": "Failed to load page",
		})
	}

	var currentBlock *domain.PageBlock
	for _, b := range pageWithBlocks.Blocks {
		if b.ID == blockID {
			currentBlock = b
			break
		}
	}
	if currentBlock == nil {
		return sse(c).MarshalAndPatchSignals(map[string]string{
			"error": "Block not found",
		})
	}

	var raw map[string]any
	if err := readSignals(c, &raw); err != nil {
		return sse(c).MarshalAndPatchSignals(map[string]string{
			"error": "Invalid request",
		})
	}

	// Build content map based on the block type from the database
	content := buildBlockContent(blockID, currentBlock.BlockType, raw)
	contentJSON, err := json.Marshal(content)
	if err != nil {
		return sse(c).MarshalAndPatchSignals(map[string]string{
			"error": "Failed to encode block content",
		})
	}

	currentBlock.Content = contentJSON
	if _, err := h.pageService.SaveBlock(c.Request().Context(), currentBlock); err != nil {
		return sse(c).MarshalAndPatchSignals(map[string]string{
			"error": "Failed to save block",
		})
	}

	// Toggle back to view mode — the saved signal values persist in the frontend
	sse(c).MarshalAndPatchSignals(map[string]any{
		fmt.Sprintf("block_%d_editing", blockID): false,
	})
	return nil
}

// buildBlockContent extracts block content from Datastar signals based on block type.
func buildBlockContent(blockID int64, blockType string, signals map[string]any) map[string]any {
	content := map[string]any{}
	prefix := fmt.Sprintf("block_%d_", blockID)

	// Helper: get signal value as string, defaulting to empty
	get := func(name string) string {
		if v, ok := signals[prefix+name].(string); ok {
			return v
		}
		return ""
	}

	switch blockType {
	case "hero", "cta":
		content["headline"] = get("headline")
		content["subheadline"] = get("subheadline")
		if blockType == "hero" {
			content["cta_text"] = get("cta_text")
			content["cta_url"] = get("cta_url")
		} else {
			content["button_text"] = get("button_text")
			content["button_url"] = get("button_url")
		}
	case "features", "testimonials", "episodes_showcase":
		content["section_title"] = get("section_title")
	case "footer":
		content["text"] = get("text")
	}

	return content
}
