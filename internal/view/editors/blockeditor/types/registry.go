package blockType

import blockEditor "github.com/gracchi-stdio/castogo/internal/view/editors/blockeditor"

func init() {
	blockEditor.Register(&blockEditor.BlockType{
		Name:  "cta",
		Label: "Call to Action",
		Fields: []blockEditor.Field{
			{Name: "headline", Label: "Headline", Kind: blockEditor.FieldTextInput},
			{Name: "description", Label: "Description", Kind: blockEditor.FieldTextArea},
			{Name: "button_text", Label: "Button Text", Kind: blockEditor.FieldTextInput},
			{Name: "button_url", Label: "Button URL", Kind: blockEditor.FieldTextInput},
		},
		Summary: "$%[1]headline !== '' ? $%[1]headline : ($%[1]description !== '' ? $%[1]description : 'Empty block')",
	})

	blockEditor.Register(&blockEditor.BlockType{
		Name:  "episodes_showcase",
		Label: "Episodes",
		Fields: []blockEditor.Field{
			{Name: "section_title", Label: "Section Title", Kind: blockEditor.FieldTextInput},
			{Name: "section_description", Label: "Section Description", Kind: blockEditor.FieldTextArea},
			{Name: "max_episodes", Label: "Max Episodes", Kind: blockEditor.FieldNumber},
			{Name: "display_mode", Label: "Display Mode", Kind: blockEditor.FieldDropdown},
		},
		Summary: "$%[1]section_title !== '' ? $%[1]section_title : 'Empty block'",
	})

	blockEditor.Register(&blockEditor.BlockType{
		Name:  "testimonials",
		Label: "Testimonials",
		Fields: []blockEditor.Field{
			{Name: "section_title", Label: "Section Title", Kind: blockEditor.FieldTextInput},
			{Name: "section_description", Label: "Section Description", Kind: blockEditor.FieldTextArea},
		},
		Items: []blockEditor.ItemList{{
			Signal: "items",
			Label:  "Testimonials",
			Fields: []blockEditor.ItemField{
				{Name: "quote", Label: "Quote", Kind: blockEditor.FieldTextArea},
				{Name: "author", Label: "Author", Kind: blockEditor.FieldTextInput},
				{Name: "role", Label: "Role", Kind: blockEditor.FieldTextInput},
				{Name: "avatar_url", Label: "Avatar URL", Kind: blockEditor.FieldTextInput},
			},
		}},
		Summary: "$%[1]section_title !== '' ? $%[1]section_title : 'Empty block'",
	})

	blockEditor.Register(&blockEditor.BlockType{
		Name:  "footer",
		Label: "Footer",
		Fields: []blockEditor.Field{
			{Name: "copyright", Label: "Copyright", Kind: blockEditor.FieldTextInput},
			{Name: "text", Label: "Text", Kind: blockEditor.FieldTextArea},
		},
		Items: []blockEditor.ItemList{
			{
				Signal: "links",
				Label:  "Navigation Links",
				Fields: []blockEditor.ItemField{
					{Name: "label", Label: "Label", Kind: blockEditor.FieldTextInput},
					{Name: "url", Label: "URL", Kind: blockEditor.FieldTextInput},
				},
			},
			{
				Signal: "social_links",
				Label:  "Social Links",
				Fields: []blockEditor.ItemField{
					{Name: "platform", Label: "Platform", Kind: blockEditor.FieldTextInput},
					{Name: "url", Label: "URL", Kind: blockEditor.FieldTextInput},
				},
			},
		},
		Summary: "$%[1]copyright !== '' ? $%[1]copyright : 'Empty block'",
	})

	blockEditor.Register(&blockEditor.BlockType{
		Name:  "prose",
		Label: "Rich Text",
		Fields: []blockEditor.Field{
			{Name: "body", Label: "Content", Kind: blockEditor.FieldTextArea},
		},
		Summary: "$%[1]body !== '' ? $%[1]body.substring(0, 50) + ($%[1]body.length > 50 ? '...' : '') : 'Empty block'",
	})
}
