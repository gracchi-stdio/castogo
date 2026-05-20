# Pages CMS Plan — Block-Based Multi-Page System

## Overview

Transform the single landing page into a full page-based CMS where:

- **Pages** are first-class entities with slug, path, layout type, publish state, and parent/child nesting
- **Blocks** (renamed from "sections") are ordered content units belonging to a page
- Pages can nest under other pages for URL paths like `/about/team`
- A **layout type** on each page determines which template renders it
- **Theme scopes** let admin and public pages share components but have distinct visual identity

---

## Schema

### `landing_pages` (new table)

```sql
CREATE TABLE landing_pages (
    id               BIGSERIAL PRIMARY KEY,
    slug             TEXT NOT NULL,
    path             TEXT NOT NULL UNIQUE,      -- materialized: 'home', 'about', 'about/team'
    title            TEXT NOT NULL,
    parent_id        BIGINT REFERENCES landing_pages(id) ON DELETE CASCADE,
    layout_type      TEXT NOT NULL DEFAULT 'landing',
    is_published     BOOLEAN NOT NULL DEFAULT false,
    meta_title       TEXT,
    meta_description TEXT,
    sort_order       INTEGER NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Prevents two siblings from having the same slug
CREATE UNIQUE INDEX idx_landing_pages_slug_parent
    ON landing_pages (slug, COALESCE(parent_id, 0));
```

### `landing_page_blocks` (renamed from `landing_page_sections`)

```sql
CREATE TABLE landing_page_blocks (
    id          BIGSERIAL PRIMARY KEY,
    page_id     BIGINT NOT NULL REFERENCES landing_pages(id) ON DELETE CASCADE,
    block_key   TEXT NOT NULL,
    content     TEXT NOT NULL DEFAULT '{}',
    is_visible  BOOLEAN NOT NULL DEFAULT true,
    sort_order  INTEGER NOT NULL DEFAULT 0,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Unique block key per page
CREATE UNIQUE INDEX idx_landing_page_blocks_page_key
    ON landing_page_blocks (page_id, block_key);
```

---

## Design Decisions

### Home Page

- Has reserved `slug = 'home'`, `path = 'home'` in the database
- Public route is hardcoded `GET /` — no URL ever shows `/home`
- `pageResolver` ignores paths starting with reserved slugs

### Reserved Slugs (blocklist)

```go
var reservedSlugs = map[string]bool{
    "home":        true,
    "admin":       true,
    "login":       true,
    "logout":      true,
    "register":    true,
    "feed":        true,
    "healthcheck": true,
    "assets":      true,
}
```

- Admin page creation rejects reserved slugs
- `pageResolver` checks first path segment against blocklist — blocks entire subtree

### Materialized `path` Column

- Stored as `'about/team'` — single indexed lookup for public resolver
- Computed on create: `parent.path + "/" + slug` (or just `slug` for top-level)
- Recomputed on slug/parent change: `UPDATE descendant paths WHERE path LIKE old_path || '/%'`
- Read-heavy (every pageview) vs write-rare (admin only) — denormalization wins

### Layout Types

Start with two, extend later:

- `landing` — block-based layout (hero → features → episodes → cta → footer)
- `text` — simple body content (for about, contact, privacy policy, etc.)

More layouts (blog, episode_detail) added by extending the enum in rendering logic.

### Nesting

- `parent_id` is nullable (NULL = top-level page)
- Max depth: 2 levels (e.g., `/about/team`) — enforced in service layer
- Admin displays tree view with parent/child relationships

---

## Theme & Template Architecture

### Theme Scopes (CSS)

Admin and public pages share the same component library (`internal/view/components/`) but use different CSS custom property values:

```
assets/css/
├── app.css              ← entry point, imports everything
├── theme.css            ← shadcn/ui tokens (stays as-is — admin/default theme)
├── theme-public.css     ← public-facing overrides (new)
└── public.css           ← public-specific styles (new)
```

`theme-public.css` overrides the same CSS custom property names under `.theme-public`:

```css
.theme-public {
  --primary: 262 83% 58%;
  --primary-foreground: 0 0% 98%;
  --radius: 0.75rem;
}
```

Tailwind v4's `@theme` block already maps tokens like `--color-primary` → `hsl(var(--primary))`. The cascade means `.theme-public` children automatically pick up new values. Zero component duplication.

### Template Structure

```
internal/view/
├── layout/
│   ├── base_layout.templ      ← shared <html>/<head>/<body> shell
│   ├── admin_layout.templ     ← admin chrome (sidebar, header)
│   └── public_layout.templ    ← public chrome (nav, footer, .theme-public wrapper)
├── pageview/                  ← page rendering templates
│   ├── landing.templ          ← block-based landing layout
│   ├── text.templ             ← simple text/markdown layout
│   └── blocks/                ← reusable block renderers
│       ├── hero.templ
│       ├── features.templ
│       ├── cta.templ
│       └── ...
├── pageadminview/             ← admin UI for managing pages + blocks
│   ├── page_list.templ
│   ├── page_form.templ
│   └── block_editor.templ
├── components/                ← shared UI components (button, card, icon, etc.)
│   └── ...                    ← MUST be used for ALL UI elements
└── landingview/               ← removed after migration
```

### Theming vs Templates — Two Separate Axes

- **Theme** = colors, radii, spacing, typography (CSS custom property scope)
- **Template** = layout structure, which blocks render where (Templ component selection)

A `landing` template under the public theme looks different from the same template under admin — but the template code is identical. Only the CSS tokens change.

---

## Component Rules

All UI elements across both admin and public views MUST use the shared component library in `internal/view/components/`. This includes:

- `button` — all buttons (variant, size, icon support)
- `card` — content cards
- `icon` — Lucide icons via `icon.Lucide(name, class)`
- `input`, `textarea`, `select` — form inputs
- `dialog` — modals
- `dropdown` — dropdown menus
- `tabs` — tab navigation
- `breadcrumb` — navigation breadcrumbs
- `sidebar` — sidebar navigation (admin)
- `sheet` — slide-in panels
- `popover`, `tooltip` — contextual overlays
- `label`, `checkbox` — form elements
- `loading` — loading indicators
- `errorbanner`, `fielderror` — error display
- `form` — form wrappers
- `avatar` — user avatars
- `brand` — brand/mascot assets
- `themetoggle` — dark/light mode switch
- `dateinput`, `datepicker`, `calendar` — date selection

When building new page templates (landing, text, blocks), always use these components instead of raw HTML elements. This ensures theme scope changes propagate correctly.

---

## Route Structure

```go
// Public routes
e.GET("/", publicHandler.homePage)              // hardcoded — loads page slug "home"
e.GET("/*pageSlug", publicHandler.pageResolver)  // dynamic — resolves /about, /about/team

// Admin routes (protected)
adminGroup.GET("/pages", adminHandler.pageList)
adminGroup.GET("/pages/create", adminHandler.pageCreatePage)
adminGroup.POST("/pages/create", adminHandler.pageCreateAction)
adminGroup.GET("/pages/:id/edit", adminHandler.pageEditPage)
adminGroup.POST("/pages/:id/edit", adminHandler.pageUpdateAction)
adminGroup.DELETE("/pages/:id", adminHandler.pageDeleteAction)
adminGroup.POST("/pages/:id/blocks/:blockKey", adminHandler.blockSaveAction)
```

---

## Steps (Build Order)

| Step | What | Details |
|------|------|---------|
| **1** | Migration | Create `landing_pages` table, rename `landing_page_sections` → `landing_page_blocks` with `page_id` FK, add unique indexes |
| **2** | Seed | Create "Home" page row, assign existing sections to it via `page_id` |
| **3** | sqlc queries | Page CRUD (Create, GetByID, GetByPath, GetBySlug, List, Update, Delete), block queries scoped by `page_id`, descendant path update query |
| **4** | Domain types | `Page`, `CreatePageInput`, `UpdatePageInput`, `PageBlock`, `PageWithBlocks`, rename existing section types, add `reservedSlugs` blocklist, add `LayoutType` constants |
| **5** | Repository | `PageRepository` interface + postgres impl. Migrate `LandingPageRepository` to work with `page_id` scope |
| **6** | Service | `PageService` — CreatePage (compute path, validate slug not reserved), GetByPath, GetBySlug, UpdatePage (recompute paths for descendants), DeletePage, GetPageWithBlocks, block CRUD |
| **7** | Theme scope | `theme-public.css` + `public.css` in `assets/css/`, update `app.css` imports |
| **8** | Public layout | `public_layout.templ` — wraps BaseLayout, adds `.theme-public`, public nav (renders from page tree), public footer |
| **9** | Page templates | `pageview/landing.templ` (refactored from current `landingview/`), `pageview/text.templ`, block renderers in `pageview/blocks/` |
| **10** | Public handler | `GET /` hardcoded home, `GET /*pageSlug` resolver with reserved slug guard, path resolution |
| **11** | Admin handler + views | Page CRUD handlers, page list/create/edit/delete templates, per-page block editor in `pageadminview/` |
| **12** | Integration | Wire `PageService` in `main.go`, update `AdminHandler` to use new service, add admin sidebar nav entry for Pages |
| **13** | Cleanup | Remove old `LandingPageService`, old `landingview/`, old `landingadminview/`, old migration references |

---

## Key Files Changed Per Step

### Step 1 — Migration
- New: `sql/migrations/011_create_landing_pages.sql`
- New: `sql/migrations/012_seed_landing_pages_home.sql`

### Step 2 — Seed
- (included in step 1 migrations)

### Step 3 — sqlc queries
- New: `sql/queries/pages.sql`
- Modified: `sql/queries/landing_sections.sql` → renamed to `sql/queries/page_blocks.sql`
- Run: `just generate`

### Step 4 — Domain types
- New: `internal/domain/page.go`
- Modified: `internal/domain/landing_page.go` → rename types, add `PageID` to block types
- Modified: `internal/domain/errors.go` — add `ErrReservedSlug`, `ErrMaxDepth`, `ErrDuplicatePath`

### Step 5 — Repository
- New: `internal/repository/page.go` (interface)
- New: `internal/repository/postgres/page_postgres.go` (implementation)
- Modified: `internal/repository/landing_page.go` → update interface to use `page_id`
- Modified: `internal/repository/postgres/landing_page_postgres.go` → update to use `page_id`

### Step 6 — Service
- New: `internal/service/page_service.go`
- Modified: `internal/service/landing_page_service.go` → simplified to use page scope

### Step 7 — Theme scope
- New: `assets/css/theme-public.css`
- New: `assets/css/public.css`
- Modified: `assets/css/app.css` — add imports

### Step 8 — Public layout
- New: `internal/view/layout/public_layout.templ`

### Step 9 — Page templates
- New: `internal/view/pageview/landing.templ`
- New: `internal/view/pageview/text.templ`
- New: `internal/view/pageview/blocks/hero.templ`
- New: `internal/view/pageview/blocks/features.templ`
- New: `internal/view/pageview/blocks/cta.templ`
- New: `internal/view/pageview/blocks/testimonials.templ`
- New: `internal/view/pageview/blocks/episodes_showcase.templ`
- New: `internal/view/pageview/blocks/footer.templ`

### Step 10 — Public handler
- Modified: `internal/handler/public.go` — add homePage, pageResolver handlers
- Modified: route registration

### Step 11 — Admin handler + views
- New: `internal/handler/admin_pages.go`
- New: `internal/view/pageadminview/page_list.templ`
- New: `internal/view/pageadminview/page_form.templ`
- New: `internal/view/pageadminview/block_editor.templ`
- Modified: `internal/handler/admin.go` — add PageService dependency, register page routes

### Step 12 — Integration
- Modified: `cmd/server/main.go` — wire PageService, update handler constructors

### Step 13 — Cleanup
- Removed: `internal/view/landingview/`
- Removed: `internal/view/landingadminview/`
- Removed: `internal/service/landing_page_service.go` (merged into page_service)
- Removed: `internal/domain/landing_page.go` (merged into page.go)
