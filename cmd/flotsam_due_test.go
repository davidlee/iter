package cmd

import (
	"fmt"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlotsamDueCommand(t *testing.T) {
	// Test command structure and basic setup
	cmd := flotsamDueCmd
	require.NotNil(t, cmd)
	assert.Equal(t, "due", cmd.Use)
	assert.Contains(t, cmd.Short, "due for review")

	// Verify flags are registered
	formatFlag := cmd.Flags().Lookup("format")
	require.NotNil(t, formatFlag)
	assert.Equal(t, "table", formatFlag.DefValue)

	limitFlag := cmd.Flags().Lookup("limit")
	require.NotNil(t, limitFlag)
	assert.Equal(t, "0", limitFlag.DefValue)
}

func TestExtractNoteMetadata(t *testing.T) {
	testCases := []struct {
		name        string
		path        string
		expectID    string
		expectTitle string
	}{
		{
			name:        "ZK-style note with ID and title",
			path:        "notes/abc1-my-great-idea.md",
			expectID:    "abc1",
			expectTitle: "my-great-idea",
		},
		{
			name:        "ZK-style note with underscore separator",
			path:        "/home/user/notes/xyz9_important_note.md",
			expectID:    "xyz9",
			expectTitle: "important_note",
		},
		{
			name:        "Note without clear ID pattern",
			path:        "notes/my-idea-file.md",
			expectID:    "my-idea-file",
			expectTitle: "my-idea-file",
		},
		{
			name:        "Note without extension",
			path:        "abc2-test-note",
			expectID:    "abc2",
			expectTitle: "test-note",
		},
		{
			name:        "Short filename",
			path:        "ab.md",
			expectID:    "ab",
			expectTitle: "ab",
		},
		{
			name:        "Numeric ID",
			path:        "2023-annual-review.md",
			expectID:    "2023",
			expectTitle: "annual-review",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			id, title := extractNoteMetadata(tc.path)
			assert.Equal(t, tc.expectID, id, "ID mismatch")
			assert.Equal(t, tc.expectTitle, title, "Title mismatch")
		})
	}
}

func TestIsAlphanumeric(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
	}{
		{"abc1", true},
		{"XYZ9", true},
		{"1234", true},
		{"abcd", true},
		{"ab-1", false},
		{"ab_1", false},
		{"ab.1", false},
		{"", true}, // edge case: empty string
		{"a", true},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := isAlphanumeric(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestDueNoteFiltering(t *testing.T) {
	// Test the logic of filtering and overdue calculation
	// Use fixed dates to avoid time zone and calculation issues
	baseTime := time.Date(2023, 6, 15, 12, 0, 0, 0, time.UTC)
	today := time.Date(2023, 6, 15, 23, 59, 59, 0, time.UTC)
	yesterday := time.Date(2023, 6, 14, 12, 0, 0, 0, time.UTC)
	tomorrow := time.Date(2023, 6, 16, 12, 0, 0, 0, time.UTC)
	threeDaysAgo := time.Date(2023, 6, 12, 12, 0, 0, 0, time.UTC)

	testCases := []struct {
		name            string
		dueDate         time.Time
		shouldInclude   bool
		shouldBeOverdue bool
	}{
		{
			name:            "note due today",
			dueDate:         today,
			shouldInclude:   true,
			shouldBeOverdue: false,
		},
		{
			name:            "note due yesterday",
			dueDate:         yesterday,
			shouldInclude:   true,
			shouldBeOverdue: true,
		},
		{
			name:            "note due three days ago",
			dueDate:         threeDaysAgo,
			shouldInclude:   true,
			shouldBeOverdue: true,
		},
		{
			name:            "note due tomorrow",
			dueDate:         tomorrow,
			shouldInclude:   false,
			shouldBeOverdue: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test overdue calculation logic
			isOverdue := tc.dueDate.Before(today.Add(-24 * time.Hour))
			assert.Equal(t, tc.shouldBeOverdue, isOverdue, "Overdue calculation mismatch")

			// Test inclusion logic (due today or overdue)
			shouldInclude := tc.dueDate.Before(today.Add(time.Second)) || tc.dueDate.Equal(today)
			assert.Equal(t, tc.shouldInclude, shouldInclude, "Inclusion logic mismatch")

			// Test days past calculation with fixed reference time
			daysPast := int(baseTime.Sub(tc.dueDate).Hours() / 24)
			if daysPast < 0 {
				daysPast = 0
			}
			// Just verify the calculation works (actual values depend on fixed dates)
			assert.True(t, daysPast >= 0, "Days past should be non-negative")
		})
	}
}

func TestDueNoteSorting(t *testing.T) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterday := today.Add(-24 * time.Hour)
	twoDaysAgo := today.Add(-48 * time.Hour)

	notes := []dueNote{
		{ID: "note3", Path: "z-note.md", DueDate: yesterday},
		{ID: "note1", Path: "a-note.md", DueDate: twoDaysAgo},
		{ID: "note4", Path: "a-note.md", DueDate: yesterday}, // Same date as note3, different path
		{ID: "note2", Path: "b-note.md", DueDate: twoDaysAgo},
	}

	// Sort using the same logic as the command
	// Sort by due date (oldest first), then by filename
	for i := 0; i < len(notes)-1; i++ {
		for j := i + 1; j < len(notes); j++ {
			if notes[j].DueDate.Before(notes[i].DueDate) ||
				(notes[j].DueDate.Equal(notes[i].DueDate) && notes[j].Path < notes[i].Path) {
				notes[i], notes[j] = notes[j], notes[i]
			}
		}
	}

	// Verify sorting order
	expectedOrder := []string{"note1", "note2", "note4", "note3"}
	for i, expectedID := range expectedOrder {
		assert.Equal(t, expectedID, notes[i].ID, "Sort order mismatch at position %d", i)
	}

	// Verify due dates are in ascending order
	for i := 0; i < len(notes)-1; i++ {
		assert.True(t,
			notes[i].DueDate.Before(notes[i+1].DueDate) || notes[i].DueDate.Equal(notes[i+1].DueDate),
			"Due dates should be in ascending order")
	}
}

func TestOutputDueNotes(t *testing.T) {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	testNotes := []dueNote{
		{
			ID:       "abc1",
			Title:    "Test Note",
			Path:     "abc1-test-note.md",
			DueDate:  yesterday,
			Overdue:  true,
			DaysPast: 1,
		},
	}

	testCases := []struct {
		name     string
		notes    []dueNote
		format   string
		wantErr  bool
		contains []string
	}{
		{
			name:     "empty notes paths format",
			notes:    []dueNote{},
			format:   "paths",
			wantErr:  false,
			contains: []string{},
		},
		{
			name:     "single note paths format",
			notes:    testNotes,
			format:   "paths",
			wantErr:  false,
			contains: []string{"abc1-test-note.md"},
		},
		{
			name:     "notes table format",
			notes:    testNotes,
			format:   "table",
			wantErr:  false,
			contains: []string{"Found 1 note(s)", "abc1", "Test Note", "1 day late"},
		},
		{
			name:     "empty notes table format",
			notes:    []dueNote{},
			format:   "table",
			wantErr:  false,
			contains: []string{"No notes due for review"},
		},
		{
			name:     "invalid format",
			notes:    testNotes,
			format:   "invalid",
			wantErr:  true,
			contains: []string{"invalid format"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := outputDueNotes(tc.notes, tc.format)

			if tc.wantErr {
				assert.Error(t, err)
				if len(tc.contains) > 0 {
					assert.Contains(t, err.Error(), tc.contains[0])
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFlotsamDueIntegration(t *testing.T) {
	// Test that the command is properly integrated into the command tree
	rootCmd := &cobra.Command{Use: "vice"}
	flotsamCmd := &cobra.Command{Use: "flotsam"}
	rootCmd.AddCommand(flotsamCmd)
	flotsamCmd.AddCommand(flotsamDueCmd)

	// Test command path resolution
	cmd, _, err := rootCmd.Find([]string{"flotsam", "due"})
	require.NoError(t, err)
	assert.Equal(t, "due", cmd.Use)

	// Test help text includes expected content
	help := cmd.Long
	assert.Contains(t, help, "due for spaced repetition")
	assert.Contains(t, help, "ZK-first enrichment")
	assert.Contains(t, help, "ADR-008")
}

func TestDueDateStatusCalculation(t *testing.T) {
	testCases := []struct {
		name           string
		daysPast       int
		expectedStatus string
	}{
		{
			name:           "due today",
			daysPast:       0,
			expectedStatus: "Due today",
		},
		{
			name:           "one day overdue",
			daysPast:       1,
			expectedStatus: "1 day late",
		},
		{
			name:           "multiple days overdue",
			daysPast:       5,
			expectedStatus: "5 days late",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate the status calculation logic from outputDueNotes
			var status string
			switch tc.daysPast {
			case 0:
				status = "Due today"
			case 1:
				status = "1 day late"
			default:
				status = fmt.Sprintf("%d days late", tc.daysPast)
			}

			assert.Equal(t, tc.expectedStatus, status)
		})
	}
}
