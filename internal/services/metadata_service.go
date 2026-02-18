package services

import (
	"fmt"
	"log/slog"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

type MetadataService struct{}

func NewMetadataService() *MetadataService {
	return &MetadataService{}
}

// PageMeta holds scraped metadata from a web page
type PageMeta struct {
	Title       string
	Description string
	IconURL     string
}

// FetchPageMeta fetches a URL and extracts the title, description and icon using Colly.
func (s *MetadataService) FetchPageMeta(rawURL string) (PageMeta, error) {
	slog.Info("meta: fetching page metadata", "url", rawURL)

	isYouTube := strings.Contains(strings.ToLower(rawURL), "youtube.com") ||
		strings.Contains(strings.ToLower(rawURL), "youtu.be")

	var meta PageMeta
	baseURL := rawURL

	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
		colly.MaxDepth(1),
	)
	c.SetRequestTimeout(10 * time.Second)

	// ── Title extraction ─────────────────────────────────────────────────────

	// For YouTube, prefer twitter:title (clean, no channel prefix or " - YouTube" suffix)
	if isYouTube {
		c.OnHTML(`meta[name="twitter:title"]`, func(e *colly.HTMLElement) {
			if meta.Title == "" {
				val := strings.TrimSpace(e.Attr("content"))
				if val != "" {
					slog.Info("meta: youtube twitter:title hit", "value", val)
					meta.Title = cleanText(val)
				}
			}
		})
		c.OnHTML(`title`, func(e *colly.HTMLElement) {
			if meta.Title == "" {
				val := strings.TrimSpace(e.Text)
				if val != "" {
					slog.Info("meta: youtube <title> hit", "value", val)
					meta.Title = cleanText(val)
				}
			}
		})
	}

	// General: og:title first, then twitter:title, then <title>
	c.OnHTML(`meta[property="og:title"]`, func(e *colly.HTMLElement) {
		if meta.Title == "" {
			val := strings.TrimSpace(e.Attr("content"))
			if val != "" {
				slog.Info("meta: og:title hit", "value", val)
				meta.Title = cleanText(val)
			}
		}
	})
	c.OnHTML(`meta[name="twitter:title"]`, func(e *colly.HTMLElement) {
		if meta.Title == "" {
			val := strings.TrimSpace(e.Attr("content"))
			if val != "" {
				slog.Info("meta: twitter:title hit", "value", val)
				meta.Title = cleanText(val)
			}
		}
	})
	c.OnHTML(`title`, func(e *colly.HTMLElement) {
		if meta.Title == "" {
			val := strings.TrimSpace(e.Text)
			if val != "" {
				slog.Info("meta: <title> hit", "value", val)
				meta.Title = cleanText(val)
			}
		}
	})

	// ── Description extraction ────────────────────────────────────────────────

	c.OnHTML(`meta[property="og:description"]`, func(e *colly.HTMLElement) {
		if meta.Description == "" {
			val := strings.TrimSpace(e.Attr("content"))
			if val != "" {
				slog.Info("meta: og:description hit", "value", val)
				meta.Description = cleanText(val)
			}
		}
	})
	c.OnHTML(`meta[name="description"]`, func(e *colly.HTMLElement) {
		if meta.Description == "" {
			val := strings.TrimSpace(e.Attr("content"))
			if val != "" {
				slog.Info("meta: description hit", "value", val)
				meta.Description = cleanText(val)
			}
		}
	})
	c.OnHTML(`meta[name="twitter:description"]`, func(e *colly.HTMLElement) {
		if meta.Description == "" {
			val := strings.TrimSpace(e.Attr("content"))
			if val != "" {
				slog.Info("meta: twitter:description hit", "value", val)
				meta.Description = cleanText(val)
			}
		}
	})

	// ── Icon extraction ───────────────────────────────────────────────────────

	c.OnHTML(`link[rel~="icon"]`, func(e *colly.HTMLElement) {
		if meta.IconURL == "" {
			href := strings.TrimSpace(e.Attr("href"))
			if href != "" {
				slog.Info("meta: icon link hit", "href", href)
				meta.IconURL = resolveURL(href, baseURL)
			}
		}
	})
	c.OnHTML(`link[rel="apple-touch-icon"]`, func(e *colly.HTMLElement) {
		if meta.IconURL == "" {
			href := strings.TrimSpace(e.Attr("href"))
			if href != "" {
				slog.Info("meta: apple-touch-icon hit", "href", href)
				meta.IconURL = resolveURL(href, baseURL)
			}
		}
	})
	c.OnHTML(`meta[property="og:image"]`, func(e *colly.HTMLElement) {
		if meta.IconURL == "" {
			val := strings.TrimSpace(e.Attr("content"))
			if val != "" {
				slog.Info("meta: og:image hit", "value", val)
				meta.IconURL = resolveURL(val, baseURL)
			}
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		slog.Error("meta: colly request error", "url", rawURL, "status", r.StatusCode, "error", err)
	})

	c.OnResponse(func(r *colly.Response) {
		slog.Info("meta: response received", "url", rawURL, "status", r.StatusCode, "bytes", len(r.Body))
	})

	if err := c.Visit(rawURL); err != nil {
		slog.Error("meta: colly visit failed", "url", rawURL, "error", err)
		return PageMeta{}, fmt.Errorf("failed to fetch URL: %w", err)
	}

	// Final fallback for icon: /favicon.ico at domain root
	if meta.IconURL == "" {
		parsed, err := url.Parse(rawURL)
		if err == nil {
			meta.IconURL = parsed.Scheme + "://" + parsed.Host + "/favicon.ico"
		}
	}

	// Known-site icon overrides (sites with non-standard or broken favicon discovery)
	if strings.Contains(strings.ToLower(rawURL), "leetcode.com") {
		slog.Info("meta: leetcode icon override", "url", rawURL)
		meta.IconURL = "https://assets.leetcode.com/static_assets/public/icons/favicon.ico"
	}

	slog.Info("meta: extracted metadata", "url", rawURL, "title", meta.Title, "description", meta.Description, "iconURL", meta.IconURL)
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
