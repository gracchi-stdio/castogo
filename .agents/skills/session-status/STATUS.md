# Castogo — session status

Last compacted: 2026-11-07 (Step 0 finished; dead queries KEPT per user
decision; only Bunny smoke + commit remain for Step 6)

## Landed decisions

- **Dead page queries KEPT (user, 2026-11-07)** — GetDescendants,
  GetPageSiblings, GetChildrenCount, GetPagePathAndChildrenCountByID,
  CountPageWithoutParent stay in BOTH corpora for future child-page
  features. Do NOT re-delete in cleanups (overrides step-6 scoping finding).
- sqlite JSON columns auto-shape via sqlc overrides + `domain.Scan`/`Value`
  (`AudioMetadata`, `PageMetadata`; `page_blocks.content` → `json.RawMessage`).
  - `Value()` returns string — pgx writes it verbatim for json/jsonb
    (verified in pgtype/json.go); sqlite stores TEXT.
  - nil `*T` param → SQL NULL on both stacks (pgx `isNilDriverValuer`,
    database/sql `callValuerValue` guard).
  - `go_type` override beats `sqlc.narg` → metadata narg is non-pointer;
    Update nil = zero-value overwrite (mirrors pg wrapper, commented).
- Wrapper conventions, `internal/repository/sqlite`:
  - `helpers.go`: `sqliteTimeLayout`, `timeTo/FromSQLite` (strict),
    `timePtrTo/FromSQLite`, `sqliteTime` (lenient — **updated_at only**),
    `orZero`, `toPtr`.
  - `xFromRow` + `xFromList` per repo; errors wrap `"verb object: %w"`.
  - No-rows sentinel: `sql.ErrNoRows` → `domain.ErrNotFound`.
- Nav semantics (022): `nav_order NULL` = hidden, else ascending display
  order, **tie-break path ASC** — `GetPublishedTopLevelPages` sorts
  `nav_order ASC, path ASC` in both corpora (was `sort_order`).

## Repo state (uncommitted, step 6 work)

- Corpora: `sql/queries/sqlite/{podcast_config,page,block}.sql` (A–C);
  pg+sqlite nav fix applied. `just generate` clean; build + vet clean
  (2026-11-07); restored queries present in both generated packages.
- Wrappers: `config_sqlite.go`, `page_sqlite.go` (E–F). Wiring:
  `cmd/cli/pages.go` (`pages-list`) registered (G).
- Part H local smoke DONE (user): `pages-list file:/tmp/castogo-local.db`
  → 1 page, hidden/published//sample.
- sqlite migrations on Bunny: 001 applied (step 2); **002_users_id pending**
  — DROP+recreate users, wipes any rows there.

## Anytype (Notebook space `sr7yvu`)

- Step 6 doc: "🧩 Step 6 — PodcastConfig + Page (sqlite) + nav query fix +
  Scan/Value" id `bafyreicxu…tpynq3u` — durable instructions. NOTE: its
  "delete the 5 dead queries" instruction is superseded by the user
  decision above; doc not yet annotated.

## Next

1. Bunny smoke (user runs, URL has authToken): `migrate` (applies 002 —
   users reset) then `pages-list --database-url libsql://…`
2. Step 6 SP tick (`wrtDCFL9bmkWfzQ28Kcpx`) + commit tree
