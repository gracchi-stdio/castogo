package handler

import "testing"

func TestBuildBlockContentFeaturesUsesSectionTitle(t *testing.T) {
	signals := map[string]any{
		"block_12_section_title":       "Why listeners stay",
		"block_12_section_description": "A short explanation",
		"block_12_item_0_title":        "Simple publishing",
		"block_12_item_0_description":  "Ship episodes quickly",
	}

	content := buildBlockContent(12, "features", signals)
	if got := content["section_title"]; got != "Why listeners stay" {
		t.Fatalf("section_title = %v, want %q", got, "Why listeners stay")
	}

	items, ok := content["items"].([]map[string]any)
	if !ok || len(items) != 1 {
		t.Fatalf("items = %#v, want one feature", content["items"])
	}
	if got := items[0]["title"]; got != "Simple publishing" {
		t.Fatalf("item title = %v, want %q", got, "Simple publishing")
	}
}

func TestValidateBlockContentRequiresPrimaryFields(t *testing.T) {
	tests := []struct {
		blockType string
		content   map[string]any
	}{
		{blockType: "hero", content: map[string]any{}},
		{blockType: "cta", content: map[string]any{}},
		{blockType: "features", content: map[string]any{}},
		{blockType: "episodes_showcase", content: map[string]any{}},
		{blockType: "testimonials", content: map[string]any{}},
		{blockType: "footer", content: map[string]any{}},
		{blockType: "prose", content: map[string]any{}},
	}

	for _, test := range tests {
		t.Run(test.blockType, func(t *testing.T) {
			if err := validateBlockContent(test.blockType, test.content); err == nil {
				t.Fatalf("validateBlockContent(%q) accepted empty content", test.blockType)
			}
		})
	}
}

func TestValidateBlockContentRequiresFeatureItemTitle(t *testing.T) {
	content := map[string]any{
		"section_title": "Benefits",
		"items": []any{
			map[string]any{"description": "Missing title"},
		},
	}

	if err := validateBlockContent("features", content); err == nil {
		t.Fatal("validateBlockContent accepted a feature without a title")
	}
}
