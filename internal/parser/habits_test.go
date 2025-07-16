package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"davidlee/vice/internal/models"
)

func TestHabitParser_ParseYAML(t *testing.T) {
	parser := NewHabitParser()

	t.Run("valid simple boolean habits schema", func(t *testing.T) {
		yamlData := `
version: "1.0.0"
created_date: "2024-01-01"
habits:
  - title: "Morning Meditation"
    position: 1
    habit_type: "simple"
    field_type:
      type: "boolean"
    scoring_type: "manual"
    prompt: "Did you meditate this morning?"
  - title: "Daily Exercise"
    position: 2
    habit_type: "simple"
    field_type:
      type: "boolean"
    scoring_type: "automatic"
    criteria:
      description: "Exercise completed"
      condition:
        equals: true
`

		schema, err := parser.ParseYAML([]byte(yamlData))
		require.NoError(t, err)
		require.NotNil(t, schema)

		assert.Equal(t, "1.0.0", schema.Version)
		assert.Equal(t, "2024-01-01", schema.CreatedDate)
		assert.Len(t, schema.Habits, 2)

		// Check first habit
		habit1 := schema.Habits[0]
		assert.Equal(t, "Morning Meditation", habit1.Title)
		assert.Equal(t, "morning_meditation", habit1.ID) // Auto-generated
		assert.Equal(t, 1, habit1.Position)
		assert.Equal(t, models.SimpleHabit, habit1.HabitType)
		assert.Equal(t, models.BooleanFieldType, habit1.FieldType.Type)
		assert.Equal(t, models.ManualScoring, habit1.ScoringType)
		assert.Equal(t, "Did you meditate this morning?", habit1.Prompt)

		// Check second habit
		habit2 := schema.Habits[1]
		assert.Equal(t, "Daily Exercise", habit2.Title)
		assert.Equal(t, "daily_exercise", habit2.ID)
		assert.Equal(t, models.AutomaticScoring, habit2.ScoringType)
		require.NotNil(t, habit2.Criteria)
		assert.Equal(t, "Exercise completed", habit2.Criteria.Description)
		require.NotNil(t, habit2.Criteria.Condition.Equals)
		assert.True(t, *habit2.Criteria.Condition.Equals)
	})

	t.Run("schema with custom IDs", func(t *testing.T) {
		yamlData := `
version: "1.0.0"
habits:
  - title: "Morning Meditation"
    id: "custom_meditation"
    position: 1
    habit_type: "simple"
    field_type:
      type: "boolean"
    scoring_type: "manual"
`

		schema, err := parser.ParseYAML([]byte(yamlData))
		require.NoError(t, err)

		assert.Equal(t, "custom_meditation", schema.Habits[0].ID)
	})

	t.Run("invalid YAML syntax", func(t *testing.T) {
		yamlData := `
version: "1.0.0"
habits:
  - title: "Test"
    invalid_yaml: [unclosed
`

		_, err := parser.ParseYAML([]byte(yamlData))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse YAML")
	})

	t.Run("unknown field in strict mode", func(t *testing.T) {
		yamlData := `
version: "1.0.0"
unknown_field: "should cause error"
habits: []
`

		_, err := parser.ParseYAML([]byte(yamlData))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse YAML")
	})

	t.Run("schema validation failure", func(t *testing.T) {
		yamlData := `
version: "1.0.0"
habits:
  - title: "Test Habit"
    position: 1
    habit_type: "simple"
    field_type:
      type: "boolean"
    # Missing scoring_type for simple habit
`

		_, err := parser.ParseYAML([]byte(yamlData))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "schema validation failed")
		assert.Contains(t, err.Error(), "scoring_type is required")
	})

	t.Run("positions auto-assigned when duplicated", func(t *testing.T) {
		yamlData := `
version: "1.0.0"
habits:
  - title: "Habit 1"
    position: 1
    habit_type: "simple"
    field_type:
      type: "boolean"
    scoring_type: "manual"
  - title: "Habit 2"
    position: 1  # Duplicate position - should be auto-corrected
    habit_type: "simple"
    field_type:
      type: "boolean"
    scoring_type: "manual"
`

		schema, err := parser.ParseYAML([]byte(yamlData))
		require.NoError(t, err)
		require.NotNil(t, schema)

		// Positions should be auto-assigned based on order
		assert.Equal(t, 1, schema.Habits[0].Position)
		assert.Equal(t, 2, schema.Habits[1].Position)
	})

	t.Run("elastic habit with numeric criteria", func(t *testing.T) {
		yamlData := `
version: "1.0.0"
created_date: "2024-01-01"
habits:
  - title: "Exercise Duration"
    position: 1
    habit_type: "elastic"
    field_type:
      type: "duration"
    scoring_type: "automatic"
    mini_criteria:
      description: "Minimum exercise"
      condition:
        greater_than_or_equal: 15
    midi_criteria:
      description: "Target exercise"
      condition:
        greater_than_or_equal: 30
    maxi_criteria:
      description: "Excellent exercise"
      condition:
        greater_than_or_equal: 60
    prompt: "How many minutes did you exercise today?"
`

		schema, err := parser.ParseYAML([]byte(yamlData))
		require.NoError(t, err)
		require.NotNil(t, schema)

		assert.Len(t, schema.Habits, 1)
		habit := schema.Habits[0]

		assert.Equal(t, "Exercise Duration", habit.Title)
		assert.Equal(t, models.ElasticHabit, habit.HabitType)
		assert.Equal(t, models.DurationFieldType, habit.FieldType.Type)
		assert.Equal(t, models.AutomaticScoring, habit.ScoringType)

		// Check mini criteria
		require.NotNil(t, habit.MiniCriteria)
		assert.Equal(t, "Minimum exercise", habit.MiniCriteria.Description)
		require.NotNil(t, habit.MiniCriteria.Condition.GreaterThanOrEqual)
		assert.Equal(t, 15.0, *habit.MiniCriteria.Condition.GreaterThanOrEqual)

		// Check midi criteria
		require.NotNil(t, habit.MidiCriteria)
		assert.Equal(t, "Target exercise", habit.MidiCriteria.Description)
		require.NotNil(t, habit.MidiCriteria.Condition.GreaterThanOrEqual)
		assert.Equal(t, 30.0, *habit.MidiCriteria.Condition.GreaterThanOrEqual)

		// Check maxi criteria
		require.NotNil(t, habit.MaxiCriteria)
		assert.Equal(t, "Excellent exercise", habit.MaxiCriteria.Description)
		require.NotNil(t, habit.MaxiCriteria.Condition.GreaterThanOrEqual)
		assert.Equal(t, 60.0, *habit.MaxiCriteria.Condition.GreaterThanOrEqual)
	})

	t.Run("elastic habit with unsigned int criteria", func(t *testing.T) {
		yamlData := `
version: "1.0.0"
habits:
  - title: "Daily Steps"
    position: 1
    habit_type: "elastic"
    field_type:
      type: "unsigned_int"
      unit: "steps"
    scoring_type: "automatic"
    mini_criteria:
      condition:
        greater_than_or_equal: 5000
    midi_criteria:
      condition:
        greater_than_or_equal: 10000
    maxi_criteria:
      condition:
        greater_than_or_equal: 15000
    prompt: "How many steps did you take today?"
`

		schema, err := parser.ParseYAML([]byte(yamlData))
		require.NoError(t, err)
		require.NotNil(t, schema)

		habit := schema.Habits[0]
		assert.Equal(t, "Daily Steps", habit.Title)
		assert.Equal(t, models.ElasticHabit, habit.HabitType)
		assert.Equal(t, models.UnsignedIntFieldType, habit.FieldType.Type)
		assert.Equal(t, "steps", habit.FieldType.Unit)

		// Verify all criteria were parsed
		assert.NotNil(t, habit.MiniCriteria)
		assert.NotNil(t, habit.MidiCriteria)
		assert.NotNil(t, habit.MaxiCriteria)

		// Check numeric values
		assert.Equal(t, 5000.0, *habit.MiniCriteria.Condition.GreaterThanOrEqual)
		assert.Equal(t, 10000.0, *habit.MidiCriteria.Condition.GreaterThanOrEqual)
		assert.Equal(t, 15000.0, *habit.MaxiCriteria.Condition.GreaterThanOrEqual)
	})

	t.Run("elastic habit with manual scoring", func(t *testing.T) {
		yamlData := `
version: "1.0.0"
habits:
  - title: "Reading Quality"
    position: 1
    habit_type: "elastic"
    field_type:
      type: "text"
    scoring_type: "manual"
    prompt: "What did you read today and how much did you enjoy it?"
    help_text: "Describe your reading experience"
`

		schema, err := parser.ParseYAML([]byte(yamlData))
		require.NoError(t, err)
		require.NotNil(t, schema)

		habit := schema.Habits[0]
		assert.Equal(t, "Reading Quality", habit.Title)
		assert.Equal(t, models.ElasticHabit, habit.HabitType)
		assert.Equal(t, models.TextFieldType, habit.FieldType.Type)
		assert.Equal(t, models.ManualScoring, habit.ScoringType)

		// Manual scoring elastic habits don't require criteria
		assert.Nil(t, habit.MiniCriteria)
		assert.Nil(t, habit.MidiCriteria)
		assert.Nil(t, habit.MaxiCriteria)
	})

	t.Run("elastic habit missing required criteria", func(t *testing.T) {
		yamlData := `
version: "1.0.0"
habits:
  - title: "Exercise Duration"
    position: 1
    habit_type: "elastic"
    field_type:
      type: "duration"
    scoring_type: "automatic"
    # Missing mini_criteria - should fail validation
    midi_criteria:
      condition:
        greater_than_or_equal: 30
    maxi_criteria:
      condition:
        greater_than_or_equal: 60
`

		_, err := parser.ParseYAML([]byte(yamlData))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "mini_criteria is required")
	})

	t.Run("elastic habit criteria ordering validation", func(t *testing.T) {
		// Valid ordering: mini ≤ midi ≤ maxi
		yamlData := `
version: "1.0.0"
habits:
  - title: "Exercise Duration"
    position: 1
    habit_type: "elastic"
    field_type:
      type: "duration"
    scoring_type: "automatic"
    mini_criteria:
      condition:
        greater_than_or_equal: 15
    midi_criteria:
      condition:
        greater_than_or_equal: 30
    maxi_criteria:
      condition:
        greater_than_or_equal: 60
`

		schema, err := parser.ParseYAML([]byte(yamlData))
		require.NoError(t, err)
		require.NotNil(t, schema)
	})

	t.Run("elastic habit invalid criteria ordering - mini > midi", func(t *testing.T) {
		yamlData := `
version: "1.0.0"
habits:
  - title: "Exercise Duration"
    position: 1
    habit_type: "elastic"
    field_type:
      type: "duration"
    scoring_type: "automatic"
    mini_criteria:
      condition:
        greater_than_or_equal: 45  # Invalid: 45 > 30
    midi_criteria:
      condition:
        greater_than_or_equal: 30
    maxi_criteria:
      condition:
        greater_than_or_equal: 60
`

		_, err := parser.ParseYAML([]byte(yamlData))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "mini criteria value")
		assert.Contains(t, err.Error(), "must be ≤ midi criteria value")
	})

	t.Run("elastic habit invalid criteria ordering - midi > maxi", func(t *testing.T) {
		yamlData := `
version: "1.0.0"
habits:
  - title: "Daily Steps"
    position: 1
    habit_type: "elastic"
    field_type:
      type: "unsigned_int"
      unit: "steps"
    scoring_type: "automatic"
    mini_criteria:
      condition:
        greater_than_or_equal: 5000
    midi_criteria:
      condition:
        greater_than_or_equal: 15000  # Invalid: 15000 > 10000
    maxi_criteria:
      condition:
        greater_than_or_equal: 10000
`

		_, err := parser.ParseYAML([]byte(yamlData))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "midi criteria value")
		assert.Contains(t, err.Error(), "must be ≤ maxi criteria value")
	})

	t.Run("elastic habit non-numeric field type - no ordering validation", func(t *testing.T) {
		yamlData := `
version: "1.0.0"
habits:
  - title: "Reading Quality"
    position: 1
    habit_type: "elastic"
    field_type:
      type: "text"
    scoring_type: "automatic"
    mini_criteria:
      condition:
        greater_than_or_equal: 100  # Nonsensical for text, but should not fail ordering validation
    midi_criteria:
      condition:
        greater_than_or_equal: 50
    maxi_criteria:
      condition:
        greater_than_or_equal: 25
`

		schema, err := parser.ParseYAML([]byte(yamlData))
		require.NoError(t, err)
		require.NotNil(t, schema)
		// Should pass because text fields don't have ordering validation
	})
}

func TestHabitParser_LoadFromFile(t *testing.T) {
	parser := NewHabitParser()

	t.Run("load valid habits file", func(t *testing.T) {
		// Create temporary file
		tempDir := t.TempDir()
		habitsFile := filepath.Join(tempDir, "habits.yml")

		yamlContent := `
version: "1.0.0"
created_date: "2024-01-01"
habits:
  - title: "Test Habit"
    position: 1
    habit_type: "simple"
    field_type:
      type: "boolean"
    scoring_type: "manual"
`

		err := os.WriteFile(habitsFile, []byte(yamlContent), 0o600)
		require.NoError(t, err)

		// Load and parse
		schema, err := parser.LoadFromFile(habitsFile)
		require.NoError(t, err)
		require.NotNil(t, schema)

		assert.Equal(t, "1.0.0", schema.Version)
		assert.Len(t, schema.Habits, 1)
		assert.Equal(t, "Test Habit", schema.Habits[0].Title)
	})

	t.Run("file does not exist", func(t *testing.T) {
		_, err := parser.LoadFromFile("/nonexistent/habits.yml")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "habits file not found")
	})

	t.Run("file read permission error", func(t *testing.T) {
		tempDir := t.TempDir()
		habitsFile := filepath.Join(tempDir, "unreadable.yml")

		// Create file and remove read permission
		err := os.WriteFile(habitsFile, []byte("test"), 0o000)
		require.NoError(t, err)

		_, err = parser.LoadFromFile(habitsFile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read habits file")
	})
}

func TestHabitParser_SaveToFile(t *testing.T) {
	parser := NewHabitParser()

	t.Run("save valid schema", func(t *testing.T) {
		tempDir := t.TempDir()
		habitsFile := filepath.Join(tempDir, "habits.yml")

		schema := &models.Schema{
			Version:     "1.0.0",
			CreatedDate: "2024-01-01",
			Habits: []models.Habit{
				{
					Title:     "Test Habit",
					Position:  1,
					HabitType: models.SimpleHabit,
					FieldType: models.FieldType{
						Type: models.BooleanFieldType,
					},
					ScoringType: models.ManualScoring,
				},
			},
		}

		err := parser.SaveToFile(schema, habitsFile)
		require.NoError(t, err)

		// Verify file was created and can be loaded back
		loadedSchema, err := parser.LoadFromFile(habitsFile)
		require.NoError(t, err)

		assert.Equal(t, schema.Version, loadedSchema.Version)
		assert.Equal(t, schema.CreatedDate, loadedSchema.CreatedDate)
		assert.Len(t, loadedSchema.Habits, 1)
		assert.Equal(t, "Test Habit", loadedSchema.Habits[0].Title)
	})

	t.Run("save invalid schema", func(t *testing.T) {
		tempDir := t.TempDir()
		habitsFile := filepath.Join(tempDir, "habits.yml")

		// Invalid schema (missing version)
		schema := &models.Schema{
			Habits: []models.Habit{},
		}

		err := parser.SaveToFile(schema, habitsFile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot save invalid schema")
	})

	t.Run("write permission error", func(t *testing.T) {
		// Try to write to root directory (should fail)
		schema := &models.Schema{
			Version: "1.0.0",
			Habits:  []models.Habit{},
		}

		err := parser.SaveToFile(schema, "/root/habits.yml")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write habits file")
	})
}

func TestHabitParser_CreateSampleSchema(t *testing.T) {
	parser := NewHabitParser()

	schema := parser.CreateSampleSchema()
	require.NotNil(t, schema)

	// Validate the sample schema
	err := schema.Validate()
	assert.NoError(t, err)

	// Check basic properties
	assert.Equal(t, "1.0.0", schema.Version)
	assert.Equal(t, "2024-01-01", schema.CreatedDate)
	assert.Len(t, schema.Habits, 3)

	// Check that all habits are simple boolean habits
	for _, habit := range schema.Habits {
		assert.Equal(t, models.SimpleHabit, habit.HabitType)
		assert.Equal(t, models.BooleanFieldType, habit.FieldType.Type)
		assert.Equal(t, models.ManualScoring, habit.ScoringType)
		assert.NotEmpty(t, habit.Title)
		assert.NotEmpty(t, habit.Prompt)
		assert.Greater(t, habit.Position, 0)
	}

	// Verify unique positions
	positions := make(map[int]bool)
	for _, habit := range schema.Habits {
		assert.False(t, positions[habit.Position], "Duplicate position found")
		positions[habit.Position] = true
	}
}

func TestHabitParser_ValidateFile(t *testing.T) {
	parser := NewHabitParser()

	t.Run("valid file", func(t *testing.T) {
		tempDir := t.TempDir()
		habitsFile := filepath.Join(tempDir, "habits.yml")

		yamlContent := `
version: "1.0.0"
habits:
  - title: "Test Habit"
    position: 1
    habit_type: "simple"
    field_type:
      type: "boolean"
    scoring_type: "manual"
`

		err := os.WriteFile(habitsFile, []byte(yamlContent), 0o600)
		require.NoError(t, err)

		err = parser.ValidateFile(habitsFile)
		assert.NoError(t, err)
	})

	t.Run("invalid file", func(t *testing.T) {
		tempDir := t.TempDir()
		habitsFile := filepath.Join(tempDir, "habits.yml")

		yamlContent := `
version: "1.0.0"
habits:
  - title: "Test Habit"
    position: 1
    habit_type: "simple"
    field_type:
      type: "boolean"
    # Missing scoring_type - should fail validation
`

		err := os.WriteFile(habitsFile, []byte(yamlContent), 0o600)
		require.NoError(t, err)

		err = parser.ValidateFile(habitsFile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "scoring_type is required")
	})
}

func TestGetHabitByID(t *testing.T) {
	schema := &models.Schema{
		Habits: []models.Habit{
			{
				ID:    "habit1",
				Title: "Habit 1",
			},
			{
				ID:    "habit2",
				Title: "Habit 2",
			},
		},
	}

	t.Run("existing habit", func(t *testing.T) {
		habit, found := GetHabitByID(schema, "habit1")
		assert.True(t, found)
		require.NotNil(t, habit)
		assert.Equal(t, "Habit 1", habit.Title)
	})

	t.Run("non-existing habit", func(t *testing.T) {
		habit, found := GetHabitByID(schema, "nonexistent")
		assert.False(t, found)
		assert.Nil(t, habit)
	})

	t.Run("nil schema", func(t *testing.T) {
		habit, found := GetHabitByID(nil, "habit1")
		assert.False(t, found)
		assert.Nil(t, habit)
	})
}

func TestGetHabitsByType(t *testing.T) {
	schema := &models.Schema{
		Habits: []models.Habit{
			{HabitType: models.SimpleHabit, Title: "Simple 1"},
			{HabitType: models.ElasticHabit, Title: "Elastic 1"},
			{HabitType: models.SimpleHabit, Title: "Simple 2"},
			{HabitType: models.InformationalHabit, Title: "Info 1"},
		},
	}

	t.Run("get simple habits", func(t *testing.T) {
		habits := GetHabitsByType(schema, models.SimpleHabit)
		assert.Len(t, habits, 2)
		assert.Equal(t, "Simple 1", habits[0].Title)
		assert.Equal(t, "Simple 2", habits[1].Title)
	})

	t.Run("get elastic habits", func(t *testing.T) {
		habits := GetHabitsByType(schema, models.ElasticHabit)
		assert.Len(t, habits, 1)
		assert.Equal(t, "Elastic 1", habits[0].Title)
	})

	t.Run("no matching habits", func(t *testing.T) {
		// Create schema with no informational habits
		simpleSchema := &models.Schema{
			Habits: []models.Habit{
				{HabitType: models.SimpleHabit, Title: "Simple 1"},
			},
		}

		habits := GetHabitsByType(simpleSchema, models.InformationalHabit)
		assert.Empty(t, habits)
	})

	t.Run("nil schema", func(t *testing.T) {
		habits := GetHabitsByType(nil, models.SimpleHabit)
		assert.Nil(t, habits)
	})
}

func TestGetSimpleBooleanHabits(t *testing.T) {
	schema := &models.Schema{
		Habits: []models.Habit{
			{
				HabitType: models.SimpleHabit,
				FieldType: models.FieldType{Type: models.BooleanFieldType},
				Title:     "Simple Boolean 1",
			},
			{
				HabitType: models.SimpleHabit,
				FieldType: models.FieldType{Type: models.UnsignedIntFieldType},
				Title:     "Simple Numeric",
			},
			{
				HabitType: models.ElasticHabit,
				FieldType: models.FieldType{Type: models.BooleanFieldType},
				Title:     "Elastic Boolean",
			},
			{
				HabitType: models.SimpleHabit,
				FieldType: models.FieldType{Type: models.BooleanFieldType},
				Title:     "Simple Boolean 2",
			},
		},
	}

	t.Run("get simple boolean habits", func(t *testing.T) {
		habits := GetSimpleBooleanHabits(schema)
		assert.Len(t, habits, 2)
		assert.Equal(t, "Simple Boolean 1", habits[0].Title)
		assert.Equal(t, "Simple Boolean 2", habits[1].Title)
	})

	t.Run("nil schema", func(t *testing.T) {
		habits := GetSimpleBooleanHabits(nil)
		assert.Nil(t, habits)
	})

	t.Run("no simple boolean habits", func(t *testing.T) {
		emptySchema := &models.Schema{
			Habits: []models.Habit{
				{
					HabitType: models.ElasticHabit,
					FieldType: models.FieldType{Type: models.BooleanFieldType},
					Title:     "Elastic Habit",
				},
			},
		}

		habits := GetSimpleBooleanHabits(emptySchema)
		assert.Empty(t, habits)
	})
}
