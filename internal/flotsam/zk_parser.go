// Copyright 2024 The zk-org Authors
// Copyright 2024 David Holsgrove
//
// This file contains code copied and adapted from zk (https://github.com/zk-org/zk)
// Original source: internal/core/note_parse.go
// Licensed under GPLv3
//
// SPDX-License-Identifier: GPL-3.0-only

package flotsam

import (
	"crypto/sha256"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/relvacode/iso8601"
	"gopkg.in/djherbis/times.v1"
)

// NoteContent holds the data parsed from the note content.
// Copied from zk with modifications for flotsam use.
type NoteContent struct {
	// Title is the heading of the note.
	Title string
	// Lead is the opening paragraph or section of the note.
	Lead string
	// Body is the content of the note, including the Lead but without the Title.
	Body string
	// Tags is the list of tags found in the note content.
	Tags []string
	// Links is the list of outbound links found in the note.
	Links []Link
	// Additional metadata. For example, extracted from a YAML frontmatter.
	Metadata map[string]interface{}
}

// Note: Link and LinkType are defined in zk_links.go

// ParseNoteContent parses a note's raw content into its components.
// Adapted from zk's note parsing logic for flotsam use.
func ParseNoteContent(content string) (*NoteContent, error) {
	// For now, this is a simplified implementation
	// TODO: Implement proper frontmatter parsing and link extraction
	
	result := &NoteContent{
		Title:    extractTitle(content),
		Lead:     extractLead(content),
		Body:     content,
		Tags:     []string{},
		Links:    []Link{},
		Metadata: make(map[string]interface{}),
	}
	
	return result, nil
}

// extractTitle extracts the title from markdown content.
// Looks for the first H1 heading.
func extractTitle(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(line[2:])
		}
	}
	return ""
}

// extractLead extracts the lead paragraph from markdown content.
// Returns the first paragraph after the title.
func extractLead(content string) string {
	lines := strings.Split(content, "\n")
	foundTitle := false
	var leadLines []string
	inLead := false
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Skip title line
		if strings.HasPrefix(line, "# ") {
			foundTitle = true
			continue
		}
		
		// Skip empty lines after title until we find content
		if foundTitle && line == "" && !inLead {
			continue
		}
		
		// Start collecting lead when we find non-empty content after title
		if foundTitle && line != "" && !inLead {
			inLead = true
			leadLines = append(leadLines, line)
			continue
		}
		
		// Continue collecting lead if we're in the lead paragraph
		if inLead && line != "" && !strings.HasPrefix(line, "#") {
			leadLines = append(leadLines, line)
			continue
		}
		
		// Stop at empty line or next heading
		if inLead && (line == "" || strings.HasPrefix(line, "#")) {
			break
		}
	}
	
	return strings.Join(leadLines, " ")
}

// IsURL returns whether the given string is a valid URL.
// Copied from zk's util/strings package.
func IsURL(s string) bool {
	_, err := url.ParseRequestURI(s)
	if err != nil {
		return false
	}

	u, err := url.Parse(s)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}

	return true
}

// CreationDateFrom extracts creation date from metadata or falls back to file times.
// Adapted from zk's creationDateFrom function.
func CreationDateFrom(metadata map[string]interface{}, times times.Timespec) time.Time {
	// Read the creation date from the YAML frontmatter `date` or `created-at` key.
	for _, key := range []string{"created-at", "date"} {
		if dateVal, ok := metadata[key]; ok {
			if dateStr, ok := dateVal.(string); ok {
				if time, err := iso8601.ParseString(dateStr); err == nil {
					return time
				}
				// Omitting the `T` is common
				if time, err := time.Parse("2006-01-02 15:04:05", dateStr); err == nil {
					return time
				}
				if time, err := time.Parse("2006-01-02 15:04", dateStr); err == nil {
					return time
				}
			}
		}
	}

	if times.HasBirthTime() {
		return times.BirthTime().UTC()
	}

	return time.Now().UTC()
}

// CalculateChecksum calculates SHA256 checksum for content.
// Copied from zk's checksum calculation.
func CalculateChecksum(content []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(content))
}

// ExtractRelativePath converts absolute path to relative path.
// Simplified version of zk's RelPath functionality.
func ExtractRelativePath(absPath, basePath string) (string, error) {
	return filepath.Rel(basePath, absPath)
}