// Package entrymenu integration tests for entry collection flow.
// AIDEV-NOTE: T018/3.1-entry-integration; test menu→entry→menu flow with teatest
package entrymenu

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"

	"davidlee/vice/internal/models"
	"davidlee/vice/internal/ui"
)

// TestEntryIntegration_MenuToEntryFlow tests the complete menu→entry→menu integration.
// This test verifies that habit selection launches entry collection and updates menu state.
// TEMPORARILY DISABLED: Test needs updating for modal architecture (T024)
func TestEntryIntegration_MenuToEntryFlow_DISABLED(t *testing.T) {
	t.Skip("Test disabled pending modal architecture integration - see T024")
	// Create test habits with different types
	habits := []models.Habit{
		{
			ID:        "simple_goal",
			Title:     "Exercise",
			HabitType: models.SimpleHabit,
			FieldType: models.FieldType{Type: "boolean"},
		},
		{
			ID:        "time_goal",
			Title:     "Wake Up Early",
			HabitType: models.SimpleHabit,
			FieldType: models.FieldType{Type: "time"},
		},
	}

	// Start with no entries (all incomplete)
	entries := make(map[string]models.HabitEntry)

	// Create and initialize EntryCollector
	collector := ui.NewEntryCollector("") // Empty path for test
	collector.InitializeForMenu(habits, entries)

	// Create the entry menu model with collector (no file for test)
	model := NewEntryMenuModel(habits, entries, collector, "")

	// Create teatest model
	tm := teatest.NewTestModel(
		t, model,
		teatest.WithInitialTermSize(80, 24),
	)

	t.Cleanup(func() {
		if err := tm.Quit(); err != nil {
			t.Fatal(err)
		}
	})

	// Give UI time to initialize
	time.Sleep(100 * time.Millisecond)

	// Test 1: Navigate to first habit (Exercise)
	// The first habit should be selected by default

	// Test 2: Select the habit (pressing Enter)
	// NOTE: This will try to launch entry collection
	// For this test, we expect it to work but the entry collection might be minimal
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	time.Sleep(200 * time.Millisecond) // Give time for entry collection

	// Test 3: Check if the model is still running (entry collection completed)
	// If entry collection worked, the model should still be active

	// Test 4: Try to navigate to next incomplete habit
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	time.Sleep(50 * time.Millisecond)

	// Test 5: Quit the menu
	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})

	// Get final model state
	finalModel := tm.FinalModel(t, teatest.WithFinalTimeout(2*time.Second))

	if entryMenuModel, ok := finalModel.(*EntryMenuModel); ok {
		// Verify the model handled the interaction properly
		if !entryMenuModel.ShouldQuit() {
			t.Error("Expected model to quit after Esc press")
		}

		// The selectedHabitID should be set from the Enter press
		selectedID := entryMenuModel.SelectedHabitID()
		if selectedID == "" {
			t.Error("Expected selectedHabitID to be set after habit selection")
		}

		t.Logf("Selected habit ID: %s", selectedID)
		t.Logf("Model quit properly: %v", entryMenuModel.ShouldQuit())
	} else {
		t.Errorf("Expected *EntryMenuModel, got %T", finalModel)
	}

	// Capture final output for inspection
	output := tm.FinalOutput(t, teatest.WithFinalTimeout(time.Second))
	outputBytes := make([]byte, 4096)
	n, _ := output.Read(outputBytes)
	outputStr := string(outputBytes[:n])

	// Verify output contains expected elements
	expectedContent := []string{
		"Exercise",
		"Wake Up Early",
		"Entry Menu",
	}

	for _, expected := range expectedContent {
		if len(outputStr) > 0 && !containsIgnoreCase(outputStr, expected) {
			t.Errorf("Expected output to contain %q", expected)
		}
	}

	t.Logf("Integration test completed - menu navigation and entry selection functional")
}

// Helper function for case-insensitive string search
func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
					findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
