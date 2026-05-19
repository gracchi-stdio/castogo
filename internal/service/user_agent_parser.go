package service

import (
	"regexp"
	"strings"
)

type ParsedUA struct {
	Service string // e.g. "Apple Podcasts", "Spotify", "Browser"
	App     string // e.g. "Apple Podcasts", "Chrome", "Firefox"
	Device  string // e.g. "desktop", "smartphone", "tablet", "bot"
	OS      string // e.g. "Windows", "macOS", "iOS", "Android"
	IsBot   bool   // whether the user agent is identified as a bot/crawler
}

type uaRule struct {
	pattern string // case-insensitive substring
	service string
	app     string
}

// Podcast app rules — checked in order, first match wins.
// More specific patterns come before generic ones.
var appRules = []uaRule{
	{"applecoremedia", "Apple Podcasts", "Apple Podcasts"},
	{"itms", "Apple Podcasts", "Apple Podcasts"},
	{"spotify", "Spotify", "Spotify"},
	{"google podcast", "Google Podcasts", "Google Podcasts"},
	{"pocket casts", "Pocket Casts", "Pocket Casts"},
	{"overcast", "Overcast", "Overcast"},
	{"castbox", "CastBox", "CastBox"},
	{"castro", "Castro", "Castro"}, // after castbox so CastBox matches first
	{"podcastaddict", "Podcast Addict", "Podcast Addict"},
	{"podcastguru", "PodcastGuru", "PodcastGuru"},
	{"antennapod", "AntennaPod", "AntennaPod"},
	{"podbean", "Podbean", "Podbean"},
	{"stitcher", "Stitcher", "Stitcher"},
	{"sonos", "Sonos", "Sonos"},
	{"tunein", "TuneIn", "TuneIn"},
	{"iheartradio", "iHeartRadio", "iHeartRadio"},
}

// Browser rules — checked after podcast apps if no app matched.
var browserRules = []uaRule{
	{"edg/", "Browser", "Edge"},
	{"chrome/", "Browser", "Chrome"},
	{"firefox/", "Browser", "Firefox"},
	{"safari/", "Browser", "Safari"},
	{"safari", "Browser", "Safari"}, // iPad UA: "...like Mac OS X" without versioned Safari
}

// botPatterns are compiled regexps for known bot/crawler user agents.
// Uses word boundaries and specific patterns to avoid false positives
// on innocent strings like device identifiers or app names.
var botPatterns = []*regexp.Regexp{
	regexp.MustCompile(`\bbot\b`),        // Googlebot, Bingbot, etc. (word boundary)
	regexp.MustCompile(`\bcrawl`),         // crawler, Googlebot-Mobile (word boundary)
	regexp.MustCompile(`spider`),          // YandexBot, Baiduspider
	regexp.MustCompile(`\bslurp\b`),       // Yahoo Slurp
	regexp.MustCompile(`\bcurl/`),         // curl/8.1.2
	regexp.MustCompile(`\bwget/`),         // wget/1.21
	regexp.MustCompile(`python-requests`), // python-requests/2.31
	regexp.MustCompile(`facebookexternalhit`), // Facebook crawler
}

func ParseUserAgent(ua string) ParsedUA {
	uaLower := strings.ToLower(ua)

	// Check bots first — short-circuit, don't bother with app detection
	if isBot(uaLower) {
		return ParsedUA{
			Service: "Bot",
			App:     "Bot",
			Device:  "bot",
			OS:      inferOS(uaLower),
			IsBot:   true,
		}
	}

	// Check podcast apps
	for _, rule := range appRules {
		if strings.Contains(uaLower, rule.pattern) {
			return ParsedUA{
				Service: rule.service,
				App:     rule.app,
				Device:  inferDevice(uaLower),
				OS:      inferOS(uaLower),
			}
		}
	}

	// Check browsers
	for _, rule := range browserRules {
		if strings.Contains(uaLower, rule.pattern) {
			return ParsedUA{
				Service: rule.service,
				App:     rule.app,
				Device:  inferDevice(uaLower),
				OS:      inferOS(uaLower),
			}
		}
	}

	// Unknown — still detect device and OS
	return ParsedUA{
		Service: "Unknown",
		App:     "Unknown",
		Device:  inferDevice(uaLower),
		OS:      inferOS(uaLower),
	}
}

func inferDevice(ua string) string {
	// iPad before iPhone before generic checks
	if strings.Contains(ua, "ipad") {
		return "tablet"
	}
	if strings.Contains(ua, "iphone") {
		return "smartphone"
	}
	// Android: "Mobile" keyword distinguishes phones from tablets
	if strings.Contains(ua, "android") {
		if strings.Contains(ua, "mobile") {
			return "smartphone"
		}
		return "tablet"
	}
	// Anything else is probably a desktop browser
	return "desktop"
}

func inferOS(ua string) string {
	switch {
	case strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad"):
		return "iOS"
	case strings.Contains(ua, "android"):
		return "Android"
	case strings.Contains(ua, "windows"):
		return "Windows"
	case strings.Contains(ua, "mac os x") || strings.Contains(ua, "macintosh"):
		return "macOS"
	case strings.Contains(ua, "linux") || strings.Contains(ua, "x11"):
		return "Linux"
	default:
		return "Unknown"
	}
}

func isBot(ua string) bool {
	for _, pattern := range botPatterns {
		if pattern.MatchString(ua) {
			return true
		}
	}
	return false
}
