package domain

import (
	"regexp"
	"strings"
)

var (
	// legalSuffixes matches common corporate suffixes at end of string.
	legalSuffixes = regexp.MustCompile(`(?i)\s*(,?\s*)(Inc\.?|LLC|Ltd\.?|Corp\.?|Corporation|Co\.?|Company|Group|Holdings|PLC|GmbH|S\.A\.?|Pty\.?|L\.?P\.?|N\.?V\.?|S\.?A\.?S\.?|AG)\s*$`)
	// nonAlphaNum matches characters that aren't alphanumeric or spaces.
	nonAlphaNum = regexp.MustCompile(`[^a-z0-9 ]`)
	// multiSpace collapses multiple spaces into one.
	multiSpace = regexp.MustCompile(`\s+`)
	// slugNonAlpha matches characters that aren't lowercase alphanumeric or hyphens.
	slugNonAlpha = regexp.MustCompile(`[^a-z0-9-]`)
	// multiHyphen collapses multiple hyphens into one.
	multiHyphen = regexp.MustCompile(`-+`)
)

// NormalizeCompanyName produces a canonical form for fuzzy matching.
// "Google, Inc." → "google", "JPMorgan Chase & Co." → "jpmorgan chase".
func NormalizeCompanyName(name string) string {
	s := strings.TrimSpace(name)
	s = legalSuffixes.ReplaceAllString(s, "")
	s = strings.ToLower(s)
	s = nonAlphaNum.ReplaceAllString(s, " ")
	s = multiSpace.ReplaceAllString(s, " ")
	s = strings.TrimSpace(s)
	return s
}

// Slugify produces a URL-safe slug for ATS board token probing.
// "Stripe, Inc." → "stripe", "My Cool Startup" → "my-cool-startup".
func Slugify(name string) string {
	s := NormalizeCompanyName(name)
	s = strings.ReplaceAll(s, " ", "-")
	s = slugNonAlpha.ReplaceAllString(s, "")
	s = multiHyphen.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}
