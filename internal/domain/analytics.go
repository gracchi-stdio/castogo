package domain

type RawLogEntry struct {
	StatusCode  int
	Timestamp   int64 // milliseconds since epoch
	BytesSent   int64
	ClientIP    string
	URL         string
	UserAgent   string
	CountryCode string
}
