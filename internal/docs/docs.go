// Package docs loads embedded markdown documentation and renders it to HTML.
//
// The markdown source itself is embedded by the root-level package
// github.com/gracchi-stdio/castogo/docs (see docs/docs.go); this package wraps
// that embed.FS with typed Section/Doc values, frontmatter parsing, and goldmark
// rendering. It has no database dependency — it is a static content loader, not
// a domain service.
package docs

import (
	"errors"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"

	docfs "github.com/gracchi-stdio/castogo/docs"
)

// Section is a documentation section identifier (matches the URL segment).
type Section string

const (
	SectionDeveloper Section = "developer"
	SectionUser      Section = "user"
)

// DocMeta is a listing entry — everything needed to render the left nav, no body.
type DocMeta struct {
	Slug        string
	Title       string
	Description string
	Order       int
}

// Doc is a fully-loaded document with its rendered HTML body.
type Doc struct {
	DocMeta
	// HTMLBody is already-rendered HTML, safe to emit via templ.Raw.
	HTMLBody string
}

var (
	ErrUnknownSection = errors.New("docs: unknown section")
	ErrNotFound       = errors.New("docs: slug not found")
)

// sectionDir maps the URL section id to the embedded subdirectory. The route
// uses developer/user (URL-friendly); the on-disk dirs are dev/user.
func sectionDir(s Section) (string, error) {
	switch s {
	case SectionDeveloper:
		return "dev", nil
	case SectionUser:
		return "user", nil
	}
	return "", ErrUnknownSection
}

// List returns the section's docs ordered by frontmatter.order ASC, then title
// ASC. The result is always non-nil; an empty folder yields an empty slice.
func List(section Section) ([]DocMeta, error) {
	dir, err := sectionDir(section)
	if err != nil {
		return nil, err
	}
	entries, err := fs.ReadDir(docfs.FS, dir)
	if err != nil {
		// Unreachable in practice (the dir is embedded), but stay nil-safe.
		return []DocMeta{}, nil
	}

	out := []DocMeta{}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		slug := strings.TrimSuffix(e.Name(), ".md")
		raw, err := docfs.FS.ReadFile(path.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		_, fm, _ := renderAndParse(string(raw)) // best-effort frontmatter
		title := asString(fm["title"])
		if title == "" {
			title = slug // graceful fallback for docs with no frontmatter
		}
		out = append(out, DocMeta{
			Slug:        slug,
			Title:       title,
			Description: asString(fm["description"]),
			Order:       asInt(fm["order"]),
		})
	}

	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Order != out[j].Order {
			return out[i].Order < out[j].Order
		}
		return out[i].Title < out[j].Title
	})
	return out, nil
}

// Load returns a single rendered document. It returns ErrNotFound for an
// unknown slug and ErrUnknownSection for a bad section.
func Load(section Section, slug string) (*Doc, error) {
	dir, err := sectionDir(section)
	if err != nil {
		return nil, err
	}
	if !isValidSlug(slug) {
		return nil, ErrNotFound
	}
	raw, err := docfs.FS.ReadFile(path.Join(dir, slug+".md"))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("docs: read %s/%s: %w", dir, slug, err)
	}

	body, fm, _ := renderAndParse(string(raw))
	title := asString(fm["title"])
	if title == "" {
		title = slug
	}
	return &Doc{
		DocMeta: DocMeta{
			Slug:        slug,
			Title:       title,
			Description: asString(fm["description"]),
			Order:       asInt(fm["order"]),
		},
		HTMLBody: body,
	}, nil
}

// isValidSlug allows only [a-z0-9-]. This also rejects "/" and "..", which is
// defense-in-depth — embed.FS is already bounded to the embedded root.
func isValidSlug(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		ok := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-'
		if !ok {
			return false
		}
	}
	return true
}
