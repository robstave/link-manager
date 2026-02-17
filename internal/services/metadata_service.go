package services

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type MetadataService struct{}

func NewMetadataService() *MetadataService {
	return &MetadataService{}
}

// FetchTitle fetches a URL and extracts the <title> tag
func (s *MetadataService) FetchTitle(url string) (string, error) {
	// Create a client with a timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Make the request
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	// Read up to 50KB (enough for <head> in most pages)
	limitedReader := io.LimitReader(resp.Body, 50*1024)
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Extract title using regex
	titleRegex := regexp.MustCompile(`(?i)<title[^>]*>(.*?)</title>`)
	matches := titleRegex.FindStringSubmatch(string(body))
	if len(matches) < 2 {
		return "", fmt.Errorf("no title found")
	}

	// Clean up the title (decode HTML entities, trim whitespace)
	title := strings.TrimSpace(matches[1])
	title = strings.ReplaceAll(title, "\n", " ")
	title = strings.ReplaceAll(title, "\r", " ")
	
	// Basic HTML entity decoding
	title = strings.ReplaceAll(title, "&amp;", "&")
	title = strings.ReplaceAll(title, "&lt;", "<")
	title = strings.ReplaceAll(title, "&gt;", ">")
	title = strings.ReplaceAll(title, "&quot;", "\"")
	title = strings.ReplaceAll(title, "&#39;", "'")

	return title, nil
}
