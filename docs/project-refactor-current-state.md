---
name: Refactor current state (cross-session sync)
description: Where the layout/component refactor stands right now — what's done, what's pending decision, and pointers to the Anytype pages with full plans. Read this first when resuming the refactor.
type: project
originSessionId: 0b1abbe6-cd33-4de3-8d85-01a1457b95cf
---
This memory captures the state of the layout/component refactor as of 2026-06-18. Read this first when resuming on any laptop. Detailed plans live in Anytype (synced separately); this file is the index + current decision point.

**Why:** The refactor spans multiple sessions and laptops. Re-investigating the datastarui fork context and audit findings each session wastes hours. This file + the Anytype pages should be enough to resume.

**How to apply:** when resuming the refactor, read this file, then read the four Anytype pages (tagged `castogo` + `layout` + `refactor`) for full detail. The form abstraction decision (Section 4 below) is the immediate blocker.

---

## 1. Essential background (don't re-investigate)

**Castogo forked `github.com/coreycole/datastarui` in commit `24de92e` (Apr 24, 2026).** The fork point is `git show 24de92e`. After the fork:
- 22 components copied into `internal/view/components/` with import paths rewritten to `internal/view/utils`
- `internal/view/utils/` vendored from `datastarui/utils/` (5 shared files: anchor.go, data_class.go, expressions.go, signals.go, tailwind_merge.go)
- Castogo added `brand_tokens.go`; dropped upstream's `connect_errors.go`, `context.go`, `device.go` (app-specific)
- 5 castogo-original components: `errorbanner`, `fielderror`, `loading`, `brand`, `imageupload`
- Upstream source saved at `/home/tg/Downloads/datastarui-main/datastarui-main/` for diffing (may not exist on other laptop — clone datastarui if needed)

**Implication:** the 22 forked components are a blacklist (see `feedback-datastarui-component-blacklist.md`). Don't modify their logic — only mechanical fixes (import paths, generated files). Fork-fixing creates divergence pain.

---

## 2. Anytype pages (in Notebook space, tagged castogo + layout [+ refactor])

1. **Castogo Layout & View Refactor Report** 🏗️ — original audit, 25 findings
2. **Castogo Toast, Forms & Swup Investigation** 🔄 — toast design, form-agnostic AdminFormPage sketch, Swup revised assessment
3. **Castogo Component Library Code Quality Audit** 🔬 — rigorous component-layer read
4. **Castogo Refactor Plan (Revised)** 🎯 — revised tiers after datastarui discovery

Read in that order if refreshing context.

---

## 3. What's done in the most recent session (2026-06-18)

- ✅ Saved `feedback-datastarui-component-blacklist.md` — 22 components off-limits for logic changes
- ✅ Saved `project-block-types-future-refactor.md` — block editor registry deferred
- ✅ **Fixed dropdown imports** — `internal/view/components/dropdown/{dropdown.templ, variants.go, expressions.go}` now import `internal/view/utils` instead of `coreycole/datastarui/utils`. dropdown_templ.go regenerated via `just generate`. Build passes.
- ✅ Investigated utils conflict — confirmed it's a vendored copy, not a parallel implementation. Diff details in Section 1 above.
- ✅ User ran `go mod tidy` — `datastarui` dependency removed from go.mod.

---

## 4. IMMEDIATE DECISION POINT — form layout abstraction

The user explicitly said: *"I want to decide on form general layout and abstraction before moving to the particular"* (the "particular" being `imageupload`). **Do not start implementing form changes or imageupload refactor until the user picks an option.**

### Three options presented

| Option | Description |
|---|---|
| **1. Follow datastarui `forms/` pattern** | Each form is its own component in `internal/view/forms/<name>/`. Pages become thin callers. Upstream-idiomatic but each form component is ~200 lines for complex forms. |
| **2. Layout with slots (`AdminFormPage`)** | Layout owns sticky bar + chrome. Caller owns signals + form + fields. Flexible, simpler migration, more boilerplate per page. |
| **3. Hybrid (recommended)** | `forms/` directory for self-contained forms (settings, episode, login) following datastarui pattern. Complex forms (page_form with tabs + block editor) compose shared helpers directly. Two patterns because castogo has two kinds of forms. |

### Specific decisions inside Option 3 (also need user confirmation)

- Sticky bar + submit button INSIDE form component (datastarui-idiomatic)
- Use `$fetching` auto-indicator from `form.Form` instead of per-form `$uploading`/`$loading_status`
- Field errors inline via `form.FormMessage`; form-level errors via toast
- `forms/` directory for: settings, episode_new, login (self-contained)
- Page-level composition for: page_form (tabs), block editor (multi-form)

### What was recommended to the user

Option 3 (hybrid). User had not yet responded when this memory was saved.

---

## 5. Other refactor tiers (pending, lower priority)

These are pending but NOT the immediate focus. From the Revised Plan (Anytype page 4):

- **Tier 1 — castogo-specific component cleanup.** Bring `errorbanner`/`fielderror`/`loading`/`imageupload` in line with datastarui's `SignalManager` pattern. **Imageupload is the worst offender** — JS built by Go string concat with no escaping. User wants to do this AFTER the form abstraction decision.
- **Tier 2 — new castogo components.** `FormField`, `Toast`, `StatusBadge`, `EmptyState`, `RowActions`, `BackButton`, `NewButton`, `StickyFormActions`. Dependent on form abstraction decision.
- **Tier 3 — page conventions.** `AdminFormPage` layout, migrate error call-sites to `$toast`, delete per-page `<errorbanner>`.
- **Tier 4 — navigation.** Dedupe `navigateAdmin`/`navigatePublic` OR adopt Swup. Independent track.
- **Tier 5 — upstream notes.** Document upstream issues, no local fixes.

---

## 6. Reserved / deferred

- **Block type registry refactor** — see `project-block-types-future-refactor.md`. Deferred per user direction. Surface when relevant; don't bring up unsolicited.
- **Public theme toggle** — quick UX win (~30 min), included in all options. Can be done anytime.

---

## 7. How to resume

1. Read this memory file.
2. Optionally read the four Anytype pages for full plan detail.
3. Ask the user which option they picked for form abstraction (Option 1, 2, or 3).
4. If Option 3, confirm the four specific decisions in Section 4.
5. Then plan implementation of the chosen path — likely starting with `forms/settings/settings.templ` as the test case (well-understood, mid-complexity).
6. **Do NOT** start imageupload refactor until form abstraction is decided — that's downstream of the decision.

---

*Last updated 2026-06-18. If this file is more than a few days old, the user may have made progress — ask before assuming this state is current.*
