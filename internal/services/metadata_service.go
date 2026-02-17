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

	resp, err := client.Get(rawURL)
	if err != nil {
		slog.Error("meta: HTTP request failed", "url", rawURL, "error", err)
		return PageMeta{}, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	slog.Info("meta: received response", "url", rawURL, "status", resp.StatusCode)

	limitedReader := io.LimitReader(resp.Body, 50*1024)
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		slog.Error("meta: failed to read response body", "url", rawURL, "error", err)
		return PageMeta{}, fmt.Errorf("failed to read response: %w", err)
	}

	slog.Info("meta: read body", "url", rawURL, "bytes", len(body))

	html := string(body)
	meta := PageMeta{
		Title:   extractTitle(html),
		IconURL: extractIcon(html, rawURL),
	}

	slog.Info("meta: extracted metadata", "url", rawURL, "title", meta.Title, "iconURL", meta.IconURL)
	return meta, nil
}

// FetchTitle fetches a URL and extracts the <title> tag (kept for backward compat)
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

func extractTitle(html string) string {
	titleRegex := regexp.MustCompile(`(?i)<title[^>]*>(.*?)</title>`)
	matches := titleRegex.FindStringSubmatch(html)
	if len(matches) < 2 {
		return ""
	}

	title := strings.TrimSpace(matches[1])
	title = strings.ReplaceAll(title, "\n", " ")
	title = strings.ReplaceAll(title, "\r", " ")

	// Basic HTML entity decoding
	title = strings.ReplaceAll(title, "&amp;", "&")
	title = strings.ReplaceAll(title, "&lt;", "<")
	title = strings.ReplaceAll(title, "&gt;", ">")
	title = strings.ReplaceAll(title, "&quot;", "\"")
	title = strings.ReplaceAll(title, "&#39;", "'")

	return title
}

func extractIcon(html, rawURL string) string {
	// Try <link rel="icon" ...> or <link rel="shortcut icon" ...>
	iconRegex := regexp.MustCompile(`(?i)<link[^>]+rel\s*=\s*["'](?:shortcut\s+)?icon["'][^>]+href\s*=\s*["']([^"']+)["']`)
	matches := iconRegex.FindStringSubmatch(html)
	if len(matches) < 2 {
		// Try alternate order: href before rel
		iconRegex2 := regexp.MustCompile(`(?i)<link[^>]+href\s*=\s*["']([^"']+)["'][^>]+rel\s*=\s*["'](?:shortcut\s+)?icon["']`)
		matches = iconRegex2.FindStringSubmatch(html)
	}

	if len(matches) >= 2 {
		iconHref := strings.TrimSpace(matches[1])
		return resolveURL(iconHref, rawURL)
	}

	// Fallback: try /favicon.ico at the domain root
	parsed, err := url.Parse(rawURL)
	if err == nil {
		return parsed.Scheme + "://" + parsed.Host + "/favicon.ico"
	}
	return ""
}

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
	// Relative URL
	return parsed.Scheme + "://" + parsed.Host + "/" + href
}
