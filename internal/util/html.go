package util

import (
	"regexp"
	"strings"
)

var (
	tagRe    = regexp.MustCompile(`<[^>]+>`)
	brRe     = regexp.MustCompile(`(?i)<br\s*/?>`)
	pRe      = regexp.MustCompile(`(?i)<p\s*/?>`)
	entityRe = regexp.MustCompile(`&[a-z]+;|&#[0-9]+;`)
)

var entities = map[string]string{
	"&amp;":  "&",
	"&lt;":   "<",
	"&gt;":   ">",
	"&quot;": `"`,
	"&#x27;": "'",
	"&apos;": "'",
	"&nbsp;": " ",
}

// StripHTML converts HTML to plain text suitable for terminal display.
func StripHTML(s string) string {
	// Replace block tags with newlines before stripping.
	s = brRe.ReplaceAllString(s, "\n")
	s = pRe.ReplaceAllString(s, "\n\n")
	// Strip remaining tags.
	s = tagRe.ReplaceAllString(s, "")
	// Decode entities.
	s = entityRe.ReplaceAllStringFunc(s, func(e string) string {
		if v, ok := entities[strings.ToLower(e)]; ok {
			return v
		}
		return e
	})
	return strings.TrimSpace(s)
}
