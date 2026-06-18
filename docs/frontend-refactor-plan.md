# Frontend Refactor Plan

Living plan for the layout, form, navigation, and frontend TS refactor. Capture of decisions, current state, and next steps across sessions. Read alongside `docs/project-refactor-current-state.md` (anytype index) and `docs/feedback-admin-ui-consistency.md` (audit context).

## Goals

1. **Form consistency** — five form pages use three different patterns (signals-only without `<form>`, raw `<form enctype=multipart>`, `@form.Form` wrapper). Field errors, submit indicators, sticky action bars re-implemented per page.
2. **Navigation duplication** — `navigateAdmin` and `navigatePublic` in `assets/js/app.js` are near-identical (~120 lines combined), plus two click-predicates and two link-updaters. Swup v4 with native View Transitions could collapse to a plugin config.
3. **Shared component gaps** — no `StatusBadge`, `RowActions`, `DataTable`, `EmptyState`, `BackButton`, `NewButton`. Pages hand-roll these.
4. **Frontend TS migration** — bootstrap TypeScript, extract `app.js` into typed semantic modules. Toast is the pilot.

## Foundational decisions (made in earlier sessions)

### Toast transport stays `ExecuteScript`

Earlier exploration considered replacing the `window.pushToast(...)` call-string with a signal-based transport (`PatchSignals` + `data-effect`). After reading the Datastar LLM system prompt reference and verifying backend call-sites:

- `sse.ExecuteScript(js)` IS Datastar's sanctioned primitive for backend-to-frontend JS execution.
- The ~50 `out.ExecuteScript(toastScript(...))` call-sites are already idiomatic. Zero transport change.
- `data-effect` could enable signal-driven toast, but stacked toasts with per-toast timers fight signal replace/merge semantics. Not worth it for transient stacks.
- Public banner (Track C) is the right place to use signal transport — single replace-semantics message.

### `#build` is an instance method, not static

Settled during toast TS migration. `#build` is called via `this.#build(...)`, so `this` is the instance and it CAN read `this.#opts`. A `static` method has `this` bound to the class and can't see instance state.

**Constructor ordering rule:** assign all fields before calling any instance method that reads them.

### JS private (`#`) over TS `private` keyword

Used for `Toast.#element`, `#timer`, `#opts`, `#exit` and `ToastManager.#toasts`, `#stackID`. Runtime-enforced, not bypassable via `as any`, hidden from `JSON.stringify`. Use plain public fields for configuration; use TS `private`/`protected` only when you need subclass access (since `#` blocks subclasses).

### Toast class layering

- `Toast` owns everything about one notification: `#element`, `#opts`, `#timer`, `#exit`. Methods: `mount`, `dismiss`, `#build`, `#getVariantClasses`.
- `ToastManager` owns only the collection: `#toasts`, `#stackID`. Methods: `push`, `clear`, `#remove`.

Per-toast concerns (variant classes, timeout decision) live on `Toast`. Manager doesn't know about per-variant timeout rules.

### Layout split is shape 1b (multi-container swap)

`admin_layout.templ` splits into:
- **Persistent chrome zone** — `<header data-admin-chrome>` containing the sidebar trigger, theme toggle, and user dropdown. Survives SPA nav.
- **Swap zone 1** — `<div data-admin-header>` inside the persistent header, holds title + actions. Swaps per page.
- **Swap zone 2** — `<main data-admin-content style="view-transition-name: admin-content">`. Swaps per page.

`navigateAdmin` extends to swap both `[data-admin-header]` and `[data-admin-content]` from the fetched doc. Maps directly onto Swup's `containers` config later — no layout rework if/when Swup lands.

## Current state

### Toast TS migration — in flight

File: `assets/js/toast.ts` + `assets/js/types.ts`.

Class skeleton present (`Toast` + `ToastManager`) with `#build` and `#getVariantClasses` filled in. Dead function-based code at the bottom of `toast.ts` (lines ~118+) still present — delete once class methods absorb all behavior.

**Known issues to fix:**
1. Constructor ordering — `#build` reads `this.#opts` before assignment. Fix: assign `this.#opts = opts;` first, then call `#build(onDismiss)` (drop redundant `opts` parameter from `#build`).
2. `mount(parent)` is empty — needs: prepend `#element` to parent, compute timeout from `#opts.variant`/`#opts.timeoutMs`, store handle on `#timer`, schedule `this.dismiss()`.
3. `dismiss()` is empty — needs: cancel `#timer`, guard on `#exit`, toggle `toast-enter` → `toast-exit`, schedule `#element.remove()` after `TOAST_EXIT_DURATION`.
4. `push` doesn't check `opts.message` before constructing — move guard out of `#build` into `push`.

**Not yet done:**
- `assets/js/types.ts` may not exist — create with `ToastVariant` and `ToastOptions`.
- `tsconfig.json` at repo root — Vite handles TS via esbuild, but editor IntelliSense and `tsc --noEmit` need this. Settings: `"module": "ESNext"`, `"moduleResolution": "bundler"`, `"target": "ES2022"`, `"lib": ["ES2022", "DOM", "DOM.Iterable"]`, `"strict": false` (tighten per-module later), `"allowJs": true`, `"noEmit": true`, `"skipLibCheck": true`, `"types": ["vite/client"]`.
- `app.js` → `app.ts` rename + Vite input + dev-mode script src in `base_layout.templ`.
- Wire `window.pushToast = (o: ToastOptions) => toastManager.push(o);` (shim — backend contract unchanged).
- Type `window.pushToast` via `global.d.ts` or `types.ts`: `declare global { interface Window { pushToast: (opts: ToastOptions) => void } }`.
- Import `{ toastManager }` from `app.ts`.

### Step 0 layout split — paused

Defined in full (see `internal/view/layout/admin_layout.templ` target structure below). Paused until toast TS migration lands so we don't fork `app.js` mid-refactor.

Target for `admin_layout.templ`:
```
<div class="flex-1 flex flex-col min-w-0">                         ← persistent column
    <header data-admin-chrome class="sticky top-0 z-10 ...">       ← PERSISTS
        @SidebarTrigger                                            ← persists
        <div data-admin-header class="flex items-center gap-2 flex-1">  ← SWAPS
            <h2 class="text-lg font-semibold flex-1">{ title }</h2>
            <div class="flex items-center gap-2">
                if actions != nil { @actions }
            </div>
        </div>
        @ThemeToggle                                               ← persists
        @DropdownMenu (user)                                       ← persists
    </header>
    <main data-admin-content style="view-transition-name: admin-content" class="flex-1 p-4">  ← SWAPS
        { children... }
    </main>
</div>
```

JS edit target (`assets/js/app.ts` near `navigateAdmin`):
```ts
const ADMIN_HEADER_SELECTOR = "[data-admin-header]";
// inside runWithOptionalViewTransition callback, after currentContent.replaceWith(nextContent):
const nextHeader = nextDocument.querySelector(ADMIN_HEADER_SELECTOR);
const currentHeader = document.querySelector(ADMIN_HEADER_SELECTOR);
if (nextHeader && currentHeader) {
    currentHeader.replaceWith(nextHeader);
}
```

## Track A — Layout Split + Form Abstraction

### A.0 Reconcile with reality (read-only)
- Re-read `internal/view/components/form/form.templ` — confirm `SignalsWithFormId`, `FormMessage`, `$fetching` indicator all present.
- Resolve the merge conflict at `internal/view/settingview/settings_page.templ:7-11` (`UU` in git status).

### A.1 Add two thin helpers
- `FormField(args FormFieldArgs) { children... }` in `internal/view/components/form/form_field.templ` — composes `FormLabel` + input (children) + `FormMessage`. ~30 lines of composition.
- `StickyActions(actions templ.Component) { children... }` — sticky submit/cancel bar. One file, used by every form page.

### A.2 Pilot 1 — `forms/settings/settings.templ`
- Move `settings_page.templ` body into `internal/view/forms/settings/settings.templ`.
- Replace per-field error scaffolding with `@form.FormField(...)`.
- Replace page-local submit indicator with `$fetching` from `form.Form`.
- **Checkpoint:** build, click through settings save, field errors render inline, success toast fires.

### A.3 Pilot 2 — `forms/login/login.templ`
- Same shape, JSON-content (not multipart). Validates pattern works for both content types.

### A.4 Decision to defer: `page_form.templ` and `episode_new_page.templ`
- Complex forms (tabs, block editor, multipart uploads). Per Option 3, page-level composition — call `StickyActions` and `FormField` directly without wrapping in a `forms/` component.
- After A.2/A.3 land, do only if wanted.

## Track B — Shared Components & Consistency Cleanup

Each is a small, independent file. Pick order by what blocks current work.

| Step | Component / change | File | Notes |
|---|---|---|---|
| B.1 | `StatusBadge` | `internal/view/components/statusbadge/` | variant map for draft/scheduled/published/archived |
| B.2 | `RowActions` | `internal/view/components/rowactions/` | wraps icon-button actions in table rows |
| B.3 | `EmptyState` | `internal/view/components/emptystate/` | icon + heading + description + optional CTA |
| B.4 | `BackButton`, `NewButton` | `.../backbutton/`, `.../newbutton/` | thin wrappers around `button.Button` with icon + href |
| B.5 | `DataTable` | `internal/view/components/datatable/` | header + rows slot; defer until B.1/B.2 done |
| B.6 | Flatten `PublicLayoutData` | `internal/domain/` + callers | struct has accreted fields — audit and trim |
| B.7 | Delete dead code | various | commented `<img>` lines, unused signal declarations, stale `errorbanner` call-sites |

**Pacing:** scaffold on request, wire into pages. Don't do B.7 until A.2/A.3 land (otherwise call-sites still live).

## Track C — Public Banner Mechanism

Constraints:
- Must survive SPA navigation (like `#toast-stack` does for admin).
- Must NOT live on body-level signal scope (keep public DOM clean).
- Must NOT be a forked-component change (blacklist).

**Sketch:**
- Banner container in `internal/view/layout/public_layout.templ` (castogo-original, editable) positioned outside `[data-public-content]`. Mirror `#toast-stack` pattern.
- `pushBanner({message, variant})` in `app.ts` — sibling of `pushToast`, scoped to public banner element. ~40 lines, mostly cloned.
- `bannerScript(message, variant)` helper in `helpers.go` returning `window.pushBanner(...)`. Sibling of `toastScript`.
- Migrate public call-sites: grep for `toastScript` in public handlers, switch the ones that should be banner. Admin stays on toast.

Defer until Track A pilot 1 lands — form pattern establishes where messages render (inline field vs. banner vs. toast), informing which call-sites move.

## Track D — Swup Integration Analysis (decision input, not committed)

### What Swup replaces in `app.ts`

| Current code | Lines | Swup equivalent |
|---|---|---|
| `shouldHandleAdminNavigation` | ~40 | `new Swup({ plugins: [new SwupPreload()] })` |
| `shouldHandlePublicNavigation` | ~46 | same plugin; `RESERVED_PATHS` becomes `route` filter or `data-no-swup` |
| `navigateAdmin` | ~60 | gone — Swup fetches, parses, swaps containers, manages history |
| `navigatePublic` | ~63 | same; `init`/`destroy` hooks move to `visit:start`/`visit:end` |
| `updateActiveNavLink` | ~13 | `@swup/head-plugin` or 10-line `content:replace` listener |
| `updateActivePublicNavLink` | ~13 | same |
| `runWithOptionalViewTransition` | ~18 | gone — `@swup/native-view-transitions-plugin` reuses existing CSS |
| popstate handler | ~27 | gone — Swup manages history |
| search form SPA interception | ~16 | `form-submit` hook or `data-no-swup` opt-out |

**Net:** ~250 lines → ~40 lines of Swup config + 2–3 lifecycle listeners for `initBlockSorter`/`initEpisodePlayers`/`initPublicNavScroll`.

### What stays

- `pushToast` / `pushBanner` (Swup doesn't do UI messages).
- `initBlockSorter`, `initEpisodePlayers`, `destroyActivePlayer`, `initPublicNavScroll` — move into Swup's `content:replace` (init) and `visit:end`/`content:replace` (destroy) hooks.
- `RESERVED_PATHS` — either `routes` option or individual links get `data-no-swup`.

### Plugin stack (all first-party `@swup/*`)

```
swup                           core page transitions
@swup/preload-plugin           prefetch on hover/intersection
@swup/native-view-transitions  { native: true } — reuses existing CSS
@swup/head-plugin              title + meta updates
@swup/scroll-plugin            scroll restoration
```

### Cost / risk

- New runtime dependency (~10KB gzipped for core + plugins).
- **Datastar interaction:** Swup swaps innerHTML on `content:replace`. Datastar's signal store is global (lives on `document`), so signals survive. Datastar-watched elements inside swapped containers re-initialize — same as current `navigateAdmin` behavior, no regression.
- CDN vs bundle: Swup can be CDN (`unpkg.com/swup@4`) or bundled via Vite. Bundling is cleaner for production cache-busting.
- `view-transition-name: admin-content` — existing CSS works as-is with `@swup/native-view-transitions`. No CSS changes.

### Decision frame

| If… | Then… |
|---|---|
| Cleanest `app.ts`, don't mind a dep | Adopt Swup as Track D, after A+B |
| Zero new deps | Skip Swup. Dedupe `navigateAdmin`/`navigatePublic` into one `navigate(url, {scope, init, destroy})` — saves ~100 lines |
| Unsure | Throwaway branch with Swup wired against one route (settings) to feel the DX |

## Recommended order

1. **TS bootstrap** — `tsconfig.json`, `app.ts` rename, `types.ts`, Vite config + `base_layout.templ` updates.
2. **Toast TS migration** — finish `Toast.mount`/`dismiss`, wire `window.pushToast`, delete dead functions, verify call-sites untouched.
3. **Step 0 layout split** — `admin_layout.templ` restructure + `navigateAdmin` multi-container swap.
4. **A.0 → A.1 → A.2** — form pattern pilot.
5. **B.1–B.4** — small components, interleaved with A.
6. **A.3** — login pilot.
7. **Track C** — public banner.
8. **B.5–B.7** — DataTable, flatten `PublicLayoutData`, dead-code sweep.
9. **Track D** — Swup decision (or dedupe).

## Verification

Per step:
- `just generate` after any `*.templ` change (regenerates `*_templ.go`).
- `just dev` — Vite HMR + templ watch + air. Click through affected route.
- For form changes: trigger validation error (empty required field) → `FormMessage` renders inline; trigger success → toast/banner fires; `$fetching` indicator shows on submit.
- For Swup: navigate admin↔admin, admin↔public, public↔public, back/forward, search submit, logout — each transitions smoothly with no console errors. `destroyActivePlayer` must still fire on leaving an episode page.
- For toast TS: trigger each variant (info/success/error) — auto-dismiss at correct timeout, manual close works, no regression in admin handlers.

## Critical files

- `internal/view/components/form/form.templ` — foundation; `FormField` composes its children.
- `internal/view/forms/settings/settings.templ` — pilot 1 (new).
- `internal/view/forms/login/login.templ` — pilot 2 (new).
- `internal/view/layout/public_layout.templ` — banner container (Track C).
- `internal/handler/helpers.go` — `toastScript`, future `bannerScript`.
- `assets/js/app.ts` — Swup target (Track D), `pushBanner` sibling (Track C), toast import.
- `assets/js/toast.ts` — TS toast module (in flight).
- `assets/js/types.ts` — TS types (in flight).
- `internal/handler/admin_settings.go`, `admin_pages.go`, `admin_episodes.go` — call-sites that may shift as forms migrate.

## Out of scope

- `imageupload` refactor — deferred until form abstraction lands.
- Block type registry — deferred indefinitely.
- 22 forked datastarui components (sidebar, dropdown, button, etc.) — logic changes blacklisted; compose only.
- `admin_layout.templ` IS in scope (Step 0).
