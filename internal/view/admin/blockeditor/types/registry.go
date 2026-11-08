package blockType

import "github.com/gracchi-stdio/castogo/internal/view/admin/blockeditor"

func init() {
	blockeditor.Register(&blockeditor.BlockType{
		Name:  "cta",
		Label: "Call to Action",
		Fields: []blockeditor.Field{
			{Name: "headline", Label: "Headline", Kind: blockeditor.FieldTextInput},
			{Name: "description", Label: "Description", Kind: blockeditor.FieldTextArea},
			{Name: "button_text", Label: "Button Text", Kind: blockeditor.FieldTextInput},
			{Name: "button_url", Label: "Button URL", Kind: blockeditor.FieldTextInput},
		},
		Summary: "$%[1]headline !== '' ? $%[1]headline : ($%[1]description !== '' ? $%[1]description : 'Empty block')",
	})

	blockeditor.Register(&blockeditor.BlockType{
		Name:  "episodes_showcase",
		Label: "Episodes",
		Fields: []blockeditor.Field{
			{Name: "section_title", Label: "Section Title", Kind: blockeditor.FieldTextInput},
			{Name: "section_description", Label: "Section Description", Kind: blockeditor.FieldTextArea},
			{Name: "max_episodes", Label: "Max Episodes", Kind: blockeditor.FieldNumber},
			{Name: "display_mode", Label: "Display Mode", Kind: blockeditor.FieldDropdown},
		},
		Summary: "$%[1]section_title !== '' ? $%[1]section_title : 'Empty block'",
	})

	blockeditor.Register(&blockeditor.BlockType{
		Name:  "testimonials",
		Label: "Testimonials",
		Fields: []blockeditor.Field{
			{Name: "section_title", Label: "Section Title", Kind: blockeditor.FieldTextInput},
			{Name: "section_description", Label: "Section Description", Kind: blockeditor.FieldTextArea},
		},
		Items: []blockeditor.ItemList{{
			Signal: "items",
			Label:  "Testimonials",
			Fields: []blockeditor.ItemField{
				{Name: "quote", Label: "Quote", Kind: blockeditor.FieldTextArea},
				{Name: "author", Label: "Author", Kind: blockeditor.FieldTextInput},
				{Name: "role", Label: "Role", Kind: blockeditor.FieldTextInput},
				{Name: "avatar_url", Label: "Avatar URL", Kind: blockeditor.FieldTextInput},
			},
		}},
		Summary: "$%[1]section_title !== '' ? $%[1]section_title : 'Empty block'",
	})

	blockeditor.Register(&blockeditor.BlockType{
		Name:  "footer",
		Label: "Footer",
		Fields: []blockeditor.Field{
			{Name: "copyright", Label: "Copyright", Kind: blockeditor.FieldTextInput},
			{Name: "text", Label: "Text", Kind: blockeditor.FieldTextArea},
		},
		Items: []blockeditor.ItemList{
			{
				Signal: "links",
				Label:  "Navigation Links",
				Fields: []blockeditor.ItemField{
					{Name: "label", Label: "Label", Kind: blockeditor.FieldTextInput},
					{Name: "url", Label: "URL", Kind: blockeditor.FieldTextInput},
				},
			},
			{
				Signal: "social_links",
				Label:  "Social Links",
				Fields: []blockeditor.ItemField{
					{Name: "platform", Label: "Platform", Kind: blockeditor.FieldTextInput},
					{Name: "url", Label: "URL", Kind: blockeditor.FieldTextInput},
				},
			},
		},
		Summary: "$%[1]copyright !== '' ? $%[1]copyright : 'Empty block'",
	})

	blockeditor.Register(&blockeditor.BlockType{
		Name:  "prose",
		Label: "Rich Text",
		Fields: []blockeditor.Field{
			{Name: "body", Label: "Content", Kind: blockeditor.FieldTextArea},
		},
		Summary: "$%[1]body !== '' ? $%[1]body.substring(0, 50) + ($%[1]body.length > 50 ? '...' : '') : 'Empty block'",
	})
}
