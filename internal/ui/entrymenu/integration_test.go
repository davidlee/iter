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
// Tests: menu navigation → goal selection → mock entry flow
// Purpose: Evaluate teatest framework utility vs current unit testing approach
func TestEntryMenuIntegration_POC(t *testing.T) {
	// Create test goals and entries
	goals := []models.Goal{
		{ID: "goal1", Title: "Exercise", GoalType: models.SimpleGoal},
		{ID: "goal2", Title: "Read", GoalType: models.SimpleGoal},
		{ID: "goal3", Title: "Meditate", GoalType: models.SimpleGoal},
	}

	entries := map[string]models.GoalEntry{
		"goal1": {
			GoalID: "goal1",
			Status: models.EntryCompleted,
		},
		// goal2 and goal3 have no entries (incomplete)
	}

	// Create the entry menu model
	model := NewEntryMenuModelForTesting(goals, entries)

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

	// Test 1: Navigate to first incomplete goal (should be "Read")
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")}) // next incomplete
	time.Sleep(50 * time.Millisecond)

	// Test 2: Try to select the goal (pressing Enter)
	// NOTE: Currently this sets selectedGoalID but doesn't quit (Phase 3.1 not implemented)
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	time.Sleep(50 * time.Millisecond)

	// Test 3: Manually quit to complete the test
	tm.Send(tea.KeyMsg{Type: tea.KeyEsc}) // Should trigger quit
	time.Sleep(50 * time.Millisecond)

	// Test 4: Verify the model's behavior
	finalModel := tm.FinalModel(t, teatest.WithFinalTimeout(time.Second))
	
	if entryMenuModel, ok := finalModel.(*EntryMenuModel); ok {
		selectedID := entryMenuModel.SelectedGoalID()
		if selectedID != "goal2" {
			t.Errorf("Expected selected goal to be 'goal2', got: %s", selectedID)
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
	t.Logf("Output preview: %q", outputStr[:min(200, len(outputStr))])
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}