# Castogo — session status

Last compacted: 2026-11-07 (refreshed at "hello" — tree had advanced past prior
compact; Part H smoke not yet verified run)

## Landed decisions

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

## Repo state (uncommitted, step 6 work)

- Query corpora done: `sql/queries/sqlite/{podcast_config,page,block}.sql`
  (Parts A–C).
- Wrappers done: `config_sqlite.go`, `page_sqlite.go` (Parts E–F).
- Wiring done: `cmd/cli/pages.go` (`pages-list` cmd) registered in
  `main.go`; connect/libsql/postgres touched (Part G).
- `go build ./...` + `go vet` clean as of 2026-11-07.
- No local `.db` file found → Part H smoke (`pages-list` vs local file DB)
  likely not yet run.
- Step 0/0.5.2 (stale pg nav query, dead queries, `sqlc.arg`) — unknown if done.

## Anytype (Notebook space `sr7yvu`)

- Step 5 + Step 6 docs are the durable instructions (Step 6: Parts A–H).
- Not re-checked this session; consult before continuing.

## Next

1. Part H: run `pages-list` against a local file DB (create temp .db, apply
   sqlite migrations, seed a page) — confirm output shape
2. Same smoke against Bunny (remote libSQL)
3. Step 0 + 0.5.2 leftovers if still pending (pg nav query fix, dead queries)
