package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"davidlee/iter/internal/models"
)

func TestGoalParser_ParseYAML(t *testing.T) {
	parser := NewGoalParser()

	t.Run("valid simple boolean goals schema", func(t *testing.T) {
		yamlData := `
version: "1.0.0"
created_date: "2024-01-01"
goals:
  - title: "Morning Meditation"
    position: 1
    goal_type: "simple"
    field_type:
      type: "boolean"
    scoring_type: "manual"
    prompt: "Did you meditate this morning?"
  - title: "Daily Exercise"
    position: 2
    goal_type: "simple"
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
		assert.Len(t, schema.Goals, 2)

		// Check first goal
		goal1 := schema.Goals[0]
		assert.Equal(t, "Morning Meditation", goal1.Title)
		assert.Equal(t, "morning_meditation", goal1.ID) // Auto-generated
		assert.Equal(t, 1, goal1.Position)
		assert.Equal(t, models.SimpleGoal, goal1.GoalType)
		assert.Equal(t, models.BooleanFieldType, goal1.FieldType.Type)
		assert.Equal(t, models.ManualScoring, goal1.ScoringType)
		assert.Equal(t, "Did you meditate this morning?", goal1.Prompt)

		// Check second goal
		goal2 := schema.Goals[1]
		assert.Equal(t, "Daily Exercise", goal2.Title)
		assert.Equal(t, "daily_exercise", goal2.ID)
		assert.Equal(t, models.AutomaticScoring, goal2.ScoringType)
		require.NotNil(t, goal2.Criteria)
		assert.Equal(t, "Exercise completed", goal2.Criteria.Description)
		require.NotNil(t, goal2.Criteria.Condition.Equals)
		assert.True(t, *goal2.Criteria.Condition.Equals)
	})

	t.Run("schema with custom IDs", func(t *testing.T) {
		yamlData := `
version: "1.0.0"
goals:
  - title: "Morning Meditation"
    id: "custom_meditation"
    position: 1
    goal_type: "simple"
    field_type:
      type: "boolean"
    scoring_type: "manual"
`

		schema, err := parser.ParseYAML([]byte(yamlData))
		require.NoError(t, err)

		assert.Equal(t, "custom_meditation", schema.Goals[0].ID)
	})

	t.Run("invalid YAML syntax", func(t *testing.T) {
		yamlData := `
version: "1.0.0"
goals:
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
goals: []
`

		_, err := parser.ParseYAML([]byte(yamlData))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse YAML")
	})

	t.Run("schema validation failure", func(t *testing.T) {
		yamlData := `
version: "1.0.0"
goals:
  - title: "Test Goal"
    position: 1
    goal_type: "simple"
    field_type:
      type: "boolean"
    # Missing scoring_type for simple goal
`

		_, err := parser.ParseYAML([]byte(yamlData))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "schema validation failed")
		assert.Contains(t, err.Error(), "scoring_type is required")
	})

	t.Run("duplicate goal positions", func(t *testing.T) {
		yamlData := `
version: "1.0.0"
goals:
  - title: "Goal 1"
    position: 1
    goal_type: "simple"
    field_type:
      type: "boolean"
    scoring_type: "manual"
  - title: "Goal 2"
    position: 1  # Duplicate position
    goal_type: "simple"
    field_type:
      type: "boolean"
    scoring_type: "manual"
`

		_, err := parser.ParseYAML([]byte(yamlData))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate goal position")
	})
}

func TestGoalParser_LoadFromFile(t *testing.T) {
	parser := NewGoalParser()

	t.Run("load valid goals file", func(t *testing.T) {
		// Create temporary file
		tempDir := t.TempDir()
		goalsFile := filepath.Join(tempDir, "goals.yml")

		yamlContent := `
version: "1.0.0"
created_date: "2024-01-01"
goals:
  - title: "Test Goal"
    position: 1
    goal_type: "simple"
    field_type:
      type: "boolean"
    scoring_type: "manual"
`

		err := os.WriteFile(goalsFile, []byte(yamlContent), 0o600)
		require.NoError(t, err)

		// Load and parse
		schema, err := parser.LoadFromFile(goalsFile)
		require.NoError(t, err)
		require.NotNil(t, schema)

		assert.Equal(t, "1.0.0", schema.Version)
		assert.Len(t, schema.Goals, 1)
		assert.Equal(t, "Test Goal", schema.Goals[0].Title)
	})

	t.Run("file does not exist", func(t *testing.T) {
		_, err := parser.LoadFromFile("/nonexistent/goals.yml")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "goals file not found")
	})

	t.Run("file read permission error", func(t *testing.T) {
		tempDir := t.TempDir()
		goalsFile := filepath.Join(tempDir, "unreadable.yml")

		// Create file and remove read permission
		err := os.WriteFile(goalsFile, []byte("test"), 0o000)
		require.NoError(t, err)

		_, err = parser.LoadFromFile(goalsFile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read goals file")
	})
}

func TestGoalParser_SaveToFile(t *testing.T) {
	parser := NewGoalParser()

	t.Run("save valid schema", func(t *testing.T) {
		tempDir := t.TempDir()
		goalsFile := filepath.Join(tempDir, "goals.yml")

		schema := &models.Schema{
			Version:     "1.0.0",
			CreatedDate: "2024-01-01",
			Goals: []models.Goal{
				{
					Title:    "Test Goal",
					Position: 1,
					GoalType: models.SimpleGoal,
					FieldType: models.FieldType{
						Type: models.BooleanFieldType,
					},
					ScoringType: models.ManualScoring,
				},
			},
		}

		err := parser.SaveToFile(schema, goalsFile)
		require.NoError(t, err)

		// Verify file was created and can be loaded back
		loadedSchema, err := parser.LoadFromFile(goalsFile)
		require.NoError(t, err)

		assert.Equal(t, schema.Version, loadedSchema.Version)
		assert.Equal(t, schema.CreatedDate, loadedSchema.CreatedDate)
		assert.Len(t, loadedSchema.Goals, 1)
		assert.Equal(t, "Test Goal", loadedSchema.Goals[0].Title)
	})

	t.Run("save invalid schema", func(t *testing.T) {
		tempDir := t.TempDir()
		goalsFile := filepath.Join(tempDir, "goals.yml")

		// Invalid schema (missing version)
		schema := &models.Schema{
			Goals: []models.Goal{},
		}

		err := parser.SaveToFile(schema, goalsFile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot save invalid schema")
	})

	t.Run("write permission error", func(t *testing.T) {
		// Try to write to root directory (should fail)
		schema := &models.Schema{
			Version: "1.0.0",
			Goals:   []models.Goal{},
		}

		err := parser.SaveToFile(schema, "/root/goals.yml")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write goals file")
	})
}

func TestGoalParser_CreateSampleSchema(t *testing.T) {
	parser := NewGoalParser()

	schema := parser.CreateSampleSchema()
	require.NotNil(t, schema)

	// Validate the sample schema
	err := schema.Validate()
	assert.NoError(t, err)

	// Check basic properties
	assert.Equal(t, "1.0.0", schema.Version)
	assert.Equal(t, "2024-01-01", schema.CreatedDate)
	assert.Len(t, schema.Goals, 3)

	// Check that all goals are simple boolean goals
	for _, goal := range schema.Goals {
		assert.Equal(t, models.SimpleGoal, goal.GoalType)
		assert.Equal(t, models.BooleanFieldType, goal.FieldType.Type)
		assert.Equal(t, models.ManualScoring, goal.ScoringType)
		assert.NotEmpty(t, goal.Title)
		assert.NotEmpty(t, goal.Prompt)
		assert.Greater(t, goal.Position, 0)
	}

	// Verify unique positions
	positions := make(map[int]bool)
	for _, goal := range schema.Goals {
		assert.False(t, positions[goal.Position], "Duplicate position found")
		positions[goal.Position] = true
	}
}

func TestGoalParser_ValidateFile(t *testing.T) {
	parser := NewGoalParser()

	t.Run("valid file", func(t *testing.T) {
		tempDir := t.TempDir()
		goalsFile := filepath.Join(tempDir, "goals.yml")

		yamlContent := `
version: "1.0.0"
goals:
  - title: "Test Goal"
    position: 1
    goal_type: "simple"
    field_type:
      type: "boolean"
    scoring_type: "manual"
`

		err := os.WriteFile(goalsFile, []byte(yamlContent), 0o600)
		require.NoError(t, err)

		err = parser.ValidateFile(goalsFile)
		assert.NoError(t, err)
	})

	t.Run("invalid file", func(t *testing.T) {
		tempDir := t.TempDir()
		goalsFile := filepath.Join(tempDir, "goals.yml")

		yamlContent := `
version: "1.0.0"
goals:
  - title: "Test Goal"
    position: 0  # Invalid position
    goal_type: "simple"
    field_type:
      type: "boolean"
    scoring_type: "manual"
`

		err := os.WriteFile(goalsFile, []byte(yamlContent), 0o600)
		require.NoError(t, err)

		err = parser.ValidateFile(goalsFile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "goal position must be positive")
	})
}

func TestGetGoalByID(t *testing.T) {
	schema := &models.Schema{
		Goals: []models.Goal{
			{
				ID:    "goal1",
				Title: "Goal 1",
			},
			{
				ID:    "goal2",
				Title: "Goal 2",
			},
		},
	}

	t.Run("existing goal", func(t *testing.T) {
		goal, found := GetGoalByID(schema, "goal1")
		assert.True(t, found)
		require.NotNil(t, goal)
		assert.Equal(t, "Goal 1", goal.Title)
	})

	t.Run("non-existing goal", func(t *testing.T) {
		goal, found := GetGoalByID(schema, "nonexistent")
		assert.False(t, found)
		assert.Nil(t, goal)
	})

	t.Run("nil schema", func(t *testing.T) {
		goal, found := GetGoalByID(nil, "goal1")
		assert.False(t, found)
		assert.Nil(t, goal)
	})
}

func TestGetGoalsByType(t *testing.T) {
	schema := &models.Schema{
		Goals: []models.Goal{
			{GoalType: models.SimpleGoal, Title: "Simple 1"},
			{GoalType: models.ElasticGoal, Title: "Elastic 1"},
			{GoalType: models.SimpleGoal, Title: "Simple 2"},
			{GoalType: models.InformationalGoal, Title: "Info 1"},
		},
	}

	t.Run("get simple goals", func(t *testing.T) {
		goals := GetGoalsByType(schema, models.SimpleGoal)
		assert.Len(t, goals, 2)
		assert.Equal(t, "Simple 1", goals[0].Title)
		assert.Equal(t, "Simple 2", goals[1].Title)
	})

	t.Run("get elastic goals", func(t *testing.T) {
		goals := GetGoalsByType(schema, models.ElasticGoal)
		assert.Len(t, goals, 1)
		assert.Equal(t, "Elastic 1", goals[0].Title)
	})

	t.Run("no matching goals", func(t *testing.T) {
		// Create schema with no informational goals
		simpleSchema := &models.Schema{
			Goals: []models.Goal{
				{GoalType: models.SimpleGoal, Title: "Simple 1"},
			},
		}

		goals := GetGoalsByType(simpleSchema, models.InformationalGoal)
		assert.Empty(t, goals)
	})

	t.Run("nil schema", func(t *testing.T) {
		goals := GetGoalsByType(nil, models.SimpleGoal)
		assert.Nil(t, goals)
	})
}

func TestGetSimpleBooleanGoals(t *testing.T) {
	schema := &models.Schema{
		Goals: []models.Goal{
			{
				GoalType:  models.SimpleGoal,
				FieldType: models.FieldType{Type: models.BooleanFieldType},
				Title:     "Simple Boolean 1",
			},
			{
				GoalType:  models.SimpleGoal,
				FieldType: models.FieldType{Type: models.UnsignedIntFieldType},
				Title:     "Simple Numeric",
			},
			{
				GoalType:  models.ElasticGoal,
				FieldType: models.FieldType{Type: models.BooleanFieldType},
				Title:     "Elastic Boolean",
			},
			{
				GoalType:  models.SimpleGoal,
				FieldType: models.FieldType{Type: models.BooleanFieldType},
				Title:     "Simple Boolean 2",
			},
		},
	}

	t.Run("get simple boolean goals", func(t *testing.T) {
		goals := GetSimpleBooleanGoals(schema)
		assert.Len(t, goals, 2)
		assert.Equal(t, "Simple Boolean 1", goals[0].Title)
		assert.Equal(t, "Simple Boolean 2", goals[1].Title)
	})

	t.Run("nil schema", func(t *testing.T) {
		goals := GetSimpleBooleanGoals(nil)
		assert.Nil(t, goals)
	})

	t.Run("no simple boolean goals", func(t *testing.T) {
		emptySchema := &models.Schema{
			Goals: []models.Goal{
				{
					GoalType:  models.ElasticGoal,
					FieldType: models.FieldType{Type: models.BooleanFieldType},
					Title:     "Elastic Goal",
				},
			},
		}

		goals := GetSimpleBooleanGoals(emptySchema)
		assert.Empty(t, goals)
	})
}
