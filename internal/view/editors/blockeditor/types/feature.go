package blockType

import blockEditor "github.com/gracchi-stdio/castogo/internal/view/editors/blockeditor"

func init() {
	blockEditor.Register(&blockEditor.BlockType{
		Name:  "feature",
		Label: "Feature",
		Fields: []blockEditor.Field{
			{Name: "title", Label: "Title", Kind: blockEditor.FieldTextInput},
			{Name: "description", Label: "Description", Kind: blockEditor.FieldTextArea},
		},
		Items: []blockEditor.ItemList{{
			Signal: "items",
			Label:  "Features",
			Fields: []blockEditor.ItemField{
				{Name: "icon", Label: "Icon URL", Kind: blockEditor.FieldTextInput},
				{Name: "title", Label: "Title", Kind: blockEditor.FieldTextInput},
				{Name: "description", Label: "Description", Kind: blockEditor.FieldTextArea},
			},
		}},
		Summary: "$%[1]section_title !== '' ? $%[1]section_title : 'Empty block'",
	})
}
