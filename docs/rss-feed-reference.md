# RSS Feed — Technical Reference

## Why We're Building Our Own (Not Using `jbub/podcasts`)

Evaluated `github.com/jbub/podcasts` (24 stars, last release Aug 2021, ~300 lines total).

**What we borrowed from its design:**
- `Feed.Write(w io.Writer)` — writes XML directly to response writer, no intermediate string allocation
- `xml.Encoder` with `Indent("", "  ")` — pretty-printed output
- `xml.Header` prefix — `<?xml version="1.0" encoding="UTF-8"?>`
- Functional options pattern for feed-level metadata
- CDATA wrapping for HTML content
- `PubDate` as custom type with `MarshalXML` for RFC 2822 formatting
- `Duration` as custom type with `MarshalXML` for `HH:MM:SS` formatting

**Why we didn't use it directly:**

| Missing feature | Why it matters |
|---|---|
| No Podcast 2.0 namespace | `<podcast:locked>` prevents unauthorized re-hosting |
| No `<itunes:episode>` | Apple requires episode numbering |
| No `<itunes:season>` | Season support |
| No `<content:encoded>` | HTML descriptions don't work without this |
| No `<itunes:image>` per item | Per-episode cover art |
| No `<itunes:type>` | Episodic vs serial podcast type |
| Wrong enclosure MIME type | Uses `"MP3"` instead of `"audio/mpeg"` — fails Apple validation |
| Unmaintained | 4 years, 24 stars — we'd be maintaining a fork anyway |

The entire package is 3 files (~300 lines). Writing our own gives full control and understanding.

---

## Go `encoding/xml` Crash Course

### Basic Marshaling

```go
type Person struct {
    XMLName xml.Name `xml:"person"`
    Name    string   `xml:"name"`
    Age     int      `xml:"age,attr"`     // "attr" = XML attribute, not element
}
// → <person age="30"><name>Alice</name></person>
```

### Key `xml` struct tag options

| Tag | Example | Output |
|---|---|---|
| Element (default) | `xml:"title"` | `<title>value</title>` |
| Attribute | `xml:"id,attr"` | `<item id="value">` |
| Omit if empty | `xml:",omitempty"` | Skipped when zero value |
| Raw inner XML | `xml:",innerxml"` | Injected as-is |
| CDATA | `xml:",cdata"` | `<![CDATA[value]]>` |
| No name (inherit parent) | `xml:",chardata"` | Text content of parent |
| Dash (skip field) | `xml:"-"` | Not serialized at all |

### Namespaces in Go XML

Go uses the **prefix** in struct tags for namespaced elements:

```go
type Channel struct {
    XMLName xml.Name `xml:"channel"`
    Title   string   `xml:"title"`                    // <title>
    Author  string   `xml:"itunes:author,omitempty"`  // <itunes:author>
}
```

The namespace URI is declared on the root `<rss>` element:

```go
type Feed struct {
    XMLName       xml.Name `xml:"rss"`
    Version       string   `xml:"version,attr"`
    ItunesNS      string   `xml:"xmlns:itunes,attr"`
    PodcastNS     string   `xml:"xmlns:podcast,attr"`
    ContentNS     string   `xml:"xmlns:content,attr"`
    Channel       *Channel
}
```

This produces:

```xml
<rss version="2.0"
     xmlns:itunes="http://www.itunes.com/dtds/podcast-1.0.dtd"
     xmlns:podcast="https://podcastindex.org/namespace/1.0"
     xmlns:content="http://purl.org/rss/1.0/modules/content/">
  <channel>
    <title>...</title>
    <itunes:author>...</itunes:author>
  </channel>
</rss>
```

### Custom MarshalXML (for special formatting)

Some fields need non-standard XML representation. Go lets you control this via `MarshalXML`:

```go
// PubDate must be RFC 2822 format: "Mon, 02 Jan 2006 15:04:05 -0700"
type PubDate struct {
    time.Time
}

func (p PubDate) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
    e.EncodeToken(start)                              // <pubDate>
    e.EncodeToken(xml.CharData(p.Format(rfc2822)))    // Mon, 15 Jan 2025 10:00:00 +0000
    e.EncodeToken(xml.EndElement{Name: start.Name})   // </pubDate>
    return nil
}
```

### Writing to `io.Writer` (not marshaling to string)

```go
func (f *Feed) Write(w io.Writer) error {
    w.Write([]byte(xml.Header))           // <?xml version="1.0" encoding="UTF-8"?>\n
    enc := xml.NewEncoder(w)
    enc.Indent("", "  ")                   // pretty print
    return enc.Encode(f)                   // writes all XML tokens
}
```

This is better than `xml.Marshal` because:
- No intermediate `[]byte` or `string` allocation
- Streams directly to the HTTP response writer
- Works with Echo's `c.Response()` (v5 returns the raw `http.ResponseWriter`)

---

## Namespace URIs We Need

```go
const (
    NSItunes  = "http://www.itunes.com/dtds/podcast-1.0.dtd"
    NSPodcast = "https://podcastindex.org/namespace/1.0"
    NSContent = "http://purl.org/rss/1.0/modules/content"
    NSGoogle  = "http://www.google.com/schemas/play-podcasts/1.0"  // optional
)
```

---

## MIME Types for Enclosures

```go
const (
    MIME_mp3  = "audio/mpeg"     // NOT "MP3" or "audio/mp3"
    MIME_m4a  = "audio/x-m4a"    // AAC audio
    MIME_wav  = "audio/wav"
    MIME_ogg  = "audio/ogg"
    MIME_flac = "audio/flac"
)
```

Apple validation expects exact MIME types. `audio/mpeg` for MP3 is the most common.

---

## Date Formats

```go
const (
    RFC2822 = "Mon, 02 Jan 2006 15:04:05 -0700"  // <pubDate>
)
```

RSS 2.0 spec requires RFC 2822 for `<pubDate>`. Go's `time.RFC1123Z` is equivalent.

---

## Field Mapping Reference

### podcast_config → Channel

| DB column | XML element | Notes |
|---|---|---|
| `title` | `<title>` | Required |
| `description` | `<description>` | Plain text |
| `description` | `<itunes:summary>` | Can be longer, CDATA |
| `site_url` | `<link>` | Podcast website URL |
| `language` | `<language>` | ISO 639 (`en-us`, `fr-fr`, etc.) |
| `copyright` | `<copyright>` | Free text |
| `author_name` | `<itunes:author>` | Show author |
| `author_email` | `<managingEditor>` | Standard RSS |
| `owner_name` | `<itunes:owner><itunes:name>` | iTunes directory owner |
| `owner_email` | `<itunes:owner><itunes:email>` | iTunes directory owner |
| `cover_image_url` | `<itunes:image href="...">` | 3000x3000 max, JPEG/PNG |
| `cover_image_url` | `<image><url>` | Standard RSS `<image>` block |
| `category` | `<itunes:category text="...">` | Apple category list |
| `subcategory` | Nested `<itunes:category>` | Inside parent category |

### Episode → Item

| Episode field | XML element | Notes |
|---|---|---|
| `title` | `<title>` | Required |
| `slug` | `<link>` | Episode page URL (if public pages exist) |
| `description` | `<description>` | Plain text summary |
| `description` | `<content:encoded>` | Full HTML description, CDATA |
| `audio_source_url` | `<enclosure url="..." length="..." type="audio/mpeg">` | Required, `length` = file size in bytes |
| `audio_metadata.file_size` | `<enclosure length="...">` | Bytes, not KB/MB |
| `audio_metadata.mimetype` | `<enclosure type="...">` | `audio/mpeg` for MP3 |
| `duration` | `<itunes:duration>` | `HH:MM:SS` format |
| `episode_number` | `<itunes:episode>` | Integer |
| `explicit` | `<itunes:explicit>` | `true` or `false` |
| `cover_image_url` | `<itunes:image href="...">` | Per-episode cover (optional) |
| `publish_at` or `created_at` | `<pubDate>` | RFC 2822 |
| slug + episode_number | `<guid isPermaLink="false">` | Must be unique, must never change |

### Static Elements (hardcoded or config-driven)

| Element | Value | Notes |
|---|---|---|
| `<podcast:locked>` | `yes` | Prevent unauthorized re-hosting |
| `<itunes:type>` | `episodic` | or `serial` for serialized shows |
| `<generator>` | `Podlog v1.0` | Identifies the feed generator |

---

## Apple Validation Requirements (Minimum Valid Feed)

1. `<title>` — required
2. `<description>` — required (or `<itunes:summary>`)
3. `<link>` — required, must be valid URL
4. `<language>` — required, ISO 639
5. `<itunes:image>` — required, min 1400x1400, max 3000x3000, JPEG or PNG
6. `<itunes:category>` — required, must be from Apple's approved list
7. `<itunes:explicit>` — required on both channel and item level
8. `<enclosure>` — required on every `<item>`, must have `url`, `length`, `type`
9. `<guid>` — required on every `<item>`, must be unique and permanent
10. `<pubDate>` — required on every `<item>`

Source: https://help.apple.com/itc/podcasts_connect/#/itcb54353390

---

## File Structure (What We'll Build)

```
internal/domain/feed.go           — XML types (Feed, Channel, Item, Enclosure, etc.)
internal/service/feed_service.go  — Builds feed from DB data
internal/handler/feed.go          — Echo handler: GET /feed/podcast.xml
```
