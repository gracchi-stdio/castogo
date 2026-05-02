package domain

import "time"

type RawLogEntry struct {
	StatusCode  int
	Timestamp   int64 // milliseconds since epoch
	BytesSent   int64
	ClientIP    string
	URL         string
	UserAgent   string
	CountryCode string
	RequestID   string // unique server RequestID
}

type ParsedLogEntry struct {
	EpisodeID   int64
	PodcastID   int64
	Timestamp   time.Time
	BytesSent   int64
	ClientIP    string
	UserAgent   string
	CountryCode string
	Date        time.Time // date only, for daily aggregation
	Service     string    // e.g. "playback", "feed"
	App         string    // e.g. "web", "mobile"
	Device      string    // e.g. "desktop", "smartphone", "tablet"
	OS          string    // e.g. "Windows", "iOS", "Android"
	IsBot       bool      // whether the user agent is identified as a bot/crawler
}

type EpisodeMetadata struct {
	EpisodeID   int64
	PodcastID   int64
	FileSize    int64
	Duration    int64 // in seconds
	PublishedAt time.Time
}
