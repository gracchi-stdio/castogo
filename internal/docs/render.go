package docs

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

// md renders documentation markdown with:
//   - meta.Meta: parses a leading YAML frontmatter block (stripped from output)
//   - extension.GFM: tables, strikethrough, autolinks, task lists
//   - AutoHeadingID: adds id attributes to headings (anchors/TOC later)
//   - html.WithUnsafe: passes raw HTML through (docs are author-controlled,
//     baked into the binary — same trust model as pageview ProseBlock)
var md = goldmark.New(
	goldmark.WithExtensions(meta.Meta, extension.GFM),
	goldmark.WithParserOptions(parser.WithAutoHeadingID()),
	goldmark.WithRendererOptions(html.WithUnsafe()),
)

// renderAndParse runs a single Convert pass over src: the rendered HTML body
// (frontmatter removed) goes to the buffer, and the parsed frontmatter is
// returned as a loosely-typed map. On error it returns src unchanged so a
// malformed doc still renders something readable.
func renderAndParse(src string) (body string, frontmatter map[string]any, err error) {
	var buf bytes.Buffer
	ctx := parser.NewContext()
	if e := md.Convert([]byte(src), &buf, parser.WithContext(ctx)); e != nil {
		return src, nil, e
	}
	return buf.String(), meta.Get(ctx), nil
}

// asString / asInt coerce frontmatter scalars (yaml.v2 returns string / int).
func asString(v any) string {
	s, _ := v.(string)
	return s
}

func asInt(v any) int {
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	}
	return 0
}
