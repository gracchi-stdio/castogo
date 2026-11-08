# Castogo — session status

Last compacted: 2026-11-08 (bunny-db complete: steps 8–9 done, local + Bunny
smokes green, merged to main)

## Landed decisions

- **Dead page queries KEPT (user, 2026-11-07)** — GetDescendants,
  GetPageSiblings, GetChildrenCount, GetPagePathAndChildrenCountByID,
  CountPageWithoutParent stay in BOTH corpora for future child-page
  features. Do NOT re-delete in cleanups.
- sqlite JSON columns auto-shape via sqlc overrides + `domain.Scan`/`Value`
  (`AudioMetadata`, `PageMetadata`; `page_blocks.content` → `json.RawMessage`).
- Wrapper conventions, `internal/repository/sqlite`: strict/lenient time
  helpers, `xFromRow`, `"verb object: %w"` wrapping, `sql.ErrNoRows` →
  `domain.ErrNotFound` (predicates like `IsFileProcessed` map to `(false, nil)`
  instead).
- Nav semantics (022): `nav_order NULL` = hidden, ascending order, tie-break
  path ASC.
- **Sessions (step 8):** SQLiteStore is hand-written (no sqlc — gorilla store
  contract, opaque `data` blob); `expires_on` TEXT must match
  `datetime('now')` format exactly; `database/sql` needs `ExecContext` (pgx
  takes ctx on plain Exec — trap carried from pgstore, compiler catches).
  Invariants + symptom→suspect table live in the design walkthrough doc.
- **Server dispatch (step 9):** `cmd/server/connect.go` — `database` STRUCT
  (not interface; server assembles once, unlike the CLI's per-command
  interface), ping/close as func fields to normalize pool vs `*sql.DB`
  shapes. Scheme switch kept in lockstep with `cmd/cli/connect.go`.
- Dialect constructor names are asymmetric: pg `NewPodcastConfigRepository`/
  `NewAnalyticsRepo` vs sqlite `NewPodcastConfigRepo`/`NewAnalyticRepo`.

## Repo state

- bunny-db merged to main 2026-11-08. Full stack dialect-switchable via
  DATABASE_URL: postgres, libsql (Bunny), file. Build + vet + `go test ./...`
  green. Smokes: local file DB and Bunny both passed the login→cookie→/admin
  200 round-trip; `pages-list` + `analytics-smoke` green on Bunny (002
  applied there — users were reset, smoke user registered).

## Anytype (Notebook space `sr7yvu`, castogo collection)

- 🧩 Step 8 — SQLiteStore · 🧩 Step 9 — server dispatch + branch close
- 🔍 How the Session Store Works — design walkthrough (invariants)
- Earlier steps 4–7 + mini-steps unchanged; Step 6 doc's "delete dead
  queries" instruction still superseded by the KEPT decision above.

## Next

1. View/domains refactor — start at *Castogo Refactor Plan (Revised)* Tier 0
   (dropdown imports datastarui/utils → internal/view/utils)
2. Optional one-liners (own commits, post-merge): rename `PocastConfigRepo`
   typo; consider hash+block key pair for sessions (today secret = hash-only)
3. Someday: expired-session purge job (`expires_on < datetime('now')`)
