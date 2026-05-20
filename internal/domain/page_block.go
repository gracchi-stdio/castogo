package domain

import (
	"encoding/json"
	"time"
)

type PageMetadata struct {
	MetaTitle       string `json:"meta_title"`
	MetaDescription string `json:"meta_description"`
	MetaKeywords    string `json:"meta_keywords"`
	OGImage         string `json:"og_image"`
}

type Page struct {
	ID          int64        `json:"id"`
	Title       string       `json:"title"`
	Layout      string       `json:"layout"`
	Slug        string       `json:"slug"`
	IsPublished bool         `json:"is_published"`
	ParentID    *int64       `json:"parent_id"`
	Path        string       `json:"path"`
	Metadata    PageMetadata `json:"metadata"`
	SortOrder   int          `json:"sort_order"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

type PageBlock struct {
	ID        int64           `json:"id"`
	PageID    int64           `json:"page_id"`
	BlockType string          `json:"block_type"`
	Content   json.RawMessage `json:"content"`
	SortOrder int             `json:"sort_order"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}
