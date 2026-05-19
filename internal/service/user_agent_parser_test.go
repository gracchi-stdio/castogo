package service

import "testing"

func TestParseUserAgent_ApplePodcasts(t *testing.T) {
	ua := "AppleCoreMedia/1.0.0.0 (iPhone; U; CPU OS 16_0 like Mac OS X)"
	parsed := ParseUserAgent(ua)

	assertUA(t, parsed, "Apple Podcasts", "Apple Podcasts", "smartphone", "iOS", false)
}

func TestParseUserAgent_Spotify(t *testing.T) {
	ua := "Spotify/8.9.2 Android/33"
	parsed := ParseUserAgent(ua)

	// No "Mobile" keyword → can't distinguish from tablet
	assertUA(t, parsed, "Spotify", "Spotify", "tablet", "Android", false)
}

func TestParseUserAgent_Chrome(t *testing.T) {
	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	parsed := ParseUserAgent(ua)

	assertUA(t, parsed, "Browser", "Chrome", "desktop", "Windows", false)
}

func TestParseUserAgent_Firefox(t *testing.T) {
	ua := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:121.0) Gecko/20100101 Firefox/121.0"
	parsed := ParseUserAgent(ua)

	assertUA(t, parsed, "Browser", "Firefox", "desktop", "macOS", false)
}

func TestParseUserAgent_Bot(t *testing.T) {
	ua := "Googlebot/2.1 (+http://www.google.com/bot.html)"
	parsed := ParseUserAgent(ua)

	assertUA(t, parsed, "Bot", "Bot", "bot", "Unknown", true)
}

func TestParseUserAgent_Curl(t *testing.T) {
	ua := "curl/8.1.2"
	parsed := ParseUserAgent(ua)

	assertUA(t, parsed, "Bot", "Bot", "bot", "Unknown", true)
}

func TestParseUserAgent_PocketCasts(t *testing.T) {
	ua := "Pocket Casts"
	parsed := ParseUserAgent(ua)

	assertUA(t, parsed, "Pocket Casts", "Pocket Casts", "desktop", "Unknown", false)
}

func TestParseUserAgent_AndroidTablet(t *testing.T) {
	// Android tablet: has "android" but NOT "mobile"
	ua := "Mozilla/5.0 (Linux; Android 13) AppleWebKit/537.36 Chrome/120.0.0.0 Safari/537.36"
	parsed := ParseUserAgent(ua)

	if parsed.Device != "tablet" {
		t.Errorf("Android without Mobile should be tablet, got %q", parsed.Device)
	}
}

func TestParseUserAgent_IPad(t *testing.T) {
	ua := "Mozilla/5.0 (iPad; CPU OS 16_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148 Safari/604.1"
	parsed := ParseUserAgent(ua)

	assertUA(t, parsed, "Browser", "Safari", "tablet", "iOS", false)
}

func TestParseUserAgent_Unknown(t *testing.T) {
	ua := "SomeWeirdAgent/1.0"
	parsed := ParseUserAgent(ua)

	assertUA(t, parsed, "Unknown", "Unknown", "desktop", "Unknown", false)
}

func TestParseUserAgent_Castro(t *testing.T) {
	// " Castro" (with space) should NOT match "CastBox"
	ua := "Castro/2.7"
	parsed := ParseUserAgent(ua)

	assertUA(t, parsed, "Castro", "Castro", "desktop", "Unknown", false)
}

func TestParseUserAgent_CastBox(t *testing.T) {
	ua := "CastBox/8.3"
	parsed := ParseUserAgent(ua)

	assertUA(t, parsed, "CastBox", "CastBox", "desktop", "Unknown", false)
}

func TestParseUserAgent_Bot_NoFalsePositive(t *testing.T) {
	// "robot" contains "bot" but should NOT be flagged — \bbot\b requires word boundary
	ua := "MyRobotPlayer/1.0"
	parsed := ParseUserAgent(ua)

	assertUA(t, parsed, "Unknown", "Unknown", "desktop", "Unknown", false)
}

func TestParseUserAgent_Bot_BaiduSpider(t *testing.T) {
	ua := "Mozilla/5.0 (compatible; Baiduspider-crawler/2.0; +http://www.baidu.com/search/spider.html)"
	parsed := ParseUserAgent(ua)

	assertUA(t, parsed, "Bot", "Bot", "bot", "Unknown", true)
}

func TestParseUserAgent_Bot_FacebookCrawler(t *testing.T) {
	ua := "facebookexternalhit/1.1"
	parsed := ParseUserAgent(ua)

	assertUA(t, parsed, "Bot", "Bot", "bot", "Unknown", true)
}

func TestParseUserAgent_Bot_Wget(t *testing.T) {
	ua := "wget/1.21.3"
	parsed := ParseUserAgent(ua)

	assertUA(t, parsed, "Bot", "Bot", "bot", "Unknown", true)
}

func TestParseUserAgent_Bot_PythonRequests(t *testing.T) {
	ua := "python-requests/2.31"
	parsed := ParseUserAgent(ua)

	assertUA(t, parsed, "Bot", "Bot", "bot", "Unknown", true)
}

// Helper to reduce repetition — checks all fields in one call
func assertUA(t *testing.T, parsed ParsedUA, wantService, wantApp, wantDevice, wantOS string, wantBot bool) {
	t.Helper() // marks this as a test helper — line numbers point to the caller

	if parsed.Service != wantService {
		t.Errorf("Service: want %q, got %q", wantService, parsed.Service)
	}
	if parsed.App != wantApp {
		t.Errorf("App: want %q, got %q", wantApp, parsed.App)
	}
	if parsed.Device != wantDevice {
		t.Errorf("Device: want %q, got %q", wantDevice, parsed.Device)
	}
	if parsed.OS != wantOS {
		t.Errorf("OS: want %q, got %q", wantOS, parsed.OS)
	}
	if parsed.IsBot != wantBot {
		t.Errorf("IsBot: want %v, got %v", wantBot, parsed.IsBot)
	}
}
