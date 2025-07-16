// Package entrymenu integration tests using BubbleTea teatest framework.
// AIDEV-NOTE: poc-teatest; evaluating teatest for end-to-end UI testing vs unit tests
package entrymenu

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"

	"davidlee/vice/internal/models"
)

// TestEntryMenuIntegration_POC is a proof-of-concept integration test using teatest.
// Tests: menu navigation → habit selection → mock entry flow
// Purpose: Evaluate teatest framework utility vs current unit testing approach
func TestEntryMenuIntegration_POC(t *testing.T) {
	// Create test habits and entries
	habits := []models.Habit{
		{ID: "habit1", Title: "Exercise", HabitType: models.SimpleHabit},
		{ID: "habit2", Title: "Read", HabitType: models.SimpleHabit},
		{ID: "habit3", Title: "Meditate", HabitType: models.SimpleHabit},
	}

	entries := map[string]models.HabitEntry{
		"habit1": {
			HabitID: "habit1",
			Status:  models.EntryCompleted,
		},
		// habit2 and habit3 have no entries (incomplete)
	}

	// Create the entry menu model
	model := NewEntryMenuModelForTesting(habits, entries)

	// Create teatest model with realistic terminal size
	tm := teatest.NewTestModel(
		t, model,
		teatest.WithInitialTermSize(80, 24),
	)

	// Cleanup
	t.Cleanup(func() {
		if err := tm.Quit(); err != nil {
			t.Fatal(err)
		}
	})

	// Give UI time to initialize
	time.Sleep(100 * time.Millisecond)

	// Test 1: Navigate to first incomplete habit (should be "Read")
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")}) // next incomplete
	time.Sleep(50 * time.Millisecond)

	// Test 2: Try to select the habit (pressing Enter)
	// NOTE: Currently this sets selectedHabitID but doesn't quit (Phase 3.1 not implemented)
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	time.Sleep(50 * time.Millisecond)

	// Test 3: Manually quit to complete the test
	tm.Send(tea.KeyMsg{Type: tea.KeyEsc}) // Should trigger quit
	time.Sleep(50 * time.Millisecond)

	// Test 4: Verify the model's behavior
	finalModel := tm.FinalModel(t, teatest.WithFinalTimeout(time.Second))

	if entryMenuModel, ok := finalModel.(*EntryMenuModel); ok {
		selectedID := entryMenuModel.SelectedHabitID()
		if selectedID != "habit2" {
			t.Errorf("Expected selected habit to be 'habit2', got: %s", selectedID)
		}

		// Verify model quit properly
		if !entryMenuModel.ShouldQuit() {
			t.Error("Expected model to quit after Esc press")
		}
	} else {
		t.Errorf("Expected *EntryMenuModel, got %T", finalModel)
	}

	// Capture and examine final output for debugging
	output := tm.FinalOutput(t, teatest.WithFinalTimeout(time.Second))
	outputBytes := make([]byte, 4096)
	n, _ := output.Read(outputBytes)
	outputStr := string(outputBytes[:n])

	// Basic output validation - should contain menu elements
	if len(outputStr) == 0 {
		t.Error("Expected non-empty output from menu")
	}

	t.Logf("Final output length: %d characters", len(outputStr))
	t.Logf("Output preview: %q", outputStr[:minInt(200, len(outputStr))])
}
