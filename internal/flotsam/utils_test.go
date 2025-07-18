// Copyright (c) 2025 Vice Project
// SPDX-License-Identifier: GPL-3.0-only

package flotsam

import (
	"strings"
	"testing"
	"time"
)

// TestFormatTimestamp tests ZK-compatible timestamp formatting
func TestFormatTimestamp(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			name:     "UTC time",
			input:    time.Date(2023, 7, 15, 14, 30, 45, 0, time.UTC),
			expected: "2023-07-15T14:30:45Z",
		},
		{
			name:     "Non-UTC time converts to UTC",
			input:    time.Date(2023, 7, 15, 14, 30, 45, 0, time.FixedZone("EST", -5*3600)),
			expected: "2023-07-15T19:30:45Z",
		},
		{
			name:     "Zero time",
			input:    time.Time{},
			expected: "0001-01-01T00:00:00Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatTimestamp(tt.input)
			if result != tt.expected {
				t.Errorf("FormatTimestamp() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestFormatTimestampHuman tests human-readable timestamp formatting
func TestFormatTimestampHuman(t *testing.T) {
	input := time.Date(2023, 7, 15, 14, 30, 45, 0, time.UTC)
	expected := "2023-07-15 14:30"

	result := FormatTimestampHuman(input)
	if result != expected {
		t.Errorf("FormatTimestampHuman() = %v, want %v", result, expected)
	}
}

// TestParseTimestamp tests parsing various timestamp formats
func TestParseTimestamp(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expectErr bool
	}{
		{
			name:      "RFC3339 format",
			input:     "2023-07-15T14:30:45Z",
			expectErr: false,
		},
		{
			name:      "ZK format",
			input:     "2023-07-15T14:30:45Z",
			expectErr: false,
		},
		{
			name:      "Common format",
			input:     "2023-07-15 14:30:45",
			expectErr: false,
		},
		{
			name:      "Short format",
			input:     "2023-07-15 14:30",
			expectErr: false,
		},
		{
			name:      "Date only",
			input:     "2023-07-15",
			expectErr: false,
		},
		{
			name:      "Invalid format",
			input:     "not-a-date",
			expectErr: true,
		},
		{
			name:      "Empty string",
			input:     "",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTimestamp(tt.input)

			if tt.expectErr {
				if err == nil {
					t.Errorf("ParseTimestamp() expected error for input %v", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("ParseTimestamp() unexpected error: %v", err)
				}
				if result.IsZero() {
					t.Errorf("ParseTimestamp() returned zero time for valid input %v", tt.input)
				}
			}
		})
	}
}

// TestNowTimestamp tests current timestamp generation
func TestNowTimestamp(t *testing.T) {
	result := NowTimestamp()

	// Should be a valid timestamp format
	if _, err := ParseTimestamp(result); err != nil {
		t.Errorf("NowTimestamp() returned invalid timestamp: %v", result)
	}

	// Should contain current year
	currentYear := time.Now().Year()
	if !strings.Contains(result, string(rune(currentYear/1000+48))+string(rune((currentYear/100)%10+48))) {
		t.Errorf("NowTimestamp() doesn't contain current year: %v", result)
	}
}

// TestSanitizeTitle tests title sanitization
func TestSanitizeTitle(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Normal title",
			input:    "My Great Idea",
			expected: "My Great Idea",
		},
		{
			name:     "Title with newlines",
			input:    "My\nGreat\nIdea",
			expected: "My Great Idea",
		},
		{
			name:     "Title with tabs and returns",
			input:    "My\tGreat\rIdea",
			expected: "My Great Idea",
		},
		{
			name:     "Multiple spaces",
			input:    "My    Great     Idea",
			expected: "My Great Idea",
		},
		{
			name:     "Leading/trailing whitespace",
			input:    "  My Great Idea  ",
			expected: "My Great Idea",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Whitespace only",
			input:    "   ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeTitle(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeTitle() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestSanitizeContent tests content sanitization
func TestSanitizeContent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Normal content",
			input:    "This is normal content",
			expected: "This is normal content",
		},
		{
			name:     "Content with HTML",
			input:    "This has <script>alert('xss')</script> tags",
			expected: "This has &lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt; tags",
		},
		{
			name:     "Content with null bytes",
			input:    "This has\x00null bytes",
			expected: "This hasnull bytes",
		},
		{
			name:     "Content with control chars but preserve newlines",
			input:    "Line 1\nLine 2\tTabbed\x01control",
			expected: "Line 1\nLine 2\tTabbedcontrol",
		},
		{
			name:     "Empty content",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeContent(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeContent() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestSanitizeTag tests tag sanitization
func TestSanitizeTag(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Normal tag",
			input:    "programming",
			expected: "programming",
		},
		{
			name:     "Tag with spaces",
			input:    "machine learning",
			expected: "machine-learning",
		},
		{
			name:     "Tag with special chars",
			input:    "C++ programming!",
			expected: "c-programming",
		},
		{
			name:     "Tag with leading/trailing hyphens",
			input:    "-important-",
			expected: "important",
		},
		{
			name:     "Tag with multiple hyphens",
			input:    "very---important",
			expected: "very-important",
		},
		{
			name:     "Mixed case",
			input:    "JavaScript",
			expected: "javascript",
		},
		{
			name:     "Empty tag",
			input:    "",
			expected: "",
		},
		{
			name:     "Whitespace only",
			input:    "   ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeTag(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeTag() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestSanitizeTags tests tag array sanitization
func TestSanitizeTags(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "Normal tags",
			input:    []string{"programming", "golang", "testing"},
			expected: []string{"programming", "golang", "testing"},
		},
		{
			name:     "Tags with duplicates",
			input:    []string{"Programming", "PROGRAMMING", "programming"},
			expected: []string{"programming"},
		},
		{
			name:     "Tags with spaces and special chars",
			input:    []string{"machine learning", "C++", "web-dev"},
			expected: []string{"machine-learning", "c", "web-dev"},
		},
		{
			name:     "Tags with empty entries",
			input:    []string{"valid", "", "   ", "also-valid"},
			expected: []string{"valid", "also-valid"},
		},
		{
			name:     "Empty tag array",
			input:    []string{},
			expected: nil,
		},
		{
			name:     "Nil tag array",
			input:    nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeTags(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("SanitizeTags() length = %v, want %v", len(result), len(tt.expected))
				return
			}

			for i, tag := range result {
				if tag != tt.expected[i] {
					t.Errorf("SanitizeTags()[%d] = %v, want %v", i, tag, tt.expected[i])
				}
			}
		})
	}
}

// TestGenerateNoteFilename tests filename generation
func TestGenerateNoteFilename(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Valid ID",
			input:    "abc1",
			expected: "abc1.md",
		},
		{
			name:     "Another valid ID",
			input:    "xyz9",
			expected: "xyz9.md",
		},
		{
			name:     "Empty ID",
			input:    "",
			expected: ".md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateNoteFilename(tt.input)
			if result != tt.expected {
				t.Errorf("GenerateNoteFilename() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestExtractIDFromFilename tests ID extraction from filenames
func TestExtractIDFromFilename(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Valid markdown file",
			input:    "abc1.md",
			expected: "abc1",
		},
		{
			name:     "Non-markdown file",
			input:    "abc1.txt",
			expected: "",
		},
		{
			name:     "File without extension",
			input:    "abc1",
			expected: "",
		},
		{
			name:     "Empty filename",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractIDFromFilename(tt.input)
			if result != tt.expected {
				t.Errorf("ExtractIDFromFilename() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestIsFlotsamFile tests flotsam file detection
func TestIsFlotsamFile(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Valid flotsam file",
			input:    "abc1.md",
			expected: true,
		},
		{
			name:     "Invalid ID - too long",
			input:    "abcd1.md",
			expected: false,
		},
		{
			name:     "Invalid ID - too short",
			input:    "ab1.md",
			expected: false,
		},
		{
			name:     "Invalid ID - uppercase",
			input:    "ABC1.md",
			expected: false,
		},
		{
			name:     "Non-markdown file",
			input:    "abc1.txt",
			expected: false,
		},
		{
			name:     "Directory",
			input:    "abc1",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsFlotsamFile(tt.input)
			if result != tt.expected {
				t.Errorf("IsFlotsamFile() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestIsValidFlotsamID tests ID validation
func TestIsValidFlotsamID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Valid ID - alphanumeric",
			input:    "abc1",
			expected: true,
		},
		{
			name:     "Valid ID - all letters",
			input:    "abcd",
			expected: true,
		},
		{
			name:     "Valid ID - all numbers",
			input:    "1234",
			expected: true,
		},
		{
			name:     "Invalid ID - too short",
			input:    "ab1",
			expected: false,
		},
		{
			name:     "Invalid ID - too long",
			input:    "abcd1",
			expected: false,
		},
		{
			name:     "Invalid ID - uppercase",
			input:    "ABC1",
			expected: false,
		},
		{
			name:     "Invalid ID - special chars",
			input:    "ab-1",
			expected: false,
		},
		{
			name:     "Empty ID",
			input:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidFlotsamID(tt.input)
			if result != tt.expected {
				t.Errorf("IsValidFlotsamID() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestTruncateString tests string truncation
func TestTruncateString(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		maxLength int
		expected  string
	}{
		{
			name:      "String shorter than max",
			input:     "Short",
			maxLength: 10,
			expected:  "Short",
		},
		{
			name:      "String equal to max",
			input:     "Exactly10!",
			maxLength: 10,
			expected:  "Exactly10!",
		},
		{
			name:      "String longer than max",
			input:     "This is a very long string",
			maxLength: 10,
			expected:  "This is...",
		},
		{
			name:      "Very short max length",
			input:     "Hello",
			maxLength: 3,
			expected:  "Hel",
		},
		{
			name:      "Empty string",
			input:     "",
			maxLength: 10,
			expected:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateString(tt.input, tt.maxLength)
			if result != tt.expected {
				t.Errorf("TruncateString() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestSlugifyTitle tests title to slug conversion
func TestSlugifyTitle(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple title",
			input:    "My Great Idea",
			expected: "my-great-idea",
		},
		{
			name:     "Title with punctuation",
			input:    "Hello, World!",
			expected: "hello-world",
		},
		{
			name:     "Title with numbers",
			input:    "Version 2.0 Release",
			expected: "version-20-release",
		},
		{
			name:     "Title with multiple spaces",
			input:    "Multiple    Spaces",
			expected: "multiple-spaces",
		},
		{
			name:     "Title with hyphens",
			input:    "Already-Hyphenated",
			expected: "already-hyphenated",
		},
		{
			name:     "Empty title",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SlugifyTitle(tt.input)
			if result != tt.expected {
				t.Errorf("SlugifyTitle() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestIsEmptyOrWhitespace tests whitespace detection
func TestIsEmptyOrWhitespace(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Non-empty string",
			input:    "Hello",
			expected: false,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: true,
		},
		{
			name:     "Spaces only",
			input:    "   ",
			expected: true,
		},
		{
			name:     "Mixed whitespace",
			input:    " \t\n ",
			expected: true,
		},
		{
			name:     "String with content and whitespace",
			input:    " Hello ",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsEmptyOrWhitespace(tt.input)
			if result != tt.expected {
				t.Errorf("IsEmptyOrWhitespace() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestNormalizeWhitespace tests whitespace normalization
func TestNormalizeWhitespace(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Normal string",
			input:    "Hello world",
			expected: "Hello world",
		},
		{
			name:     "String with tabs",
			input:    "Hello\tworld",
			expected: "Hello world",
		},
		{
			name:     "String with newlines",
			input:    "Hello\nworld",
			expected: "Hello world",
		},
		{
			name:     "String with multiple spaces",
			input:    "Hello    world",
			expected: "Hello world",
		},
		{
			name:     "Mixed whitespace",
			input:    "  Hello\t\n\r   world  ",
			expected: "Hello world",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Whitespace only",
			input:    " \t\n ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeWhitespace(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeWhitespace() = %v, want %v", result, tt.expected)
			}
		})
	}
}
