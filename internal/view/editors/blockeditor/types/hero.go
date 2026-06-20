package blockType

import blockEditor "github.com/gracchi-stdio/castogo/internal/view/editors/blockeditor"

func init() {
	blockEditor.Register(&blockEditor.BlockType{
		Name:  "hero",
		Label: "Hero",
		Fields: []blockEditor.Field{
			{Name: "headline", Label: "Headline", Kind: blockEditor.FieldTextArea},
			{Name: "subheadline", Label: "Sub Headline", Kind: blockEditor.FieldTextArea},
			{Name: "cta_text", Label: "CTA Text", Kind: blockEditor.FieldTextInput},
			{Name: "cta_url", Label: "CTA URL", Kind: blockEditor.FieldTextInput},
			{Name: "background_image", Label: "Background Image", Kind: blockEditor.FieldImage},
			{Name: "overlay_opacity", Label: "Overlay Opacity", Kind: blockEditor.FieldNumber},
		},
	})
}
