package ui

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"davidlee/vice/internal/models"
	"davidlee/vice/internal/parser"
	"davidlee/vice/internal/storage"
)

func TestNewEntryCollector(t *testing.T) {
	collector := NewEntryCollector("checklists.yml")

	// Test basic initialization
	assert.NotNil(t, collector)
	assert.NotNil(t, collector.goalParser)
	assert.NotNil(t, collector.entryStorage)
	assert.NotNil(t, collector.scoringEngine)
	assert.NotNil(t, collector.flowFactory)
	assert.NotNil(t, collector.entries)
	assert.NotNil(t, collector.achievements)
	assert.NotNil(t, collector.notes)
	assert.NotNil(t, collector.statuses)
	assert.Equal(t, 0, len(collector.habits))
	assert.Equal(t, 0, len(collector.entries))
	assert.Equal(t, 0, len(collector.achievements))
	assert.Equal(t, 0, len(collector.notes))
	assert.Equal(t, 0, len(collector.statuses))
}

func TestEntryCollector_loadExistingEntries(t *testing.T) {
	t.Run("load existing entries for today", func(t *testing.T) {
		tempDir := t.TempDir()
		entriesFile := filepath.Join(tempDir, "entries.yml")

		// Create entry for today
		today := time.Now().Format("2006-01-02")
		entryLog := models.CreateEmptyEntryLog()
		dayEntry := models.DayEntry{
			Date: today,
			Habits: []models.HabitEntry{
				{
					HabitID:   "meditation",
					Value:     true,
					Notes:     "Great session",
					Status:    models.EntryCompleted,
					CreatedAt: time.Now(),
				},
				{
					HabitID:   "exercise",
					Value:     false,
					Notes:     "Too tired",
					Status:    models.EntryFailed,
					CreatedAt: time.Now(),
				},
			},
		}
		err := entryLog.AddDayEntry(dayEntry)
		require.NoError(t, err)

		entryStorage := storage.NewEntryStorage()
		err = entryStorage.SaveToFile(entryLog, entriesFile)
		require.NoError(t, err)

		// Load existing entries
		collector := NewEntryCollector("checklists.yml")
		err = collector.loadExistingEntries(entriesFile)
		require.NoError(t, err)

		// Verify entries were loaded
		assert.Equal(t, true, collector.entries["meditation"])
		assert.Equal(t, false, collector.entries["exercise"])
		assert.Equal(t, "Great session", collector.notes["meditation"])
		assert.Equal(t, "Too tired", collector.notes["exercise"])
	})

	t.Run("no existing entries for today", func(t *testing.T) {
		tempDir := t.TempDir()
		entriesFile := filepath.Join(tempDir, "entries.yml")

		// Create entry for different date
		entryLog := models.CreateEmptyEntryLog()
		dayEntry := models.DayEntry{
			Date: "2024-01-01",
			Habits: []models.HabitEntry{
				{HabitID: "meditation", Value: true, Status: models.EntryCompleted, CreatedAt: time.Now()},
			},
		}
		err := entryLog.AddDayEntry(dayEntry)
		require.NoError(t, err)

		entryStorage := storage.NewEntryStorage()
		err = entryStorage.SaveToFile(entryLog, entriesFile)
		require.NoError(t, err)

		// Load existing entries
		collector := NewEntryCollector("checklists.yml")
		err = collector.loadExistingEntries(entriesFile)
		require.NoError(t, err)

		// Verify no entries were loaded for today
		assert.Empty(t, collector.entries)
		assert.Empty(t, collector.notes)
	})

	t.Run("file does not exist", func(t *testing.T) {
		collector := NewEntryCollector("checklists.yml")
		err := collector.loadExistingEntries("/nonexistent/entries.yml")
		require.NoError(t, err)

		// Should not fail when file doesn't exist
		assert.Empty(t, collector.entries)
		assert.Empty(t, collector.notes)
	})
}

func TestEntryCollector_saveEntries(t *testing.T) {
	t.Run("save collected entries", func(t *testing.T) {
		tempDir := t.TempDir()
		entriesFile := filepath.Join(tempDir, "entries.yml")

		// Create habits for testing
		habits := []models.Habit{
			{
				ID:    "meditation",
				Title: "Morning Meditation",
			},
			{
				ID:    "exercise",
				Title: "Daily Exercise",
			},
		}

		// Setup collector with test data
		collector := NewEntryCollector("checklists.yml")
		collector.habits = habits
		collector.entries["meditation"] = true
		collector.entries["exercise"] = false
		collector.notes["meditation"] = "Peaceful session"
		collector.notes["exercise"] = ""
		collector.statuses["meditation"] = models.EntryCompleted
		collector.statuses["exercise"] = models.EntryFailed

		// Save entries
		err := collector.saveEntries(entriesFile)
		require.NoError(t, err)

		// Verify entries were saved
		entryStorage := storage.NewEntryStorage()
		entryLog, err := entryStorage.LoadFromFile(entriesFile)
		require.NoError(t, err)

		today := time.Now().Format("2006-01-02")
		assert.Len(t, entryLog.Entries, 1)
		assert.Equal(t, today, entryLog.Entries[0].Date)
		assert.Len(t, entryLog.Entries[0].Habits, 2)

		// Check meditation habit
		meditationHabit := entryLog.Entries[0].Habits[0]
		assert.Equal(t, "meditation", meditationHabit.HabitID)
		assert.Equal(t, true, meditationHabit.Value)
		assert.Equal(t, "Peaceful session", meditationHabit.Notes)
		assert.False(t, meditationHabit.CreatedAt.IsZero())

		// Check exercise habit
		exerciseHabit := entryLog.Entries[0].Habits[1]
		assert.Equal(t, "exercise", exerciseHabit.HabitID)
		assert.Equal(t, false, exerciseHabit.Value)
		assert.Equal(t, "", exerciseHabit.Notes)
		assert.False(t, exerciseHabit.CreatedAt.IsZero())
	})

	t.Run("skip unprocessed habits", func(t *testing.T) {
		tempDir := t.TempDir()
		entriesFile := filepath.Join(tempDir, "entries.yml")

		// Create habits but only process one
		habits := []models.Habit{
			{ID: "meditation", Title: "Morning Meditation"},
			{ID: "exercise", Title: "Daily Exercise"},
		}

		collector := NewEntryCollector("checklists.yml")
		collector.habits = habits
		collector.entries["meditation"] = true
		collector.statuses["meditation"] = models.EntryCompleted
		// exercise not in entries map (not processed)

		err := collector.saveEntries(entriesFile)
		require.NoError(t, err)

		// Verify only processed habit was saved
		entryStorage := storage.NewEntryStorage()
		entryLog, err := entryStorage.LoadFromFile(entriesFile)
		require.NoError(t, err)

		assert.Len(t, entryLog.Entries, 1)
		assert.Len(t, entryLog.Entries[0].Habits, 1)
		assert.Equal(t, "meditation", entryLog.Entries[0].Habits[0].HabitID)
	})
}

func TestEntryCollector_displayWelcome(t *testing.T) {
	// Create collector with test habits
	collector := NewEntryCollector("checklists.yml")
	collector.habits = []models.Habit{
		{ID: "meditation", Title: "Morning Meditation"},
		{ID: "exercise", Title: "Daily Exercise"},
	}

	// This test mainly ensures the function doesn't panic
	// In a real UI test, we'd capture stdout, but for now we just call it
	collector.displayWelcome()

	// Test passes if no panic occurs
	assert.Len(t, collector.habits, 2)
}

func TestEntryCollector_displayCompletion(t *testing.T) {
	t.Run("perfect completion", func(t *testing.T) {
		collector := NewEntryCollector("checklists.yml")
		collector.habits = []models.Habit{
			{ID: "meditation", Title: "Morning Meditation", HabitType: models.SimpleHabit},
			{ID: "exercise", Title: "Daily Exercise", HabitType: models.SimpleHabit},
		}
		collector.entries = map[string]interface{}{
			"meditation": true,
			"exercise":   true,
		}

		// This test mainly ensures the function doesn't panic
		collector.displayCompletion()

		// Test passes if no panic occurs
		assert.Len(t, collector.habits, 2)
	})

	t.Run("partial completion", func(t *testing.T) {
		collector := NewEntryCollector("checklists.yml")
		collector.habits = []models.Habit{
			{ID: "meditation", Title: "Morning Meditation", HabitType: models.SimpleHabit},
			{ID: "exercise", Title: "Daily Exercise", HabitType: models.SimpleHabit},
		}
		collector.entries = map[string]interface{}{
			"meditation": true,
			"exercise":   false,
		}

		collector.displayCompletion()
		assert.Len(t, collector.habits, 2)
	})

	t.Run("no completion", func(t *testing.T) {
		collector := NewEntryCollector("checklists.yml")
		collector.habits = []models.Habit{
			{ID: "meditation", Title: "Morning Meditation", HabitType: models.SimpleHabit},
			{ID: "exercise", Title: "Daily Exercise", HabitType: models.SimpleHabit},
		}
		collector.entries = map[string]interface{}{
			"meditation": false,
			"exercise":   false,
		}

		collector.displayCompletion()
		assert.Len(t, collector.habits, 2)
	})
}

func TestEntryCollector_timePtr(t *testing.T) {
	now := time.Now()
	ptr := timePtr(now)

	require.NotNil(t, ptr)
	assert.Equal(t, now, *ptr)
}

func TestEntryCollector_CollectTodayEntries_ErrorCases(t *testing.T) {
	t.Run("habits file not found", func(t *testing.T) {
		collector := NewEntryCollector("checklists.yml")

		err := collector.CollectTodayEntries("/nonexistent/habits.yml", "/tmp/entries.yml")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load habits")
	})

	t.Run("no simple boolean habits", func(t *testing.T) {
		tempDir := t.TempDir()
		habitsFile := filepath.Join(tempDir, "habits.yml")
		entriesFile := filepath.Join(tempDir, "entries.yml")

		// Create schema with no simple boolean habits
		schema := &models.Schema{
			Version: "1.0.0",
			Habits:  []models.Habit{},
		}

		goalParser := parser.NewHabitParser()
		err := goalParser.SaveToFile(schema, habitsFile)
		require.NoError(t, err)

		collector := NewEntryCollector("checklists.yml")
		err = collector.CollectTodayEntries(habitsFile, entriesFile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no habits found")
	})
}

func TestEntryCollector_Integration(t *testing.T) {
	t.Run("end to end workflow with mocked UI", func(t *testing.T) {
		tempDir := t.TempDir()
		habitsFile := filepath.Join(tempDir, "habits.yml")
		entriesFile := filepath.Join(tempDir, "entries.yml")

		// Create test schema with simple boolean habits
		schema := &models.Schema{
			Version: "1.0.0",
			Habits: []models.Habit{
				{
					ID:          "meditation",
					Title:       "Morning Meditation",
					Position:    1,
					Description: "10 minutes of mindfulness",
					HabitType:   models.SimpleHabit,
					FieldType:   models.FieldType{Type: models.BooleanFieldType},
					ScoringType: models.ManualScoring,
					Prompt:      "Did you meditate this morning?",
				},
				{
					ID:          "exercise",
					Title:       "Daily Exercise",
					Position:    2,
					Description: "At least 30 minutes of physical activity",
					HabitType:   models.SimpleHabit,
					FieldType:   models.FieldType{Type: models.BooleanFieldType},
					ScoringType: models.ManualScoring,
					Prompt:      "Did you exercise today?",
				},
			},
		}

		goalParser := parser.NewHabitParser()
		err := goalParser.SaveToFile(schema, habitsFile)
		require.NoError(t, err)

		// Test that collector can load habits successfully
		collector := NewEntryCollector("checklists.yml")

		// Load habits manually to test without UI interaction
		loadedSchema, err := collector.goalParser.LoadFromFile(habitsFile)
		require.NoError(t, err)

		collector.habits = parser.GetSimpleBooleanHabits(loadedSchema)
		assert.Len(t, collector.habits, 2)

		// Simulate user entries
		collector.entries["meditation"] = true
		collector.entries["exercise"] = false
		collector.notes["meditation"] = "Great start to the day"
		collector.notes["exercise"] = "Will try tomorrow"
		collector.statuses["meditation"] = models.EntryCompleted
		collector.statuses["exercise"] = models.EntryFailed

		// Test save functionality
		err = collector.saveEntries(entriesFile)
		require.NoError(t, err)

		// Verify saved data
		entryStorage := storage.NewEntryStorage()
		entryLog, err := entryStorage.LoadFromFile(entriesFile)
		require.NoError(t, err)

		today := time.Now().Format("2006-01-02")
		dayEntry, found := entryLog.GetDayEntry(today)
		require.True(t, found)

		assert.Len(t, dayEntry.Habits, 2)

		// Find and verify meditation habit
		var meditationEntry, exerciseEntry *models.HabitEntry
		for i := range dayEntry.Habits {
			switch dayEntry.Habits[i].HabitID {
			case "meditation":
				meditationEntry = &dayEntry.Habits[i]
			case "exercise":
				exerciseEntry = &dayEntry.Habits[i]
			}
		}

		require.NotNil(t, meditationEntry)
		assert.Equal(t, true, meditationEntry.Value)
		assert.Equal(t, "Great start to the day", meditationEntry.Notes)

		require.NotNil(t, exerciseEntry)
		assert.Equal(t, false, exerciseEntry.Value)
		assert.Equal(t, "Will try tomorrow", exerciseEntry.Notes)
	})
}
