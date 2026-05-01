# Podcast Distribution & Analytics Plan

## Overview

This document covers two interrelated features:

1. **RSS Feed Generation** — the universal backbone all podcast directories consume
2. **Download Analytics** — IAB 2.0-compliant tracking via CDN log processing

Both depend on one foundational truth: **one public RSS feed URL, served to all platforms.**

---

## Part 1: How Podcast Distribution Works

### The Single Feed Model

```
Your DB → RSS Feed (you generate) → Podcast directories index it
                                        ↑
                                 One-time submission per directory
```

There is no "publish to Spotify" or "publish to Apple." You generate one RSS feed, submit its URL to each directory once, and they continuously index it. All platforms read the same URL.

### What You Control vs What They Control

| You control | Platform controls |
|---|---|
| RSS feed content (episodes, metadata, cover art URLs) | When/how often they re-fetch your feed |
| Feed URL (`https://yourdomain.com/feed/podcast.xml`) | Visibility toggle in their dashboard |
| Adding new episodes to the feed | Listing/delisting your show |
| Audio file hosting (CDN) | Play counts, completion rate, demographics |

### Why Platform-Specific Feeds Don't Work

- RSS is public and unauthenticated — no reliable way to detect who's reading
- Directories share data (once Apple indexes you, Overcast finds you through Apple)
- Different feeds = fractured audience, split reviews, split rankings
- Visibility is controlled per-platform via **their** dashboard, not your feed

### The `<podcast:locked>` Tag

The Podcast Namespace defines `<podcast:locked>yes</podcast:locked>` — a blanket "do not re-host my feed without permission" signal. Respected by most modern apps. One tag, applies to all platforms.

### Directory Submission Links

Manual one-time process per platform. Your app provides the links, the user clicks through:

| Platform | Submit URL |
|---|---|
| Apple Podcasts | https://podcastsconnect.apple.com/my-podcasts/new-feed |
| Spotify | https://podcasters.spotify.com/dash/submit |
| Amazon Music | https://podcasters.amazon.com |
| YouTube Music | https://studio.youtube.com/channel/content/podcasts |
| Google Podcasts | Migrating to YouTube Music |
| Pocket Casts | https://www.pocketcasts.com/submit/ |
| Overcast | https://overcast.fm/podcasterinfo |
| Podcast Addict | https://podcastaddict.com/submit |
| Podcast Index | https://podcastindex.org/add |
| Deezer | https://podcasters.deezer.com/submission |
| Castbox | https://helpcenter.castbox.fm/portal/kb/articles/submit-my-podcast |
| Player.FM | https://player.fm/importer/feed |
| TuneIn | https://help.tunein.com/contact/add-podcast-S19TR3Sdf |
| Listen Notes | https://www.listennotes.com/submit/ |

Most smaller apps (Overcast, Castro, Pocket Casts) auto-index from Apple's directory. Once you're in Apple, you're in most places.

---

## Part 2: RSS Feed Architecture

### Design Principles

Following the 3-layer design from CLAUDE.md:

1. **DB-backed domain** — `podcast_config` + `episodes` (repository + service) — already built
2. **Feed projection** — in-memory view of publishable episodes with normalized URLs/dates
3. **Deterministic renderer** — one `encoding/xml` pipeline, no business logic

### RSS 2.0 + Namespace Extensions

The feed must support:

| Namespace | URI | Elements |
|---|---|---|
| iTunes | `http://www.itunes.com/dtds/podcast-1.0.dtd` | `<itunes:category>`, `<itunes:owner>`, `<itunes:image>`, `<itunes:explicit>`, `<itunes:author>`, `<itunes:summary>` |
| Podcast 2.0 | `https://podcastindex.org/namespace/1.0` | `<podcast:locked>`, `<podcast:guid>` |
| Google Play | `http://www.google.com/schemas/play-podcasts/1.0` | `<googleplay:category>`, `<googleplay:description>` |
| Content | `http://purl.org/rss/1.0/modules/content/` | `<content:encoded>` (full HTML description) |

### File Structure

```
internal/domain/feed.go           — XML-annotated structs for RSS serialization
internal/service/feed_service.go  — Builds feed from podcast_config + published episodes
internal/handler/feed.go          — GET /feed/podcast.xml, serves application/rss+xml
```

### Go Packages

- `encoding/xml` (stdlib) — direct control over namespace output, no magic
- No external RSS library needed — podcast RSS has specific namespace requirements that generic libraries don't handle well

### Feed URL Design

```
https://yourdomain.com/feed/podcast.xml
```

This URL is **identity** — once submitted to directories, it must never change. Registered LAST in Echo route order (before static files) so it doesn't conflict with other routes.

### Current `podcast_config` → RSS Field Mapping

| podcast_config column | RSS element |
|---|---|
| `title` | `<title>` |
| `description` | `<description>`, `<itunes:summary>` |
| `site_url` | `<link>` |
| `language` | `<language>` |
| `copyright` | `<copyright>` |
| `author_name` | `<itunes:author>` |
| `author_email` | `<managingEditor>` |
| `owner_name` | `<itunes:owner><itunes:name>` |
| `owner_email` | `<itunes:owner><itunes:email>` |
| `cover_image_url` | `<itunes:image href="...">`, `<image><url>` |
| `category` | `<itunes:category text="...">` |
| `subcategory` | `<itunes:category text="..."><itunes:category text="...">` |

### Episode → RSS `<item>` Mapping

| Episode field | RSS element |
|---|---|
| `title` | `<item><title>` |
| `slug` | `<item><link>` (public episode page URL) |
| `description` | `<item><description>`, `<content:encoded>` |
| `audio_source_url` | `<enclosure url="..." length="..." type="audio/mpeg">` |
| `duration` | `<itunes:duration>` |
| `episode_number` | `<itunes:episode>` |
| `explicit` | `<itunes:explicit>` |
| `cover_image_url` | `<itunes:image href="...">` |
| `publish_at` / `created_at` | `<pubDate>` |
| `slug` | `<guid isPermaLink="false">podcast-slug-ep-{episode_number}</guid>` |

---

## Part 3: Analytics Architecture

### The CDN Constraint

Audio files are served from Bunny.net CDN, not our Go server. The server never sees download requests. Analytics must come from **CDN log processing**, not server-side middleware.

### IAB Podcast Measurement 2.0 Compliance

Industry standard for counting downloads. Key rules:

- **24-hour deduplication window** — same IP + User-Agent + episode within 24h counts as 1 download
- **1-minute minimum** — must download ≥1 minute of audio to count
- **Byte threshold** — `bytes_threshold = file_size × (60 / duration_seconds)`
- **Exclude bots** — filter known bot IPs and user-agents
- **Apple probe exclusion** — ignore `Range: bytes=0-1` requests (Apple's connectivity check)
- **Unique listeners** — identified by IP + User-Agent combination (imperfect but industry standard)

### Bunny.net Log Delivery Methods

#### Option A: Logging API (NOT recommended)

```
GET https://logging.bunnycdn.com/{MM-DD-YY}/{pullZoneId}.log
```

- 3-day retention only — risk of data loss
- Simple HTTP GET, pipe-delimited lines
- Good for debugging, not for production analytics

#### Option B: Permanent Log Storage (RECOMMENDED — primary data source)

Logs written to Edge Storage Zone as gzip-compressed part files.

**File naming:** `pullzone-logs/<zone>/<YYYY>/<MM>/<dd>_<workerId>-<part-index>-<rand>.gzip`

**Part rotation:** Files closed and uploaded when any of:
- Part reaches 2 GB
- Part reaches 5,000,000 lines
- 120 minutes elapsed
- No writes for 1 hour
- Midnight UTC

**Pros:**
- Permanent — no data loss
- Reprocessable — fix bugs in pipeline, reprocess from scratch
- Uses existing Bunny Storage SDK
- Same pipe-delimited log format

**Cons:**
- 2-hour max delay before data appears
- Need background worker to poll for new files

#### Option C: Log Forwarding (Syslog, future enhancement)

Real-time UDP Syslog forwarding, 10-30 second delay.

**Pros:** Near real-time — could power a "currently listening" live counter

**Cons:**
- UDP — packet loss possible, no retry
- Unencrypted
- Not reliable enough for core analytics

**Recommendation:** Use for supplementary real-time features only, not as primary analytics source.

### Log Format (pipe-delimited, all options)

```
CacheStatus|StatusCode|Timestamp(ms)|BytesSent|PullZoneID|ClientIP|Referer|URL|EdgeLocation|UserAgent|RequestID|CountryCode
```

Example:
```
HIT|200|1507167062421|4506789|390|163.172.53.0|-|https://cdn.example.com/episodes/ep12.mp3|WA|AppleCoreMedia/1.0 (iPhone)|322b688bd63fb63f2babe9de30a5d262|DE
```

Key fields for analytics:
- **BytesSent** (field 4) — enables IAB byte threshold
- **URL** (field 8) — identifies which episode
- **UserAgent** (field 10) — identifies app/platform
- **CountryCode** (field 12) — GeoIP already done by Bunny (no MaxMind needed!)
- **RequestID** (field 11) — natural dedup key

**Extended logging** (contact Bunny support to enable):
- Body Bytes Sent — more accurate bytes (excluding headers)
- Range Header — critical for Apple probe detection
- Authorization Header — useful for premium podcasts

### Analytics Processing Pipeline

```
Bunny Storage Zone (gzip log parts)
         ↓
┌─────────────────────────────────────────────────────┐
│ Background Worker (every 15 minutes)                │
│                                                     │
│ 1. LIST files in pullzone-logs/<zone>/<YYYY>/<MM>/  │
│ 2. SKIP files already in processed_files table      │
│ 3. DOWNLOAD + GUNZIP new files                      │
│ 4. PARSE each pipe-delimited line                   │
│ 5. FILTER:                                          │
│    • Status not 200/206?           → SKIP           │
│    • URL not matching episodes/?   → SKIP           │
│    • Known bot User-Agent?         → SKIP           │
│ 6. ENRICH:                                         │
│    • episode_id from URL path                       │
│    • app/device/os from User-Agent (opawg data)     │
│    • country from CountryCode field (free GeoIP!)   │
│ 7. IAB DEDUP:                                       │
│    • hash = SHA1(date + ip + ua + episode_id)       │
│    • Accumulate bytes_sent per hash                 │
│    • If accumulated >= 1-minute threshold → COUNT   │
│ 8. UPSERT into 6 summary tables                    │
│ 9. MARK file as processed                          │
└─────────────────────────────────────────────────────┘
```

### CDN-Agnostic Design

```go
type LogFetcher interface {
    FetchNewLogs(ctx context.Context) ([]RawLogEntry, error)
}

type RawLogEntry struct {
    CacheStatus  string
    StatusCode   int
    Timestamp    time.Time
    BytesSent    int64
    ClientIP     string
    Referer      string
    URL          string
    EdgeLocation string
    UserAgent    string
    RequestID    string
    CountryCode  string
}
```

Implement `BunnyLogFetcher` now. Could add `CloudflareLogFetcher`, `S3LogFetcher` later. Analytics pipeline only sees `RawLogEntry`.

### Database Schema

**Processing state:**
```sql
CREATE TABLE analytics_processed_files (
    id            BIGSERIAL PRIMARY KEY,
    file_name     TEXT NOT NULL UNIQUE,
    pull_zone     TEXT NOT NULL,
    file_date     DATE NOT NULL,
    entries_count INT NOT NULL DEFAULT 0,
    processed_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

**IAB byte accumulator:**
```sql
CREATE TABLE analytics_download_accumulator (
    hash       CHAR(40) PRIMARY KEY,
    episode_id BIGINT NOT NULL,
    date       DATE NOT NULL,
    bytes_seen BIGINT NOT NULL DEFAULT 0,
    counted    BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

**6 summary tables (Castopod's design, adapted for PostgreSQL):**
```sql
-- Daily podcast totals
CREATE TABLE analytics_podcasts (
    podcast_id       BIGINT NOT NULL,
    date             DATE NOT NULL,
    duration         DECIMAL(15,3) NOT NULL DEFAULT 0,
    bandwidth        BIGINT NOT NULL DEFAULT 0,
    unique_listeners INT NOT NULL DEFAULT 1,
    hits             INT NOT NULL DEFAULT 1,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (podcast_id, date)
);

-- Per episode per day (+ age = days since publication)
CREATE TABLE analytics_podcasts_by_episode (
    podcast_id BIGINT NOT NULL,
    episode_id BIGINT NOT NULL,
    date       DATE NOT NULL,
    age        INT NOT NULL,
    hits       INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (podcast_id, episode_id, date)
);

-- Per hour of day (listening time patterns)
CREATE TABLE analytics_podcasts_by_hour (
    podcast_id BIGINT NOT NULL,
    date       DATE NOT NULL,
    hour       INT NOT NULL,
    hits       INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (podcast_id, date, hour)
);

-- Per player/app (which platform served the listener)
CREATE TABLE analytics_podcasts_by_player (
    podcast_id BIGINT NOT NULL,
    date       DATE NOT NULL,
    service    VARCHAR(128) NOT NULL DEFAULT '',
    app        VARCHAR(128) NOT NULL DEFAULT '',
    device     VARCHAR(32) NOT NULL DEFAULT '',
    os         VARCHAR(32) NOT NULL DEFAULT '',
    is_bot     BOOLEAN NOT NULL DEFAULT FALSE,
    hits       INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (podcast_id, date, service, app, device, os, is_bot)
);

-- Per country per day (Bunny gives us CountryCode for free)
CREATE TABLE analytics_podcasts_by_country (
    podcast_id   BIGINT NOT NULL,
    date         DATE NOT NULL,
    country_code VARCHAR(3) NOT NULL,
    hits         INT NOT NULL DEFAULT 1,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (podcast_id, date, country_code)
);

-- Per region per day (requires GeoIP if region-level needed)
CREATE TABLE analytics_podcasts_by_region (
    podcast_id   BIGINT NOT NULL,
    date         DATE NOT NULL,
    country_code VARCHAR(3) NOT NULL,
    region_code  VARCHAR(3) NOT NULL DEFAULT '',
    latitude     DECIMAL(8,6) DEFAULT NULL,
    longitude    DECIMAL(9,6) DEFAULT NULL,
    hits         INT NOT NULL DEFAULT 1,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (podcast_id, date, country_code, region_code)
);
```

**Note:** PostgreSQL uses `ON CONFLICT (...) DO UPDATE` instead of MySQL's `ON DUPLICATE KEY UPDATE`. This is used for all upserts.

### Dependencies We DON'T Need

| Initially thought needed | Why we don't |
|---|---|
| MaxMind GeoLite2 | Bunny provides CountryCode in log field 12 |
| `github.com/oschwald/geoip2-golang` | Same reason |
| Redis for dedup | PostgreSQL dedup table + processed_files table |
| `ipcat` IP deny list | Bunny's CDN filters most bots; we filter by User-Agent |
| Server-side audio serving middleware | Audio served from CDN, not our server |

### Dependencies We DO Need

| Package | Purpose |
|---|---|
| `encoding/xml` (stdlib) | RSS XML generation |
| `encoding/json` (stdlib) | Parse opawg user-agents data |
| `compress/gzip` (stdlib) | Decompress Bunny log files |
| `crypto/sha1` (stdlib) | Dedup hash |
| `github.com/l0wl3vel/bunny-storage-go-sdk` | Already in use — same SDK to read log files from Storage Zone |

### New Config Variables

```go
// Analytics - Bunny.net Permanent Log Storage
BunnyLogStorageEndpoint string `env:"BUNNY_LOG_STORAGE_ENDPOINT"`
BunnyLogStoragePassword  string `env:"BUNNY_LOG_STORAGE_PASSWORD"`
BunnyLogStorageZoneName  string `env:"BUNNY_LOG_STORAGE_ZONE_NAME"`
```

---

## Part 4: Platforms Module

Static platform registry (like Castopod's `Platforms.php`), stored as Go data, not in the database.

### Platform Categories

1. **Podcasting** — Apple, Spotify, Amazon, YouTube Music, Pocket Casts, etc. (~40 platforms)
2. **Social** — Bluesky, Mastodon, Discord, etc. (~20 platforms)
3. **Funding** — Patreon, Ko-fi, Buy Me a Coffee, etc. (~15 platforms)

### Data Structure

```go
type Platform struct {
    Slug      string
    Label     string
    HomeURL   string
    SubmitURL string // nil if auto-indexes or no submission process
}
```

### Database Table (user's platform links)

```sql
CREATE TABLE platform_links (
    id          BIGSERIAL PRIMARY KEY,
    podcast_id  BIGINT NOT NULL REFERENCES podcast_config(id),
    platform    TEXT NOT NULL,
    type        TEXT NOT NULL,  -- 'podcasting', 'social', 'funding'
    url         TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(podcast_id, platform)
);
```

Users store their show's URL on each platform. These generate `<podcast:social>` and `<podcast:funding>` RSS tags.

---

## Part 5: Build Stages

### Phase 1: RSS Feed (Foundation)

Everything else depends on this.

1. **RSS XML types** (`internal/domain/feed.go`)
   - Go structs with `xml:` tags for RSS 2.0 + iTunes + Podcast 2.0 namespaces
   - `RSS`, `Channel`, `Item`, `Enclosure`, `ITunesOwner`, etc.

2. **Feed service** (`internal/service/feed_service.go`)
   - Fetches `PodcastConfig` + published episodes
   - Projects into feed types (URL normalization, date formatting)
   - No business logic — pure data transformation

3. **Feed handler** (`internal/handler/feed.go`)
   - `GET /feed/podcast.xml`
   - Sets `Content-Type: application/rss+xml`
   - Caches output (optional, can add later)

4. **Feed validation**
   - Test against Podbase validator and Apple's validator
   - Golden XML fixtures for regression testing

### Phase 2: Platform Distribution

5. **Platform registry** (`internal/platform/data.go`)
   - Static Go map of all podcast/social/funding platforms
   - Labels, home URLs, submit URLs

6. **Platform links DB** (migration + repository + service)
   - Users store their show's URL per platform
   - Status tracking (submitted/approved)

7. **Distribution UI**
   - Post-creation pop-up: "Distribute your podcast"
   - Platform list with submit links + status tracking
   - RSS feed URL copy button

### Phase 3: Analytics Infrastructure

8. **Analytics tables migration**
   - All summary tables + processed_files + download_accumulator

9. **Log fetcher worker** (`internal/worker/log_fetcher.go`)
   - Background goroutine, polls Bunny Storage Zone every 15 min
   - Lists new files, downloads, gunzips

10. **Log parser** (pipe-delimited → `RawLogEntry`)
    - Parse Bunny.net log format
    - Filter non-audio, non-200/206 requests

11. **IAB processing pipeline** (`internal/service/analytics_service.go`)
    - User-Agent parsing (opawg data)
    - Byte accumulation per listener per episode per day
    - 1-minute threshold check
    - 24-hour deduplication

12. **Analytics repository** (`internal/repository/analytics_repo.go`)
    - PostgreSQL `INSERT ... ON CONFLICT DO UPDATE` upserts
    - Batch operations for efficiency

### Phase 4: Analytics UI

13. **Analytics dashboard**
    - Total downloads, unique listeners
    - By episode, by platform, by country, by time of day
    - Time range selector (7d, 30d, 90d, all time)

14. **Live counter** (optional, future)
    - Log Forwarding (Syslog) listener
    - "X people listening right now"

---

## References

- Castopod source: https://github.com/ad-aures/castopod (PHP, CodeIgniter 4)
- Bunny.net CDN Logging: https://docs.bunny.net/cdn/logging
- Bunny.net Permanent Log Storage: https://docs.bunny.net/cdn/logging/permanent-log-storage
- IAB Podcast Measurement 2.0: https://iabtechlab.com/standards/podcast-measurement-guidelines/
- Podcast Namespace: https://podcastindex.org/namespace/1.0
- Apple Podcasts RSS spec: https://help.apple.com/itc/podcasts_connect/#/itcb54353390
- opawg User Agents: https://github.com/opawg/user-agents-v2
- opawg RSS User Agents: https://github.com/opawg/podcast-rss-useragents
