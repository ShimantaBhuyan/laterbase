package fetcher

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode"
)

// Fetcher interface for fetching URL content.
type Fetcher interface {
	Fetch(ctx context.Context, url string) *FetchResult
}

// FetchResult holds the result of fetching a URL.
type FetchResult struct {
	Title       string
	ContentText string // cleaned markdown-ish text
	FaviconURL  string
	OgImageURL  string
	IsPaywalled bool
	Error       error
}

// SimpleFetcher fetches URLs using basic net/http.
type SimpleFetcher struct {
	client *http.Client
}

// NewSimpleFetcher creates a new SimpleFetcher.
func NewSimpleFetcher() *SimpleFetcher {
	return &SimpleFetcher{
		client: &http.Client{
			Timeout: 15 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return errors.New("too many redirects")
				}
				return nil
			},
		},
	}
}

var (
	titleRegex      = regexp.MustCompile(`(?i)<title[^>]*>(.*?)</title>`)
	faviconRegex    = regexp.MustCompile(`(?i)<link[^>]+rel=["']?(?:icon|shortcut\s+icon|apple-touch-icon)["'][^>]*href=["']([^"']+)["']`)
	ogImageRegex    = regexp.MustCompile(`(?i)<meta[^>]+(?:property=["']og:image["']|name=["']twitter:image["'])[^>]+content=["']([^"']+)["']`)
	scriptRegex     = regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`)
	styleRegex      = regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`)
	navRegex        = regexp.MustCompile(`(?is)<nav[^>]*>.*?</nav>`)
	footerRegex     = regexp.MustCompile(`(?is)<footer[^>]*>.*?</footer>`)
	headerRegex     = regexp.MustCompile(`(?is)<header[^>]*>.*?</header>`)
	htmlTagRegex    = regexp.MustCompile(`<[^>]*>`)
	whitespaceRegex = regexp.MustCompile(`\s+`)

	paywallKeywords = []string{
		"paywall",
		"subscribe for",
		"metered",
		"subscription required",
		"unlock for",
		"premium article",
		"subscribe now",
		"subscribe to continue",
		"this is a subscriber-only",
		"sign up for unlimited",
	}
)

// Fetch implements the Fetcher interface.
func (f *SimpleFetcher) Fetch(ctx context.Context, url string) (*FetchResult) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return &FetchResult{Error: fmt.Errorf("creating request: %w", err)}
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := f.client.Do(req)
	if err != nil {
		return &FetchResult{Error: fmt.Errorf("request failed: %w", err)}
	}
	defer resp.Body.Close()

	// Check for inaccessible status codes
	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized {
		return &FetchResult{Error: errors.New("inaccessible")}
	}

	// Read body
	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024)) // 5MB limit
	if err != nil {
		return &FetchResult{Error: fmt.Errorf("reading body: %w", err)}
	}
	bodyStr := string(bodyBytes)

	// Check for paywall
	if isPaywalled(bodyStr) {
		title := extractTitle(bodyStr)
		favicon := extractFavicon(bodyStr, url)
		ogImage := extractOgImage(bodyStr, url)
		return &FetchResult{
			Title:       title,
			FaviconURL:  favicon,
			OgImageURL:  ogImage,
			IsPaywalled: true,
		}
	}

	// Extract metadata
	title := extractTitle(bodyStr)
	favicon := extractFavicon(bodyStr, url)
	ogImage := extractOgImage(bodyStr, url)
	contentText := cleanHTML(bodyStr)

	return &FetchResult{
		Title:       title,
		ContentText: contentText,
		FaviconURL:  favicon,
		OgImageURL:  ogImage,
	}
}

// isPaywalled checks if the content contains paywall keywords.
func isPaywalled(content string) bool {
	lower := strings.ToLower(content)
	for _, kw := range paywallKeywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

// extractTitle extracts the page title from HTML.
func extractTitle(html string) string {
	matches := titleRegex.FindStringSubmatch(html)
	if len(matches) > 1 {
		title := strings.TrimSpace(matches[1])
		// Decode common HTML entities
		title = strings.ReplaceAll(title, "&amp;", "&")
		title = strings.ReplaceAll(title, "&quot;", "\"")
		title = strings.ReplaceAll(title, "&#39;", "'")
		title = strings.ReplaceAll(title, "&lt;", "<")
		title = strings.ReplaceAll(title, "&gt;", ">")
		return title
	}
	return ""
}

// extractFavicon extracts the favicon URL from HTML.
func extractFavicon(html, baseURL string) string {
	matches := faviconRegex.FindStringSubmatch(html)
	if len(matches) > 1 {
		favicon := matches[1]
		if strings.HasPrefix(favicon, "//") {
			// Protocol-relative URL
			if strings.HasPrefix(baseURL, "https://") {
				return "https:" + favicon
			}
			return "http:" + favicon
		}
		if strings.HasPrefix(favicon, "/") {
			// Relative to root
			return baseURL + favicon
		}
		if strings.HasPrefix(favicon, "http://") || strings.HasPrefix(favicon, "https://") {
			return favicon
		}
		// Relative path
		return baseURL + "/" + favicon
	}
	// Fallback
	return baseURL + "/favicon.ico"
}

// extractOgImage extracts the Open Graph image URL from HTML.
func extractOgImage(html, baseURL string) string {
	matches := ogImageRegex.FindStringSubmatch(html)
	if len(matches) > 1 {
		img := matches[1]
		if strings.HasPrefix(img, "//") {
			if strings.HasPrefix(baseURL, "https://") {
				return "https:" + img
			}
			return "http:" + img
		}
		if strings.HasPrefix(img, "/") {
			base := baseURL
			if idx := strings.Index(base, "//"); idx != -1 {
				if slashIdx := strings.Index(base[idx+2:], "/"); slashIdx != -1 {
					base = base[:idx+2+slashIdx]
				}
			}
			return base + img
		}
		if strings.HasPrefix(img, "http://") || strings.HasPrefix(img, "https://") {
			return img
		}
		return baseURL + "/" + img
	}
	return ""
}

// cleanHTML strips unwanted tags and returns the cleaned text.
func cleanHTML(html string) string {
	// Remove script, style, nav, footer, header blocks
	html = scriptRegex.ReplaceAllString(html, " ")
	html = styleRegex.ReplaceAllString(html, " ")
	html = navRegex.ReplaceAllString(html, " ")
	html = footerRegex.ReplaceAllString(html, " ")
	html = headerRegex.ReplaceAllString(html, " ")

	// Replace <br>, <p>, <li> with newlines for readability
	html = regexp.MustCompile(`(?i)<br\s*/?>`).ReplaceAllString(html, "\n")
	html = regexp.MustCompile(`(?i)</p>`).ReplaceAllString(html, "\n\n")
	html = regexp.MustCompile(`(?i)</li>`).ReplaceAllString(html, "\n")
	html = regexp.MustCompile(`(?i)</tr>`).ReplaceAllString(html, "\n")
	html = regexp.MustCompile(`(?i)</h[1-6]>`).ReplaceAllString(html, "\n\n")

	// Strip remaining HTML tags
	html = htmlTagRegex.ReplaceAllString(html, " ")

	// Decode common HTML entities
	html = strings.ReplaceAll(html, "&nbsp;", " ")
	html = strings.ReplaceAll(html, "&amp;", "&")
	html = strings.ReplaceAll(html, "&lt;", "<")
	html = strings.ReplaceAll(html, "&gt;", ">")
	html = strings.ReplaceAll(html, "&quot;", "\"")
	html = strings.ReplaceAll(html, "&#39;", "'")

	// Collapse whitespace
	html = whitespaceRegex.ReplaceAllString(html, " ")

	// Trim leading/trailing whitespace
	html = strings.TrimSpace(html)

	// Truncate to 3000 words
	html = truncateWords(html, 3000)

	return html
}

// truncateWords truncates text to at most maxWords words.
func truncateWords(text string, maxWords int) string {
	if maxWords <= 0 {
		return ""
	}

	words := strings.Fields(text)
	if len(words) <= maxWords {
		return text
	}

	// Find the last word boundary before maxWords
	return strings.Join(words[:maxWords], " ")
}

// IsSpace reports whether a rune is whitespace.
func IsSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == '\v' || r == '\f' ||
		unicode.Is(unicode.Zs, r)
}
