package docs

import (
	"strings"
	"testing"
)

func TestListDeveloper(t *testing.T) {
	items, err := List(SectionDeveloper)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("want 1 item, got %d", len(items))
	}
	got := items[0]
	if got.Slug != "getting-started" {
		t.Errorf("Slug = %q, want getting-started", got.Slug)
	}
	if got.Title != "Getting Started" {
		t.Errorf("Title = %q, want \"Getting Started\"", got.Title)
	}
	if got.Order != 1 {
		t.Errorf("Order = %d, want 1", got.Order)
	}
}

func TestLoadRendersBody(t *testing.T) {
	doc, err := Load(SectionDeveloper, "getting-started")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if doc.Title != "Getting Started" {
		t.Errorf("Title = %q", doc.Title)
	}
	// GFM table should render as <table>, heading as <h1.
	for _, want := range []string{"<table>", "<h1", "Common Commands"} {
		if !strings.Contains(doc.HTMLBody, want) {
			t.Errorf("HTMLBody missing %q\n%s", want, doc.HTMLBody)
		}
	}
}

func TestLoadUnknownSlug(t *testing.T) {
	if _, err := Load(SectionDeveloper, "does-not-exist"); err != ErrNotFound {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

func TestLoadRejectsTraversal(t *testing.T) {
	for _, bad := range []string{"..", "../something", "a/b", "A", ""} {
		if _, err := Load(SectionDeveloper, bad); err != ErrNotFound {
			t.Errorf("Load(%q): want ErrNotFound, got %v", bad, err)
		}
	}
}
