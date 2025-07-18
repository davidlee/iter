package entrymenu

import (
	"testing"
	"time"

	"github.com/davidlee/vice/internal/models"
)

func TestNavigationHelper_FindNextIncompleteHabit(t *testing.T) {
	helper := NewNavigationHelper()

	habits := []models.Habit{
		{ID: "habit1", Title: "Completed Habit"},
		{ID: "habit2", Title: "Incomplete Habit 1"},
		{ID: "habit3", Title: "Skipped Habit"},
		{ID: "habit4", Title: "Incomplete Habit 2"},
	}

	entries := map[string]models.HabitEntry{
		"habit1": {HabitID: "habit1", Status: models.EntryCompleted, CreatedAt: time.Now()},
		"habit3": {HabitID: "habit3", Status: models.EntrySkipped, CreatedAt: time.Now()},
	}

	tests := []struct {
		name         string
		currentIndex int
		expectedNext int
	}{
		{
			name:         "from start, find first incomplete",
			currentIndex: 0,
			expectedNext: 1, // habit2 is incomplete
		},
		{
			name:         "from incomplete habit, find next incomplete",
			currentIndex: 1,
			expectedNext: 3, // habit4 is incomplete
		},
		{
			name:         "from last habit, wrap to first incomplete",
			currentIndex: 3,
			expectedNext: 1, // wrap around to habit2
		},
		{
			name:         "from skipped habit, find next incomplete",
			currentIndex: 2,
			expectedNext: 3, // habit4 is incomplete
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := helper.FindNextIncompleteHabit(habits, entries, tt.currentIndex)
			if result != tt.expectedNext {
				t.Errorf("Expected next incomplete habit at index %d, got %d", tt.expectedNext, result)
			}
		})
	}
}

func TestNavigationHelper_FindPreviousIncompleteHabit(t *testing.T) {
	helper := NewNavigationHelper()

	habits := []models.Habit{
		{ID: "habit1", Title: "Incomplete Habit 1"},
		{ID: "habit2", Title: "Completed Habit"},
		{ID: "habit3", Title: "Skipped Habit"},
		{ID: "habit4", Title: "Incomplete Habit 2"},
	}

	entries := map[string]models.HabitEntry{
		"habit2": {HabitID: "habit2", Status: models.EntryCompleted, CreatedAt: time.Now()},
		"habit3": {HabitID: "habit3", Status: models.EntrySkipped, CreatedAt: time.Now()},
	}

	tests := []struct {
		name         string
		currentIndex int
		expectedPrev int
	}{
		{
			name:         "from end, find previous incomplete",
			currentIndex: 3,
			expectedPrev: 0, // habit1 is incomplete
		},
		{
			name:         "from incomplete habit, find previous incomplete",
			currentIndex: 0,
			expectedPrev: 3, // wrap around to habit4
		},
		{
			name:         "from middle, find previous incomplete",
			currentIndex: 2,
			expectedPrev: 0, // habit1 is incomplete
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := helper.FindPreviousIncompleteHabit(habits, entries, tt.currentIndex)
			if result != tt.expectedPrev {
				t.Errorf("Expected previous incomplete habit at index %d, got %d", tt.expectedPrev, result)
			}
		})
	}
}

func TestNavigationHelper_GetVisibleHabitsAfterFilter(t *testing.T) {
	helper := NewNavigationHelper()

	habits := []models.Habit{
		{ID: "habit1", Title: "Completed Habit"},
		{ID: "habit2", Title: "Failed Habit"},
		{ID: "habit3", Title: "Skipped Habit"},
		{ID: "habit4", Title: "Incomplete Habit"},
	}

	entries := map[string]models.HabitEntry{
		"habit1": {HabitID: "habit1", Status: models.EntryCompleted, CreatedAt: time.Now()},
		"habit2": {HabitID: "habit2", Status: models.EntryFailed, CreatedAt: time.Now()},
		"habit3": {HabitID: "habit3", Status: models.EntrySkipped, CreatedAt: time.Now()},
	}

	tests := []struct {
		name        string
		filterState FilterState
		expectedIDs []string
	}{
		{
			name:        "no filter",
			filterState: FilterNone,
			expectedIDs: []string{"habit1", "habit2", "habit3", "habit4"},
		},
		{
			name:        "hide skipped",
			filterState: FilterHideSkipped,
			expectedIDs: []string{"habit1", "habit2", "habit4"},
		},
		{
			name:        "hide previous",
			filterState: FilterHidePrevious,
			expectedIDs: []string{"habit3", "habit4"},
		},
		{
			name:        "hide skipped and previous",
			filterState: FilterHideSkippedAndPrevious,
			expectedIDs: []string{"habit4"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := helper.GetVisibleHabitsAfterFilter(habits, entries, tt.filterState)

			if len(result) != len(tt.expectedIDs) {
				t.Errorf("Expected %d visible habits, got %d", len(tt.expectedIDs), len(result))
				return
			}

			for i, habit := range result {
				if habit.ID != tt.expectedIDs[i] {
					t.Errorf("Expected habit ID %s at index %d, got %s", tt.expectedIDs[i], i, habit.ID)
				}
			}
		})
	}
}

func TestNavigationHelper_ShouldAutoSelectNext(t *testing.T) {
	helper := NewNavigationHelper()

	tests := []struct {
		name               string
		returnBehavior     ReturnBehavior
		justCompletedEntry bool
		expected           bool
	}{
		{
			name:               "return to menu, just completed",
			returnBehavior:     ReturnToMenu,
			justCompletedEntry: true,
			expected:           false,
		},
		{
			name:               "return to next habit, just completed",
			returnBehavior:     ReturnToNextHabit,
			justCompletedEntry: true,
			expected:           true,
		},
		{
			name:               "return to next habit, not completed",
			returnBehavior:     ReturnToNextHabit,
			justCompletedEntry: false,
			expected:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := helper.ShouldAutoSelectNext(tt.returnBehavior, tt.justCompletedEntry)
			if result != tt.expected {
				t.Errorf("Expected %t, got %t", tt.expected, result)
			}
		})
	}
}

func TestNavigationHelper_GetFilterDescription(t *testing.T) {
	helper := NewNavigationHelper()

	tests := []struct {
		filterState FilterState
		expected    string
	}{
		{FilterNone, "showing all habits"},
		{FilterHideSkipped, "hiding skipped habits"},
		{FilterHidePrevious, "hiding completed/failed habits"},
		{FilterHideSkippedAndPrevious, "hiding skipped and completed/failed habits"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := helper.GetFilterDescription(tt.filterState)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestNavigationEnhancer_SelectNextIncompleteHabit(t *testing.T) {
	habits := []models.Habit{
		{ID: "habit1", Title: "Completed Habit"},
		{ID: "habit2", Title: "Incomplete Habit"},
		{ID: "habit3", Title: "Another Incomplete Habit"},
	}

	entries := map[string]models.HabitEntry{
		"habit1": {HabitID: "habit1", Status: models.EntryCompleted, CreatedAt: time.Now()},
	}

	model := NewEntryMenuModelForTesting(habits, entries)
	model.width = 80
	model.height = 24

	// Start at habit1 (completed)
	model.list.Select(0)

	// Select next incomplete habit
	model.navEnhancer.SelectNextIncompleteHabit(model)

	// Should now be at habit2 (incomplete)
	if model.list.Index() != 1 {
		t.Errorf("Expected selection at index 1, got %d", model.list.Index())
	}
}

func TestNavigationEnhancer_UpdateListAfterFilterChange(t *testing.T) {
	habits := []models.Habit{
		{ID: "habit1", Title: "Completed Habit"},
		{ID: "habit2", Title: "Skipped Habit"},
		{ID: "habit3", Title: "Incomplete Habit"},
	}

	entries := map[string]models.HabitEntry{
		"habit1": {HabitID: "habit1", Status: models.EntryCompleted, CreatedAt: time.Now()},
		"habit2": {HabitID: "habit2", Status: models.EntrySkipped, CreatedAt: time.Now()},
	}

	model := NewEntryMenuModelForTesting(habits, entries)
	model.width = 80
	model.height = 24

	// Apply filter to hide skipped habits
	model.filterState = FilterHideSkipped
	model.navEnhancer.UpdateListAfterFilterChange(model)

	// Should have 2 items (completed and incomplete)
	items := model.list.Items()
	if len(items) != 2 {
		t.Errorf("Expected 2 items after filter, got %d", len(items))
	}

	// Should auto-select the incomplete habit (habit3)
	selectedItem := model.list.SelectedItem()
	if menuItem, ok := selectedItem.(EntryMenuItem); ok {
		if menuItem.Habit.ID != "habit3" {
			t.Errorf("Expected habit3 to be selected, got %s", menuItem.Habit.ID)
		}
	} else {
		t.Error("Expected EntryMenuItem to be selected")
	}
}

func TestHabitInfo_StatusMethods(t *testing.T) {
	tests := []struct {
		name         string
		hasEntry     bool
		status       models.EntryStatus
		isComplete   bool
		isIncomplete bool
		isSkipped    bool
		isFailed     bool
	}{
		{
			name:         "no entry",
			hasEntry:     false,
			isComplete:   false,
			isIncomplete: true,
			isSkipped:    false,
			isFailed:     false,
		},
		{
			name:         "completed entry",
			hasEntry:     true,
			status:       models.EntryCompleted,
			isComplete:   true,
			isIncomplete: false,
			isSkipped:    false,
			isFailed:     false,
		},
		{
			name:         "skipped entry",
			hasEntry:     true,
			status:       models.EntrySkipped,
			isComplete:   false,
			isIncomplete: false,
			isSkipped:    true,
			isFailed:     false,
		},
		{
			name:         "failed entry",
			hasEntry:     true,
			status:       models.EntryFailed,
			isComplete:   false,
			isIncomplete: false,
			isSkipped:    false,
			isFailed:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			habitInfo := &HabitInfo{
				HasEntry: tt.hasEntry,
			}

			if tt.hasEntry {
				habitInfo.Entry = models.HabitEntry{Status: tt.status}
			}

			if habitInfo.IsComplete() != tt.isComplete {
				t.Errorf("IsComplete() = %t, expected %t", habitInfo.IsComplete(), tt.isComplete)
			}
			if habitInfo.IsIncomplete() != tt.isIncomplete {
				t.Errorf("IsIncomplete() = %t, expected %t", habitInfo.IsIncomplete(), tt.isIncomplete)
			}
			if habitInfo.IsSkipped() != tt.isSkipped {
				t.Errorf("IsSkipped() = %t, expected %t", habitInfo.IsSkipped(), tt.isSkipped)
			}
			if habitInfo.IsFailed() != tt.isFailed {
				t.Errorf("IsFailed() = %t, expected %t", habitInfo.IsFailed(), tt.isFailed)
			}
		})
	}
}

func TestEntryMenuKeyMap_HelpMethods(t *testing.T) {
	keyMap := DefaultEntryMenuKeyMap()

	// Test short help
	shortHelp := keyMap.GetShortHelp()
	if len(shortHelp) != 8 { // up, down, select, next incomplete, return behavior, filter skipped, filter previous, quit
		t.Errorf("Expected 8 short help bindings, got %d", len(shortHelp))
	}

	// Test full help
	fullHelp := keyMap.GetFullHelp()
	if len(fullHelp) != 3 {
		t.Errorf("Expected 3 groups in full help, got %d", len(fullHelp))
	}

	// Check navigation group has 3 bindings (up, down, select)
	if len(fullHelp[0]) != 3 {
		t.Errorf("Expected 3 navigation bindings, got %d", len(fullHelp[0]))
	}

	// Check menu controls group has 4 bindings (return, filter skipped, filter previous, clear filters)
	if len(fullHelp[1]) != 4 {
		t.Errorf("Expected 4 menu control bindings, got %d", len(fullHelp[1]))
	}

	// Check exit group has 1 binding
	if len(fullHelp[2]) != 1 {
		t.Errorf("Expected 1 exit binding, got %d", len(fullHelp[2]))
	}
}
