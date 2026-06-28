package handler

type pageUpdateInput struct {
	Title       string  `json:"title" validate:"required"`
	Slug        string  `json:"slug"`
	Layout      string  `json:"page_layout"`
	ParentID    float64 `json:"parent_id"`
	IsPublished struct {
		Checked bool `json:"checked"`
	} `json:"is_published"`
}

type pageCreateInput struct {
	Title     string  `json:"title" validate:"required"`
	Slug      string  `json:"slug"`
	Layout    string  `json:"page_layout"`
	ParentID  float64 `json:"parent_id"`
	ShowInNav struct {
		Checked bool `json:"checked"`
	} `json:"show_in_nav"`
	IsPublished struct {
		Checked bool `json:"checked"`
	} `json:"is_published"`
}
