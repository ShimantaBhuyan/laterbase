package content

import (
	"regexp"
	"strings"
)

var (
	// allowedTags are the HTML tags we preserve in sanitized output.
	allowedTags = map[string]bool{
		"p": true, "br": true, "b": true, "i": true,
		"a": true, "ul": true, "ol": true, "li": true,
		"pre": true, "code": true, "blockquote": true,
	}

	// allowedAttrs are attributes we preserve on allowed tags.
	allowedAttrs = map[string]map[string]bool{
		"a": {"href": true, "target": true},
	}

	tagRegex   = regexp.MustCompile(`(?i)<(/?)(\w+)([^>]*)>`)
	attrRegex  = regexp.MustCompile(`(\w+)\s*=\s*"([^"]*)"`)
)

// SanitizeHTML strips all HTML tags except a safe subset, preserving allowed attributes.
func SanitizeHTML(html string) string {
	// Process tags
	result := tagRegex.ReplaceAllStringFunc(html, func(match string) string {
		parts := tagRegex.FindStringSubmatch(match)
		if len(parts) < 4 {
			return match
		}

		slash := parts[1]
		tagName := strings.ToLower(parts[2])
		attrs := parts[3]

		if !allowedTags[tagName] {
			// Strip disallowed tags entirely
			return ""
		}

		if slash == "/" {
			return "</" + tagName + ">"
		}

		// Build allowed attributes
		allowed := ""
		if tagName == "a" {
			attrMatches := attrRegex.FindAllStringSubmatch(attrs, -1)
			for _, attr := range attrMatches {
				if len(attr) >= 3 {
					attrName := strings.ToLower(attr[1])
					attrValue := attr[2]
					if allowedAttrs["a"][attrName] {
						if attrName == "target" && attrValue != "_blank" {
							continue
						}
						allowed += " " + attrName + "=\"" + attrValue + "\""
					}
				}
			}
		}

		return "<" + tagName + allowed + ">"
	})

	return result
}

// TruncateWords truncates text to at most maxWords words.
func TruncateWords(text string, maxWords int) string {
	if maxWords <= 0 {
		return ""
	}

	words := strings.Fields(text)
	if len(words) <= maxWords {
		return text
	}

	return strings.Join(words[:maxWords], " ")
}
