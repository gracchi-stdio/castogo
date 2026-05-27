package handler

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/gracchi-stdio/castogo/internal/config"
	"github.com/gracchi-stdio/castogo/internal/domain"
	"github.com/gracchi-stdio/castogo/internal/service"
	"github.com/gracchi-stdio/castogo/internal/view/pageadminview"
	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"
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

	return echo.WrapHandler(templ.Handler(pageadminview.PageFormPage(getSharedData(c), nil, parentPages, nil, false, "")))(c)
}

type pageCreateInput struct {
	Title    string  `json:"title" validate:"required"`
	Slug     string  `json:"slug"`
	Layout   string  `json:"page_layout"`
	ParentID float64 `json:"parent_id"`
}

type pageUpdateInput struct {
	Title       string  `json:"title" validate:"required"`
	Slug        string  `json:"slug"`
	Layout      string  `json:"page_layout"`
	ParentID    float64 `json:"parent_id"`
	IsPublished struct {
		Checked bool `json:"checked"`
	} `json:"is_published"`
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
		if errors.Is(err, domain.ErrHomepageExists) {
			return sse(c).MarshalAndPatchSignals(map[string]string{
				"slug_error": "A homepage already exists — only one root page may have an empty slug",
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

	defaultTab := c.QueryParam("tab")
	if defaultTab == "" {
		defaultTab = "settings"
	}

	return echo.WrapHandler(templ.Handler(pageadminview.PageFormPage(
		getSharedData(c),
		pageWithBlocks.Page,
		parentPages,
		pageWithBlocks.Blocks,
		true,
		defaultTab,
	)))(c)
}

func (h *AdminHandler) pageUpdateAction(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid page ID")
	}

	var rawSignals map[string]any
	if err := readSignals(c, &rawSignals); err != nil {
		return sse(c).MarshalAndPatchSignals(map[string]string{
			"error": "Invalid request",
		})
	}

	var raw pageUpdateInput
	rawBytes, _ := json.Marshal(rawSignals)
	json.Unmarshal(rawBytes, &raw)

	if err := validate.Struct(raw); err != nil {
		sse(c).MarshalAndPatchSignals(fieldValidationErrors(err))
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
		return sse(c).MarshalAndPatchSignals(map[string]string{
			"error": "Failed to update page",
		})
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

	// Close all block edit modes
	patchSignals := map[string]any{
		"error": "",
	}
	if pwb != nil {
		for _, block := range pwb.Blocks {
			patchSignals[fmt.Sprintf("block_%d_editing", block.ID)] = false
		}
	}
	sse(c).MarshalAndPatchSignals(patchSignals)
	return nil
}

func (h *AdminHandler) blockReorderAction(c echo.Context) error {
	pageID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid page ID")
	}

	var raw map[string]any
	if err := readSignals(c, &raw); err != nil {
		return sse(c).MarshalAndPatchSignals(map[string]string{
			"error": "Invalid request",
		})
	}

	blockIDsRaw, ok := raw["block_ids"].([]any)
	if !ok {
		return sse(c).MarshalAndPatchSignals(map[string]string{
			"error": "Missing block order",
		})
	}

	blockIDs := make([]int64, 0, len(blockIDsRaw))
	for _, v := range blockIDsRaw {
		if id, ok := v.(float64); ok {
			blockIDs = append(blockIDs, int64(id))
		}
	}

	if err := h.pageService.ReorderBlocks(c.Request().Context(), pageID, blockIDs); err != nil {
		return sse(c).MarshalAndPatchSignals(map[string]string{
			"error": "Failed to reorder blocks",
		})
	}

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

	sse(c).ExecuteScript(fmt.Sprintf("window.navigateAdmin(%q)", fmt.Sprintf("/admin/pages/%d/edit?tab=blocks", pageID)))
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

	sse(c).ExecuteScript(fmt.Sprintf("window.navigateAdmin(%q)", fmt.Sprintf("/admin/pages/%d/edit?tab=blocks", pageID)))
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

	cleanSignalString := func(v string) string {
		v = strings.TrimSpace(v)
		if len(v) >= 2 {
			if (v[0] == '\'' && v[len(v)-1] == '\'') || (v[0] == '"' && v[len(v)-1] == '"') {
				v = v[1 : len(v)-1]
			}
		}
		return v
	}

	get := func(name string) string {
		if v, ok := signals[prefix+name].(string); ok {
			return cleanSignalString(v)
		}
		return ""
	}

	getInt := func(name string) int {
		if v, ok := signals[prefix+name]; ok {
			switch n := v.(type) {
			case float64:
				return int(n)
			case string:
				i, _ := strconv.Atoi(n)
				return i
			}
		}
		return 0
	}

	switch blockType {
	case "hero":
		content["headline"] = get("headline")
		content["subheadline"] = get("subheadline")
		content["cta_text"] = get("cta_text")
		content["cta_url"] = get("cta_url")
		content["background_image"] = get("background_image")
	case "cta":
		content["headline"] = get("headline")
		content["description"] = get("description")
		content["button_text"] = get("button_text")
		content["button_url"] = get("button_url")
	case "features":
		content["section_title"] = get("section_title")
		content["section_description"] = get("section_description")
		items := []map[string]any{}
		for i := 0; ; i++ {
			icon := get(fmt.Sprintf("item_%d_icon", i))
			title := get(fmt.Sprintf("item_%d_title", i))
			desc := get(fmt.Sprintf("item_%d_description", i))
			if title == "" && icon == "" && desc == "" {
				break
			}
			items = append(items, map[string]any{
				"icon": icon, "title": title, "description": desc,
			})
		}
		content["items"] = items
	case "episodes_showcase":
		content["section_title"] = get("section_title")
		content["section_description"] = get("section_description")
		if max := getInt("max_episodes"); max > 0 {
			content["max_episodes"] = max
		}
		if mode := get("display_mode"); mode != "" {
			content["display_mode"] = mode
		} else {
			content["display_mode"] = "grid"
		}
	case "testimonials":
		content["section_title"] = get("section_title")
		content["section_description"] = get("section_description")
		items := []map[string]any{}
		for i := 0; ; i++ {
			quote := get(fmt.Sprintf("item_%d_quote", i))
			author := get(fmt.Sprintf("item_%d_author", i))
			role := get(fmt.Sprintf("item_%d_role", i))
			avatarURL := get(fmt.Sprintf("item_%d_avatar_url", i))
			if quote == "" && author == "" {
				break
			}
			items = append(items, map[string]any{
				"quote": quote, "author": author, "role": role, "avatar_url": avatarURL,
			})
		}
		content["items"] = items
	case "footer":
		content["text"] = get("text")
		content["copyright"] = get("copyright")
		links := []map[string]any{}
		for i := 0; ; i++ {
			label := get(fmt.Sprintf("link_%d_label", i))
			url := get(fmt.Sprintf("link_%d_url", i))
			if label == "" && url == "" {
				break
			}
			links = append(links, map[string]any{"label": label, "url": url})
		}
		content["links"] = links
		socialLinks := []map[string]any{}
		for i := 0; ; i++ {
			platform := get(fmt.Sprintf("social_%d_platform", i))
			url := get(fmt.Sprintf("social_%d_url", i))
			if platform == "" && url == "" {
				break
			}
			socialLinks = append(socialLinks, map[string]any{"platform": platform, "url": url})
		}
		content["social_links"] = socialLinks
	case "prose":
		content["body"] = get("body")
	}

	return content
}

func (h *AdminHandler) blockUploadImage(c echo.Context) error {
	if err := c.Request().ParseMultipartForm(10 << 20); err != nil {
		sse(c).MarshalAndPatchSignals(map[string]string{
			"error": "Failed to parse form data",
		})
		return nil
	}

	signalName := c.FormValue("signal_name")
	if signalName == "" {
		sse(c).MarshalAndPatchSignals(map[string]string{
			"error": "Missing signal name",
		})
		return nil
	}

	file, header, err := c.Request().FormFile("image_file")
	if err != nil {
		sse(c).MarshalAndPatchSignals(map[string]string{
			"error": "Please select a file to upload",
		})
		return nil
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
		sse(c).MarshalAndPatchSignals(map[string]string{
			"error": "Invalid file type. Please upload a JPG, PNG, or WebP image.",
		})
		return nil
	}

	b := make([]byte, 4)
	rand.Read(b)
	filename := fmt.Sprintf("%s/block_img_%x%s", strings.ToLower(strings.TrimSpace(config.Cfg.AppName)), b, ext)

	url, err := h.storageService.UploadFile(c.Request().Context(), file, filename)
	if err != nil {
		sse(c).MarshalAndPatchSignals(map[string]string{
			"error": "Failed to upload image. Please try again.",
		})
		return nil
	}

	sse(c).MarshalAndPatchSignals(map[string]string{
		signalName: url,
	})
	return nil
}

func (h *AdminHandler) blockAddItemAction(c echo.Context) error {
	pageID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid page ID")
	}

	blockID, err := strconv.ParseInt(c.Param("blockId"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid block ID")
	}

	pwb, err := h.pageService.GetPageWithBlocks(c.Request().Context(), pageID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Page not found")
	}

	var block *domain.PageBlock
	for _, b := range pwb.Blocks {
		if b.ID == blockID {
			block = b
			break
		}
	}
	if block == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Block not found")
	}

	var content map[string]any
	if err := json.Unmarshal(block.Content, &content); err != nil {
		content = map[string]any{}
	}

	itemType := c.QueryParam("type")
	switch itemType {
	case "feature":
		items := toSlice(content["items"])
		items = append(items, map[string]any{"icon": "", "title": "", "description": ""})
		content["items"] = items
	case "testimonial":
		items := toSlice(content["items"])
		items = append(items, map[string]any{"quote": "", "author": "", "role": "", "avatar_url": ""})
		content["items"] = items
	case "link":
		items := toSlice(content["links"])
		items = append(items, map[string]any{"label": "", "url": ""})
		content["links"] = items
	case "social":
		items := toSlice(content["social_links"])
		items = append(items, map[string]any{"platform": "", "url": ""})
		content["social_links"] = items
	default:
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid item type")
	}

	block.Content, _ = json.Marshal(content)
	if _, err := h.pageService.SaveBlock(c.Request().Context(), block); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to save block")
	}

	// Re-render items via SSE patch (no full page reload)
	html, err := pageadminview.RenderItemsFragment(pageID, block, itemType)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to render items")
	}

	containerID := pageadminview.ItemsContainerID(blockID, itemType)
	signals := map[string]any{
		fmt.Sprintf("block_%d_editing", blockID): true,
	}
	for k, v := range pageadminview.NewItemSignals(block, itemType) {
		signals[k] = v
	}

	sse(c).MarshalAndPatchSignals(signals)
	sse(c).PatchElements(html, datastar.WithSelectorID(containerID), datastar.WithModeInner())
	return nil
}

func (h *AdminHandler) blockRemoveItemAction(c echo.Context) error {
	pageID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid page ID")
	}

	blockID, err := strconv.ParseInt(c.Param("blockId"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid block ID")
	}

	index, err := strconv.Atoi(c.Param("index"))
	if err != nil || index < 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid index")
	}

	pwb, err := h.pageService.GetPageWithBlocks(c.Request().Context(), pageID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Page not found")
	}

	var block *domain.PageBlock
	for _, b := range pwb.Blocks {
		if b.ID == blockID {
			block = b
			break
		}
	}
	if block == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Block not found")
	}

	var content map[string]any
	if err := json.Unmarshal(block.Content, &content); err != nil {
		content = map[string]any{}
	}

	itemType := c.QueryParam("type")
	switch itemType {
	case "feature", "testimonial":
		items := toSlice(content["items"])
		if index < len(items) {
			content["items"] = append(items[:index], items[index+1:]...)
		}
	case "link":
		items := toSlice(content["links"])
		if index < len(items) {
			content["links"] = append(items[:index], items[index+1:]...)
		}
	case "social":
		items := toSlice(content["social_links"])
		if index < len(items) {
			content["social_links"] = append(items[:index], items[index+1:]...)
		}
	default:
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid item type")
	}

	block.Content, _ = json.Marshal(content)
	if _, err := h.pageService.SaveBlock(c.Request().Context(), block); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to save block")
	}

	// Re-render items via SSE patch (no full page reload)
	html, err := pageadminview.RenderItemsFragment(pageID, block, itemType)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to render items")
	}

	containerID := pageadminview.ItemsContainerID(blockID, itemType)
	signals := map[string]any{
		fmt.Sprintf("block_%d_editing", blockID): true,
	}
	for k, v := range pageadminview.AllItemSignals(block, itemType) {
		signals[k] = v
	}

	sse(c).MarshalAndPatchSignals(signals)
	sse(c).PatchElements(html, datastar.WithSelectorID(containerID), datastar.WithModeInner())
	return nil
}

// toSlice converts a JSON array to []any, returning nil if not an array.
func toSlice(v any) []any {
	if v == nil {
		return nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	return arr
}
