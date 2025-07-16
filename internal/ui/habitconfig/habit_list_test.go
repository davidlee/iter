package habitconfig

import (
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"davidlee/vice/internal/models"
)

func TestHabitItem_FilterValue(t *testing.T) {
	t.Run("combines title and habit type for filtering", func(t *testing.T) {
		habit := models.Habit{
			Title:     "Morning Meditation",
			HabitType: models.SimpleHabit,
		}
		item := HabitItem{Habit: habit}

		filterValue := item.FilterValue()
		assert.Equal(t, "Morning Meditation simple", filterValue)
	})

	t.Run("handles different habit types", func(t *testing.T) {
		testCases := []struct {
			habitType    models.HabitType
			expectedType string
		}{
			{models.SimpleHabit, "simple"},
			{models.ElasticHabit, "elastic"},
			{models.InformationalHabit, "informational"},
			{models.ChecklistHabit, "checklist"},
		}

		for _, tc := range testCases {
			habit := models.Habit{
				Title:     "Test Habit",
				HabitType: tc.habitType,
			}
			item := HabitItem{Habit: habit}

			filterValue := item.FilterValue()
			assert.Contains(t, filterValue, tc.expectedType)
		}
	})
}

func TestHabitItem_Title(t *testing.T) {
	t.Run("returns emoji + habit title", func(t *testing.T) {
		habit := models.Habit{
			Title:     "Daily Exercise",
			HabitType: models.SimpleHabit,
		}
		item := HabitItem{Habit: habit}

		title := item.Title()
		assert.Equal(t, "✅ Daily Exercise", title)
	})
}

func TestHabitItem_Description(t *testing.T) {
	t.Run("formats with indentation for alignment", func(t *testing.T) {
		habit := models.Habit{
			Title:       "Daily Exercise",
			Description: "30 minutes of physical activity",
			HabitType:   models.SimpleHabit,
		}
		item := HabitItem{Habit: habit}

		description := item.Description()
		assert.Equal(t, "   30 minutes of physical activity", description)
	})

	t.Run("returns empty string for empty description", func(t *testing.T) {
		habit := models.Habit{
			Title:       "Daily Exercise",
			Description: "",
			HabitType:   models.SimpleHabit,
		}
		item := HabitItem{Habit: habit}

		description := item.Description()
		assert.Equal(t, "", description)
	})
}

func TestHabitItem_getHabitTypeEmoji(t *testing.T) {
	t.Run("returns correct emoji for each habit type", func(t *testing.T) {
		testCases := []struct {
			habitType     models.HabitType
			expectedEmoji string
		}{
			{models.SimpleHabit, "✅"},
			{models.ElasticHabit, "🎯"},
			{models.InformationalHabit, "📊"},
			{models.ChecklistHabit, "📝"},
		}

		for _, tc := range testCases {
			habit := models.Habit{
				HabitType: tc.habitType,
			}
			item := HabitItem{Habit: habit}

			emoji := item.getHabitTypeEmoji()
			assert.Equal(t, tc.expectedEmoji, emoji)
		}
	})

	t.Run("returns question mark for unknown habit type", func(t *testing.T) {
		habit := models.Habit{
			HabitType: models.HabitType("unknown"),
		}
		item := HabitItem{Habit: habit}

		emoji := item.getHabitTypeEmoji()
		assert.Equal(t, "❓", emoji)
	})
}

func TestNewHabitListModel(t *testing.T) {
	t.Run("creates model with habits", func(t *testing.T) {
		habits := []models.Habit{
			{
				ID:        "meditation",
				Title:     "Morning Meditation",
				HabitType: models.SimpleHabit,
			},
			{
				ID:        "exercise",
				Title:     "Daily Exercise",
				HabitType: models.ElasticHabit,
			},
		}

		model := NewHabitListModel(habits)

		assert.NotNil(t, model)
		assert.Equal(t, habits, model.habits)
		assert.Equal(t, 2, len(model.list.Items()))
		assert.Equal(t, "Habits", model.list.Title)
		assert.True(t, model.list.FilteringEnabled())
	})

	t.Run("handles empty habit list", func(t *testing.T) {
		habits := []models.Habit{}

		model := NewHabitListModel(habits)

		assert.NotNil(t, model)
		assert.Equal(t, 0, len(model.list.Items()))
	})
}

func TestHabitListModel_Modal(t *testing.T) {
	t.Run("modal starts closed", func(t *testing.T) {
		habits := []models.Habit{
			{Title: "Test Habit", HabitType: models.SimpleHabit},
		}
		model := NewHabitListModel(habits)

		assert.False(t, model.showModal)
	})

	t.Run("enter key opens modal", func(t *testing.T) {
		habits := []models.Habit{
			{Title: "Test Habit", HabitType: models.SimpleHabit},
		}
		model := NewHabitListModel(habits)
		model.width = 80
		model.height = 24

		// Simulate enter key press
		msg := tea.KeyMsg{Type: tea.KeyEnter}
		updatedModel, _ := model.Update(msg)
		model = updatedModel.(*HabitListModel)

		assert.True(t, model.showModal)
	})

	t.Run("space key opens modal", func(t *testing.T) {
		habits := []models.Habit{
			{Title: "Test Habit", HabitType: models.SimpleHabit},
		}
		model := NewHabitListModel(habits)
		model.width = 80
		model.height = 24

		// Simulate space key press
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
		updatedModel, _ := model.Update(msg)
		model = updatedModel.(*HabitListModel)

		assert.True(t, model.showModal)
	})

	t.Run("escape key closes modal", func(t *testing.T) {
		habits := []models.Habit{
			{Title: "Test Habit", HabitType: models.SimpleHabit},
		}
		model := NewHabitListModel(habits)
		model.showModal = true

		// Simulate escape key press
		msg := tea.KeyMsg{Type: tea.KeyEsc}
		updatedModel, _ := model.Update(msg)
		model = updatedModel.(*HabitListModel)

		assert.False(t, model.showModal)
	})

	t.Run("q key closes modal", func(t *testing.T) {
		habits := []models.Habit{
			{Title: "Test Habit", HabitType: models.SimpleHabit},
		}
		model := NewHabitListModel(habits)
		model.showModal = true

		// Simulate 'q' key press
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
		updatedModel, _ := model.Update(msg)
		model = updatedModel.(*HabitListModel)

		assert.False(t, model.showModal)
	})

	t.Run("non-modal keys ignored when modal is open", func(t *testing.T) {
		habits := []models.Habit{
			{Title: "Test Habit", HabitType: models.SimpleHabit},
		}
		model := NewHabitListModel(habits)
		model.showModal = true

		// Simulate 'e' key press (edit key should be ignored when modal is open)
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}}
		updatedModel, cmd := model.Update(msg)
		model = updatedModel.(*HabitListModel)

		assert.True(t, model.showModal) // Modal should still be open
		assert.Nil(t, cmd)              // Should not trigger any command
	})

	t.Run("getSelectedHabit returns correct habit", func(t *testing.T) {
		habits := []models.Habit{
			{Title: "First Habit", HabitType: models.SimpleHabit},
			{Title: "Second Habit", HabitType: models.ElasticHabit},
		}
		model := NewHabitListModel(habits)

		// Default selection should be first habit
		selectedHabit := model.getSelectedHabit()
		assert.NotNil(t, selectedHabit)
		assert.Equal(t, "First Habit", selectedHabit.Title)
	})

	t.Run("getSelectedHabit handles empty list", func(t *testing.T) {
		habits := []models.Habit{}
		model := NewHabitListModel(habits)

		selectedHabit := model.getSelectedHabit()
		assert.Nil(t, selectedHabit)
	})
}

func TestHabitListKeyMap(t *testing.T) {
	t.Run("default keybindings are properly configured", func(t *testing.T) {
		keys := DefaultHabitListKeyMap()

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
		assert.Equal(t, 6, len(shortHelp)) // Up, Down, ShowDetail, Edit, Delete, Quit

		fullHelp := keys.FullHelp()
		assert.Equal(t, 2, len(fullHelp))
		assert.Equal(t, 4, len(fullHelp[0]))
		assert.Equal(t, 4, len(fullHelp[1]))
	})

	t.Run("custom keybindings can be set", func(t *testing.T) {
		habits := []models.Habit{
			{Title: "Test Habit", HabitType: models.SimpleHabit},
		}
		model := NewHabitListModel(habits)

		// Create custom keybindings
		customKeys := DefaultHabitListKeyMap()
		customKeys.Quit = key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "exit"),
		)

		// Apply custom keybindings
		model = model.WithKeyMap(customKeys)

		assert.Equal(t, customKeys, model.keys)
	})
}

func TestHabitListModel_Keybindings(t *testing.T) {
	t.Run("quit key triggers quit command", func(t *testing.T) {
		habits := []models.Habit{
			{Title: "Test Habit", HabitType: models.SimpleHabit},
		}
		model := NewHabitListModel(habits)

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
		habits := []models.Habit{
			{Title: "Test Habit", HabitType: models.SimpleHabit},
		}
		model := NewHabitListModel(habits)

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

	t.Run("operation keys work correctly", func(t *testing.T) {
		habits := []models.Habit{
			{Title: "Test Habit", HabitType: models.SimpleHabit, ID: "test-1"},
		}
		model := NewHabitListModel(habits)

		// Edit and delete keys should trigger quit (to initiate operation)
		editMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}}
		updatedModel, cmd := model.Update(editMsg)
		assert.NotNil(t, updatedModel)
		assert.NotNil(t, cmd) // Should issue tea.Quit command

		deleteMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
		updatedModel, cmd = model.Update(deleteMsg)
		assert.NotNil(t, updatedModel)
		assert.NotNil(t, cmd) // Should issue tea.Quit command

		// Search key is not yet implemented
		searchMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
		updatedModel, cmd = model.Update(searchMsg)
		assert.NotNil(t, updatedModel)
		assert.Nil(t, cmd) // No command should be issued yet
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
