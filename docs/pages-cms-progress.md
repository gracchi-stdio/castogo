# Pages CMS — Build Progress

Tracking the step-by-step implementation of the block-based multi-page CMS for Castogo.
Full plan: `docs/pages-cms-plan.md`

---

## Completed Steps

### Step 1 — Migration ✅
- `sql/migrations/011_drop_landing_page_sections.sql` — dropped old table
- `sql/migrations/012_create_pages_blocks.sql` — created `pages` and `page_blocks` tables
- Materialized `path` column with UNIQUE constraint
- Self-referential `parent_id` FK
- Unique indexes: `(slug, COALESCE(parent_id, 0))` for siblings, `(page_id, block_type)` for blocks

### Step 3 — sqlc Queries ✅
- `sql/queries/page.sql` — full page CRUD + GetChildren, GetDescendants, UpdateDescendantPaths, GetChildrenCount, GetPagePathAndChildrenCountByID, CountPageWithoutParent, GetPageSiblings
- `sql/queries/block.sql` — block CRUD (Create, GetByPageID, Update, Delete)
- Removed old `landing_sections.sql`

### Step 4 — Domain Types ✅
- `internal/domain/page_block.go` — Page, PageBlock, PageMetadata
- `internal/domain/errors.go` — ErrReservedSlug, ErrDuplicatePath, ErrMaxDepth, ErrInvalidParent
- sqlc.yaml updated to column-level JSONB overrides (PageMetadata, AudioMetadata, RawMessage)

### Step 5 — Repository ✅
- `internal/repository/page.go` — PageRepository interface
- `internal/repository/postgres/page_postgres.go` — full implementation with toDomain converters
- Removed old `landing_page_postgres.go`
- Fixed `episode_postgres.go` for new sqlc JSONB overrides

### Step 6 — Service ✅
- `internal/service/page_service.go` — PageService with business logic
- CreatePage: reserved slug check, path computation, max depth enforcement, duplicate key error translation
- UpdatePage: partial updates, path recomputation, descendant path update
- GetPageWithBlocks: composite page + blocks loader
- SaveBlock / DeleteBlock: block CRUD

---

## Remaining Steps

### Step 7 — Theme Scope
- Create `assets/css/theme-public.css` — CSS custom property overrides under `.theme-public`
- Create `assets/css/public.css` — public-specific styles
- Update `assets/css/app.css` imports

### Step 8 — Public Layout
- Create `internal/view/layout/public_layout.templ` — wraps BaseLayout, adds `.theme-public` div, public nav, footer

### Step 9 — Page Templates
- Create `internal/view/pageview/landing.templ` — block-based layout
- Create `internal/view/pageview/text.templ` — simple text layout
- Create `internal/view/pageview/blocks/*.templ` — hero, features, cta, testimonials, episodes_showcase, footer

### Step 10 — Public Handler
- Modify `internal/handler/public.go` — homePage handler (slug "home"), pageResolver handler
- Route registration: `GET /` hardcoded, `GET /*pageSlug` dynamic with reserved slug guard

### Step 11 — Admin Handler + Views
- Create `internal/handler/admin_pages.go` — page CRUD handlers
- Create `internal/view/pageadminview/` — page list, page form, block editor

### Step 12 — Integration
- Modify `cmd/server/main.go` — wire PageService, update handler constructors
- Add admin sidebar nav entry for Pages
- Seed home page with existing block data

### Step 13 — Cleanup
- Remove old landing page code: landingview/, landingadminview/, landing_page_service.go, landing_page.go, landing_page.go (repo)
