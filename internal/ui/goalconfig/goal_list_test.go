package goalconfig

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"davidlee/iter/internal/models"
)

func TestGoalItem_FilterValue(t *testing.T) {
	t.Run("combines title and goal type for filtering", func(t *testing.T) {
		goal := models.Goal{
			Title:    "Morning Meditation",
			GoalType: models.SimpleGoal,
		}
		item := GoalItem{Goal: goal}

		filterValue := item.FilterValue()
		assert.Equal(t, "Morning Meditation simple", filterValue)
	})

	t.Run("handles different goal types", func(t *testing.T) {
		testCases := []struct {
			goalType     models.GoalType
			expectedType string
		}{
			{models.SimpleGoal, "simple"},
			{models.ElasticGoal, "elastic"},
			{models.InformationalGoal, "informational"},
			{models.ChecklistGoal, "checklist"},
		}

		for _, tc := range testCases {
			goal := models.Goal{
				Title:    "Test Goal",
				GoalType: tc.goalType,
			}
			item := GoalItem{Goal: goal}

			filterValue := item.FilterValue()
			assert.Contains(t, filterValue, tc.expectedType)
		}
	})
}

func TestGoalItem_Title(t *testing.T) {
	t.Run("returns goal title", func(t *testing.T) {
		goal := models.Goal{
			Title: "Daily Exercise",
		}
		item := GoalItem{Goal: goal}

		title := item.Title()
		assert.Equal(t, "Daily Exercise", title)
	})
}

func TestGoalItem_Description(t *testing.T) {
	t.Run("formats as ID | Type | Status", func(t *testing.T) {
		goal := models.Goal{
			ID:       "exercise",
			Title:    "Daily Exercise",
			GoalType: models.SimpleGoal,
		}
		item := GoalItem{Goal: goal}

		description := item.Description()
		assert.Equal(t, "exercise | simple | Simple", description)
	})

	t.Run("handles different goal types correctly", func(t *testing.T) {
		testCases := []struct {
			goalType       models.GoalType
			expectedStatus string
		}{
			{models.SimpleGoal, "Simple"},
			{models.ElasticGoal, "Elastic"},
			{models.InformationalGoal, "Info"},
			{models.ChecklistGoal, "Checklist"},
		}

		for _, tc := range testCases {
			goal := models.Goal{
				ID:       "test",
				GoalType: tc.goalType,
			}
			item := GoalItem{Goal: goal}

			description := item.Description()
			assert.Contains(t, description, tc.expectedStatus)
		}
	})
}

func TestGoalItem_getGoalStatus(t *testing.T) {
	t.Run("returns Simple for simple goals", func(t *testing.T) {
		goal := models.Goal{
			GoalType: models.SimpleGoal,
		}
		item := GoalItem{Goal: goal}

		status := item.getGoalStatus()
		assert.Equal(t, "Simple", status)
	})

	t.Run("includes scoring type for simple goals when present", func(t *testing.T) {
		goal := models.Goal{
			GoalType:    models.SimpleGoal,
			ScoringType: models.AutomaticScoring,
		}
		item := GoalItem{Goal: goal}

		status := item.getGoalStatus()
		assert.Equal(t, "Simple (automatic)", status)
	})

	t.Run("returns appropriate status for each goal type", func(t *testing.T) {
		testCases := []struct {
			goalType       models.GoalType
			expectedStatus string
		}{
			{models.ElasticGoal, "Elastic"},
			{models.InformationalGoal, "Info"},
			{models.ChecklistGoal, "Checklist"},
		}

		for _, tc := range testCases {
			goal := models.Goal{
				GoalType: tc.goalType,
			}
			item := GoalItem{Goal: goal}

			status := item.getGoalStatus()
			assert.Equal(t, tc.expectedStatus, status)
		}
	})

	t.Run("returns Unknown for unrecognized goal type", func(t *testing.T) {
		goal := models.Goal{
			GoalType: models.GoalType("unknown"),
		}
		item := GoalItem{Goal: goal}

		status := item.getGoalStatus()
		assert.Equal(t, "Unknown", status)
	})
}

func TestNewGoalListModel(t *testing.T) {
	t.Run("creates model with goals", func(t *testing.T) {
		goals := []models.Goal{
			{
				ID:       "meditation",
				Title:    "Morning Meditation",
				GoalType: models.SimpleGoal,
			},
			{
				ID:       "exercise",
				Title:    "Daily Exercise",
				GoalType: models.ElasticGoal,
			},
		}

		model := NewGoalListModel(goals)

		assert.NotNil(t, model)
		assert.Equal(t, goals, model.goals)
		assert.Equal(t, 2, len(model.list.Items()))
		assert.Equal(t, "Goals", model.list.Title)
		assert.True(t, model.list.FilteringEnabled())
	})

	t.Run("handles empty goal list", func(t *testing.T) {
		goals := []models.Goal{}

		model := NewGoalListModel(goals)

		assert.NotNil(t, model)
		assert.Equal(t, 0, len(model.list.Items()))
	})
}

func TestTruncateOrPad(t *testing.T) {
	t.Run("truncates long text with ellipsis", func(t *testing.T) {
		result := truncateOrPad("This is a very long text", 10)
		assert.Equal(t, "This is...", result)
	})

	t.Run("pads short text with spaces", func(t *testing.T) {
		result := truncateOrPad("Short", 10)
		assert.Equal(t, "Short     ", result)
	})

	t.Run("returns exact text when length matches", func(t *testing.T) {
		result := truncateOrPad("ExactFit12", 10)
		assert.Equal(t, "ExactFit12", result)
	})

	t.Run("handles edge case with width less than ellipsis", func(t *testing.T) {
		result := truncateOrPad("Long text", 2)
		assert.Equal(t, "Lo", result)
	})
}
