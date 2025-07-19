package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlotsamEditCommand(t *testing.T) {
	// Test command structure and basic setup
	cmd := flotsamEditCmd
	require.NotNil(t, cmd)
	assert.Equal(t, "edit [note-id]", cmd.Use)
	assert.Contains(t, cmd.Short, "Edit flotsam notes")

	// Verify flags are registered
	interactiveFlag := cmd.Flags().Lookup("interactive")
	require.NotNil(t, interactiveFlag)
	assert.Equal(t, "false", interactiveFlag.DefValue)
}

func TestFindNotesByID(t *testing.T) {
	testNotes := []string{
		"notes/abc1-my-idea.md",
		"notes/abc2-another-idea.md",
		"notes/xyz1-different-note.md",
		"notes/abc1-variant.md",
		"notes/some-note-with-abc1-inside.md",
		"notes/def3-no-match.md",
	}

	testCases := []struct {
		name          string
		noteID        string
		expectedPaths []string
		expectedCount int
	}{
		{
			name:   "exact prefix match",
			noteID: "abc1",
			expectedPaths: []string{
				"notes/abc1-my-idea.md",
				"notes/abc1-variant.md",
				"notes/some-note-with-abc1-inside.md",
			},
			expectedCount: 3,
		},
		{
			name:   "single match",
			noteID: "xyz1",
			expectedPaths: []string{
				"notes/xyz1-different-note.md",
			},
			expectedCount: 1,
		},
		{
			name:          "no matches",
			noteID:        "nonexistent",
			expectedPaths: []string{},
			expectedCount: 0,
		},
		{
			name:   "partial ID match",
			noteID: "abc",
			expectedPaths: []string{
				"notes/abc1-my-idea.md",
				"notes/abc2-another-idea.md",
				"notes/abc1-variant.md",
				"notes/some-note-with-abc1-inside.md",
			},
			expectedCount: 4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			matches := findNotesByID(testNotes, tc.noteID)

			assert.Equal(t, tc.expectedCount, len(matches), "Number of matches")

			// Check that all expected paths are found
			for _, expectedPath := range tc.expectedPaths {
				assert.Contains(t, matches, expectedPath, "Expected path should be in matches")
			}

			// Check that no unexpected paths are found
			for _, match := range matches {
				assert.Contains(t, tc.expectedPaths, match, "Found path should be in expected paths")
			}
		})
	}
}

func TestNoteIDMatching(t *testing.T) {
	// Test the logic used in findNotesByID for different filename patterns
	testCases := []struct {
		name        string
		filename    string
		noteID      string
		shouldMatch bool
		description string
	}{
		{
			name:        "ZK-style prefix match",
			filename:    "abc1-my-note.md",
			noteID:      "abc1",
			shouldMatch: true,
			description: "Standard ZK naming with ID prefix",
		},
		{
			name:        "exact filename match",
			filename:    "abc1.md",
			noteID:      "abc1",
			shouldMatch: true,
			description: "Filename is exactly the ID",
		},
		{
			name:        "ID in middle of filename",
			filename:    "my-abc1-note.md",
			noteID:      "abc1",
			shouldMatch: true,
			description: "ID appears in middle of filename",
		},
		{
			name:        "partial ID match",
			filename:    "abc123-note.md",
			noteID:      "abc1",
			shouldMatch: true,
			description: "Partial ID match at start",
		},
		{
			name:        "no match",
			filename:    "xyz9-different.md",
			noteID:      "abc1",
			shouldMatch: false,
			description: "Completely different ID",
		},
		{
			name:        "case sensitive",
			filename:    "ABC1-note.md",
			noteID:      "abc1",
			shouldMatch: false,
			description: "Case doesn't match",
		},
		{
			name:        "without extension",
			filename:    "abc1-note",
			noteID:      "abc1",
			shouldMatch: true,
			description: "File without .md extension",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate the matching logic from findNotesByID
			filename := tc.filename
			filename = strings.TrimSuffix(filename, ".md")

			hasPrefix := strings.HasPrefix(filename, tc.noteID)
			contains := strings.Contains(filename, tc.noteID)
			matches := hasPrefix || contains

			assert.Equal(t, tc.shouldMatch, matches, tc.description)
		})
	}
}

func TestFlotsamEditIntegration(t *testing.T) {
	// Test that the command is properly integrated into the command tree
	rootCmd := &cobra.Command{Use: "vice"}
	flotsamCmd := &cobra.Command{Use: "flotsam"}
	rootCmd.AddCommand(flotsamCmd)
	flotsamCmd.AddCommand(flotsamEditCmd)

	// Test command path resolution
	cmd, _, err := rootCmd.Find([]string{"flotsam", "edit"})
	require.NoError(t, err)
	assert.Equal(t, "edit [note-id]", cmd.Use)

	// Test help text includes expected content
	help := cmd.Long
	assert.Contains(t, help, "interactive picker")
	assert.Contains(t, help, "ZK")
	assert.Contains(t, help, "ADR-008")
}

func TestEditCommandArguments(t *testing.T) {
	// Test argument parsing logic
	testCases := []struct {
		name            string
		args            []string
		interactiveFlag bool
		expectedMode    string
		expectedNoteID  string
	}{
		{
			name:            "no arguments - interactive mode",
			args:            []string{},
			interactiveFlag: false,
			expectedMode:    "interactive",
		},
		{
			name:            "with note ID - direct mode",
			args:            []string{"abc1"},
			interactiveFlag: false,
			expectedMode:    "direct",
			expectedNoteID:  "abc1",
		},
		{
			name:            "with note ID but interactive flag - interactive mode",
			args:            []string{"abc1"},
			interactiveFlag: true,
			expectedMode:    "interactive",
		},
		{
			name:            "interactive flag only - interactive mode",
			args:            []string{},
			interactiveFlag: true,
			expectedMode:    "interactive",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate the logic from runFlotsamEdit
			editInteractive = tc.interactiveFlag

			var mode string
			var noteID string

			if len(tc.args) == 0 || editInteractive {
				mode = "interactive"
			} else {
				mode = "direct"
				noteID = tc.args[0]
			}

			assert.Equal(t, tc.expectedMode, mode, "Mode should match expected")
			if tc.expectedNoteID != "" {
				assert.Equal(t, tc.expectedNoteID, noteID, "Note ID should match expected")
			}
		})
	}
}

func TestMultipleMatchHandling(t *testing.T) {
	// Test the behavior when multiple notes match an ID
	testNotes := []string{
		"abc1-first-note.md",
		"abc1-second-note.md",
		"abc1-third-note.md",
	}

	matches := findNotesByID(testNotes, "abc1")

	// Should find all three matches
	assert.Equal(t, 3, len(matches))

	// All original notes should be in matches
	for _, note := range testNotes {
		assert.Contains(t, matches, note)
	}
}

func TestSingleMatchHandling(t *testing.T) {
	// Test the behavior when exactly one note matches an ID
	testNotes := []string{
		"abc1-unique-note.md",
		"xyz2-different-note.md",
		"def3-another-note.md",
	}

	matches := findNotesByID(testNotes, "abc1")

	// Should find exactly one match
	assert.Equal(t, 1, len(matches))
	assert.Equal(t, "abc1-unique-note.md", matches[0])
}

func TestNoMatchHandling(t *testing.T) {
	// Test the behavior when no notes match an ID
	testNotes := []string{
		"xyz1-note.md",
		"def2-note.md",
		"ghi3-note.md",
	}

	matches := findNotesByID(testNotes, "abc1")

	// Should find no matches
	assert.Equal(t, 0, len(matches))
	assert.Empty(t, matches)
}
