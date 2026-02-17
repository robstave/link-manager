package services

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type MetadataService struct{}

func NewMetadataService() *MetadataService {
	return &MetadataService{}
}

// PageMeta holds scraped metadata from a web page
type PageMeta struct {
	Title   string
	IconURL string
}

// FetchPageMeta fetches a URL and extracts the title and icon
func (s *MetadataService) FetchPageMeta(rawURL string) (PageMeta, error) {
	slog.Info("meta: fetching page metadata", "url", rawURL)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		slog.Error("meta: failed to create request", "url", rawURL, "error", err)
		return PageMeta{}, fmt.Errorf("failed to create request: %w", err)
	}

	// Many sites block or degrade responses without a browser-like user agent
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; bookmarkbot/1.0)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		slog.Error("meta: HTTP request failed", "url", rawURL, "error", err)
		return PageMeta{}, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	slog.Info("meta: received response", "url", rawURL, "status", resp.StatusCode)

	// 100KB — enough to capture <head> on even verbose sites
	limitedReader := io.LimitReader(resp.Body, 250*1024)
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		slog.Error("meta: failed to read response body", "url", rawURL, "error", err)
		return PageMeta{}, fmt.Errorf("failed to read response: %w", err)
	}

	slog.Info("meta: read body", "url", rawURL, "bytes", len(body))

	html := string(body)
	meta := PageMeta{
		Title:   extractTitle(html, rawURL),
		IconURL: extractIcon(html, rawURL),
	}

	slog.Info("meta: extracted metadata", "url", rawURL, "title", meta.Title, "iconURL", meta.IconURL)
	return meta, nil
}

// FetchTitle fetches a URL and extracts the title (kept for backward compat)
func (s *MetadataService) FetchTitle(rawURL string) (string, error) {
	meta, err := s.FetchPageMeta(rawURL)
	if err != nil {
		return "", err
	}
	if meta.Title == "" {
		return "", fmt.Errorf("no title found")
	}
	return meta.Title, nil
}

// extractTitle tries og:title first (usually cleaner), then falls back to <title>
// For YouTube, it prioritises the <title> tag.
func extractTitle(html string, rawURL string) string {

	slog.Info("meta: extracting title----", "url", rawURL)

	// Log a small snippet of the body to help debug
	snippetLen := 500
	if len(html) < snippetLen {
		snippetLen = len(html)
	}
	slog.Info("meta: body snippet", "snippet", html[:snippetLen])

	// YouTube-specific priority: use standard <title> tag first
	isYouTube := strings.Contains(strings.ToLower(rawURL), "youtube.com") || strings.Contains(strings.ToLower(rawURL), "youtu.be")
	if isYouTube {
		slog.Info("meta: isYoutube", "url", rawURL)
		// Use (?is) so that . matches newlines
		titleRegex := regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)

		matches := titleRegex.FindStringSubmatch(html)
		slog.Info("meta: youtube search result", "count", len(matches), "url", rawURL)
		if len(matches) > 0 {
			slog.Info("meta: youtube matches", "matches", matches)
		}
		if len(matches) >= 2 {
			slog.Info("meta: youtube success", "url", rawURL)
			return cleanText(matches[1])
		}

		// Also check twitter:title for YouTube specifically
		twitterRegex := regexp.MustCompile(`(?i)<meta[^>]+name\s*=\s*["']twitter:title["'][^>]+content\s*=\s*["']([^"']+)["']`)
		if twitterMatches := twitterRegex.FindStringSubmatch(html); len(twitterMatches) >= 2 {
			slog.Info("meta: youtube twitter:title success", "url", rawURL)
			return cleanText(twitterMatches[1])
		}

		slog.Info("meta: isYoutube failed xxx", "url", rawURL)
	}

	// og:title — content before property
	ogRegex := regexp.MustCompile(`(?i)<meta[^>]+property\s*=\s*["']og:title["'][^>]+content\s*=\s*["']([^"']+)["']`)
	matches := ogRegex.FindStringSubmatch(html)
	slog.Info("meta: ogRegex search result", "count", len(matches), "url", rawURL)
	if len(matches) > 0 {
		slog.Info("meta: ogRegex matches", "matches", matches)
	}
	if len(matches) >= 2 {
		slog.Info("meta: og:title success", "url", rawURL)
		return cleanText(matches[1])
	}

	// og:title — content after property (attribute order varies)
	ogRegex2 := regexp.MustCompile(`(?i)<meta[^>]+content\s*=\s*["']([^"']+)["'][^>]+property\s*=\s*["']og:title["']`)
	matches2 := ogRegex2.FindStringSubmatch(html)
	slog.Info("meta: ogRegex2 search result", "count", len(matches2), "url", rawURL)
	if len(matches2) > 0 {
		slog.Info("meta: ogRegex2 matches", "matches", matches2)
	}
	if len(matches2) >= 2 {
		slog.Info("meta: og:title (variant) success", "url", rawURL)
		return cleanText(matches2[1])
	}

	// twitter:title — similar to og:title, often present on news/social sites
	twitterRegex := regexp.MustCompile(`(?i)<meta[^>]+name\s*=\s*["']twitter:title["'][^>]+content\s*=\s*["']([^"']+)["']`)
	matchesTwitter := twitterRegex.FindStringSubmatch(html)
	slog.Info("meta: twitterRegex search result", "count", len(matchesTwitter), "url", rawURL)
	if len(matchesTwitter) > 0 {
		slog.Info("meta: twitterRegex matches", "matches", matchesTwitter)
	}
	if len(matchesTwitter) >= 2 {
		slog.Info("meta: twitter:title success", "url", rawURL)
		return cleanText(matchesTwitter[1])
	}

	// twitter:title — content after name (variant)
	twitterRegex2 := regexp.MustCompile(`(?i)<meta[^>]+content\s*=\s*["']([^"']+)["'][^>]+name\s*=\s*["']twitter:title["']`)
	matchesTwitter2 := twitterRegex2.FindStringSubmatch(html)
	slog.Info("meta: twitterRegex2 search result", "count", len(matchesTwitter2), "url", rawURL)
	if len(matchesTwitter2) > 0 {
		slog.Info("meta: twitterRegex2 matches", "matches", matchesTwitter2)
	}
	if len(matchesTwitter2) >= 2 {
		slog.Info("meta: twitter:title (variant) success", "url", rawURL)
		return cleanText(matchesTwitter2[1])
	}

	// Standard <title> tag as final fallback
	// Use (?is) so that . matches newlines
	titleRegex := regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)
	matchesStandard := titleRegex.FindStringSubmatch(html)
	slog.Info("meta: standard title search result", "count", len(matchesStandard), "url", rawURL)
	if len(matchesStandard) > 0 {
		slog.Info("meta: standard title matches", "matches", matchesStandard)
	}
	if len(matchesStandard) >= 2 {
		slog.Info("meta: standard title success", "url", rawURL)
		return cleanText(matchesStandard[1])
	}

	return ""
}

// extractIcon tries several strategies in order of preference
func extractIcon(html, rawURL string) string {
	// Ordered patterns — first match wins
	patterns := []string{
		// rel="icon" with quotes, href after rel
		`(?i)<link[^>]+rel\s*=\s*["'][^"']*icon[^"']*["'][^>]+href\s*=\s*["']([^"']+)["']`,
		// rel="icon" with quotes, href before rel
		`(?i)<link[^>]+href\s*=\s*["']([^"']+)["'][^>]+rel\s*=\s*["'][^"']*icon[^"']*["']`,
		// apple-touch-icon — high res, good fallback
		`(?i)<link[^>]+rel\s*=\s*["']apple-touch-icon["'][^>]+href\s*=\s*["']([^"']+)["']`,
		// og:image — last resort before /favicon.ico, gives something visual
		`(?i)<meta[^>]+property\s*=\s*["']og:image["'][^>]+content\s*=\s*["']([^"']+)["']`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(html); len(matches) >= 2 {
			href := strings.TrimSpace(matches[1])
			return resolveURL(href, rawURL)
		}
	}

	// Final fallback: /favicon.ico at the domain root
	parsed, err := url.Parse(rawURL)
	if err == nil {
		return parsed.Scheme + "://" + parsed.Host + "/favicon.ico"
	}

	return ""
}

// resolveURL turns relative hrefs into absolute URLs
func resolveURL(href, base string) string {
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}

	parsed, err := url.Parse(base)
	if err != nil {
		return href
	}

	if strings.HasPrefix(href, "//") {
		return parsed.Scheme + ":" + href
	}

	if strings.HasPrefix(href, "/") {
		return parsed.Scheme + "://" + parsed.Host + href
	}

	// Relative URL — resolve against base path
	return parsed.Scheme + "://" + parsed.Host + "/" + href
}

// cleanText normalises whitespace and decodes common HTML entities
func cleanText(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\t", " ")

	// Collapse multiple spaces
	spaceRegex := regexp.MustCompile(`\s{2,}`)
	s = spaceRegex.ReplaceAllString(s, " ")

	// Common HTML entities
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&quot;", "\"")
	s = strings.ReplaceAll(s, "&#39;", "'")
	s = strings.ReplaceAll(s, "&apos;", "'")
	s = strings.ReplaceAll(s, "&nbsp;", " ")
	s = strings.ReplaceAll(s, "&#8211;", "–")
	s = strings.ReplaceAll(s, "&#8212;", "—")
	s = strings.ReplaceAll(s, "&#8216;", "'")
	s = strings.ReplaceAll(s, "&#8217;", "'")
	s = strings.ReplaceAll(s, "&#8220;", "\"")
	s = strings.ReplaceAll(s, "&#8221;", "\"")

	return s
}
