# Podlog

Self-hosted podcast hosting platform. Manages podcast metadata, episodes (with draft/scheduled/published/archived workflow), users, and auto-generates RSS feeds compliant with Apple/Spotify/Google namespaces. Module: `github.com/gracchi-stdio/castogo`

## Tech Stack

- Go 1.26.2, Echo v4, PostgreSQL 17, sqlc, pgx/v5
- Frontend: Tailwind v4 (via Vite), Datastar (runtime), Templ (type-safe HTML templates)
- Config: `github.com/caarlos0/env/v11` — global `config.Cfg`, `config.Cfg.IsDev`
- Task runner: `just` (justfile) — run `just` to list commands
- Frontend build: Vite + Bun — `just dev` starts Vite HMR + templ watch + air (Go live reload)

## Architecture

```
cmd/server/main.go           — entry point
internal/config/             — env-based config
internal/db/                 — sqlc generated (gitignored, run `just generate`)
internal/domain/             — domain types + errors
internal/repository/         — interfaces + postgres implementations
internal/service/             — business logic
internal/handler/            — HTTP handlers + route registration
  ├─ helpers.go              — shared: sse(), readSignals(), validate, fieldValidationErrors
  ├─ logger.go               — colored request logger middleware
  ├─ middleware.go            — AuthMiddleware, redirectLogin
  ├─ auth.go                 — login/session handler
  ├─ admin.go                — admin dashboard handler
  ├─ clock.go                — SSE clock example
  └─ filter_example.go       — SSE filter example
internal/view/               — Templ templates (*_templ.go generated, gitignored)
  └─ vite_assets.go          — reads Vite manifest for production asset paths
sql/migrations/              — numbered: 001_, 002_, 003_
sql/queries/                 — sqlc query source of truth
assets/                      — frontend source (Vite entry point)
  ├─ js/app.js               — JS entry point (imports CSS, view transitions logic)
  └─ css/                    — Tailwind v4 CSS (theme tokens, animations)
public/                      — served statically by Echo (registered LAST)
  ├─ assets/                 — Vite production build output (hashed JS/CSS)
  └─ .vite/manifest.json     — Vite manifest for asset resolution
```

## Echo v4 Rules

- Handler signature: `func(c echo.Context) error`
- Built on standard `net/http` — no adaptor needed for Datastar SDK or any stdlib middleware
- Request: `c.Request()` returns `*http.Request`, `c.Response()` returns `*echo.Response` (wraps `http.ResponseWriter`)
- JSON binding: `c.Bind(&x)` (reads from JSON body, query params, or path params)
- Templ pages: `echo.WrapHandler(templ.Handler(view.Page()))`
- Route groups: `e.Group("/admin", middleware)` for protected routes
- Static files: `e.Static("/", "public")` — register LAST so it doesn't swallow routes
- Import: `github.com/labstack/echo/v4`, middleware: `github.com/labstack/echo/v4/middleware`
- Sessions: `github.com/labstack/echo-contrib/session` (gorilla/sessions)

## Echo-First

Use Echo's built-in middleware before writing from scratch. Check these first:
- `github.com/labstack/echo/v4/middleware/` — CORS, CSRF, logger, recover, rate limiter, compress, static, redirect, basic auth, JWT, key auth
- `github.com/labstack/echo-contrib/` — session (gorilla/sessions), prometheus, jaeger, cassandra
- `github.com/gorilla/sessions` — session stores (cookie, redis, filesystem)

If Echo has a middleware for it, use it. Don't reinvent authentication, rate limiting, CORS, CSRF protection, etc.

## Datastar + Echo Pattern

Datastar's SDK (`datastar.NewSSE`) calls `http.ResponseController.Flush()` which conflicts with Echo's `Response.Flush()` (calls `WriteHeader` if not committed). Solution: pass the **raw writer** `c.Response().Writer` instead of Echo's `Response` wrapper.

Shared helpers in `internal/handler/helpers.go`:

```go
// sse returns a Datastar SSE generator wired through Echo's raw response writer.
func sse(c echo.Context) *datastar.ServerSentEventGenerator {
    return datastar.NewSSE(c.Response().Writer, c.Request())
}

// readSignals reads Datastar signals from the request body into target.
func readSignals(c echo.Context, target any) error {
    return datastar.ReadSignals(c.Request(), target)
}
```

Usage in handlers:
```go
sse(c).MarshalAndPatchSignals(map[string]string{"error": "message"})
sse(c).Redirect("/admin")
sse(c).PatchElements(html)
```

Session saving before SSE (Set-Cookie header committed together with SSE response):
```go
sess.Save(c.Request(), c.Response().Writer)
sse(c).Redirect("/admin")
```

## Session & Auth

- Sessions: `session.Get("session", c)` from `echo-contrib/session`
- UUID stored as string in session: `sess.Values["user_id"] = user.ID.String()`
  - `gob` (used by securecookie) doesn't support `[16]byte` — must store as string
- Session options: `Path: "/"`, `HttpOnly: true`, `SameSite: Lax`, `MaxAge: 30 days`
- Auth middleware checks session → loads user → sets `c.Set("user", user)`
- Unauthenticated requests redirect to `/login` via `c.Redirect(302, "/login")`
- Login page redirects to `/admin` if user already signed in

## Learning Project

This is a guided learning project, not vibe coding. For each new feature or step:
- Explain what we're building and why before writing code
- Introduce concepts as they come up (don't dump everything at once)
- Stop after each meaningful step for review and questions
- Let the user write or modify code themselves when it's educational

## Database

- `internal/db/` is generated by sqlc — never edit directly
- Edit `sql/queries/*.sql`, then run `just generate`
- Sessions via `echo-contrib/session` (gorilla/sessions) — cookie store for now, PostgreSQL later
- sqlc config: `emit_json_tags`, `emit_empty_slices`, `emit_pointers_for_null_types`

## Domain Errors

Defined in `internal/domain/errors.go`:
- `ErrNotFound`, `ErrUnauthorized`, `ErrDuplicateSlug`, `ErrInvalidInput`

## Cross-Platform

Dev on Ubuntu 24 + Windows 11. Use `filepath.Join()` for paths, Docker for Postgres.



## Shared Components (`internal/view/components/`)

All UI elements across both admin and public views MUST use the shared component library. Never write raw `<button>`, `<input>`, `<select>`, etc. in templates. This ensures theme scope changes propagate correctly and maintains visual consistency.

Available components:
- `button` — all buttons (variants: default, destructive, outline, secondary, ghost, link; sizes: sm, lg, icon)
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

Usage pattern — each component has:
- `args.go` — typed struct for component properties (e.g., `ButtonArgs`)
- `component.templ` — Templ template
- `variants.go` (optional) — variant/size class maps
- `component_templ.go` — generated (gitignored)

Import and call in templates:
```go
@button.Button(button.ButtonArgs{Variant: "default", Size: "sm"}) {
    { "Click me" }
}
```

## CSS Architecture

- Source files live in `assets/css/` — `app.css` (entry point, imports Tailwind) and `theme.css` (design tokens)
- Vite processes CSS via `@tailwindcss/vite` plugin — Tailwind v4 with `@import "tailwindcss"`
- Production: Vite builds to `public/assets/app-[hash].css` (resolved via manifest)
- Dev: Vite dev server serves CSS with HMR on `localhost:3000`
- shadcn/ui design tokens in `theme.css` — CSS custom properties mapped via `@theme` block

## Theme Scopes (Admin vs Public)

Admin and public pages share the same component library but have distinct visual identity via CSS custom property scopes:

- `theme.css` — default shadcn/ui tokens (used by admin)
- `theme-public.css` — overrides under `.theme-public` class (used by public pages)
- `public.css` — public-specific styles

Public layout wraps content in `<div class="theme-public">`. All Tailwind utilities (`bg-primary`, `text-foreground`, etc.) automatically resolve to the public theme's token values inside this scope. No component duplication needed.

## Vite + Asset Pipeline

- `vite.config.ts` — Tailwind v4 plugin, manifest output to `public/`
- Dev mode (`config.Cfg.IsDev`): `base_layout.templ` loads from Vite dev server at `localhost:3000`
- Production: `vite_assets.go` reads `.vite/manifest.json` to resolve hashed filenames
- Run `just dev` for dev mode (Vite HMR + templ watch + air), `just frontend-build` for production build

## Datastar + Templ Rules

- Signal values are JS expressions — strings must be quoted: `data-signals:status="'all'"` not `"all"`
- Shoelace web components need manual signal wiring: `data-on:sl-change="$signal = el.value"` (not `data-bind`)
- Backend is source of truth — signals are for user input and temp UI state only
- Keep element IDs stable across SSE swaps for morphing and CSS transitions
- Expressions use `$` prefix for signals: `$count++`, `$name.toUpperCase()`
- Datastar Go SDK: `sse(c)` for writing, `readSignals(c, &target)` for reading — both in `helpers.go`
- SDK convenience methods: `sse(c).Redirect(url)`, `sse(c).MarshalAndPatchSignals(map)`, `sse(c).PatchElements(html)`
- Full Datastar reference: `docs/datastar-reference.md`

## Pages CMS

Block-based multi-page system. Pages are entities with slug, materialized path, layout type, and parent/child nesting. Blocks (renamed from "sections") are ordered content units belonging to a page.

Key rules:
- Home page: configurable via `homepage_id` in `podcast_config` table; admin picks a page in Settings; `GET /` reads this setting (returns 404 if not set)
- Reserved slugs (`admin`, `login`, `logout`, `register`, `feed`, `healthcheck`, `assets`) — cannot be used for page creation, blocked in `pageResolver`
- Materialized `path` column — single indexed lookup for public resolver, recomputed on slug/parent change
- Layout types: `landing` (block-based), `text` (simple body) — extensible
- Max nesting depth: 2 levels

Full plan: `docs/pages-cms-plan.md`

## CSRF

Not needed yet. `SameSite: Lax` + JSON-only `fetch()` via Datastar provides adequate protection. Add when introducing traditional form endpoints or cross-origin API access.

## RSS Feed Architecture (future)

When we build the RSS feed layer, use a 3-layer design:
1. **DB-backed domain** — podcast_config, episodes (repository + service)
2. **Feed projection** — in-memory view of publishable episodes with normalized URLs/dates
3. **Deterministic renderer** — one `encoding/xml` pipeline, no business logic

Keep mutable admin model separate from read-only feed model. Test the serialization boundary with golden XML fixtures.

## Reference Docs

- `docs/pages-cms-plan.md` — Pages CMS full plan, schema, steps, architecture
- `docs/datastar-reference.md` — Datastar SDK, attributes, expressions, patching, animations
- `docs/shoelace-drawer-menu-details.md` — Drawer, Menu, Menu Item, Details component reference
- `docs/open-props-reference.md` — Open Props design tokens, colors, spacing, typography, Shoelace integration
