# Settings Save & iTunes Categories Plan

## Overview

Two tasks:
1. Implement `settingsSave` handler — make the Save Settings button work
2. Add iTunes category/subcategory dropdowns with cascade (Approach B: SSE round-trip)

Also discussed: page reorganization for when pages need page-specific sub-templates.

---

## Page Reorganization (Future Cleanup)

### Problem

Page-specific sub-components end up at the same level as `base_layout.templ` and `icons.go`. When `settings_page.templ` needs a category select helper, it shouldn't sit next to top-level layouts.

### Proposed Structure

```
internal/view/
├── components/           # shared building blocks (button, input, select, card...)
│   ├── button/
│   ├── select/
│   ├── input/
│   └── ...
├── shared/               # shared page-level pieces (error_banner, loading)
├── utils/                # signal helpers, tw_merge
├── layout/
│   ├── base.templ        # BaseLayout()
│   └── admin.templ       # AdminLayout()
├── icons.go
├── vite_assets.go
├── login/
│   └── page.templ        # LoginPage()
├── register/
│   └── page.templ        # RegisterPage()
├── dashboard/
│   └── page.templ        # DashboardPage()
├── settings/
│   ├── page.templ        # SettingsPage()
│   └── category_select.templ   # page-specific sub-template
├── episodes/
│   ├── list.templ
│   └── new.templ
└── examples/             # or delete these
    ├── tiktak.templ
    ├── filter.templ
    └── add_sub.templ
```

### Handler Import Change

Before:
```go
import "github.com/gracchi-stdio/castogo/internal/view"
templ.Handler(view.SettingsPage(data, config))
```

After:
```go
import "github.com/gracchi-stdio/castogo/internal/view/settings"
templ.Handler(settings.Page(data, config))
```

### Recommendation

Do this as a **dedicated cleanup step** after the settings + categories work. Don't mix new feature work with moving files around.

---

## iTunes Categories — Implementation Plan (Approach B: SSE Round-Trip)

### Architecture Decision

**Option A chosen:** Server-owned Go struct is the single source of truth.

- Go struct defines all categories + subcategories
- Server validates on save
- Templ renders the select options server-side
- Subcategory cascade via SSE: when category changes, a GET request fetches the new subcategory options as server-rendered HTML

Why not purely client-side: the RSS feed outputs `<itunes:category>` directly. An invalid category means Apple rejects the feed. Catch bad data at save time, not at feed output time.

Why not embed JSON + client-side DOM: the project's pattern is server-rendered templates with Datastar SSE patches. Approach B stays consistent.

---

### Step 1: Create `internal/domain/categories.go`

Define a Go `map[string][]string` where:
- Key = top-level category name (exact string Apple expects in RSS)
- Value = slice of subcategory names (empty slice if no subcategories)

Also add two helpers:
- `IsValidCategory(category, subcategory string) bool` — checks if the combination is valid
- `SubcategoriesFor(category string) []string` — returns subcategories for a given category

#### Full Apple Category List

| Category | Subcategories |
|---|---|
| Arts | Books, Design, Fashion & Beauty, Food, Performing Arts, Visual Arts |
| Business | Careers, Entrepreneurship, Investing, Management, Marketing, Non-Profit |
| Comedy | Comedy Interviews, Improv, Stand-Up |
| Education | Courses, How To, Language Learning, Self-Improvement |
| Fiction | Comedy Fiction, Drama, Science Fiction |
| Government | *(none)* |
| History | *(none)* |
| Health & Fitness | Alternative Health, Fitness, Medicine, Mental Health, Nutrition, Sexuality |
| Kids & Family | Education for Kids, Family, Parenting, Pets & Animals, Stories for Kids |
| Leisure | Animation & Manga, Automotive, Aviation, Crafts, Games, Hobbies, Home & Garden, Video Games |
| Music | Music Commentary, Music History, Music Interviews |
| News | Business News, Daily News, Entertainment News, News Commentary, Politics, Sports News, Tech News |
| Religion & Spirituality | Buddhism, Christianity, Hinduism, Islam, Judaism, Religion, Spirituality |
| Science | Astronomy, Chemistry, Earth Sciences, Life Sciences, Mathematics, Natural Sciences, Nature, Physics, Social Sciences |
| Society & Culture | Documentary, Education, History, Personal Journals, Philosophy, Places & Travel, Relationships |
| Sports | Baseball, Basketball, Cricket, Fantasy Sports, Football, Golf, Hockey, Rugby, Running, Soccer, Swimming, Tennis, Volleyball, Wrestling |
| Technology | *(none)* |
| True Crime | *(none)* |
| TV & Film | After Shows, Film History, Film Interviews, Film Reviews, TV Reviews |

---

### Step 2: Add Subcategory SSE Endpoint

**File:** `internal/handler/admin_settings.go`

Add handler method `settingsSubcategories(c *echo.Context) error`:

1. Read `category` from query params: `c.QueryParam("category")`
2. Look up subcategories from the domain map using `SubcategoriesFor()`
3. Return SSE `PatchElements` with a re-rendered subcategory select component

Register the route: `g.GET("/settings/subcategories", h.settingsSubcategories)`

---

### Step 3: Update Settings Page Template

**File:** `internal/view/settings_page.templ`

Replace the two plain `<input>` fields for category/subcategory with `selectcomponent.Select`:

- **Category select:** all 19 top-level categories as `Options`. `OnChange` triggers `@get('/admin/settings/subcategories?category=' + $category_select.value)` to update the subcategory dropdown.
- **Subcategory select:** its content gets replaced by the SSE response from Step 2.

The category select needs to pass the options from the Go map. The template receives the categories data as a parameter or accesses the domain package directly.

---

### Step 4: Implement `settingsSave` Handler

**File:** `internal/handler/admin_settings.go`

Fill in the existing stub:

1. Read form values with `c.FormValue()` (the form uses `@post` with `contentType: 'form'`)
2. Validate category/subcategory using `IsValidCategory()`
3. Build `domain.UpdatePodcastConfig` with string pointers:
   - Non-empty field → pointer to the string
   - Empty field → nil pointer (COALESCE keeps existing DB value)
4. Call `h.settingsService.UpdatePodcastConfig()`
5. Respond with SSE: success signal or error signal

#### Key Detail: Empty vs Unchanged Fields

The form sends all fields every time. To avoid overwriting existing values with empty strings, convert:
- Empty string → `nil` pointer (SQL `COALESCE` keeps current value)
- Non-empty string → `&value` pointer

Example:
```go
func stringPtrIfNonEmpty(s string) *string {
    if s == "" {
        return nil
    }
    return &s
}
```

#### Validation

- Title should be required
- Category (if provided) must pass `IsValidCategory()`
- If category has subcategories and one is selected, it must be valid for that category
- If category has no subcategories, subcategory must be empty

---

## Relevant Files

| File | Role |
|---|---|
| `internal/domain/categories.go` | **NEW** — category map + validation |
| `internal/domain/podcast_config.go` | `PodcastConfig` struct with Category/Subcategory fields |
| `internal/service/settings_service.go` | `GetPodcastConfig`, `UpdatePodcastConfig` |
| `internal/handler/admin_settings.go` | `settingsSave` stub, `settingsUploadCoverImage`, new `settingsSubcategories` |
| `internal/view/settings_page.templ` | Settings form with current text inputs for category |
| `internal/view/components/select/` | Existing select component (args, variants, expressions) |
| `sql/queries/podcast_config.sql` | `UpdatePodcastConfig` with COALESCE pattern |
| `internal/handler/helpers.go` | `sse()`, `readSignals()`, `parseInt64()` |
