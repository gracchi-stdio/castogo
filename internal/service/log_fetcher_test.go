package service

// Test files live in the same package (package service) so they can access
// unexported functions and types. The file must be named *_test.go — Go's
// test runner only looks for that pattern.

// WHY THESE IMPORTS:
//   "testing"        — Go's built-in test framework (provides *testing.T)
//   "context"        — context.Background() for the fetcher call
//   "compress/gzip"  — to gzip-compress the fake response (Bunny sends gzip)
//   "net/http"       — HTTP types (ResponseWriter, Request)
//   "net/http/httptest" — fake HTTP server for testing
//   "time"           — time.Date() to create a specific date
//   "strings"        — to check the request path

import (
	"compress/gzip"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Test function naming convention: func TestXxx(t *testing.T)
// - Must start with "Test" (uppercase)
// - Xxx describes what's being tested
// - Receives *testing.T (the test's control handle)
//
// Run with: go test ./internal/service/ -run TestBunnyLogFetcher -v

// A sample Bunny CDN log line (pipe-delimited, 12 fields).
// This is what the real Bunny API returns — we use it in our fake server.
const testLogLine = "HIT|200|1507167062421|412|390|163.172.53.0|-|https://cdn.example.com/episodes/ep1.mp3|WA|Mozilla/5.0|abc123|DE"

func TestBunnyLogFetcher_FetchEntries(t *testing.T) {
	// ============================================================
	// STEP 1: Create a fake HTTP server
	// ============================================================
	// httptest.NewServer starts a real HTTP server on a random port.
	// The handler function runs every time the fetcher makes a request.
	// This replaces the real Bunny API — we control exactly what comes back.

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// --- VERIFY: The fetcher sent the right request ---

		// Check AccessKey header (Bunny requires this)
		if r.Header.Get("AccessKey") != "test-api-key" {
			t.Errorf("expected AccessKey 'test-api-key', got '%s'", r.Header.Get("AccessKey"))
		}

		// Check URL path format: /{MM}-{DD}-{YY}/{pullZoneID}.log
		// For date Jan 15, 2026 → path should be "/01-15-26/test-zone.log"
		if !strings.HasSuffix(r.URL.Path, ".log") {
			t.Errorf("expected path ending in .log, got %s", r.URL.Path)
		}

		// --- RESPOND: Send gzip-compressed log data ---

		// The real Bunny API sends gzip-compressed responses.
		// gzip.NewWriter wraps the ResponseWriter — everything written
		// to `gz` gets compressed before being sent to the client.
		gz := gzip.NewWriter(w)
		gz.Write([]byte(testLogLine + "\n"))
		gz.Close() // MUST close — flushes the gzip compressor
	}))
	defer server.Close() // Shut down the fake server when test ends

	// ============================================================
	// STEP 2: Create the fetcher, pointing at our fake server
	// ============================================================
	// server.URL is something like "http://127.0.0.1:54321"
	// We pass it as baseURL instead of "https://logging.bunnycdn.com"
	// This is dependency injection — the fetcher doesn't know it's talking to a fake.

	fetcher := NewBunnyLogFetcher(
		server.URL,      // baseURL: fake server instead of real Bunny
		"test-api-key",  // accessKey: any string, handler checks it above
		"test-zone",     // pullZoneID: appears in the URL path
		server.Client(), // httpClient: configured to talk to the test server
	)

	// ============================================================
	// STEP 3: Call the method under test
	// ============================================================
	testDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	entries, err := fetcher.FetchEntries(context.Background(), testDate)

	// ============================================================
	// STEP 4: Assert the results
	// ============================================================

	// t.Fatalf stops the test immediately — use when further checks would crash.
	// t.Errorf reports failure but continues — use for non-critical checks.
	// The %v format verb prints any value in a human-readable way.

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	// Check individual fields — the parser should have extracted these
	// from the pipe-delimited test log line.
	entry := entries[0]

	if entry.StatusCode != 200 {
		t.Errorf("expected StatusCode 200, got %d", entry.StatusCode)
	}
	if entry.Timestamp != 1507167062421 {
		t.Errorf("expected Timestamp 1507167062421, got %d", entry.Timestamp)
	}
	if entry.BytesSent != 412 {
		t.Errorf("expected BytesSent 412, got %d", entry.BytesSent)
	}
	if entry.ClientIP != "163.172.53.0" {
		t.Errorf("expected ClientIP '163.172.53.0', got '%s'", entry.ClientIP)
	}
	if entry.URL != "https://cdn.example.com/episodes/ep1.mp3" {
		t.Errorf("expected URL 'https://cdn.example.com/episodes/ep1.mp3', got '%s'", entry.URL)
	}
	if entry.UserAgent != "Mozilla/5.0" {
		t.Errorf("expected UserAgent 'Mozilla/5.0', got '%s'", entry.UserAgent)
	}
	if entry.RequestID != "abc123" {
		t.Errorf("expected RequestID 'abc123', got '%s'", entry.RequestID)
	}
	if entry.CountryCode != "DE" {
		t.Errorf("expected CountryCode 'DE', got '%s'", entry.CountryCode)
	}
}
