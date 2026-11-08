package blockType

import "github.com/gracchi-stdio/castogo/internal/view/admin/blockeditor"

func init() {
	blockeditor.Register(&blockeditor.BlockType{
		Name:  "features",
		Label: "Features",
		Fields: []blockeditor.Field{
			{Name: "section_title", Label: "Section Title", Kind: blockeditor.FieldTextInput},
			{Name: "section_description", Label: "Section Description", Kind: blockeditor.FieldTextArea},
		},
		Items: []blockeditor.ItemList{{
			Signal: "items",
			Label:  "Features",
			Fields: []blockeditor.ItemField{
				{Name: "icon", Label: "Icon URL", Kind: blockeditor.FieldTextInput},
				{Name: "title", Label: "Title", Kind: blockeditor.FieldTextInput},
				{Name: "description", Label: "Description", Kind: blockeditor.FieldTextArea},
			},
		}},
		Summary: "$%[1]section_title !== '' ? $%[1]section_title : 'Empty block'",
	})
}
