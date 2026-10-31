package domain

import (
	"regexp"
	"strings"
)

var slugRegex = regexp.MustCompile(`[^a-z0-9]+`)

// Slugify converts a handle into a URL-safe, lowercase slug.
func Slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = slugRegex.ReplaceAllString(s, "-")

	s = strings.Trim(s, "-")
	if s == "" {
		s = "organizer"
	}

	return s
}
