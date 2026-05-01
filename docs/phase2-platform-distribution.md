# Phase 2: Platform Distribution Plan

## Goal

Build a **platform distribution UI** that helps podcasters get their show listed on directories (Apple, Spotify, Amazon, etc.) by providing:

1. Their **RSS feed URL** with a copy-to-clipboard button
2. A **directory of podcast platforms** with direct submit links
3. A way to store their **per-platform show URLs** (e.g., "Here's my show on Spotify")

This is NOT an API integration — directory submission is always manual. We're building a helpful guide and link tracker.

---

## Architecture Overview

```
internal/domain/platform.go       — Types: Platform, PlatformLink, PlatformType
internal/domain/platform_data.go  — Static Go maps of ~40 podcast/social/funding platforms
sql/migrations/006_...            — platform_links table
sql/queries/platform_links.sql    — sqlc queries for platform_links CRUD
internal/repository/platform_link_repo.go  — Interface
internal/repository/postgres/platform_link_repo.go — Implementation
internal/service/distribution_service.go   — Business logic
internal/handler/admin_distribution.go     — Handler + SSE endpoints
internal/view/distribution_page.templ      — Distribution UI page
cmd/server/main.go              — Wire everything
```

---

## Step-by-Step Build Order

### Step 1: Domain Types + Static Platform Registry

We put everything in `internal/domain/platform.go` to avoid splitting types across packages. The domain package currently only imports `time` — platform types are simple strings and structs with no external dependencies, so this keeps things cohesive.

**File:** `internal/domain/platform.go`

```go
package domain

type PlatformType string

const (
    PlatformTypePodcasting PlatformType = "podcasting"
    PlatformTypeSocial     PlatformType = "social"
    PlatformTypeFunding    PlatformType = "funding"
)

// Platform is static metadata — no database, defined in code.
type Platform struct {
    Slug      string       // e.g., "apple-podcasts"
    Label     string       // e.g., "Apple Podcasts"
    HomeURL   string       // e.g., "https://www.apple.com/apple-podcasts/"
    SubmitURL string       // e.g., "https://podcastsconnect.apple.com/" (empty if auto-indexes)
    Type      PlatformType // podcasting, social, or funding
}

// PlatformLink is a user's stored URL for a platform (database-backed).
type PlatformLink struct {
    ID         int64
    PodcastID  int64
    Platform   string    // slug matching a Platform in the registry
    Type       string    // "podcasting", "social", "funding"
    URL        string    // user's show/page URL on that platform
    CreatedAt  time.Time
    UpdatedAt  time.Time
}

type CreatePlatformLink struct {
    PodcastID int64
    Platform  string
    Type      string
    URL       string
}

type UpdatePlatformLink struct {
    ID  int64
    URL *string  // only URL is editable
}
```

**File:** `internal/domain/platform_data.go`

Static Go maps of platform metadata. Separate file for data volume (this will be hundreds of lines of platform entries).

```go
package domain

// All podcasting directories with submit URLs
var PodcastingPlatforms = map[string]Platform{
    "apple-podcasts": {Slug: "apple-podcasts", Label: "Apple Podcasts", HomeURL: "https://www.apple.com/apple-podcasts/", SubmitURL: "https://podcastsconnect.apple.com/my-podcasts/new-feed", Type: PlatformTypePodcasting},
    "spotify":        {Slug: "spotify", Label: "Spotify", HomeURL: "https://www.spotify.com/", SubmitURL: "https://podcasters.spotify.com/dash/submit", Type: PlatformTypePodcasting},
    // ... ~15-20 entries
}

var SocialPlatforms = map[string]Platform{ /* ... */ }
var FundingPlatforms = map[string]Platform{ /* ... */ }

// PlatformsByType returns all platforms of a given type, sorted by label.
func PlatformsByType(t PlatformType) []Platform { /* ... */ }
```

The data maps:
- `PodcastingPlatforms` — ~15-20 major directories with submit URLs (Apple, Spotify, Amazon, YouTube Music, Google Podcasts, Pocket Casts, Overcast, Castro, Podchaser, iHeartRadio, TuneIn, Deezer, Podbean, Pandora, Samsung, etc.)
- `SocialPlatforms` — Bluesky, Mastodon, X/Twitter, Instagram, Facebook, TikTok, YouTube, LinkedIn, Discord, Reddit, Threads
- `FundingPlatforms` — Patreon, Ko-fi, Buy Me a Coffee, PayPal, GitHub Sponsors, Liberapay

Each map is `map[string]Platform` keyed by slug.

**Why static data in domain?** The `internal/domain/` package already holds all our types. The platform data is pure Go constants — no database, no imports. Keeping `Platform`, `PlatformType`, and `PlatformLink` together means one import for the service layer. Matches Castopod's approach (static data, not DB-driven).

**Analytics naming note:** The platform slugs defined here (e.g., `"apple-podcasts"`, `"spotify"`) should align with the `service` column values in the future `analytics_podcasts_by_player` table (Phase 3). When we parse User-Agents using opawg data, the platform identification will map back to these slugs. Keep slug naming consistent.

### Step 2: Database Migration

**File:** `sql/migrations/006_create_platform_links.sql`

```sql
CREATE TABLE platform_links (
    id          BIGSERIAL PRIMARY KEY,
    podcast_id  BIGINT NOT NULL REFERENCES podcast_config(id) ON DELETE CASCADE,
    platform    TEXT    NOT NULL,
    type        TEXT    NOT NULL,  -- 'podcasting', 'social', 'funding'
    url         TEXT    NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(podcast_id, platform)
);

CREATE INDEX idx_platform_links_podcast_id ON platform_links(podcast_id);
```

The `UNIQUE(podcast_id, platform)` constraint ensures one link per platform per podcast. `ON DELETE CASCADE` on the FK means if the podcast config is deleted, links go away.

### Step 3: sqlc Queries

**File:** `sql/queries/platform_links.sql`

```sql
-- name: CreatePlatformLink :one
INSERT INTO platform_links (podcast_id, platform, type, url)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetPlatformLink :one
SELECT * FROM platform_links
WHERE podcast_id = $1 AND platform = $2;

-- name: ListPlatformLinks :many
SELECT * FROM platform_links
WHERE podcast_id = $1
ORDER BY type, platform;

-- name: UpdatePlatformLink :one
UPDATE platform_links
SET url = COALESCE($2, url),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeletePlatformLink :exec
DELETE FROM platform_links WHERE id = $1;

-- name: UpsertPlatformLink :one
INSERT INTO platform_links (podcast_id, platform, type, url)
VALUES ($1, $2, $3, $4)
ON CONFLICT (podcast_id, platform)
DO UPDATE SET url = EXCLUDED.url, updated_at = NOW()
RETURNING *;
```

Then run `just generate` to produce `internal/db/platform_links.sql.go`.

### Step 4: Repository Layer

**Interface:** `internal/repository/platform_link_repo.go`

```go
type PlatformLinkRepository interface {
    Upsert(ctx context.Context, podcastID int64, platform, platformType, url string) (*domain.PlatformLink, error)
    ListByPodcast(ctx context.Context, podcastID int64) ([]*domain.PlatformLink, error)
    Delete(ctx context.Context, id int64) error
}
```

**Implementation:** `internal/repository/postgres/platform_link_repo.go`

Follows the same pattern as `podcast_config_repo.go` — wraps `db.Queries`, converts to domain types via helper.

### Step 5: Distribution Service

**File:** `internal/service/distribution_service.go`

```go
type DistributionService struct {
    linkRepo    repository.PlatformLinkRepository
    configRepo  repository.PodcastConfigRepository
}

// DistributionData is everything the distribution page template needs.
type DistributionData struct {
    FeedURL  string              // e.g., "https://yoursite.com/feed/podcast.xml"
    Sections []PlatformSection   // grouped platforms with link status
}

// PlatformSection groups platforms by type for the UI.
type PlatformSection struct {
    Type      PlatformType
    Label     string             // "Podcast Directories", "Social", "Funding"
    Platforms []PlatformWithLink // static platform + user's stored link (if any)
}

// PlatformWithLink merges a static platform with the user's stored link.
type PlatformWithLink struct {
    Platform domain.Platform
    Link     *domain.PlatformLink // nil if not connected
    Connected bool
}

func (s *DistributionService) GetDistributionData(ctx context.Context) (*DistributionData, error)
func (s *DistributionService) SaveLink(ctx context.Context, podcastID int64, platform, platformType, url string) error
func (s *DistributionService) RemoveLink(ctx context.Context, id int64) error
```

The key concept: `GetDistributionData` merges static platform data with stored user links. For each platform, the UI checks:
- `Connected == true` → show stored URL, Edit/Remove buttons
- `Platform.SubmitURL != ""` → show Submit button
- Neither? → just show the platform card without action buttons

### Step 6: Handler

**File:** `internal/handler/admin_distribution.go`

Routes (registered on the admin group):
| Method | Path | Purpose |
|---|---|---|
| GET | `/admin/distribution` | Full page render (distribution page) |
| POST (SSE) | `/admin/distribution/link` | Save a platform link (upsert) |
| DELETE (SSE) | `/admin/distribution/link` | Remove a platform link |

The SSE endpoints use Datastar patterns from `helpers.go`:
- `readSignals(c, &input)` to read form data
- `sse(c).PatchElements(html)` to update the UI inline
- `sse(c).MarshalAndPatchSignals(map)` for feedback

Add to `AdminHandler`:
- `distributionService` field
- Route registration in `RegisterRoutes`
- Wire in `cmd/server/main.go`

### Step 7: Distribution UI

**File:** `internal/view/distribution_page.templ`

The page has three sections:

#### Section 1: RSS Feed URL (top, prominent)

```
┌─────────────────────────────────────────────┐
│ 📡 Your RSS Feed URL                        │
│                                             │
│ https://yoursite.com/feed/podcast.xml       │
│                                     [Copy]  │
│                                             │
│ Submit this URL to any podcast directory.   │
│ Most platforms auto-index from Apple.       │
└─────────────────────────────────────────────┘
```

- Copy-to-clipboard via Datastar action: `navigator.clipboard.writeText($feedUrl)`
- A `data-on:click` on the copy button that sets a brief "Copied!" signal

#### Section 2: Podcast Directories (main section)

A card grid, one card per platform:

```
┌──────────────────────┐  ┌──────────────────────┐
│ 🎧 Apple Podcasts    │  │ 🎧 Spotify           │
│                      │  │                      │
│ [Submit →] [Connect] │  │ [Submit →] [Connect] │
│                      │  │                      │
│ ✅ Connected         │  │ 🔲 Not connected     │
│ apple.com/...        │  │                      │
└──────────────────────┘  └──────────────────────┘
```

Each card shows:
- Platform icon/logo (or fallback icon)
- Platform name
- **Submit** button → opens platform's submit URL in new tab
- **Connect** button → opens a dialog/sheet to paste the user's show URL
- If connected: shows the stored URL with an **Edit** and **Remove** button

The "Connect" flow:
1. User clicks "Connect" → a Shoelace `<sl-dialog>` slides in
2. User pastes their show URL on that platform
3. POST to `/admin/distribution/link` via SSE
4. Server upserts the link, patches the card to show "Connected" state

#### Section 3: Social & Funding Links (bottom, collapsible)

Similar cards for social platforms and funding platforms. These are secondary — shown in a collapsible section or tabs.

For social links, these would eventually generate `<podcast:social>` tags in the RSS feed.
For funding links, these would generate `<podcast:funding>` tags.

### Step 8: Sidebar Navigation Update

**File:** `internal/view/admin_layout.templ`

Add "Distribution" to the sidebar under the "Podcast" section:

```go
{
    Title: "Podcast",
    Items: []sidebar.SidebarItem{
        {Title: "All Episodes", Href: "/admin/episodes", Icon: "list"},
        {Title: "New Episode", Href: "/admin/episodes/create", Icon: "circle-plus"},
        {Title: "Distribution", Href: "/admin/distribution", Icon: "radio"},  // NEW
    },
},
```

### Step 9: Wiring in main.go

**File:** `cmd/server/main.go`

```go
// Repositories
platformLinkRepo := postgres.NewPlatformLinkRepo(db)

// Services
distributionService := service.NewDistributionService(platformLinkRepo, podcastRepo)

// Admin handler — add distributionService param
adminHandler := handler.NewAdminHandler(
    storageService,
    episodeService,
    audioProcessor,
    settingsService,
    distributionService,  // NEW
)
```

---

## Out of Scope for Phase 2

These are deliberately deferred:

- **RSS `<podcast:social>` and `<podcast:funding>` tags** — will be added when we integrate platform links into the feed service (can be a quick follow-up)
- **Platform status polling** (checking if a show is actually listed) — requires platform-specific API integrations
- **Auto-submission** — most platforms don't have APIs for this
- **Analytics per platform** — depends on Phase 3/4 (Bunny log parsing + User-Agent detection)
- **Platform icons/logos** — can use Shoelace icons or emoji for now, proper SVGs later

---

## Future Integration: RSS Feed Enhancement

Once platform links exist, we can enhance the RSS feed:

```xml
<!-- From funding platform links -->
<podcast:funding url="https://patreon.com/myshow">Support on Patreon</podcast:funding>

<!-- From social platform links -->
<podcast:social url="https://bsky.app/profile/myshow" platform="bluesky">Follow on Bluesky</podcast:social>
```

This is a small addition to `feed_service.go` — query platform links alongside podcast config, filter by type, add to channel struct.

---

## Estimated Complexity

| Step | Effort | Notes |
|---|---|---|
| 1. Domain types + platform data | Medium | Types are easy; ~40 platform entries is tedious but straightforward |
| 2. Migration | Small | Single table |
| 3. sqlc queries | Small | 6 queries, standard CRUD + upsert |
| 4. Repository | Small | Interface + postgres impl, follows existing pattern |
| 5. Service | Medium | Merge logic (static data + stored links), DistributionData struct |
| 6. Handler | Small | 3 endpoints, standard SSE pattern |
| 7. UI | Large | Most work — distribution page with cards, dialogs, SSE interactions |
| 8. Sidebar | Trivial | One line addition |
| 9. Wiring | Small | Follow existing pattern in main.go |

**Recommended build order:** 1 → 2 → 3 → 4 → 5 → 8 → 9 → 6 → 7

Build the data foundation first, wire it up, then build the UI last.
