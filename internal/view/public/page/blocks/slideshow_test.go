package blocks

import (
	"strings"
	"testing"
)

func TestParseSlideTime(t *testing.T) {
	cases := []struct {
		in   string
		want int
		ok   bool
	}{
		{"0", 0, true},         // bare seconds
		{"45", 45, true},       //
		{"00:1", 1, true},      // short form M:S
		{"03:12", 192, true},   // MM:SS
		{"1:02:03", 3723, true}, // H:MM:SS
		{"", 0, false},          //
		{"ab", 0, false},        // non-numeric
		{"1:2:3:4", 0, false},   // too many parts
		{"-1", 0, false},        // negative
		{"ab", 0, false},       // non-numeric
		{"1:2:3:4", 0, false},  // too many parts
		{"-1", 0, false},       // negative
	}
	for _, c := range cases {
		got, ok := parseSlideTime(c.in)
		if ok != c.ok || got != c.want {
			t.Errorf("parseSlideTime(%q) = (%d, %v), want (%d, %v)", c.in, got, ok, c.want, c.ok)
		}
	}
}

func TestParseSlides(t *testing.T) {
	// The exact deck shape from manual testing: leading separator, lenient
	// time directives, an empty-bodied timed slide, and an untimed slide.
	src := "---\ntime: 0\n---\n\n### Begin\n\n---\ntime: 00:1\n---\n\n![Paul Gauguin](https://example.com/gauguin.jpg)\n"

	slides := parseSlides(src)
	if len(slides) != 2 {
		t.Fatalf("parseSlides returned %d slides, want 2 (empty-bodied slide dropped): %+v", len(slides), slides)
	}

	// Slide 0: "### Begin" — the frontmatter "time: 0" attaches to it.
	if slides[0].timeSec == nil || *slides[0].timeSec != 0 {
		t.Errorf("slide 0 timeSec = %v, want 0 (carried from frontmatter block)", slides[0].timeSec)
	}
	if !strings.Contains(string(slides[0].html), "<h3>Begin</h3>") {
		t.Errorf("slide 0 html = %q, want rendered <h3>Begin</h3>", slides[0].html)
	}

	// Slide 1: image slide, revealed at 1s ("00:1").
	if slides[1].timeSec == nil || *slides[1].timeSec != 1 {
		t.Errorf("slide 1 timeSec = %v, want 1", slides[1].timeSec)
	}
	if !strings.Contains(string(slides[1].html), `<img src="https://example.com/gauguin.jpg"`) {
		t.Errorf("slide 1 html = %q, want the markdown image rendered", slides[1].html)
	}
}

func TestParseSlidesEmptyAndBlank(t *testing.T) {
	if got := parseSlides(""); len(got) != 0 {
		t.Errorf("parseSlides(\"\") = %d slides, want 0", len(got))
	}
	if got := parseSlides("---\n\n---\n\n---\n"); len(got) != 0 {
		t.Errorf("all-blank deck = %d slides, want 0", len(got))
	}

	// No separators at all — the whole source is one slide.
	slides := parseSlides("just text")
	if len(slides) != 1 || !strings.Contains(string(slides[0].html), "just text") {
		t.Errorf("single-slide deck mis-parsed: %+v", slides)
	}
}

func TestParseSlidesFitDirective(t *testing.T) {
	slides := parseSlides("{.object-cover}\n\n![Artwork](https://example.com/art.jpg)\n---\n{.object-fill}\n\nText")
	if len(slides) != 2 {
		t.Fatalf("parseSlides returned %d slides, want 2", len(slides))
	}
	if slides[0].fitClass != "slide-fit-cover" {
		t.Errorf("slide 0 fitClass = %q, want slide-fit-cover", slides[0].fitClass)
	}
	if strings.Contains(string(slides[0].html), "object-cover") {
		t.Error("fit directive leaked into rendered markdown")
	}
	if slides[1].fitClass != "slide-fit-fill" {
		t.Errorf("slide 1 fitClass = %q, want slide-fit-fill", slides[1].fitClass)
	}
}

func TestRenderSlideMarkdownComarkImageAttributes(t *testing.T) {
	html := string(renderSlideMarkdown("![Artwork](https://example.com/art.jpg){.object-contain}"))
	if strings.Contains(html, "{.object-cover}") {
		t.Fatal("Comark image attributes leaked into rendered HTML")
	}
	if !strings.Contains(html, `<img class="slide-fit-contain"`) {
		t.Fatalf("rendered image = %q, want slide-fit-contain class", html)
	}
	if !strings.Contains(html, `<figure class="slide-image"><img class="slide-fit-contain"`) {
		t.Fatalf("rendered image = %q, want semantic media figure", html)
	}

	paragraph := string(renderSlideMarkdown("A caption with ![Artwork](https://example.com/art.jpg)"))
	if strings.Contains(paragraph, `<figure class="slide-image">`) {
		t.Fatalf("mixed image paragraph = %q, should remain a paragraph", paragraph)
	}

	html = string(renderSlideMarkdown(`![Artwork](https://example.com/art.jpg){.object-cover}`))
	if !strings.Contains(html, `<img class="slide-fit-cover"`) {
		t.Fatalf("cover image = %q, want slide-fit-cover class", html)
	}

	html = string(renderSlideMarkdown(`![Artwork](https://example.com/art.jpg){class="object-contain"}`))
	if !strings.Contains(html, `<img class="slide-fit-contain"`) {
		t.Fatalf("explicit class image = %q, want slide-fit-contain class", html)
	}
}
