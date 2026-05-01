# Analytics: Log Fetcher Reference

## Why `*http.Client` as a Field

Never use `http.Get()` or `http.DefaultClient` in production. Three reasons:

1. **Timeouts** — `http.DefaultClient` has no timeout. A hung request blocks forever.
2. **Testability** — inject a client with fake transport to test without hitting real APIs.
3. **Connection reuse** — `http.Client` pools connections internally. One client, many requests.

```go
// In main.go
client := &http.Client{Timeout: 60 * time.Second}
fetcher := service.NewBunnyLogFetcher(client, apiKey, zoneID)
```

## Go Date Formatting

Go uses reference time `Mon Jan 2 15:04:05 MST 2006`, not Python-style `{MM}-{DD}-{YY}`.

```go
time.Now().Format("01-02-06")       // → "05-01-26" (MM-DD-YY)
time.Now().Format("2006-01-02")      // → "2026-05-01" (YYYY-MM-DD)
```

## Bunny Logging API

```
GET https://logging.bunnycdn.com/{MM-DD-YY}/{pull_zone_id}.log
Header: AccessKey: your-api-key
Header: Accept-Encoding: gzip
```

Response: pipe-delimited lines, same format as `ParseLogEntry`.

3-day retention. For permanent storage, use Permanent Log Storage (gzip files in Edge Storage Zone).

## Streaming Pattern for Log Parsing

Use `bufio.Scanner` to read line by line — don't load entire files into memory.

```go
scanner := bufio.NewScanner(reader)
for scanner.Scan() {
    entry, err := ParseLogEntry(scanner.Text())
    if err != nil {
        continue // skip malformed lines
    }
    entries = append(entries, *entry)
}
return entries, scanner.Err()
```

## Context Propagation

All I/O operations should accept `context.Context` as the first parameter:

```go
func (f *BunnyLogFetcher) FetchEntries(ctx context.Context, date time.Time) ([]domain.RawLogEntry, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    // ...
}
```

This allows the caller to cancel the request (timeout, shutdown, etc.).

## CDN-Agnostic Interface

The analytics service defines the interface it needs:

```go
// In analytics_service.go
type LogFetcher interface {
    FetchEntries(ctx context.Context, date time.Time) ([]domain.RawLogEntry, error)
}
```

`BunnyLogFetcher` implements it implicitly (structural typing). Future implementations (Cloudflare, S3) just need the same method signature.
