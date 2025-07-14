package goalconfig

import (
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
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
	t.Run("returns emoji + goal title", func(t *testing.T) {
		goal := models.Goal{
			Title:    "Daily Exercise",
			GoalType: models.SimpleGoal,
		}
		item := GoalItem{Goal: goal}

		title := item.Title()
		assert.Equal(t, "✅ Daily Exercise", title)
	})
}

func TestGoalItem_Description(t *testing.T) {
	t.Run("formats with indentation for alignment", func(t *testing.T) {
		goal := models.Goal{
			Title:       "Daily Exercise",
			Description: "30 minutes of physical activity",
			GoalType:    models.SimpleGoal,
		}
		item := GoalItem{Goal: goal}

		description := item.Description()
		assert.Equal(t, "   30 minutes of physical activity", description)
	})

	t.Run("returns empty string for empty description", func(t *testing.T) {
		goal := models.Goal{
			Title:       "Daily Exercise",
			Description: "",
			GoalType:    models.SimpleGoal,
		}
		item := GoalItem{Goal: goal}

		description := item.Description()
		assert.Equal(t, "", description)
	})
}

func TestGoalItem_getGoalTypeEmoji(t *testing.T) {
	t.Run("returns correct emoji for each goal type", func(t *testing.T) {
		testCases := []struct {
			goalType      models.GoalType
			expectedEmoji string
		}{
			{models.SimpleGoal, "✅"},
			{models.ElasticGoal, "🎯"},
			{models.InformationalGoal, "📊"},
			{models.ChecklistGoal, "📝"},
		}

		for _, tc := range testCases {
			goal := models.Goal{
				GoalType: tc.goalType,
			}
			item := GoalItem{Goal: goal}

			emoji := item.getGoalTypeEmoji()
			assert.Equal(t, tc.expectedEmoji, emoji)
		}
	})

	t.Run("returns question mark for unknown goal type", func(t *testing.T) {
		goal := models.Goal{
			GoalType: models.GoalType("unknown"),
		}
		item := GoalItem{Goal: goal}

		emoji := item.getGoalTypeEmoji()
		assert.Equal(t, "❓", emoji)
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

func TestGoalListModel_Modal(t *testing.T) {
	t.Run("modal starts closed", func(t *testing.T) {
		goals := []models.Goal{
			{Title: "Test Goal", GoalType: models.SimpleGoal},
		}
		model := NewGoalListModel(goals)

		assert.False(t, model.showModal)
	})

	t.Run("enter key opens modal", func(t *testing.T) {
		goals := []models.Goal{
			{Title: "Test Goal", GoalType: models.SimpleGoal},
		}
		model := NewGoalListModel(goals)
		model.width = 80
		model.height = 24

		// Simulate enter key press
		msg := tea.KeyMsg{Type: tea.KeyEnter}
		updatedModel, _ := model.Update(msg)
		model = updatedModel.(*GoalListModel)

		assert.True(t, model.showModal)
	})

	t.Run("space key opens modal", func(t *testing.T) {
		goals := []models.Goal{
			{Title: "Test Goal", GoalType: models.SimpleGoal},
		}
		model := NewGoalListModel(goals)
		model.width = 80
		model.height = 24

		// Simulate space key press
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
		updatedModel, _ := model.Update(msg)
		model = updatedModel.(*GoalListModel)

		assert.True(t, model.showModal)
	})

	t.Run("escape key closes modal", func(t *testing.T) {
		goals := []models.Goal{
			{Title: "Test Goal", GoalType: models.SimpleGoal},
		}
		model := NewGoalListModel(goals)
		model.showModal = true

		// Simulate escape key press
		msg := tea.KeyMsg{Type: tea.KeyEsc}
		updatedModel, _ := model.Update(msg)
		model = updatedModel.(*GoalListModel)

		assert.False(t, model.showModal)
	})

	t.Run("q key closes modal", func(t *testing.T) {
		goals := []models.Goal{
			{Title: "Test Goal", GoalType: models.SimpleGoal},
		}
		model := NewGoalListModel(goals)
		model.showModal = true

		// Simulate 'q' key press
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
		updatedModel, _ := model.Update(msg)
		model = updatedModel.(*GoalListModel)

		assert.False(t, model.showModal)
	})

	t.Run("non-modal keys ignored when modal is open", func(t *testing.T) {
		goals := []models.Goal{
			{Title: "Test Goal", GoalType: models.SimpleGoal},
		}
		model := NewGoalListModel(goals)
		model.showModal = true

		// Simulate 'e' key press (edit key should be ignored when modal is open)
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}}
		updatedModel, cmd := model.Update(msg)
		model = updatedModel.(*GoalListModel)

		assert.True(t, model.showModal) // Modal should still be open
		assert.Nil(t, cmd)              // Should not trigger any command
	})

	t.Run("getSelectedGoal returns correct goal", func(t *testing.T) {
		goals := []models.Goal{
			{Title: "First Goal", GoalType: models.SimpleGoal},
			{Title: "Second Goal", GoalType: models.ElasticGoal},
		}
		model := NewGoalListModel(goals)

		// Default selection should be first goal
		selectedGoal := model.getSelectedGoal()
		assert.NotNil(t, selectedGoal)
		assert.Equal(t, "First Goal", selectedGoal.Title)
	})

	t.Run("getSelectedGoal handles empty list", func(t *testing.T) {
		goals := []models.Goal{}
		model := NewGoalListModel(goals)

		selectedGoal := model.getSelectedGoal()
		assert.Nil(t, selectedGoal)
	})
}

func TestGoalListKeyMap(t *testing.T) {
	t.Run("default keybindings are properly configured", func(t *testing.T) {
		keys := DefaultGoalListKeyMap()

		// Test that keys are defined
		assert.NotNil(t, keys.Up)
		assert.NotNil(t, keys.Down)
		assert.NotNil(t, keys.ShowDetail)
		assert.NotNil(t, keys.CloseModal)
		assert.NotNil(t, keys.Edit)
		assert.NotNil(t, keys.Delete)
		assert.NotNil(t, keys.Search)
		assert.NotNil(t, keys.Quit)

		// Test help text is available
		shortHelp := keys.ShortHelp()
		assert.Equal(t, 4, len(shortHelp))

		fullHelp := keys.FullHelp()
		assert.Equal(t, 2, len(fullHelp))
		assert.Equal(t, 4, len(fullHelp[0]))
		assert.Equal(t, 4, len(fullHelp[1]))
	})

	t.Run("custom keybindings can be set", func(t *testing.T) {
		goals := []models.Goal{
			{Title: "Test Goal", GoalType: models.SimpleGoal},
		}
		model := NewGoalListModel(goals)

		// Create custom keybindings
		customKeys := DefaultGoalListKeyMap()
		customKeys.Quit = key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "exit"),
		)

		// Apply custom keybindings
		model = model.WithKeyMap(customKeys)

		assert.Equal(t, customKeys, model.keys)
	})
}

func TestGoalListModel_Keybindings(t *testing.T) {
	t.Run("quit key triggers quit command", func(t *testing.T) {
		goals := []models.Goal{
			{Title: "Test Goal", GoalType: models.SimpleGoal},
		}
		model := NewGoalListModel(goals)

		// Simulate 'q' key press
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
		_, cmd := model.Update(msg)

		// Execute the command to check if it's a quit command
		if cmd != nil {
			result := cmd()
			_, isQuitMsg := result.(tea.QuitMsg)
			assert.True(t, isQuitMsg, "Expected quit command")
		} else {
			t.Error("Expected a command to be returned")
		}
	})

	t.Run("ctrl+c triggers quit command", func(t *testing.T) {
		goals := []models.Goal{
			{Title: "Test Goal", GoalType: models.SimpleGoal},
		}
		model := NewGoalListModel(goals)

		// Simulate Ctrl+C key press
		msg := tea.KeyMsg{Type: tea.KeyCtrlC}
		_, cmd := model.Update(msg)

		// Execute the command to check if it's a quit command
		if cmd != nil {
			result := cmd()
			_, isQuitMsg := result.(tea.QuitMsg)
			assert.True(t, isQuitMsg, "Expected quit command")
		} else {
			t.Error("Expected a command to be returned")
		}
	})

	t.Run("future operation keys are handled gracefully", func(t *testing.T) {
		goals := []models.Goal{
			{Title: "Test Goal", GoalType: models.SimpleGoal},
		}
		model := NewGoalListModel(goals)

		testCases := []struct {
			key  rune
			desc string
		}{
			{'e', "edit key"},
			{'d', "delete key"},
			{'/', "search key"},
		}

		for _, tc := range testCases {
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{tc.key}}
			updatedModel, cmd := model.Update(msg)

			// Should not crash and should return the model
			assert.NotNil(t, updatedModel)
			assert.Nil(t, cmd) // No command should be issued yet
		}
	})
}

func TestRenderCriteria(t *testing.T) {
	t.Run("handles nil criteria", func(t *testing.T) {
		result := renderCriteria(nil)
		assert.Equal(t, "None", result)
	})

	t.Run("renders description only", func(t *testing.T) {
		criteria := &models.Criteria{
			Description: "Test description",
		}
		result := renderCriteria(criteria)
		assert.Equal(t, "Test description", result)
	})

	t.Run("renders numeric conditions", func(t *testing.T) {
		greaterThan := 5.0
		criteria := &models.Criteria{
			Condition: &models.Condition{
				GreaterThan: &greaterThan,
			},
		}
		result := renderCriteria(criteria)
		assert.Equal(t, "> 5.00", result)
	})

	t.Run("renders boolean conditions", func(t *testing.T) {
		equals := true
		criteria := &models.Criteria{
			Condition: &models.Condition{
				Equals: &equals,
			},
		}
		result := renderCriteria(criteria)
		assert.Equal(t, "= true", result)
	})

	t.Run("handles empty criteria", func(t *testing.T) {
		criteria := &models.Criteria{}
		result := renderCriteria(criteria)
		assert.Equal(t, "No conditions specified", result)
	})
}
