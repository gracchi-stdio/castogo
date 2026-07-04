package pageform

import "github.com/gracchi-stdio/castogo/internal/domain"

type Signals struct {
	Title       string `json:"title"`
	Slug        string `json:"slug"`
	ParentID    int64  `json:"parent_id"`
	IsPublished bool   `json:"is_published"`
	ShowInNav   bool   `json:"show_in_nav"`

	TitleError string `json:"title_error,omitempty"`
	SlugError  string `json:"slug_error,omitempty"`
}

func NewSignals(page *domain.Page) Signals {
	if page == nil {
		return Signals{}
	}
	s := Signals{
		Title:       page.Title,
		Slug:        page.Slug,
		IsPublished: page.IsPublished,
		ShowInNav:   page.ShowInNav,
	}
	if page.ParentID != nil {
		s.ParentID = *page.ParentID
	}
	return s
}
