package blockEditor

import (
	"bytes"
	"context"
	"fmt"

	"github.com/gracchi-stdio/castogo/internal/domain"
)

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
