// Copyright (c) 2025 Vice Project
// SPDX-License-Identifier: GPL-3.0-only

// Package flotsam provides utility functions for flotsam note management.
// AIDEV-NOTE: utility functions following Vice patterns and ZK compatibility
package flotsam

import (
	"fmt"
	"html"
	"regexp"
	"strings"
	"time"
	"unicode"
)

// Timestamp formatting utilities for consistent time handling across flotsam

// FormatTimestamp formats a time.Time to ZK-compatible ISO8601 string.
// AIDEV-NOTE: matches ZK frontmatter timestamp format for compatibility
func FormatTimestamp(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05Z")
}

// FormatTimestampHuman formats a time.Time to human-readable string.
// AIDEV-NOTE: for display purposes in UI components
func FormatTimestampHuman(t time.Time) string {
	return t.Format("2006-01-02 15:04")
}

// ParseTimestamp parses various timestamp formats commonly found in notes.
// AIDEV-NOTE: handles ZK formats plus common variations for robustness
func ParseTimestamp(s string) (time.Time, error) {
	// Try various formats in order of preference
	formats := []string{
		time.RFC3339,           // "2006-01-02T15:04:05Z07:00"
		"2006-01-02T15:04:05Z", // ZK format
		"2006-01-02 15:04:05",  // Common format
		"2006-01-02 15:04",     // Common short format
		"2006-01-02",           // Date only
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t.UTC(), nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse timestamp: %s", s)
}

// NowTimestamp returns current time formatted as ZK-compatible timestamp.
// AIDEV-NOTE: convenience function for consistent timestamp generation
func NowTimestamp() string {
	return FormatTimestamp(time.Now())
}

// Content sanitization utilities for safe note handling

// SanitizeTitle removes problematic characters from note titles.
// AIDEV-NOTE: ensures titles are safe for filenames and YAML frontmatter
func SanitizeTitle(title string) string {
	if title == "" {
		return ""
	}

	// Remove leading/trailing whitespace
	title = strings.TrimSpace(title)

	// Replace problematic characters with safe alternatives
	title = strings.ReplaceAll(title, "\n", " ")
	title = strings.ReplaceAll(title, "\r", " ")
	title = strings.ReplaceAll(title, "\t", " ")

	// Collapse multiple spaces into single spaces
	multiSpace := regexp.MustCompile(`\s+`)
	title = multiSpace.ReplaceAllString(title, " ")

	return title
}

// SanitizeContent removes or escapes problematic content in note body.
// AIDEV-NOTE: basic sanitization for markdown content safety
func SanitizeContent(content string) string {
	if content == "" {
		return ""
	}

	// Remove null bytes and other control characters except newlines/tabs first
	var cleaned strings.Builder
	for _, r := range content {
		if unicode.IsControl(r) && r != '\n' && r != '\t' && r != '\r' {
			continue
		}
		cleaned.WriteRune(r)
	}

	// Then escape HTML entities to prevent injection
	return html.EscapeString(cleaned.String())
}

// SanitizeTag ensures tags are safe and consistent.
// AIDEV-NOTE: follows ZK tag conventions for compatibility
func SanitizeTag(tag string) string {
	if tag == "" {
		return ""
	}

	// Remove whitespace and convert to lowercase
	tag = strings.ToLower(strings.TrimSpace(tag))

	// Replace spaces and problematic chars with hyphens
	tag = regexp.MustCompile(`[^\w\-]+`).ReplaceAllString(tag, "-")

	// Remove leading/trailing hyphens and collapse multiple hyphens
	tag = strings.Trim(tag, "-")
	tag = regexp.MustCompile(`-+`).ReplaceAllString(tag, "-")

	return tag
}

// SanitizeTags processes a slice of tags and returns cleaned, deduplicated tags.
// AIDEV-NOTE: ensures tag consistency across flotsam notes
func SanitizeTags(tags []string) []string {
	if len(tags) == 0 {
		return nil
	}

	seen := make(map[string]bool)
	var result []string

	for _, tag := range tags {
		cleaned := SanitizeTag(tag)
		if cleaned != "" && !seen[cleaned] {
			seen[cleaned] = true
			result = append(result, cleaned)
		}
	}

	return result
}

// File and path utilities for flotsam note management

// GenerateNoteFilename creates a safe filename for a flotsam note.
// AIDEV-NOTE: ZK-compatible filename pattern using ID + .md extension
func GenerateNoteFilename(id string) string {
	return fmt.Sprintf("%s.md", id)
}

// ExtractIDFromFilename extracts the ID from a flotsam note filename.
// AIDEV-NOTE: reverse of GenerateNoteFilename for file discovery
func ExtractIDFromFilename(filename string) string {
	if !strings.HasSuffix(filename, ".md") {
		return ""
	}
	return strings.TrimSuffix(filename, ".md")
}

// IsFlotsamFile checks if a filename is a valid flotsam note file.
// AIDEV-NOTE: filters files during directory scanning
func IsFlotsamFile(filename string) bool {
	if !strings.HasSuffix(filename, ".md") {
		return false
	}

	id := ExtractIDFromFilename(filename)
	return IsValidFlotsamID(id)
}

// IsValidFlotsamID validates that an ID matches flotsam/ZK format requirements.
// AIDEV-NOTE: ZK-compatible ID validation - 4-char alphanumeric lowercase
func IsValidFlotsamID(id string) bool {
	if id == "" {
		return false
	}
	matched, _ := regexp.MatchString(`^[a-z0-9]{4}$`, id)
	return matched
}

// String manipulation utilities for note processing

// TruncateString truncates a string to maxLength, adding ellipsis if truncated.
// AIDEV-NOTE: for display purposes in UI lists and previews
func TruncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}

	if maxLength <= 3 {
		return s[:maxLength]
	}

	return s[:maxLength-3] + "..."
}

// SlugifyTitle converts a title to a URL-safe slug.
// AIDEV-NOTE: useful for generating alternative filename patterns if needed
func SlugifyTitle(title string) string {
	if title == "" {
		return ""
	}

	// Convert to lowercase and replace spaces/punctuation with hyphens
	slug := strings.ToLower(title)
	slug = regexp.MustCompile(`[^\w\s\-]+`).ReplaceAllString(slug, "")
	slug = regexp.MustCompile(`[\s\-]+`).ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")

	return slug
}

// Helper utilities for common operations

// IsEmptyOrWhitespace checks if a string is empty or contains only whitespace.
// AIDEV-NOTE: validation helper following Vice patterns
func IsEmptyOrWhitespace(s string) bool {
	return strings.TrimSpace(s) == ""
}

// NormalizeWhitespace normalizes whitespace in a string for consistent processing.
// AIDEV-NOTE: cleanup utility for user input processing
func NormalizeWhitespace(s string) string {
	// Replace various whitespace chars with spaces
	s = regexp.MustCompile(`[\t\r\n]+`).ReplaceAllString(s, " ")

	// Collapse multiple spaces
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")

	return strings.TrimSpace(s)
}
