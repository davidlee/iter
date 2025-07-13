package ui

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"davidlee/iter/internal/models"
	"davidlee/iter/internal/parser"
	"davidlee/iter/internal/storage"
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
	assert.Equal(t, 0, len(collector.goals))
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
			Goals: []models.GoalEntry{
				{
					GoalID:    "meditation",
					Value:     true,
					Notes:     "Great session",
					Status:    models.EntryCompleted,
					CreatedAt: time.Now(),
				},
				{
					GoalID:    "exercise",
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
			Goals: []models.GoalEntry{
				{GoalID: "meditation", Value: true, Status: models.EntryCompleted, CreatedAt: time.Now()},
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

		// Create goals for testing
		goals := []models.Goal{
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
		collector.goals = goals
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
		assert.Len(t, entryLog.Entries[0].Goals, 2)

		// Check meditation goal
		meditationGoal := entryLog.Entries[0].Goals[0]
		assert.Equal(t, "meditation", meditationGoal.GoalID)
		assert.Equal(t, true, meditationGoal.Value)
		assert.Equal(t, "Peaceful session", meditationGoal.Notes)
		assert.False(t, meditationGoal.CreatedAt.IsZero())

		// Check exercise goal
		exerciseGoal := entryLog.Entries[0].Goals[1]
		assert.Equal(t, "exercise", exerciseGoal.GoalID)
		assert.Equal(t, false, exerciseGoal.Value)
		assert.Equal(t, "", exerciseGoal.Notes)
		assert.False(t, exerciseGoal.CreatedAt.IsZero())
	})

	t.Run("skip unprocessed goals", func(t *testing.T) {
		tempDir := t.TempDir()
		entriesFile := filepath.Join(tempDir, "entries.yml")

		// Create goals but only process one
		goals := []models.Goal{
			{ID: "meditation", Title: "Morning Meditation"},
			{ID: "exercise", Title: "Daily Exercise"},
		}

		collector := NewEntryCollector("checklists.yml")
		collector.goals = goals
		collector.entries["meditation"] = true
		collector.statuses["meditation"] = models.EntryCompleted
		// exercise not in entries map (not processed)

		err := collector.saveEntries(entriesFile)
		require.NoError(t, err)

		// Verify only processed goal was saved
		entryStorage := storage.NewEntryStorage()
		entryLog, err := entryStorage.LoadFromFile(entriesFile)
		require.NoError(t, err)

		assert.Len(t, entryLog.Entries, 1)
		assert.Len(t, entryLog.Entries[0].Goals, 1)
		assert.Equal(t, "meditation", entryLog.Entries[0].Goals[0].GoalID)
	})
}

func TestEntryCollector_displayWelcome(t *testing.T) {
	// Create collector with test goals
	collector := NewEntryCollector("checklists.yml")
	collector.goals = []models.Goal{
		{ID: "meditation", Title: "Morning Meditation"},
		{ID: "exercise", Title: "Daily Exercise"},
	}

	// This test mainly ensures the function doesn't panic
	// In a real UI test, we'd capture stdout, but for now we just call it
	collector.displayWelcome()

	// Test passes if no panic occurs
	assert.Len(t, collector.goals, 2)
}

func TestEntryCollector_displayCompletion(t *testing.T) {
	t.Run("perfect completion", func(t *testing.T) {
		collector := NewEntryCollector("checklists.yml")
		collector.goals = []models.Goal{
			{ID: "meditation", Title: "Morning Meditation", GoalType: models.SimpleGoal},
			{ID: "exercise", Title: "Daily Exercise", GoalType: models.SimpleGoal},
		}
		collector.entries = map[string]interface{}{
			"meditation": true,
			"exercise":   true,
		}

		// This test mainly ensures the function doesn't panic
		collector.displayCompletion()

		// Test passes if no panic occurs
		assert.Len(t, collector.goals, 2)
	})

	t.Run("partial completion", func(t *testing.T) {
		collector := NewEntryCollector("checklists.yml")
		collector.goals = []models.Goal{
			{ID: "meditation", Title: "Morning Meditation", GoalType: models.SimpleGoal},
			{ID: "exercise", Title: "Daily Exercise", GoalType: models.SimpleGoal},
		}
		collector.entries = map[string]interface{}{
			"meditation": true,
			"exercise":   false,
		}

		collector.displayCompletion()
		assert.Len(t, collector.goals, 2)
	})

	t.Run("no completion", func(t *testing.T) {
		collector := NewEntryCollector("checklists.yml")
		collector.goals = []models.Goal{
			{ID: "meditation", Title: "Morning Meditation", GoalType: models.SimpleGoal},
			{ID: "exercise", Title: "Daily Exercise", GoalType: models.SimpleGoal},
		}
		collector.entries = map[string]interface{}{
			"meditation": false,
			"exercise":   false,
		}

		collector.displayCompletion()
		assert.Len(t, collector.goals, 2)
	})
}

func TestEntryCollector_timePtr(t *testing.T) {
	now := time.Now()
	ptr := timePtr(now)

	require.NotNil(t, ptr)
	assert.Equal(t, now, *ptr)
}

func TestEntryCollector_CollectTodayEntries_ErrorCases(t *testing.T) {
	t.Run("goals file not found", func(t *testing.T) {
		collector := NewEntryCollector("checklists.yml")

		err := collector.CollectTodayEntries("/nonexistent/goals.yml", "/tmp/entries.yml")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load goals")
	})

	t.Run("no simple boolean goals", func(t *testing.T) {
		tempDir := t.TempDir()
		goalsFile := filepath.Join(tempDir, "goals.yml")
		entriesFile := filepath.Join(tempDir, "entries.yml")

		// Create schema with no simple boolean goals
		schema := &models.Schema{
			Version: "1.0.0",
			Goals:   []models.Goal{},
		}

		goalParser := parser.NewGoalParser()
		err := goalParser.SaveToFile(schema, goalsFile)
		require.NoError(t, err)

		collector := NewEntryCollector("checklists.yml")
		err = collector.CollectTodayEntries(goalsFile, entriesFile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no goals found")
	})
}

func TestEntryCollector_Integration(t *testing.T) {
	t.Run("end to end workflow with mocked UI", func(t *testing.T) {
		tempDir := t.TempDir()
		goalsFile := filepath.Join(tempDir, "goals.yml")
		entriesFile := filepath.Join(tempDir, "entries.yml")

		// Create test schema with simple boolean goals
		schema := &models.Schema{
			Version: "1.0.0",
			Goals: []models.Goal{
				{
					ID:          "meditation",
					Title:       "Morning Meditation",
					Position:    1,
					Description: "10 minutes of mindfulness",
					GoalType:    models.SimpleGoal,
					FieldType:   models.FieldType{Type: models.BooleanFieldType},
					ScoringType: models.ManualScoring,
					Prompt:      "Did you meditate this morning?",
				},
				{
					ID:          "exercise",
					Title:       "Daily Exercise",
					Position:    2,
					Description: "At least 30 minutes of physical activity",
					GoalType:    models.SimpleGoal,
					FieldType:   models.FieldType{Type: models.BooleanFieldType},
					ScoringType: models.ManualScoring,
					Prompt:      "Did you exercise today?",
				},
			},
		}

		goalParser := parser.NewGoalParser()
		err := goalParser.SaveToFile(schema, goalsFile)
		require.NoError(t, err)

		// Test that collector can load goals successfully
		collector := NewEntryCollector("checklists.yml")

		// Load goals manually to test without UI interaction
		loadedSchema, err := collector.goalParser.LoadFromFile(goalsFile)
		require.NoError(t, err)

		collector.goals = parser.GetSimpleBooleanGoals(loadedSchema)
		assert.Len(t, collector.goals, 2)

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

		assert.Len(t, dayEntry.Goals, 2)

		// Find and verify meditation goal
		var meditationEntry, exerciseEntry *models.GoalEntry
		for i := range dayEntry.Goals {
			switch dayEntry.Goals[i].GoalID {
			case "meditation":
				meditationEntry = &dayEntry.Goals[i]
			case "exercise":
				exerciseEntry = &dayEntry.Goals[i]
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
