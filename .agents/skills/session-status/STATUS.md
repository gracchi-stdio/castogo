# Castogo — session status

Last compacted: 2026-11-07 (step 7 local smoke GREEN; steps 6+7 down to
Bunny run + commit)

## Landed decisions

- **Dead page queries KEPT (user, 2026-11-07)** — GetDescendants,
  GetPageSiblings, GetChildrenCount, GetPagePathAndChildrenCountByID,
  CountPageWithoutParent stay in BOTH corpora. Do NOT re-delete.
- sqlite JSON columns auto-shape via sqlc overrides + `domain.Scan`/`Value`.
- Wrapper conventions, `internal/repository/sqlite`: helpers (`sqliteTimeLayout`,
  `timeTo/FromSQLite`, strict; `sqliteTime` lenient updated_at only;
  `sqliteDateLayout`/`dateToSQLite` — DATE-as-TEXT "2006-01-02" UTC),
  `xFromRow`/`xFromList`, errors wrap `"verb object: %w"`, no-rows →
  `domain.ErrNotFound` EXCEPT predicates (`IsFileProcessed` → false,nil).
- Nav semantics (022): nav_order NULL=hidden, ascending, tie-break path ASC.
- Analytics upserts twin 1:1; sqlite ON CONFLICT/DO UPDATE/EXCLUDED/RETURNING
  semantics == PG; verified empirically through modernc (upsert arithmetic
  accumulates across repeated calls — repro'd while debugging).
- pg analytics ctor renamed `NewAnalyticsPostgres` → `NewAnalyticsRepo`
  (user, naming symmetry); server call site updated.

## Repo state (uncommitted, steps 6+7 complete locally)

- Step 6: corpora+wrappers+wiring+pages-list local smoke DONE.
- Step 7: analytics.sql corpus, analytics_sqlite.go wrapper, dateToSQLite,
  CLI wiring both dialects, `analytics-smoke` cmd — ALL DONE.
- Smoke verified (2026-11-07): all 9 interface methods + sqlite3/python
  arithmetic check — `2|2048`, `251|1`, `1`, `1|3`, `1`, `1` ✓
  (CLI command is `migrate`, NOT db-migrate).
- Debug detour resolved: smoke's second UpsertPodcastDaily call had been
  removed (looked like a dup) → (1,1024); restored, all green. No driver bug.
- go.mod diagnostic: `github.com/google/uuid should be indirect` — run
  `go mod tidy` before commit (leftover from step 4.5).
- sqlite migrations on Bunny: 001 applied; 002_users_id pending (=users
  table reset there).

## Anytype (Notebook space `sr7yvu`)

- Step 7 doc: id `bafyreibfojv4z4dbgo3sz643fpc3f25jn6b2gxlzjbiejmn6ib3iwagti4`
  — Part F corrected (migrate cmd name, bytes_seen column, python3 fallback).
- Step 6 doc: completion note appended. Step 6 id `bafyreicxu…tpynq3u`.

## Next

1. Bunny run (steps 6+7): `migrate` (applies 002 — users reset) →
   `pages-list` → `analytics-smoke` against libsql:// URL
2. `go mod tidy` + commit tree (steps 6+7, maybe one commit or split 6/7)
3. SP ticks (step 6 `wrtDCFL9bmkWfzQ28Kcpx`, step 7 `os5YvLT2Q01q-R0VXC2M6`)
