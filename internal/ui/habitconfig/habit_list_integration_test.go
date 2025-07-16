package habitconfig

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"davidlee/vice/internal/models"
	"davidlee/vice/internal/parser"
)

func TestHabitConfigurator_ListHabits_Integration(t *testing.T) {
	t.Run("lists habits from existing file", func(t *testing.T) {
		// Create temporary directory
		tempDir := t.TempDir()

		habitsFile := filepath.Join(tempDir, "habits.yml")

		// Create sample habits file
		habits := []models.Habit{
			{
				ID:          "meditation",
				Title:       "Morning Meditation",
				Description: "Daily mindfulness practice",
				HabitType:   models.SimpleHabit,
				FieldType:   models.FieldType{Type: models.BooleanFieldType},
				ScoringType: models.ManualScoring,
			},
			{
				ID:          "exercise",
				Title:       "Daily Exercise",
				Description: "Physical fitness routine",
				HabitType:   models.ElasticHabit,
				FieldType:   models.FieldType{Type: models.DurationFieldType},
				ScoringType: models.ManualScoring, // Use manual scoring to avoid criteria requirement
			},
			{
				ID:          "mood",
				Title:       "Mood Tracking",
				Description: "Track daily emotional state",
				HabitType:   models.InformationalHabit,
				FieldType:   models.FieldType{Type: models.UnsignedIntFieldType},
			},
		}

		schema := models.Schema{
			Version:     "1.0",
			CreatedDate: "2024-01-01",
			Habits:      habits,
		}

		// Save habits file
		habitParser := parser.NewHabitParser()
		err := habitParser.SaveToFile(&schema, habitsFile)
		require.NoError(t, err)

		// Test that file loading works (this exercises the ListHabits loading logic)
		// Note: We can't easily test the bubbletea program in unit tests, but we can verify
		// the configurator loads habits correctly and would launch the UI
		loadedSchema, err := habitParser.LoadFromFile(habitsFile)
		require.NoError(t, err)
		assert.Equal(t, 3, len(loadedSchema.Habits))
		assert.Equal(t, "meditation", loadedSchema.Habits[0].ID)
		assert.Equal(t, "exercise", loadedSchema.Habits[1].ID)
		assert.Equal(t, "mood", loadedSchema.Habits[2].ID)

		// Verify the configurator can create a model with the loaded habits
		listModel := NewHabitListModel(loadedSchema.Habits)
		assert.NotNil(t, listModel)
		assert.Equal(t, 3, len(listModel.habits))
		assert.Equal(t, 3, len(listModel.list.Items()))
	})

	t.Run("handles non-existent habits file gracefully", func(t *testing.T) {
		// Create temporary directory
		tempDir := t.TempDir()

		nonExistentFile := filepath.Join(tempDir, "nonexistent.yml")

		// Test that ListHabits handles missing file gracefully
		configurator := NewHabitConfigurator()
		err := configurator.ListHabits(nonExistentFile)

		// Should not error, should handle gracefully with user message
		assert.NoError(t, err)
	})

	t.Run("handles empty habits file gracefully", func(t *testing.T) {
		// Create temporary directory
		tempDir := t.TempDir()

		habitsFile := filepath.Join(tempDir, "empty_habits.yml")

		// Create empty habits file
		schema := models.Schema{
			Version:     "1.0",
			CreatedDate: "2024-01-01",
			Habits:      []models.Habit{}, // Empty habits list
		}

		habitParser := parser.NewHabitParser()
		err := habitParser.SaveToFile(&schema, habitsFile)
		require.NoError(t, err)

		// Test ListHabits with empty file
		configurator := NewHabitConfigurator()
		err = configurator.ListHabits(habitsFile)

		// Should handle empty list gracefully
		assert.NoError(t, err)
	})
}

func TestHabitListModel_WithRealHabitData(t *testing.T) {
	t.Run("creates list model with realistic habit variety", func(t *testing.T) {
		// Test with realistic habit data covering all types
		habits := []models.Habit{
			{
				ID:          "meditation",
				Title:       "Morning Meditation",
				Description: "Daily mindfulness practice",
				HabitType:   models.SimpleHabit,
				FieldType:   models.FieldType{Type: models.BooleanFieldType},
				ScoringType: models.ManualScoring,
			},
			{
				ID:          "exercise_duration",
				Title:       "Exercise Duration",
				Description: "Track workout time",
				HabitType:   models.ElasticHabit,
				FieldType:   models.FieldType{Type: models.DurationFieldType},
				ScoringType: models.ManualScoring,
			},
			{
				ID:          "mood_rating",
				Title:       "Daily Mood Rating",
				Description: "Track emotional state",
				HabitType:   models.InformationalHabit,
				FieldType:   models.FieldType{Type: models.UnsignedIntFieldType},
			},
			{
				ID:          "morning_routine",
				Title:       "Morning Routine Checklist",
				Description: "Complete morning tasks",
				HabitType:   models.ChecklistHabit,
				FieldType:   models.FieldType{Type: models.ChecklistFieldType},
				ScoringType: models.AutomaticScoring,
			},
		}

		model := NewHabitListModel(habits)

		assert.NotNil(t, model)
		assert.Equal(t, 4, len(model.habits))
		assert.Equal(t, 4, len(model.list.Items()))

		// Verify each habit type is properly represented
		items := model.list.Items()

		// Test first item (SimpleHabit)
		item0, ok := items[0].(HabitItem)
		require.True(t, ok)
		assert.Equal(t, "Morning Meditation", item0.Habit.Title)
		assert.Equal(t, "meditation", item0.Habit.ID)
		assert.Contains(t, item0.Title(), "✅")

		// Test second item (ElasticHabit)
		item1, ok := items[1].(HabitItem)
		require.True(t, ok)
		assert.Equal(t, "Exercise Duration", item1.Habit.Title)
		assert.Contains(t, item1.Title(), "🎯")

		// Test third item (InformationalHabit)
		item2, ok := items[2].(HabitItem)
		require.True(t, ok)
		assert.Equal(t, "Daily Mood Rating", item2.Habit.Title)
		assert.Contains(t, item2.Title(), "📊")

		// Test fourth item (ChecklistHabit)
		item3, ok := items[3].(HabitItem)
		require.True(t, ok)
		assert.Equal(t, "Morning Routine Checklist", item3.Habit.Title)
		assert.Contains(t, item3.Title(), "📝")
	})

	t.Run("filtering works with realistic habit data", func(t *testing.T) {
		habits := []models.Habit{
			{Title: "Morning Meditation", HabitType: models.SimpleHabit},
			{Title: "Evening Meditation", HabitType: models.SimpleHabit},
			{Title: "Exercise", HabitType: models.ElasticHabit},
			{Title: "Mood Check", HabitType: models.InformationalHabit},
		}

		model := NewHabitListModel(habits)

		// Test filtering functionality
		items := model.list.Items()

		// Verify filter values contain both title and type
		item0, ok := items[0].(HabitItem)
		require.True(t, ok)
		filterValue := item0.FilterValue()
		assert.Contains(t, filterValue, "Morning Meditation")
		assert.Contains(t, filterValue, "simple")

		// Test that meditation habits would be found by "meditation" filter
		item1, ok := items[1].(HabitItem)
		require.True(t, ok)
		filterValue1 := item1.FilterValue()
		assert.Contains(t, filterValue1, "Evening Meditation")
		assert.Contains(t, filterValue1, "simple")
	})
}
