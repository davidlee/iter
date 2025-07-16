package entrymenu

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/key"

	"davidlee/vice/internal/models"
)

func TestNewEntryMenuModelForTesting(t *testing.T) {
	// Create test goals
	goals := []models.Goal{
		{
			ID:          "goal1",
			Title:       "Test Goal 1",
			Description: "Test description",
			GoalType:    models.SimpleGoal,
		},
		{
			ID:          "goal2",
			Title:       "Test Goal 2",
			Description: "Another test",
			GoalType:    models.ElasticGoal,
		},
	}

	// Create test entries
	entries := map[string]models.GoalEntry{
		"goal1": {
			GoalID:    "goal1",
			Value:     true,
			Status:    models.EntryCompleted,
			CreatedAt: time.Now(),
		},
	}

	// Create model
	model := NewEntryMenuModelForTesting(goals, entries)

	// Verify model creation
	if model == nil {
		t.Fatal("Expected model to be created")
	}

	if len(model.goals) != 2 {
		t.Errorf("Expected 2 goals, got %d", len(model.goals))
	}

	if len(model.entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(model.entries))
	}

	if model.filterState != FilterNone {
		t.Errorf("Expected FilterNone, got %v", model.filterState)
	}

	if model.returnBehavior != ReturnToMenu {
		t.Errorf("Expected ReturnToMenu, got %v", model.returnBehavior)
	}

	if model.viewRenderer == nil {
		t.Error("Expected viewRenderer to be initialized")
	}
}

func TestEntryMenuItemStatusColors(t *testing.T) {
	tests := []struct {
		name          string
		hasEntry      bool
		entryStatus   models.EntryStatus
		expectedColor string
	}{
		{
			name:          "no entry",
			hasEntry:      false,
			expectedColor: "250", // light grey
		},
		{
			name:          "completed",
			hasEntry:      true,
			entryStatus:   models.EntryCompleted,
			expectedColor: "214", // gold
		},
		{
			name:          "failed",
			hasEntry:      true,
			entryStatus:   models.EntryFailed,
			expectedColor: "88", // dark red
		},
		{
			name:          "skipped",
			hasEntry:      true,
			entryStatus:   models.EntrySkipped,
			expectedColor: "240", // dark grey
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := EntryMenuItem{
				Goal: models.Goal{
					ID:       "test",
					Title:    "Test Goal",
					GoalType: models.SimpleGoal,
				},
				HasEntry:    tt.hasEntry,
				EntryStatus: tt.entryStatus,
			}

			color := item.getStatusColor()
			if string(color) != tt.expectedColor {
				t.Errorf("Expected color %s, got %s", tt.expectedColor, color)
			}
		})
	}
}

func TestFilterToggling(t *testing.T) {
	model := NewEntryMenuModelForTesting([]models.Goal{}, map[string]models.GoalEntry{})

	// Test skipped filter toggling
	model.toggleSkippedFilter()
	if model.filterState != FilterHideSkipped {
		t.Errorf("Expected FilterHideSkipped, got %v", model.filterState)
	}

	model.toggleSkippedFilter()
	if model.filterState != FilterNone {
		t.Errorf("Expected FilterNone, got %v", model.filterState)
	}

	// Test previous filter toggling
	model.togglePreviousFilter()
	if model.filterState != FilterHidePrevious {
		t.Errorf("Expected FilterHidePrevious, got %v", model.filterState)
	}

	// Test combined filtering
	model.toggleSkippedFilter()
	if model.filterState != FilterHideSkippedAndPrevious {
		t.Errorf("Expected FilterHideSkippedAndPrevious, got %v", model.filterState)
	}
}

func TestReturnBehaviorToggling(t *testing.T) {
	model := NewEntryMenuModelForTesting([]models.Goal{}, map[string]models.GoalEntry{})

	// Initial state
	if model.returnBehavior != ReturnToMenu {
		t.Errorf("Expected ReturnToMenu, got %v", model.returnBehavior)
	}

	// Toggle to next goal
	model.toggleReturnBehavior()
	if model.returnBehavior != ReturnToNextGoal {
		t.Errorf("Expected ReturnToNextGoal, got %v", model.returnBehavior)
	}

	// Toggle back to menu
	model.toggleReturnBehavior()
	if model.returnBehavior != ReturnToMenu {
		t.Errorf("Expected ReturnToMenu, got %v", model.returnBehavior)
	}
}

func TestShouldFilterOut(t *testing.T) {
	tests := []struct {
		name         string
		filterState  FilterState
		hasEntry     bool
		entryStatus  models.EntryStatus
		shouldFilter bool
	}{
		{
			name:         "no filter, no entry",
			filterState:  FilterNone,
			hasEntry:     false,
			shouldFilter: false,
		},
		{
			name:         "hide skipped, has skipped entry",
			filterState:  FilterHideSkipped,
			hasEntry:     true,
			entryStatus:  models.EntrySkipped,
			shouldFilter: true,
		},
		{
			name:         "hide skipped, has completed entry",
			filterState:  FilterHideSkipped,
			hasEntry:     true,
			entryStatus:  models.EntryCompleted,
			shouldFilter: false,
		},
		{
			name:         "hide previous, has completed entry",
			filterState:  FilterHidePrevious,
			hasEntry:     true,
			entryStatus:  models.EntryCompleted,
			shouldFilter: true,
		},
		{
			name:         "hide previous, has failed entry",
			filterState:  FilterHidePrevious,
			hasEntry:     true,
			entryStatus:  models.EntryFailed,
			shouldFilter: true,
		},
		{
			name:         "hide previous, has skipped entry",
			filterState:  FilterHidePrevious,
			hasEntry:     true,
			entryStatus:  models.EntrySkipped,
			shouldFilter: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a navigation helper to test filtering logic
			helper := NewNavigationHelper()

			// Create a goal and entry for testing
			goal := models.Goal{ID: "test", Title: "Test Goal"}
			entries := make(map[string]models.GoalEntry)

			if tt.hasEntry {
				entries["test"] = models.GoalEntry{
					GoalID: "test",
					Status: tt.entryStatus,
				}
			}

			// Get visible goals to test filter logic
			goals := []models.Goal{goal}
			visibleGoals := helper.GetVisibleGoalsAfterFilter(goals, entries, tt.filterState)

			shouldFilter := len(visibleGoals) == 0
			if shouldFilter != tt.shouldFilter {
				t.Errorf("Expected shouldFilter %v, got %v", tt.shouldFilter, shouldFilter)
			}
		})
	}
}

func TestEntryMenuModel_View(t *testing.T) {
	goals := []models.Goal{
		{
			ID:       "goal1",
			Title:    "Test Goal",
			GoalType: models.SimpleGoal,
		},
	}

	entries := map[string]models.GoalEntry{
		"goal1": {
			GoalID:    "goal1",
			Status:    models.EntryCompleted,
			CreatedAt: time.Now(),
		},
	}

	model := NewEntryMenuModelForTesting(goals, entries)

	// Set dimensions for proper rendering
	model.width = 80
	model.height = 24

	// Test view rendering
	view := model.View()

	// Should contain progress information
	if !strings.Contains(view, "1/1 completed") {
		t.Errorf("Expected view to contain progress info, got: %s", view)
	}

	// Should contain return behavior in footer
	if !strings.Contains(view, "Return:") {
		t.Errorf("Expected view to contain return behavior in footer, got: %s", view)
	}
}

func TestEntryMenuModel_ViewWithFilters(t *testing.T) {
	goals := []models.Goal{
		{ID: "goal1", Title: "Completed Goal", GoalType: models.SimpleGoal},
		{ID: "goal2", Title: "Skipped Goal", GoalType: models.SimpleGoal},
	}

	entries := map[string]models.GoalEntry{
		"goal1": {GoalID: "goal1", Status: models.EntryCompleted, CreatedAt: time.Now()},
		"goal2": {GoalID: "goal2", Status: models.EntrySkipped, CreatedAt: time.Now()},
	}

	model := NewEntryMenuModelForTesting(goals, entries)

	// Set dimensions for proper rendering
	model.width = 80
	model.height = 24

	// Enable skipped filter
	model.toggleSkippedFilter()

	view := model.View()

	// Should show filter information
	if !strings.Contains(view, "hiding skipped") {
		t.Errorf("Expected view to show filter info, got: %s", view)
	}
}

func TestEntryMenuModel_NavigationEnhancements(t *testing.T) {
	goals := []models.Goal{
		{ID: "goal1", Title: "Completed Goal", GoalType: models.SimpleGoal},
		{ID: "goal2", Title: "Incomplete Goal 1", GoalType: models.SimpleGoal},
		{ID: "goal3", Title: "Skipped Goal", GoalType: models.SimpleGoal},
		{ID: "goal4", Title: "Incomplete Goal 2", GoalType: models.SimpleGoal},
	}

	entries := map[string]models.GoalEntry{
		"goal1": {GoalID: "goal1", Status: models.EntryCompleted, CreatedAt: time.Now()},
		"goal3": {GoalID: "goal3", Status: models.EntrySkipped, CreatedAt: time.Now()},
	}

	model := NewEntryMenuModelForTesting(goals, entries)
	model.width = 80
	model.height = 24

	// Test navigation enhancer initialization
	if model.navEnhancer == nil {
		t.Error("Expected navEnhancer to be initialized")
	}

	// Test GetCurrentGoalInfo
	goalInfo := model.GetCurrentGoalInfo()
	if goalInfo == nil {
		t.Error("Expected goal info to be available")
	}

	// Test SelectFirstIncompleteGoal
	model.SelectFirstIncompleteGoal()
	selectedItem := model.list.SelectedItem()
	if menuItem, ok := selectedItem.(EntryMenuItem); ok {
		if menuItem.HasEntry {
			t.Error("Expected first incomplete goal to be selected")
		}
	}
}

func TestEntryMenuModel_ClearFilters(t *testing.T) {
	model := NewEntryMenuModelForTesting([]models.Goal{}, map[string]models.GoalEntry{})

	// Set some filters
	model.filterState = FilterHideSkippedAndPrevious

	// Clear filters
	model.clearAllFilters()

	if model.filterState != FilterNone {
		t.Errorf("Expected FilterNone after clearing, got %v", model.filterState)
	}
}

func TestEntryMenuModel_EnhancedKeybindings(t *testing.T) {
	keyMap := DefaultEntryMenuKeyMap()

	// Test that all enhanced keybindings are defined
	bindings := []struct {
		name    string
		binding key.Binding
	}{
		{"NextIncomplete", keyMap.NextIncomplete},
		{"PreviousIncomplete", keyMap.PreviousIncomplete},
		{"ClearFilters", keyMap.ClearFilters},
	}

	for _, b := range bindings {
		if len(b.binding.Keys()) == 0 {
			t.Errorf("Expected %s binding to have keys defined", b.name)
		}
		if b.binding.Help().Key == "" {
			t.Errorf("Expected %s binding to have help text", b.name)
		}
	}
}
