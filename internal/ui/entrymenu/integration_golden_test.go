// Package entrymenu golden file testing POC with teatest.
// AIDEV-NOTE: poc-golden-files; evaluating golden file testing for UI regression
package entrymenu

import (
	"io"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"

	"github.com/davidlee/vice/internal/models"
)

// TestEntryMenuGoldenFiles_POC tests golden file functionality for UI regression testing.
// Purpose: Evaluate if golden files provide value for detecting UI layout changes
func TestEntryMenuGoldenFiles_POC(t *testing.T) {
	// Create predictable test data
	habits := []models.Habit{
		{ID: "exercise", Title: "Exercise", HabitType: models.SimpleHabit},
		{ID: "read", Title: "Read", HabitType: models.SimpleHabit},
	}

	entries := map[string]models.HabitEntry{
		"exercise": {
			HabitID: "exercise",
			Status:  models.EntryCompleted,
		},
	}

	model := NewEntryMenuModelForTesting(habits, entries)

	// Use consistent terminal size for reproducible output
	tm := teatest.NewTestModel(
		t, model,
		teatest.WithInitialTermSize(60, 20),
	)

	t.Cleanup(func() {
		if err := tm.Quit(); err != nil {
			t.Fatal(err)
		}
	})

	// Let UI stabilize
	time.Sleep(100 * time.Millisecond)

	// Navigate to incomplete habit
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	time.Sleep(50 * time.Millisecond)

	// Quit to get final output
	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})

	// Get final output
	output := tm.FinalOutput(t, teatest.WithFinalTimeout(time.Second))
	outputBytes, err := io.ReadAll(output)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	// For POC, test on raw output (ANSI sequences included)
	rawOutput := string(outputBytes)

	// Golden file testing - uncomment to update golden files
	// teatest.RequireEqualOutput(t, outputBytes)

	// For POC, just validate key content is present in raw output
	expectedContent := []string{
		"Entry Menu",
		"Exercise",
		"Read",
		"completed", // Part of progress text
		"Return: menu",
	}

	for _, expected := range expectedContent {
		if !strings.Contains(rawOutput, expected) {
			t.Errorf("Expected output to contain %q, but it didn't", expected)
		}
	}

	t.Logf("Raw output length: %d characters", len(rawOutput))
	t.Logf("Contains ANSI sequences: %t", strings.Contains(rawOutput, "\x1b["))

	// Clean output for preview
	cleanOutput := stripANSI(rawOutput)
	t.Logf("Clean output preview:\n%s", cleanOutput[:minInt(300, len(cleanOutput))])
}

// stripANSI removes ANSI escape sequences for more stable golden file testing
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func stripANSI(s string) string {
	// Simple ANSI stripping - more sophisticated libraries exist
	result := strings.Builder{}
	inEscape := false

	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' || r == 'K' || r == 'H' || r == 'J' {
				inEscape = false
			}
			continue
		}
		result.WriteRune(r)
	}

	return result.String()
}
