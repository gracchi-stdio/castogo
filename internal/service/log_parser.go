package service

import (
	"bufio"
	"fmt"
	"io"
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

	return &domain.RawLogEntry{
		StatusCode:  statusCode,
		Timestamp:   timestamp,
		BytesSent:   bytesSent,
		ClientIP:    parts[5],
		URL:         parts[7],
		UserAgent:   parts[9],
		RequestID:   parts[10],
		CountryCode: parts[11],
	}, nil
}

func ParseLogEntries(reader io.Reader) ([]domain.RawLogEntry, error) {
	scanner := bufio.NewScanner(reader)
	var entries []domain.RawLogEntry
	for scanner.Scan() {
		entry, err := ParseLogEntry(scanner.Text())
		if err != nil {
			continue // skip malformed lines
		}
		entries = append(entries, *entry)
	}
	return entries, scanner.Err()
}
