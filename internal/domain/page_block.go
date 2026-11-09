package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// RawMessage is arbitrary JSON for a block's content column. It exists for
// the same reason AudioMetadata/PageMetadata do: modernc hands TEXT as
// string, libsql as []byte -- json.RawMessage implements neither Scan nor
// Value, so reading blocks would fail on local file DBs without this type.
type RawMessage []byte

// Scan implements sql.Scanner. See AudioMetadata.Scan.
func (m *RawMessage) Scan(src any) error {
	switch v := src.(type) {
	case nil:
		*m = nil
		return nil
	case string:
		*m = RawMessage(v)
		return nil
	case []byte:
		*m = append((*m)[:0], v...)
		return nil
	default:
		return fmt.Errorf("block content: unsupported scan type %T", src)
	}
}

func (m RawMessage) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return string(m), nil
}

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
