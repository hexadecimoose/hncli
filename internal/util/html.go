package util

import (
	"html"
	"regexp"
	"strings"
)

var (
	tagRe = regexp.MustCompile(`<[^>]+>`)
	brRe  = regexp.MustCompile(`(?i)<br\s*/?>`)
	pRe   = regexp.MustCompile(`(?i)<p\s*/?>`)
)

// StripHTML converts HTML to plain text suitable for terminal display.
func StripHTML(s string) string {
	// Replace block tags with newlines before stripping.
	s = brRe.ReplaceAllString(s, "\n")
	s = pRe.ReplaceAllString(s, "\n\n")
	// Strip remaining tags.
	s = tagRe.ReplaceAllString(s, "")
	// Decode all HTML entities (named, decimal &#39;, and hex &#x27;).
	s = html.UnescapeString(s)
	return strings.TrimSpace(s)
}
