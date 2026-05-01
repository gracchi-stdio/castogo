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
- **Bunny Permanent Log Storage** chosen for analytics (permanent gzip files in Edge Storage Zone)
- **IAB Podcast Measurement 2.0** compliance planned: 24-hour dedup, 1-minute minimum download, byte thresholds, bot filtering

---

## What's Next

| Phase | Scope | Status |
|---|---|---|
| Phase 1 | RSS Feed | ✅ Done |
| Phase 2 | Platform Distribution | 🔲 Next |
| Phase 3 | Analytics Infrastructure | 🔲 Future |
| Phase 4 | Analytics UI | 🔲 Future |

### Known Cleanup Items

- Fix `PunlicHandler` → `PublicHandler` typo in `internal/handler/public.go`
- Set an iTunes category in admin settings (Apple requires `<itunes:category>`)
- Test with real audio data (proper duration and file size)
- Validate against Apple's Podcasts Connect
- Consider adding `<lastBuildDate>` to channel
- `settingsSave` handler is still a stub (returns `nil`)
