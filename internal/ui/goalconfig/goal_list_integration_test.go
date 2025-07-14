package goalconfig

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"davidlee/vice/internal/models"
	"davidlee/vice/internal/parser"
)

func TestGoalConfigurator_ListGoals_Integration(t *testing.T) {
	t.Run("lists goals from existing file", func(t *testing.T) {
		// Create temporary directory
		tempDir := t.TempDir()

		goalsFile := filepath.Join(tempDir, "goals.yml")

		// Create sample goals file
		goals := []models.Goal{
			{
				ID:          "meditation",
				Title:       "Morning Meditation",
				Description: "Daily mindfulness practice",
				GoalType:    models.SimpleGoal,
				FieldType:   models.FieldType{Type: models.BooleanFieldType},
				ScoringType: models.ManualScoring,
			},
			{
				ID:          "exercise",
				Title:       "Daily Exercise",
				Description: "Physical fitness routine",
				GoalType:    models.ElasticGoal,
				FieldType:   models.FieldType{Type: models.DurationFieldType},
				ScoringType: models.ManualScoring, // Use manual scoring to avoid criteria requirement
			},
			{
				ID:          "mood",
				Title:       "Mood Tracking",
				Description: "Track daily emotional state",
				GoalType:    models.InformationalGoal,
				FieldType:   models.FieldType{Type: models.UnsignedIntFieldType},
			},
		}

		schema := models.Schema{
			Version:     "1.0",
			CreatedDate: "2024-01-01",
			Goals:       goals,
		}

		// Save goals file
		goalParser := parser.NewGoalParser()
		err := goalParser.SaveToFile(&schema, goalsFile)
		require.NoError(t, err)

		// Test that file loading works (this exercises the ListGoals loading logic)
		// Note: We can't easily test the bubbletea program in unit tests, but we can verify
		// the configurator loads goals correctly and would launch the UI
		loadedSchema, err := goalParser.LoadFromFile(goalsFile)
		require.NoError(t, err)
		assert.Equal(t, 3, len(loadedSchema.Goals))
		assert.Equal(t, "meditation", loadedSchema.Goals[0].ID)
		assert.Equal(t, "exercise", loadedSchema.Goals[1].ID)
		assert.Equal(t, "mood", loadedSchema.Goals[2].ID)

		// Verify the configurator can create a model with the loaded goals
		listModel := NewGoalListModel(loadedSchema.Goals)
		assert.NotNil(t, listModel)
		assert.Equal(t, 3, len(listModel.goals))
		assert.Equal(t, 3, len(listModel.list.Items()))
	})

	t.Run("handles non-existent goals file gracefully", func(t *testing.T) {
		// Create temporary directory
		tempDir := t.TempDir()

		nonExistentFile := filepath.Join(tempDir, "nonexistent.yml")

		// Test that ListGoals handles missing file gracefully
		configurator := NewGoalConfigurator()
		err := configurator.ListGoals(nonExistentFile)

		// Should not error, should handle gracefully with user message
		assert.NoError(t, err)
	})

	t.Run("handles empty goals file gracefully", func(t *testing.T) {
		// Create temporary directory
		tempDir := t.TempDir()

		goalsFile := filepath.Join(tempDir, "empty_goals.yml")

		// Create empty goals file
		schema := models.Schema{
			Version:     "1.0",
			CreatedDate: "2024-01-01",
			Goals:       []models.Goal{}, // Empty goals list
		}

		goalParser := parser.NewGoalParser()
		err := goalParser.SaveToFile(&schema, goalsFile)
		require.NoError(t, err)

		// Test ListGoals with empty file
		configurator := NewGoalConfigurator()
		err = configurator.ListGoals(goalsFile)

		// Should handle empty list gracefully
		assert.NoError(t, err)
	})
}

func TestGoalListModel_WithRealGoalData(t *testing.T) {
	t.Run("creates list model with realistic goal variety", func(t *testing.T) {
		// Test with realistic goal data covering all types
		goals := []models.Goal{
			{
				ID:          "meditation",
				Title:       "Morning Meditation",
				Description: "Daily mindfulness practice",
				GoalType:    models.SimpleGoal,
				FieldType:   models.FieldType{Type: models.BooleanFieldType},
				ScoringType: models.ManualScoring,
			},
			{
				ID:          "exercise_duration",
				Title:       "Exercise Duration",
				Description: "Track workout time",
				GoalType:    models.ElasticGoal,
				FieldType:   models.FieldType{Type: models.DurationFieldType},
				ScoringType: models.ManualScoring,
			},
			{
				ID:          "mood_rating",
				Title:       "Daily Mood Rating",
				Description: "Track emotional state",
				GoalType:    models.InformationalGoal,
				FieldType:   models.FieldType{Type: models.UnsignedIntFieldType},
			},
			{
				ID:          "morning_routine",
				Title:       "Morning Routine Checklist",
				Description: "Complete morning tasks",
				GoalType:    models.ChecklistGoal,
				FieldType:   models.FieldType{Type: models.ChecklistFieldType},
				ScoringType: models.AutomaticScoring,
			},
		}

		model := NewGoalListModel(goals)

		assert.NotNil(t, model)
		assert.Equal(t, 4, len(model.goals))
		assert.Equal(t, 4, len(model.list.Items()))

		// Verify each goal type is properly represented
		items := model.list.Items()

		// Test first item (SimpleGoal)
		item0, ok := items[0].(GoalItem)
		require.True(t, ok)
		assert.Equal(t, "Morning Meditation", item0.Goal.Title)
		assert.Equal(t, "meditation", item0.Goal.ID)
		assert.Contains(t, item0.Title(), "✅")

		// Test second item (ElasticGoal)
		item1, ok := items[1].(GoalItem)
		require.True(t, ok)
		assert.Equal(t, "Exercise Duration", item1.Goal.Title)
		assert.Contains(t, item1.Title(), "🎯")

		// Test third item (InformationalGoal)
		item2, ok := items[2].(GoalItem)
		require.True(t, ok)
		assert.Equal(t, "Daily Mood Rating", item2.Goal.Title)
		assert.Contains(t, item2.Title(), "📊")

		// Test fourth item (ChecklistGoal)
		item3, ok := items[3].(GoalItem)
		require.True(t, ok)
		assert.Equal(t, "Morning Routine Checklist", item3.Goal.Title)
		assert.Contains(t, item3.Title(), "📝")
	})

	t.Run("filtering works with realistic goal data", func(t *testing.T) {
		goals := []models.Goal{
			{Title: "Morning Meditation", GoalType: models.SimpleGoal},
			{Title: "Evening Meditation", GoalType: models.SimpleGoal},
			{Title: "Exercise", GoalType: models.ElasticGoal},
			{Title: "Mood Check", GoalType: models.InformationalGoal},
		}

		model := NewGoalListModel(goals)

		// Test filtering functionality
		items := model.list.Items()

		// Verify filter values contain both title and type
		item0, ok := items[0].(GoalItem)
		require.True(t, ok)
		filterValue := item0.FilterValue()
		assert.Contains(t, filterValue, "Morning Meditation")
		assert.Contains(t, filterValue, "simple")

		// Test that meditation goals would be found by "meditation" filter
		item1, ok := items[1].(GoalItem)
		require.True(t, ok)
		filterValue1 := item1.FilterValue()
		assert.Contains(t, filterValue1, "Evening Meditation")
		assert.Contains(t, filterValue1, "simple")
	})
}
