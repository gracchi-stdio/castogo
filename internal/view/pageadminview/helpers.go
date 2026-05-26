package pageadminview

import (
	"fmt"
	"strconv"

	"github.com/gracchi-stdio/castogo/internal/domain"
	selectcomponent "github.com/gracchi-stdio/castogo/internal/view/components/select"
)

type pageFormData struct {
	Title       string
	Slug        string
	Layout      string
	IsPublished bool
	ParentID    string
	IsHome      bool
	IsEdit      bool
	PageID      int64
}

func newPageFormData(page *domain.Page) pageFormData {
	d := pageFormData{
		Layout:   "landing",
		ParentID: "0",
	}
	if page != nil {
		d.Title = page.Title
		d.Slug = page.Slug
		d.Layout = page.Layout
		d.IsPublished = page.IsPublished
		d.IsHome = page.Path == "home"
		d.PageID = page.ID
		d.IsEdit = true
		if page.ParentID != nil {
			d.ParentID = strconv.FormatInt(*page.ParentID, 10)
		}
	}
	return d
}

func (d pageFormData) TitleSignal() string {
	return fmt.Sprintf("'%s'", jsEscape(d.Title))
}

func (d pageFormData) SlugSignal() string {
	return fmt.Sprintf("'%s'", jsEscape(d.Slug))
}

func (d pageFormData) LayoutSignal() string {
	return fmt.Sprintf("'%s'", d.Layout)
}

func (d pageFormData) ParentIDSignal() string {
	return d.ParentID
}

func (d pageFormData) PublishedSignal() string {
	return fmt.Sprintf("%t", d.IsPublished)
}

func parentPageOptions(parentPages []*domain.Page) []selectcomponent.SelectOptionArgs {
	opts := make([]selectcomponent.SelectOptionArgs, 0, len(parentPages))
	for _, p := range parentPages {
		if p.ParentID == nil {
			opts = append(opts, selectcomponent.SelectOptionArgs{
				Value: fmt.Sprintf("%d", p.ID),
				Label: p.Title + " (/" + p.Path + ")",
			})
		}
	}
	return opts
}

func jsEscape(s string) string {
	out := make([]byte, 0, len(s))
	for _, c := range s {
		if c == '\'' {
			out = append(out, '\\', '\'')
		} else if c == '\\' {
			out = append(out, '\\', '\\')
		} else {
			out = append(out, string(c)...)
		}
	}
	return string(out)
}
