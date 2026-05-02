# Castogo — Progress Summary

## What's Been Built

### Phase 1: RSS Feed (Foundation) ✅

A working RSS feed endpoint (`GET /feed/podcast.xml`) that produces valid XML with Apple/Spotify/Podcast 2.0 namespace support.

#### Files Created

| File | Purpose |
|---|---|
| `internal/domain/feed.go` | RSS XML types: `Enclosure`, `PubDate` (custom RFC 2822 MarshalXML), `Duration` (custom HH:MM:SS MarshalXML), `ITunesOwner`, `ITunesImage`, `ITunesCategory`, `GUID`, `Item`, `Channel`, `RSS`, `NewRSSFeed()`, `RSS.Write()` |
| `internal/service/feed_service.go` | `FeedService` with `BuildFeed()` and `buildItem()` — maps `PodcastConfig` + published episodes to `RSS` struct |
| `internal/handler/feed.go` | `RSSFeed()` handler — streams XML with `application/rss+xml; charset=utf-8` content type |
| `internal/handler/public.go` | `PunlicHandler` (typo, fix later) with route registration for `/feed/podcast.xml` |
| `cmd/server/main.go` | Wired `feedService` → `NewPublicHandler`, registered public feed route |

#### Reference Documents

| File | Purpose |
|---|---|
| `docs/distribution-analytics-plan.md` | Full planning: distribution model, RSS architecture, analytics pipeline, database schema, build stages |
| `docs/rss-feed-reference.md` | Technical reference: Go XML patterns, namespace URIs, MIME types, field mapping, Apple validation requirements |

#### Feed Validation

**Working:**
- Valid XML with all namespace declarations (itunes, podcast, content)
- Channel metadata: title, description, link, language, author, owner, image, type, `podcast:locked`, generator
- Items: title, description, GUID, pubDate, enclosure (URL + MIME type), duration, explicit
- Proper content-type: `application/rss+xml; charset=utf-8`

**Data-level gaps (not code bugs):**
- No `<itunes:category>` — database `category` column is empty
- Duration showing `00:00` — episode duration is 0 in DB
- GUID showing `ep-0` — episode number is 0

#### Bug Found & Fixed

`ListPublished(ctx, 0, 0)` passed `LIMIT 0` to PostgreSQL, returning zero rows. Fixed to `ListPublished(ctx, 500, 0)`.

### Key Architecture Decisions

- **One RSS feed URL** serves all platforms — no platform-specific feeds
- **Custom `encoding/xml` types** instead of third-party library (jbub/podcasts rejected: unmaintained, missing Podcast 2.0 namespace)
- **Bunny.net CDN** serves audio files — server never sees download requests
- **Bunny Logging API** used for analytics (3-day retention, simple HTTP GET, pipe-delimited)
- **IAB Podcast Measurement 2.0** compliance: 24-hour dedup via SHA1 hash, 1-minute minimum download via byte threshold, bot filtering via User-Agent
- **CDN-agnostic `LogFetcher` interface** — `BunnyLogFetcher` implements it, future CDN providers just need the same method signature
- **Configurable base URL** on fetcher — dependency injection enables `httptest` fake server testing
- **Graceful shutdown** — `signal.NotifyContext` + ordered cleanup (Echo → Worker → DB) with 30-second timeout
- **Rule-based User-Agent parser** — pattern table with first-match-wins, covers 16 podcast apps, 4 browsers, bot detection, device/OS inference

---

### Phase 3: Analytics Infrastructure ✅

A complete CDN log processing pipeline that fetches Bunny.net logs, filters audio requests, enforces IAB dedup rules, and upserts into 6 summary tables.

#### Files Created

| File | Purpose |
|---|---|
| `internal/domain/analytics.go` | `RawLogEntry` (8 fields from CDN log), `ParsedLogEntry` (enriched for analytics), `EpisodeMetadata` (for IAB byte threshold) |
| `internal/service/log_parser.go` | `ParseLogEntry(line)` — single line, `ParseLogEntries(reader)` — streaming via `bufio.Scanner` |
| `internal/service/log_fetcher.go` | `LogFetcher` interface + `BunnyLogFetcher` — Logging API with gzip, configurable base URL |
| `internal/service/user_agent_parser.go` | `ParseUserAgent(ua)` — rule table with 16 podcast apps, 4 browsers, bot detection, device/OS inference |
| `internal/service/analytics_service.go` | `AnalyticsService.ProcessLogsForDate()` — full pipeline: fetch → filter → enrich → IAB dedup → upsert 6 tables |
| `internal/service/analytics_worker.go` | `AnalyticsWorker` — background goroutine with `Start()`/`Stop()`, `time.Ticker`, `context.WithCancel`, `sync.WaitGroup` |
| `sql/migrations/006_create_analytics.sql` | 7 tables: `analytics_processed_files`, `analytics_download_accumulator`, `analytics_podcasts`, `_by_episode`, `_by_hour`, `_by_player`, `_by_country` |
| `sql/queries/analytics.sql` | sqlc queries: upsert accumulator, mark counted, upsert all 6 summary tables, insert/check processed files |
| `internal/repository/analytics.go` | `AnalyticsRepository` interface — 8 methods for the analytics pipeline |
| `internal/repository/postgres/analytics_postgres.go` | PostgreSQL implementation using sqlc-generated queries |
| `cmd/server/main.go` | Wired: fetcher → service → worker, `signal.NotifyContext` graceful shutdown |

#### Tests (16 passing)

| Test File | What it covers |
|---|---|
| `internal/service/log_fetcher_test.go` | `httptest.NewServer` fake Bunny API, gzip response, URL format, AccessKey header, field assertions |
| `internal/service/log_parser_test.go` | Happy path (all fields), invalid line (too few fields), invalid status code |
| `internal/service/user_agent_parser_test.go` | Apple Podcasts, Spotify, Chrome, Firefox, bots, iPad, Android tablet, Castro vs CastBox disambiguation, unknown UA |

#### Reference Documents

| File | Purpose |
|---|---|
| `docs/analytics-log-fetcher.md` | Why `*http.Client` as a field, Go date formatting, Bunny Logging API, streaming pattern, context propagation, CDN-agnostic interface |
| `docs/go-testing-reference.md` | Go testing cheat sheet: running tests, assertions, httptest pattern, gzip handler, testability, naming conventions |

#### Pipeline Flow

```
Bunny Logging API (pipe-delimited, gzip)
         ↓
BunnyLogFetcher.FetchEntries(ctx, date)
         ↓
ParseLogEntries(reader)  ← streaming, bufio.Scanner
         ↓
for each RawLogEntry:
  ├── shouldProcess?          → skip if status not 200/206
  ├── episodeMap[URL]?        → skip if not our audio file
  ├── SHA1(date+ip+ua+epID)   → IAB 24-hour dedup hash
  ├── AccumulateBytes(...)    → ON CONFLICT DO UPDATE
  ├── threshold met?          → fileSize × (60/duration) = 1-min rule
  └── Upsert 6 tables:
        ├── analytics_podcasts (daily totals + bandwidth)
        ├── analytics_podcasts_by_episode (+ age since publish)
        ├── analytics_podcasts_by_hour (listening patterns)
        ├── analytics_podcasts_by_player (service/app/device/os/bot)
        └── analytics_podcasts_by_country (geo)
```

---

## What's Next

| Phase | Scope | Status |
|---|---|---|
| Phase 1 | RSS Feed | ✅ Done |
| Phase 2 | Platform Distribution | 🔲 Planned |
| Phase 3 | Analytics Infrastructure | ✅ Done |
| Phase 4 | Analytics UI | 🔲 Next |

### Known Cleanup Items

- Fix `PunlicHandler` → `PublicHandler` typo in `internal/handler/public.go`
- Set an iTunes category in admin settings (Apple requires `<itunes:category>`)
- Test with real audio data (proper duration and file size)
- Validate against Apple's Podcasts Connect
- Consider adding `<lastBuildDate>` to channel
- `settingsSave` handler is still a stub (returns `nil`)

### Phase 3 Follow-up Items

- Upgrade from Logging API to Permanent Log Storage (infinite retention, reprocessable)
- Add `analytics_processed_files` tracking for date-level dedup (currently relies on idempotent upserts)
- Add User-Agent parsing using opawg data (more accurate than substring matching)
- Handle edge case: `time.Now()` timezone vs Bunny UTC timestamps
- Consider processing today's logs too (near real-time), not just yesterday's
- Add `BunnyLogStorageEndpoint`, `BunnyLogStoragePassword`, `BunnyLogStorageZoneName` config vars for Permanent Log Storage
