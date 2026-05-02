package service

import "testing"

func TestParseLogEntry(t *testing.T) {
	line := "HIT|200|1507167062421|412|390|163.172.53.0|-|https://cdn.example.com/ep1.mp3|WA|Mozilla/5.0|abc123|DE"

	entry, err := ParseLogEntry(line)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check all parsed fields
	if entry.StatusCode != 200 {
		t.Errorf("StatusCode: want 200, got %d", entry.StatusCode)
	}
	if entry.Timestamp != 1507167062421 {
		t.Errorf("Timestamp: want 1507167062421, got %d", entry.Timestamp)
	}
	if entry.BytesSent != 412 {
		t.Errorf("BytesSent: want 412, got %d", entry.BytesSent)
	}
	if entry.ClientIP != "163.172.53.0" {
		t.Errorf("ClientIP: want '163.172.53.0', got '%s'", entry.ClientIP)
	}
	if entry.URL != "https://cdn.example.com/ep1.mp3" {
		t.Errorf("URL: want 'https://cdn.example.com/ep1.mp3', got '%s'", entry.URL)
	}
	if entry.UserAgent != "Mozilla/5.0" {
		t.Errorf("UserAgent: want 'Mozilla/5.0', got '%s'", entry.UserAgent)
	}
	if entry.RequestID != "abc123" {
		t.Errorf("RequestID: want 'abc123', got '%s'", entry.RequestID)
	}
	if entry.CountryCode != "DE" {
		t.Errorf("CountryCode: want 'DE', got '%s'", entry.CountryCode)
	}
}

func TestParseLogEntry_InvalidLine(t *testing.T) {
	// Too few fields — should return an error
	entry, err := ParseLogEntry("HIT|200|1507167062421")

	if err == nil {
		t.Fatal("expected error for malformed line, got nil")
	}
	if entry != nil {
		t.Error("expected nil entry on error")
	}
}

func TestParseLogEntry_InvalidStatusCode(t *testing.T) {
	// 12 fields but StatusCode is not a number
	line := "HIT|notanumber|1507167062421|412|390|1.2.3.4|-|https://x.com/f.mp3|WA|UA|rid|US"

	entry, err := ParseLogEntry(line)

	if err == nil {
		t.Fatal("expected error for invalid status code, got nil")
	}
	if entry != nil {
		t.Error("expected nil entry on error")
	}
}
