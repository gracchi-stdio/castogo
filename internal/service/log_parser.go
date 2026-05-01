package service

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gracchi-stdio/castogo/internal/domain"
)

func ParseLogEntry(logLine string) (*domain.RawLogEntry, error) {
	// Log format (12 pipe-delimited fields):
	// CacheStatus|StatusCode|Timestamp(ms)|BytesSent|PullZoneID|ClientIP|Referer|URL|EdgeLocation|UserAgent|RequestID|CountryCode
	parts := strings.Split(logLine, "|")
	if len(parts) < 12 {
		return nil, fmt.Errorf("invalid log line: expected 12 fields, got %d", len(parts))
	}

	statusCode, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid status code: %v", err)
	}

	timestamp, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp: %v", err)
	}

	bytesSent, err := strconv.ParseInt(parts[3], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid bytes sent: %v", err)
	}

	// parts[0] = CacheStatus (not needed)
	// parts[4] = PullZoneID  (not needed)
	// parts[6] = Referer     (not needed)
	// parts[8] = EdgeLocation(not needed)
	// parts[10]= RequestID   (not needed)

	return &domain.RawLogEntry{
		StatusCode:  statusCode,
		Timestamp:   timestamp,
		BytesSent:   bytesSent,
		ClientIP:    parts[5],
		URL:         parts[7],
		UserAgent:   parts[9],
		CountryCode: parts[11],
	}, nil
}
