package entrymenu

import (
	"strings"
	"testing"
	"time"

	"github.com/davidlee/vice/internal/models"
)

func TestViewRenderer_RenderProgressBar(t *testing.T) {
	renderer := NewViewRenderer(80, 24)

	tests := []struct {
		name     string
		habits   []models.Habit
		entries  map[string]models.HabitEntry
		contains []string
	}{
		{
			name:     "empty habits",
			habits:   []models.Habit{},
			entries:  map[string]models.HabitEntry{},
			contains: []string{"No habits configured"},
		},
		{
			name: "no entries",
			habits: []models.Habit{
				{ID: "habit1", Title: "Test Habit 1"},
				{ID: "habit2", Title: "Test Habit 2"},
			},
			entries:  map[string]models.HabitEntry{},
			contains: []string{"0/2 completed", "0.0%", "2 remaining"},
		},
		{
			name: "mixed completion status",
			habits: []models.Habit{
				{ID: "habit1", Title: "Completed Habit"},
				{ID: "habit2", Title: "Failed Habit"},
				{ID: "habit3", Title: "Skipped Habit"},
				{ID: "habit4", Title: "Incomplete Habit"},
			},
			entries: map[string]models.HabitEntry{
				"habit1": {
					HabitID:   "habit1",
					Status:    models.EntryCompleted,
					CreatedAt: time.Now(),
				},
				"habit2": {
					HabitID:   "habit2",
					Status:    models.EntryFailed,
					CreatedAt: time.Now(),
				},
				"habit3": {
					HabitID:   "habit3",
					Status:    models.EntrySkipped,
					CreatedAt: time.Now(),
				},
			},
			contains: []string{"1/4 completed", "25.0%", "1 failed", "1 skipped", "1 remaining"},
		},
		{
			name: "all completed",
			habits: []models.Habit{
				{ID: "habit1", Title: "Habit 1"},
				{ID: "habit2", Title: "Habit 2"},
			},
			entries: map[string]models.HabitEntry{
				"habit1": {
					HabitID:   "habit1",
					Status:    models.EntryCompleted,
					CreatedAt: time.Now(),
				},
				"habit2": {
					HabitID:   "habit2",
					Status:    models.EntryCompleted,
					CreatedAt: time.Now(),
				},
			},
			contains: []string{"2/2 completed", "100.0%", "0 failed", "0 skipped", "0 remaining"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderer.renderProgress(tt.habits, tt.entries)

			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected result to contain %q, got: %s", expected, result)
				}
			}
		})
	}
}

func TestViewRenderer_RenderFilters(t *testing.T) {
	renderer := NewViewRenderer(80, 24)

	tests := []struct {
		name        string
		filterState FilterState
		expected    string
	}{
		{
			name:        "no filter",
			filterState: FilterNone,
			expected:    "",
		},
		{
			name:        "hide skipped",
			filterState: FilterHideSkipped,
			expected:    "hiding skipped",
		},
		{
			name:        "hide previous",
			filterState: FilterHidePrevious,
			expected:    "hiding previous",
		},
		{
			name:        "hide both",
			filterState: FilterHideSkippedAndPrevious,
			expected:    "hiding skipped, hiding previous",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderer.RenderFilters(tt.filterState)

			if tt.expected == "" {
				if result != "" {
					t.Errorf("Expected empty result, got: %s", result)
				}
			} else {
				if !strings.Contains(result, tt.expected) {
					t.Errorf("Expected result to contain %q, got: %s", tt.expected, result)
				}
			}
		})
	}
}

func TestViewRenderer_RenderReturnBehavior(t *testing.T) {
	renderer := NewViewRenderer(80, 24)

	tests := []struct {
		name     string
		behavior ReturnBehavior
		expected string
	}{
		{
			name:     "return to menu",
			behavior: ReturnToMenu,
			expected: "Return: menu",
		},
		{
			name:     "return to next habit",
			behavior: ReturnToNextHabit,
			expected: "Return: next habit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderer.RenderReturnBehavior(tt.behavior)

			if !strings.Contains(result, tt.expected) {
				t.Errorf("Expected result to contain %q, got: %s", tt.expected, result)
			}
		})
	}
}

func TestViewRenderer_CalculateProgressStats(t *testing.T) {
	renderer := NewViewRenderer(80, 24)

	habits := []models.Habit{
		{ID: "habit1", Title: "Completed Habit"},
		{ID: "habit2", Title: "Failed Habit"},
		{ID: "habit3", Title: "Skipped Habit"},
		{ID: "habit4", Title: "Incomplete Habit"},
	}

	entries := map[string]models.HabitEntry{
		"habit1": {
			HabitID:   "habit1",
			Status:    models.EntryCompleted,
			CreatedAt: time.Now(),
		},
		"habit2": {
			HabitID:   "habit2",
			Status:    models.EntryFailed,
			CreatedAt: time.Now(),
		},
		"habit3": {
			HabitID:   "habit3",
			Status:    models.EntrySkipped,
			CreatedAt: time.Now(),
		},
	}

	stats := renderer.calculateProgressStats(habits, entries)

	expected := ProgressStats{
		Total:     4,
		Completed: 1,
		Failed:    1,
		Skipped:   1,
		Attempted: 3,
		Remaining: 1,
	}

	if stats != expected {
		t.Errorf("Expected stats %+v, got %+v", expected, stats)
	}
}

func TestViewRenderer_RenderProgressBarVisual(t *testing.T) {
	renderer := NewViewRenderer(80, 24)

	tests := []struct {
		name          string
		completedPct  float64
		total         int
		shouldContain []string
	}{
		{
			name:          "zero percent",
			completedPct:  0.0,
			total:         4,
			shouldContain: []string{"[", "]", "0.0%"},
		},
		{
			name:          "fifty percent",
			completedPct:  50.0,
			total:         4,
			shouldContain: []string{"[", "]", "50.0%"},
		},
		{
			name:          "hundred percent",
			completedPct:  100.0,
			total:         4,
			shouldContain: []string{"[", "]", "100.0%"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderer.renderProgressBarVisual(tt.completedPct)

			for _, expected := range tt.shouldContain {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected result to contain %q, got: %s", expected, result)
				}
			}
		})
	}
}

func TestViewRenderer_RenderHeader(t *testing.T) {
	renderer := NewViewRenderer(80, 24)

	habits := []models.Habit{
		{ID: "habit1", Title: "Test Habit"},
	}

	entries := map[string]models.HabitEntry{
		"habit1": {
			HabitID:   "habit1",
			Status:    models.EntryCompleted,
			CreatedAt: time.Now(),
		},
	}

	result := renderer.RenderHeader(habits, entries, FilterHideSkipped)

	// Should contain progress information
	if !strings.Contains(result, "1/1 completed") {
		t.Errorf("Expected header to contain progress info, got: %s", result)
	}

	// Should contain filter information
	if !strings.Contains(result, "hiding skipped") {
		t.Errorf("Expected header to contain filter info, got: %s", result)
	}

	// Header should no longer contain return behavior (moved to footer)
	if strings.Contains(result, "Return:") {
		t.Errorf("Expected header to NOT contain return behavior (moved to footer), got: %s", result)
	}
}

func TestViewRenderer_ZeroWidth(t *testing.T) {
	renderer := NewViewRenderer(0, 24)

	result := renderer.renderProgressBarVisual(50.0)

	// Should handle zero width gracefully
	if result != "" {
		t.Errorf("Expected empty result for zero width, got: %s", result)
	}
}

func TestProgressStats_EdgeCases(t *testing.T) {
	renderer := NewViewRenderer(80, 24)

	tests := []struct {
		name    string
		habits  []models.Habit
		entries map[string]models.HabitEntry
		check   func(ProgressStats) bool
	}{
		{
			name:    "no habits",
			habits:  []models.Habit{},
			entries: map[string]models.HabitEntry{},
			check: func(stats ProgressStats) bool {
				return stats.Total == 0 && stats.Remaining == 0
			},
		},
		{
			name: "no entries",
			habits: []models.Habit{
				{ID: "habit1", Title: "Habit 1"},
			},
			entries: map[string]models.HabitEntry{},
			check: func(stats ProgressStats) bool {
				return stats.Total == 1 && stats.Attempted == 0 && stats.Remaining == 1
			},
		},
		{
			name: "more entries than habits",
			habits: []models.Habit{
				{ID: "habit1", Title: "Habit 1"},
			},
			entries: map[string]models.HabitEntry{
				"habit1": {HabitID: "habit1", Status: models.EntryCompleted, CreatedAt: time.Now()},
				"habit2": {HabitID: "habit2", Status: models.EntryCompleted, CreatedAt: time.Now()}, // Extra entry
			},
			check: func(stats ProgressStats) bool {
				return stats.Total == 1 && stats.Completed == 1 && stats.Attempted == 1
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := renderer.calculateProgressStats(tt.habits, tt.entries)

			if !tt.check(stats) {
				t.Errorf("Stats check failed for %s: %+v", tt.name, stats)
			}
		})
	}
}
