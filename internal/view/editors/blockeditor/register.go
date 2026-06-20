package blockEditor

import "github.com/a-h/templ"

type FieldKind int

const (
	FieldTextInput FieldKind = iota
	FieldNumber
	FieldTextArea
	FieldImage
	FieldDropdown
)

type Field struct {
	Name  string
	Label string
	Kind  FieldKind
}

type ItemField struct {
	Name  string
	Label string
	Kind  FieldKind
}

type ItemList struct {
	Signal string // "items" | "links" | "social_links"
	Label  string
	Fields []ItemField
}

type BlockType struct {
	Name    string
	Label   string
	Fields  []Field
	Items   []ItemList
	Summary string // Go template for block summary in page editor, e.g. "$%[1]stitle"
}

var registry = map[string]*BlockType{}

func Register(t *BlockType) {
	registry[t.Name] = t
}

func Lookup(name string) *BlockType {
	return registry[name]
}

func All() []*BlockType {
	out := make([]*BlockType, 0, len(registry))
	for _, t := range registry {
		out = append(out, t)
	}
	return out
}

// SignalAttrs emits data-signals attributes for a block's editable fields.
// Minimal implementation for Slice 0 (compile only); full registry-driven
// implementation lands in Slice 1.1.
func (t *BlockType) SignalAttrs(blockID int64, content map[string]any) templ.Attributes {
	return nil
}

// ContentFromSignals reads page-form POST signals and rebuilds block content.
// Minimal implementation for Slice 0 (compile only); full registry-driven
// implementation lands in Slice 1.1.
func (t *BlockType) ContentFromSignals(blockID int64, signals map[string]any) map[string]any {
	return nil
}
