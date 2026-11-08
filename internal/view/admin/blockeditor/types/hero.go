package blockType

import "github.com/gracchi-stdio/castogo/internal/view/admin/blockeditor"

func init() {
	blockeditor.Register(&blockeditor.BlockType{
		Name:  "hero",
		Label: "Hero",
		Fields: []blockeditor.Field{
			{Name: "headline", Label: "Headline", Kind: blockeditor.FieldTextArea},
			{Name: "subheadline", Label: "Sub Headline", Kind: blockeditor.FieldTextArea},
			{Name: "cta_text", Label: "CTA Text", Kind: blockeditor.FieldTextInput},
			{Name: "cta_url", Label: "CTA URL", Kind: blockeditor.FieldTextInput},
			{Name: "background_image", Label: "Background Image", Kind: blockeditor.FieldImage},
			{Name: "overlay_opacity", Label: "Overlay Opacity", Kind: blockeditor.FieldNumber},
		},
	})
}
