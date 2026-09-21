package blockEditor

import (
	"bytes"
	"context"
	"fmt"

	"github.com/gracchi-stdio/castogo/internal/domain"
)

// SignalsForBlock returns every signal the block's editor edits (fields and
// items) as plain values, for SSE patch-signals. The same keys are seeded by
// blockSignalAttrs in the pane HTML; pushing them explicitly too makes saving
// independent of the browser re-applying data-signals attributes on morphed
// content (Datastar's global store is what fetch actions serialize).
func SignalsForBlock(block *domain.PageBlock) map[string]any {
	content := parseBlockContent(block)
	pfx := blockEditPrefix(block, "")

	signals := map[string]any{
		pfx + "editing":    false,
		pfx + "block_type": block.BlockType,
	}
	set := func(name, key string) { signals[pfx+name] = strVal(content[key]) }

	switch block.BlockType {
	case "hero":
		set("headline", "headline")
		set("subheadline", "subheadline")
		set("cta_text", "cta_text")
		set("cta_url", "cta_url")
		set("background_image", "background_image")
		set("overlay_opacity", "overlay_opacity")
	case "cta":
		desc := content["description"]
		if desc == nil {
			desc = content["subheadline"]
		}
		signals[pfx+"description"] = strVal(desc)
		set("headline", "headline")
		set("button_text", "button_text")
		set("button_url", "button_url")
	case "features":
		set("section_title", "section_title")
		set("section_description", "section_description")
		for i, item := range toItems(content["items"]) {
			signals[fmt.Sprintf("%sitem_%d_icon", pfx, i)] = strVal(item["icon"])
			signals[fmt.Sprintf("%sitem_%d_title", pfx, i)] = strVal(item["title"])
			signals[fmt.Sprintf("%sitem_%d_description", pfx, i)] = strVal(item["description"])
		}
	case "episodes_showcase":
		set("section_title", "section_title")
		set("section_description", "section_description")
		signals[pfx+"max_episodes"] = numVal(content["max_episodes"])
		set("display_mode", "display_mode")
	case "testimonials":
		set("section_title", "section_title")
		set("section_description", "section_description")
		for i, item := range toItems(content["items"]) {
			signals[fmt.Sprintf("%sitem_%d_quote", pfx, i)] = strVal(item["quote"])
			signals[fmt.Sprintf("%sitem_%d_author", pfx, i)] = strVal(item["author"])
			signals[fmt.Sprintf("%sitem_%d_role", pfx, i)] = strVal(item["role"])
			signals[fmt.Sprintf("%sitem_%d_avatar_url", pfx, i)] = strVal(item["avatar_url"])
		}
	case "footer":
		set("copyright", "copyright")
		set("text", "text")
		for i, link := range toItems(content["links"]) {
			signals[fmt.Sprintf("%slink_%d_label", pfx, i)] = strVal(link["label"])
			signals[fmt.Sprintf("%slink_%d_url", pfx, i)] = strVal(link["url"])
		}
		for i, link := range toItems(content["social_links"]) {
			signals[fmt.Sprintf("%ssocial_%d_platform", pfx, i)] = strVal(link["platform"])
			signals[fmt.Sprintf("%ssocial_%d_url", pfx, i)] = strVal(link["url"])
		}
	case "prose":
		set("body", "body")
	}

	return signals
}

func numVal(v any) int {
	if f, ok := v.(float64); ok {
		return int(f)
	}
	return 0
}

// ItemsContainerID returns the DOM element ID for a block's items list.
func ItemsContainerID(blockID int64, listType string) string {
	switch listType {
	case "feature":
		return fmt.Sprintf("block-%d-features", blockID)
	case "testimonial":
		return fmt.Sprintf("block-%d-testimonials", blockID)
	case "link":
		return fmt.Sprintf("block-%d-footer-links", blockID)
	case "social":
		return fmt.Sprintf("block-%d-footer-social", blockID)
	}
	return ""
}

// RenderItemsFragment renders the items for a block's list as an HTML string for SSE patching.
func RenderItemsFragment(pageID int64, block *domain.PageBlock, listType string) (string, error) {
	var buf bytes.Buffer
	ctx := context.Background()
	content := parseBlockContent(block)
	pfx := blockEditPrefix(block, "")

	switch listType {
	case "feature":
		for i, item := range toItems(content["items"]) {
			if err := featureItemEditor(pageID, block.ID, pfx, i, item).Render(ctx, &buf); err != nil {
				return "", err
			}
		}
	case "testimonial":
		for i, item := range toItems(content["items"]) {
			if err := testimonialItemEditor(pageID, block.ID, pfx, i, item).Render(ctx, &buf); err != nil {
				return "", err
			}
		}
	case "link":
		for i, link := range toItems(content["links"]) {
			if err := footerLinkEditor(pageID, block.ID, pfx, i, link).Render(ctx, &buf); err != nil {
				return "", err
			}
		}
	case "social":
		for i, link := range toItems(content["social_links"]) {
			if err := footerSocialEditor(pageID, block.ID, pfx, i, link).Render(ctx, &buf); err != nil {
				return "", err
			}
		}
	}
	return buf.String(), nil
}

// NewItemSignals returns empty signals for the newest (last) item in a block's list.
func NewItemSignals(block *domain.PageBlock, listType string) map[string]string {
	content := parseBlockContent(block)
	pfx := blockEditPrefix(block, "")

	switch listType {
	case "feature":
		items := toItems(content["items"])
		i := len(items) - 1
		return map[string]string{
			pfx + fmt.Sprintf("item_%d_icon", i):        "",
			pfx + fmt.Sprintf("item_%d_title", i):       "",
			pfx + fmt.Sprintf("item_%d_description", i): "",
		}
	case "testimonial":
		items := toItems(content["items"])
		i := len(items) - 1
		return map[string]string{
			pfx + fmt.Sprintf("item_%d_quote", i):      "",
			pfx + fmt.Sprintf("item_%d_author", i):     "",
			pfx + fmt.Sprintf("item_%d_role", i):       "",
			pfx + fmt.Sprintf("item_%d_avatar_url", i): "",
		}
	case "link":
		links := toItems(content["links"])
		i := len(links) - 1
		return map[string]string{
			pfx + fmt.Sprintf("link_%d_label", i): "",
			pfx + fmt.Sprintf("link_%d_url", i):   "",
		}
	case "social":
		socials := toItems(content["social_links"])
		i := len(socials) - 1
		return map[string]string{
			pfx + fmt.Sprintf("social_%d_platform", i): "",
			pfx + fmt.Sprintf("social_%d_url", i):      "",
		}
	}
	return nil
}

// AllItemSignals returns all item signals from the block's DB content (for re-initialization after removal).
func AllItemSignals(block *domain.PageBlock, listType string) map[string]string {
	content := parseBlockContent(block)
	pfx := blockEditPrefix(block, "")
	signals := map[string]string{}

	switch listType {
	case "feature":
		for i, item := range toItems(content["items"]) {
			signals[pfx+fmt.Sprintf("item_%d_icon", i)] = strVal(item["icon"])
			signals[pfx+fmt.Sprintf("item_%d_title", i)] = strVal(item["title"])
			signals[pfx+fmt.Sprintf("item_%d_description", i)] = strVal(item["description"])
		}
	case "testimonial":
		for i, item := range toItems(content["items"]) {
			signals[pfx+fmt.Sprintf("item_%d_quote", i)] = strVal(item["quote"])
			signals[pfx+fmt.Sprintf("item_%d_author", i)] = strVal(item["author"])
			signals[pfx+fmt.Sprintf("item_%d_role", i)] = strVal(item["role"])
			signals[pfx+fmt.Sprintf("item_%d_avatar_url", i)] = strVal(item["avatar_url"])
		}
	case "link":
		for i, link := range toItems(content["links"]) {
			signals[pfx+fmt.Sprintf("link_%d_label", i)] = strVal(link["label"])
			signals[pfx+fmt.Sprintf("link_%d_url", i)] = strVal(link["url"])
		}
	case "social":
		for i, link := range toItems(content["social_links"]) {
			signals[pfx+fmt.Sprintf("social_%d_platform", i)] = strVal(link["platform"])
			signals[pfx+fmt.Sprintf("social_%d_url", i)] = strVal(link["url"])
		}
	}
	return signals
}

func strVal(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
