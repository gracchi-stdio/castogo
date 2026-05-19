package landingview

import "fmt"

func formatDuration(seconds int) string {
	m := seconds / 60
	s := seconds % 60
	if m >= 60 {
		h := m / 60
		m = m % 60
		return fmt.Sprintf("%dh %dm", h, m)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}

func ternary(cond bool, a, b string) string {
	if cond {
		return a
	}
	return b
}

func socialPlatformIcon(platform string) string {
	switch platform {
	case "twitter", "x":
		return "twitter"
	case "github":
		return "github"
	case "youtube":
		return "youtube"
	case "linkedin":
		return "linkedin"
	case "instagram":
		return "instagram"
	case "facebook":
		return "facebook"
	default:
		return "link"
	}
}
