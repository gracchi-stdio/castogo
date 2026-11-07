package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type PageMetadata struct {
	MetaTitle       string `json:"meta_title"`
	MetaDescription string `json:"meta_description"`
	MetaKeywords    string `json:"meta_keywords"`
	OGImage         string `json:"og_image"`
}

// Scan implements sql.Scanner so sqlc overrides can map JSON columns to this
// type on both engines. See AudioMetadata.Scan.
func (m *PageMetadata) Scan(src any) error {
	switch v := src.(type) {
	case nil:
		*m = PageMetadata{}
		return nil
	case string:
		return json.Unmarshal([]byte(v), m)
	case []byte:
		return json.Unmarshal(v, m)
	default:
		return fmt.Errorf("metadata: unsupported scan type %T", src)
	}
}

// Value implements driver.Valuer. See AudioMetadata.Value.
func (m PageMetadata) Value() (driver.Value, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

type Page struct {
	ID          int64        `json:"id"`
	Title       string       `json:"title"`
	Layout      string       `json:"layout"`
	Slug        string       `json:"slug"`
	IsPublished bool         `json:"is_published"`
	NavOrder    *int32       `json:"nav_order"`
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
