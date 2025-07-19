// Package flotsam provides Unix interop functionality for flotsam notes.
// This file contains tests for zk delegation link operations.
package flotsam

import (
	"testing"
)

// TestGetBacklinks tests the zk delegation for backlink retrieval.
// This test will skip if zk is not available.
func TestGetBacklinks(t *testing.T) {
	if !isZKAvailable() {
		t.Skip("zk command not available, skipping zk delegation test")
	}

	// Test with a non-existent note (zk may return error, which is expected)
	backlinks, err := GetBacklinks("non-existent-note.md")

	// If there's an error, it should be due to zk command failure (expected)
	// If no error, backlinks should be a valid slice
	if err == nil && backlinks == nil {
		t.Error("GetBacklinks should return empty slice, not nil when no error")
	}

	// This test mainly verifies the function exists and handles zk delegation
	t.Logf("GetBacklinks test completed - err: %v, results: %d items", err, len(backlinks))
}

// TestGetOutboundLinks tests the zk delegation for outbound link retrieval.
// This test will skip if zk is not available.
func TestGetOutboundLinks(t *testing.T) {
	if !isZKAvailable() {
		t.Skip("zk command not available, skipping zk delegation test")
	}

	// Test with a non-existent note (zk may return error, which is expected)
	outbound, err := GetOutboundLinks("non-existent-note.md")

	// If there's an error, it should be due to zk command failure (expected)
	// If no error, outbound should be a valid slice
	if err == nil && outbound == nil {
		t.Error("GetOutboundLinks should return empty slice, not nil when no error")
	}

	t.Logf("GetOutboundLinks test completed - err: %v, results: %d items", err, len(outbound))
}

// TestGetLinkedNotes tests the combined backlinks and outbound links retrieval.
// This test will skip if zk is not available.
func TestGetLinkedNotes(t *testing.T) {
	if !isZKAvailable() {
		t.Skip("zk command not available, skipping zk delegation test")
	}

	// Test with a non-existent note (both calls may fail, which is expected)
	backlinks, outbound, err := GetLinkedNotes("non-existent-note.md")

	// This test mainly verifies the function exists and handles zk delegation
	t.Logf("GetLinkedNotes test completed - err: %v, backlinks: %d, outbound: %d",
		err, len(backlinks), len(outbound))
}

// TestParseZKPathOutput tests the parsing of zk command output.
func TestParseZKPathOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty output",
			input:    "",
			expected: []string{},
		},
		{
			name:     "single path",
			input:    "note1.md",
			expected: []string{"note1.md"},
		},
		{
			name:     "multiple paths",
			input:    "note1.md\nnote2.md\nnote3.md",
			expected: []string{"note1.md", "note2.md", "note3.md"},
		},
		{
			name:     "paths with whitespace",
			input:    "  note1.md  \n  note2.md  \n",
			expected: []string{"note1.md", "note2.md"},
		},
		{
			name:     "paths with empty lines",
			input:    "note1.md\n\nnote2.md\n\n",
			expected: []string{"note1.md", "note2.md"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseZKPathOutput(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d paths, got %d", len(tt.expected), len(result))
				return
			}

			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("Expected path %q at index %d, got %q", expected, i, result[i])
				}
			}
		})
	}
}

// TestBackwardCompatibilityBuildBacklinkIndex tests that the deprecated function still works.
func TestBackwardCompatibilityBuildBacklinkIndex(t *testing.T) {
	notes := map[string]string{
		"note1": "This links to [[note2]] and [[note3]].",
		"note2": "This links to [[note3]] and [[note1]].",
		"note3": "This links to [[note1]].",
	}

	backlinks := BuildBacklinkIndex(notes)

	// note1 should be linked from note2 and note3
	if len(backlinks["note1"]) != 2 {
		t.Errorf("Expected 2 backlinks to note1, got %d", len(backlinks["note1"]))
	}

	// note2 should be linked from note1
	if len(backlinks["note2"]) != 1 {
		t.Errorf("Expected 1 backlink to note2, got %d", len(backlinks["note2"]))
	}

	// note3 should be linked from note1 and note2
	if len(backlinks["note3"]) != 2 {
		t.Errorf("Expected 2 backlinks to note3, got %d", len(backlinks["note3"]))
	}
}
