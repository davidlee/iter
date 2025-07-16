package entrymenu

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/key"

	"davidlee/vice/internal/models"
)

func TestNewEntryMenuModelForTesting(t *testing.T) {
	// Create test habits
	habits := []models.Habit{
		{
			ID:          "goal1",
			Title:       "Test Habit 1",
			Description: "Test description",
			HabitType:   models.SimpleHabit,
		},
		{
			ID:          "goal2",
			Title:       "Test Habit 2",
			Description: "Another test",
			HabitType:   models.ElasticHabit,
		},
	}

	// Create test entries
	entries := map[string]models.HabitEntry{
		"goal1": {
			HabitID:   "goal1",
			Value:     true,
			Status:    models.EntryCompleted,
			CreatedAt: time.Now(),
		},
	}

	// Create model
	model := NewEntryMenuModelForTesting(habits, entries)

	// Verify model creation
	if model == nil {
		t.Fatal("Expected model to be created")
	}

	if len(model.habits) != 2 {
		t.Errorf("Expected 2 habits, got %d", len(model.habits))
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
				Habit: models.Habit{
					ID:        "test",
					Title:     "Test Habit",
					HabitType: models.SimpleHabit,
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
	model := NewEntryMenuModelForTesting([]models.Habit{}, map[string]models.HabitEntry{})

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
	model := NewEntryMenuModelForTesting([]models.Habit{}, map[string]models.HabitEntry{})

	// Initial state
	if model.returnBehavior != ReturnToMenu {
		t.Errorf("Expected ReturnToMenu, got %v", model.returnBehavior)
	}

	// Toggle to next habit
	model.toggleReturnBehavior()
	if model.returnBehavior != ReturnToNextHabit {
		t.Errorf("Expected ReturnToNextHabit, got %v", model.returnBehavior)
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

			// Create a habit and entry for testing
			habit := models.Habit{ID: "test", Title: "Test Habit"}
			entries := make(map[string]models.HabitEntry)

			if tt.hasEntry {
				entries["test"] = models.HabitEntry{
					HabitID: "test",
					Status:  tt.entryStatus,
				}
			}

			// Get visible habits to test filter logic
			habits := []models.Habit{habit}
			visibleHabits := helper.GetVisibleHabitsAfterFilter(habits, entries, tt.filterState)

			shouldFilter := len(visibleHabits) == 0
			if shouldFilter != tt.shouldFilter {
				t.Errorf("Expected shouldFilter %v, got %v", tt.shouldFilter, shouldFilter)
			}
		})
	}
}

func TestEntryMenuModel_View(t *testing.T) {
	habits := []models.Habit{
		{
			ID:        "goal1",
			Title:     "Test Habit",
			HabitType: models.SimpleHabit,
		},
	}

	entries := map[string]models.HabitEntry{
		"goal1": {
			HabitID:   "goal1",
			Status:    models.EntryCompleted,
			CreatedAt: time.Now(),
		},
	}

	model := NewEntryMenuModelForTesting(habits, entries)

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
	habits := []models.Habit{
		{ID: "goal1", Title: "Completed Habit", HabitType: models.SimpleHabit},
		{ID: "goal2", Title: "Skipped Habit", HabitType: models.SimpleHabit},
	}

	entries := map[string]models.HabitEntry{
		"goal1": {HabitID: "goal1", Status: models.EntryCompleted, CreatedAt: time.Now()},
		"goal2": {HabitID: "goal2", Status: models.EntrySkipped, CreatedAt: time.Now()},
	}

	model := NewEntryMenuModelForTesting(habits, entries)

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
	habits := []models.Habit{
		{ID: "goal1", Title: "Completed Habit", HabitType: models.SimpleHabit},
		{ID: "goal2", Title: "Incomplete Habit 1", HabitType: models.SimpleHabit},
		{ID: "goal3", Title: "Skipped Habit", HabitType: models.SimpleHabit},
		{ID: "goal4", Title: "Incomplete Habit 2", HabitType: models.SimpleHabit},
	}

	entries := map[string]models.HabitEntry{
		"goal1": {HabitID: "goal1", Status: models.EntryCompleted, CreatedAt: time.Now()},
		"goal3": {HabitID: "goal3", Status: models.EntrySkipped, CreatedAt: time.Now()},
	}

	model := NewEntryMenuModelForTesting(habits, entries)
	model.width = 80
	model.height = 24

	// Test navigation enhancer initialization
	if model.navEnhancer == nil {
		t.Error("Expected navEnhancer to be initialized")
	}

	// Test GetCurrentHabitInfo
	goalInfo := model.GetCurrentHabitInfo()
	if goalInfo == nil {
		t.Error("Expected habit info to be available")
	}

	// Test SelectFirstIncompleteHabit
	model.SelectFirstIncompleteHabit()
	selectedItem := model.list.SelectedItem()
	if menuItem, ok := selectedItem.(EntryMenuItem); ok {
		if menuItem.HasEntry {
			t.Error("Expected first incomplete habit to be selected")
		}
	}
}

func TestEntryMenuModel_ClearFilters(t *testing.T) {
	model := NewEntryMenuModelForTesting([]models.Habit{}, map[string]models.HabitEntry{})

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
