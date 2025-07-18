package flotsam

import (
	"os"
	"testing"
	"time"

	"gopkg.in/djherbis/times.v1"
)

func TestParseNoteContent(t *testing.T) {
	content := `---
title: Test Note
tags: [test, example]
---

# Test Note

This is the lead paragraph of the note.

This is the body content that follows.
`

	result, err := ParseNoteContent(content)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Title != "Test Note" {
		t.Errorf("Expected title 'Test Note', got '%s'", result.Title)
	}

	if result.Lead != "This is the lead paragraph of the note." {
		t.Errorf("Expected lead paragraph, got '%s'", result.Lead)
	}

	if result.Body != content {
		t.Errorf("Expected body to be full content")
	}
}

func TestExtractTitle(t *testing.T) {
	tests := []struct {
		content  string
		expected string
	}{
		{"# Hello World", "Hello World"},
		{"## Not a title\n# Real Title", "Real Title"},
		{"No title here", ""},
		{"# Title with trailing spaces   ", "Title with trailing spaces"},
	}

	for _, test := range tests {
		result := extractTitle(test.content)
		if result != test.expected {
			t.Errorf("extractTitle(%q) = %q, expected %q", test.content, result, test.expected)
		}
	}
}

func TestExtractLead(t *testing.T) {
	tests := []struct {
		content  string
		expected string
	}{
		{"# Title\n\nThis is the lead.", "This is the lead."},
		{"# Title\nThis is the lead.", "This is the lead."},
		{"No title\nThis should not be lead.", ""},
		{"# Title\n\nFirst paragraph.\n\nSecond paragraph.", "First paragraph."},
	}

	for _, test := range tests {
		result := extractLead(test.content)
		if result != test.expected {
			t.Errorf("extractLead(%q) = %q, expected %q", test.content, result, test.expected)
		}
	}
}

func TestIsURL(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"https://example.com", true},
		{"http://example.com", true},
		{"ftp://example.com", true},
		{"example.com", false},
		{"not a url", false},
		{"", false},
	}

	for _, test := range tests {
		result := IsURL(test.input)
		if result != test.expected {
			t.Errorf("IsURL(%q) = %t, expected %t", test.input, result, test.expected)
		}
	}
}

func TestCreationDateFrom(t *testing.T) {
	// Test with created-at metadata - should parse from metadata
	metadata := map[string]interface{}{
		"created-at": "2024-01-01T10:00:00Z",
	}

	// Use a temporary file to get a real Timespec
	tmpFile := "/tmp/test_times"
	// AIDEV-NOTE: use 0600 permissions for temp files (security best practice)
	err := os.WriteFile(tmpFile, []byte("test"), 0600)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		// AIDEV-NOTE: always check error return from os.Remove to avoid linter warnings
		if err := os.Remove(tmpFile); err != nil {
			t.Logf("Warning: failed to remove temp file %s: %v", tmpFile, err)
		}
	}()

	fileStats, err := times.Stat(tmpFile)
	if err != nil {
		t.Fatalf("Failed to get file times: %v", err)
	}

	result := CreationDateFrom(metadata, fileStats)

	expected := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	if !result.Equal(expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}

	// Test with date metadata
	metadata = map[string]interface{}{
		"date": "2024-01-01 10:00:00",
	}

	result = CreationDateFrom(metadata, fileStats)
	if !result.Equal(expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}

	// Test fallback to file times when no metadata
	emptyMetadata := map[string]interface{}{}
	result = CreationDateFrom(emptyMetadata, fileStats)

	// Should fallback to file modification time or current time
	if result.IsZero() {
		t.Errorf("Expected non-zero time, got zero")
	}
}

func TestCalculateChecksum(t *testing.T) {
	content := []byte("test content")
	checksum := CalculateChecksum(content)

	// Should be consistent
	checksum2 := CalculateChecksum(content)
	if checksum != checksum2 {
		t.Errorf("Checksum should be consistent")
	}

	// Should be different for different content
	checksum3 := CalculateChecksum([]byte("different content"))
	if checksum == checksum3 {
		t.Errorf("Checksum should be different for different content")
	}
}

func TestExtractRelativePath(t *testing.T) {
	absPath := "/home/user/notes/test.md"
	basePath := "/home/user/notes"

	result, err := ExtractRelativePath(absPath, basePath)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result != "test.md" {
		t.Errorf("Expected 'test.md', got '%s'", result)
	}
}
