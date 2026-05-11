// ABOUTME: Defines custom template functions for HTML rendering.
// ABOUTME: Contains timeAgo for relative time and faviconURL for blog favicons.
package server

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

// timeAgo converts a time.Time to a human-readable relative time string.
// Returns empty string for nil input, handles edge cases like future times.
func timeAgo(t *time.Time) string {
	if t == nil {
		return ""
	}

	diff := time.Since(*t)

	// Handle future times
	if diff < 0 {
		return "in the future"
	}

	seconds := int64(diff.Seconds())
	minutes := int64(diff.Minutes())
	hours := int64(diff.Hours())
	days := hours / 24
	weeks := days / 7
	months := days / 30
	years := days / 365

	switch {
	case seconds < 60:
		return "just now"
	case minutes == 1:
		return "1 minute ago"
	case minutes < 60:
		return fmt.Sprintf("%d minutes ago", minutes)
	case hours == 1:
		return "1 hour ago"
	case hours < 24:
		return fmt.Sprintf("%d hours ago", hours)
	case days == 1:
		return "yesterday"
	case days < 7:
		return fmt.Sprintf("%d days ago", days)
	case weeks == 1:
		return "1 week ago"
	case weeks < 5:
		return fmt.Sprintf("%d weeks ago", weeks)
	case months == 1:
		return "1 month ago"
	case months < 12:
		return fmt.Sprintf("%d months ago", months)
	case years == 1:
		return "1 year ago"
	default:
		return fmt.Sprintf("%d years ago", years)
	}
}

// smryURL strips the protocol from an article URL and prepends the smry.ai domain
// so the article can be opened in smry.ai's reader/summarizer.
func smryURL(articleURL string) string {
	u := strings.TrimPrefix(articleURL, "https://")
	u = strings.TrimPrefix(u, "http://")
	return "https://smry.ai/" + u
}

// faviconURL builds a Google S2 favicon URL from a blog URL.
// Returns empty string if URL is invalid or has no host.
func faviconURL(blogURL string) string {
	u, err := url.Parse(blogURL)
	if err != nil || u.Host == "" {
		return ""
	}
	return fmt.Sprintf("https://www.google.com/s2/favicons?domain=%s&sz=32", u.Host)
}

// ptrEq compares a pointer to an int64 value.
// Returns true if ptr is non-nil and dereferenced value equals v.
// Used for comparing Blog.CategoryID (*int64) with Category.ID (int64).
func ptrEq(ptr *int64, v int64) bool {
	if ptr == nil {
		return false
	}
	return *ptr == v
}
