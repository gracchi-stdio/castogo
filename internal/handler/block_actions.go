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

	"github.com/gracchi-stdio/castogo/internal/config"
	"github.com/gracchi-stdio/castogo/internal/domain"
	blockEditor "github.com/gracchi-stdio/castogo/internal/view/editors/blockeditor"
	pageform "github.com/gracchi-stdio/castogo/internal/view/editors/page"
	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"
)

func (h *AdminHandler) blockReorderAction(c echo.Context) error {
	pageID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid page ID")
	}

	var raw map[string]any
	if err := readSignals(c, &raw); err != nil {
		out := sse(c)
		out.ExecuteScript(toastScript("Invalid request", "error"))
		return nil
	}

	blockIDsRaw, ok := raw["block_ids"].([]any)
	if !ok {
		out := sse(c)
		out.ExecuteScript(toastScript("Missing block order", "error"))
		return nil
	}

	blockIDs := make([]int64, 0, len(blockIDsRaw))
	for _, v := range blockIDsRaw {
		if id, ok := v.(float64); ok {
			blockIDs = append(blockIDs, int64(id))
		}
	}

	if err := h.pageService.ReorderBlocks(c.Request().Context(), pageID, blockIDs); err != nil {
		out := sse(c)
		out.ExecuteScript(toastScript("Failed to reorder blocks", "error"))
		return nil
	}

	// Persisted order diverges from any cached /edit/blocks snapshot — bust it so a
	// later tab switch re-fetches fresh.
	sse(c).ExecuteScript("window.bustBlocksCache()")
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

	saved, err := h.pageService.SaveBlock(c.Request().Context(), block)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create block")
	}

	// Re-render the list with the new block and open it in the form pane — a partial
	// SSE patch, not a reload. A Swup navigation would serve cached, pre-add HTML.
	pwb, err := h.pageService.GetPageWithBlocks(c.Request().Context(), pageID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load blocks")
	}
	listHTML, err := pageform.RenderBlockListFragment(pageID, pwb.Blocks, saved.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to render list")
	}
	formHTML, err := pageform.RenderBlockDetailFragment(pageID, saved)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to render block")
	}

	out := sse(c)
	out.MarshalAndPatchSignals(map[string]any{"active_block": saved.ID})
	out.PatchElements(listHTML, datastar.WithSelectorID("block-list"), datastar.WithModeInner())
	out.PatchElements(formHTML, datastar.WithSelectorID("block-form-pane"), datastar.WithModeInner())
	out.ExecuteScript("window.bustBlocksCache()")
	out.ExecuteScript(toastScript("Block added", "success"))
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

	// active_block (seeded by blocksShell) tells us which block is open in the form pane.
	var raw map[string]any
	_ = readSignals(c, &raw) // optional; ignore missing/empty bodies
	activeID := signalInt64(raw, "active_block")

	if err := h.pageService.DeleteBlock(c.Request().Context(), blockID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete block")
	}

	// Re-render the list without the deleted block — partial SSE patch, no reload.
	pwb, err := h.pageService.GetPageWithBlocks(c.Request().Context(), pageID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load blocks")
	}
	listHTML, err := pageform.RenderBlockListFragment(pageID, pwb.Blocks, activeID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to render list")
	}

	out := sse(c)
	out.PatchElements(listHTML, datastar.WithSelectorID("block-list"), datastar.WithModeInner())

	// If we deleted the block currently open in the form pane, clear the pane.
	if activeID != 0 && activeID == blockID {
		emptyHTML, err := pageform.RenderBlockFormEmptyFragment()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to render")
		}
		out.MarshalAndPatchSignals(map[string]any{"active_block": 0})
		out.PatchElements(emptyHTML, datastar.WithSelectorID("block-form-pane"), datastar.WithModeInner())
	}
	out.ExecuteScript("window.bustBlocksCache()")
	out.ExecuteScript(toastScript("Block deleted", "success"))
	return nil
}

// blockSelect — Blocks-tab selection: render the chosen block's form in the form pane
// and highlight its card. No navigation — the list pane stays mounted and Swup's cache
// is untouched; both panes are patched in place. (Datastar @get target.)
func (h *AdminHandler) blockSelect(c echo.Context) error {
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
		if errors.Is(err, domain.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Page not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load page")
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

	listHTML, err := pageform.RenderBlockListFragment(pageID, pwb.Blocks, blockID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to render list")
	}
	formHTML, err := pageform.RenderBlockDetailFragment(pageID, block)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to render block")
	}

	out := sse(c)
	out.MarshalAndPatchSignals(map[string]any{"active_block": blockID})
	out.PatchElements(listHTML, datastar.WithSelectorID("block-list"), datastar.WithModeInner())
	out.PatchElements(formHTML, datastar.WithSelectorID("block-form-pane"), datastar.WithModeInner())
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
		out := sse(c)
		out.ExecuteScript(toastScript("Failed to load page", "error"))
		return nil
	}

	var currentBlock *domain.PageBlock
	for _, b := range pageWithBlocks.Blocks {
		if b.ID == blockID {
			currentBlock = b
			break
		}
	}
	if currentBlock == nil {
		out := sse(c)
		out.ExecuteScript(toastScript("Block not found", "error"))
		return nil
	}

	var raw map[string]any
	if err := readSignals(c, &raw); err != nil {
		out := sse(c)
		out.ExecuteScript(toastScript("Invalid request", "error"))
		return nil
	}

	// Build content map based on the block type from the database
	content := buildBlockContent(blockID, currentBlock.BlockType, raw)
	contentJSON, err := json.Marshal(content)
	if err != nil {
		out := sse(c)
		out.ExecuteScript(toastScript("Failed to encode block content", "error"))
		return nil
	}

	currentBlock.Content = contentJSON
	if _, err := h.pageService.SaveBlock(c.Request().Context(), currentBlock); err != nil {
		out := sse(c)
		out.ExecuteScript(toastScript("Failed to save block", "error"))
		return nil
	}

	// Stay in edit mode so the user can keep iterating. Toast confirms the save.
	sse(c).ExecuteScript(toastScript("Block saved", "success"))
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
		content["overlay_opacity"] = get("overlay_opacity")
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
		out := sse(c)
		out.ExecuteScript(toastScript("Failed to parse form data", "error"))
		return nil
	}

	signalName := c.FormValue("signal_name")
	if signalName == "" {
		out := sse(c)
		out.ExecuteScript(toastScript("Missing signal name", "error"))
		return nil
	}

	file, header, err := c.Request().FormFile("image_file")
	if err != nil {
		out := sse(c)
		out.ExecuteScript(toastScript("Please select a file to upload", "error"))
		return nil
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
		out := sse(c)
		out.ExecuteScript(toastScript("Invalid file type. Please upload a JPG, PNG, or WebP image.", "error"))
		return nil
	}

	b := make([]byte, 4)
	rand.Read(b)
	filename := fmt.Sprintf("%s/block_img_%x%s", strings.ToLower(strings.TrimSpace(config.Cfg.AppName)), b, ext)

	url, err := h.storageService.UploadFile(c.Request().Context(), file, filename)
	if err != nil {
		out := sse(c)
		out.ExecuteScript(toastScript("Failed to upload image. Please try again.", "error"))
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
	html, err := blockEditor.RenderItemsFragment(pageID, block, itemType)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to render items")
	}

	containerID := blockEditor.ItemsContainerID(blockID, itemType)
	signals := map[string]any{
		fmt.Sprintf("block_%d_editing", blockID): true,
	}
	for k, v := range blockEditor.NewItemSignals(block, itemType) {
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
	html, err := blockEditor.RenderItemsFragment(pageID, block, itemType)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to render items")
	}

	containerID := blockEditor.ItemsContainerID(blockID, itemType)
	signals := map[string]any{
		fmt.Sprintf("block_%d_editing", blockID): true,
	}
	for k, v := range blockEditor.AllItemSignals(block, itemType) {
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

// signalInt64 reads an int64-valued signal from a Datastar signals map.
func signalInt64(signals map[string]any, key string) int64 {
	switch n := signals[key].(type) {
	case float64:
		return int64(n)
	case int64:
		return n
	case string:
		v, _ := strconv.ParseInt(n, 10, 64)
		return v
	}
	return 0
}
