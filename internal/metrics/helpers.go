package metrics

import "strings"

// NormalizeUserAgent — reduce raw user agent strings to a small set
func NormalizeUserAgent(ua string) string {
	ua = strings.ToLower(ua)
	switch {
	case strings.Contains(ua, "bot") || strings.Contains(ua, "spider") || strings.Contains(ua, "crawler"):
		return "bot"
	case strings.Contains(ua, "edg"):
		return "edge"
	case strings.Contains(ua, "chrome"):
		return "chrome"
	case strings.Contains(ua, "firefox"):
		return "firefox"
	case strings.Contains(ua, "safari"):
		return "safari"
	case strings.Contains(ua, "opera"):
		return "opera"
	case ua == "":
		return "unknown"
	default:
		return "other"
	}
}

// NormalizeReferrer — reduce raw referrer URLs to just the domain
func NormalizeReferrer(referrer string) string {
	if referrer == "" {
		return "direct"
	}

	// strip protocol
	referrer = strings.TrimPrefix(referrer, "https://")
	referrer = strings.TrimPrefix(referrer, "http://")

	// take only the domain (before the first /)
	if idx := strings.Index(referrer, "/"); idx != -1 {
		referrer = referrer[:idx]
	}

	// strip "www."
	referrer = strings.TrimPrefix(referrer, "www.")

	if referrer == "" {
		return "direct"
	}
	return referrer
}