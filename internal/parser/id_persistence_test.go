package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"davidlee/iter/internal/models"
)

func TestGoalParser_LoadFromFileWithIDPersistence(t *testing.T) {
	t.Run("generates and persists missing goal IDs", func(t *testing.T) {
		tempDir := t.TempDir()
		goalsFile := filepath.Join(tempDir, "goals.yml")

		// Create a goals file without IDs
		yamlContent := `version: "1.0.0"
goals:
  - title: "Morning Exercise"
    goal_type: "simple"
    field_type:
      type: "boolean"
    scoring_type: "manual"
    prompt: "Did you exercise this morning?"
  - title: "Daily Reading"
    goal_type: "simple"
    field_type:
      type: "boolean"
    scoring_type: "manual"
    prompt: "Did you read today?"
`

		err := os.WriteFile(goalsFile, []byte(yamlContent), 0o600) //nolint:gosec // Test file in temp dir
		require.NoError(t, err)

		parser := NewGoalParser()

		// Load with ID persistence enabled
		schema, err := parser.LoadFromFileWithIDPersistence(goalsFile, true)
		require.NoError(t, err)
		require.Len(t, schema.Goals, 2)

		// Verify IDs were generated
		assert.Equal(t, "morning_exercise", schema.Goals[0].ID)
		assert.Equal(t, "daily_reading", schema.Goals[1].ID)

		// Reload the file to verify IDs were persisted
		reloadedSchema, err := parser.LoadFromFile(goalsFile)
		require.NoError(t, err)
		require.Len(t, reloadedSchema.Goals, 2)

		// Verify IDs are now present in the file
		assert.Equal(t, "morning_exercise", reloadedSchema.Goals[0].ID)
		assert.Equal(t, "daily_reading", reloadedSchema.Goals[1].ID)

		// Check file contents directly
		savedContent, err := os.ReadFile(goalsFile) //nolint:gosec // Test file in temp dir
		require.NoError(t, err)
		savedYAML := string(savedContent)
		assert.Contains(t, savedYAML, "id: morning_exercise")
		assert.Contains(t, savedYAML, "id: daily_reading")
	})

	t.Run("does not modify file when IDs already exist", func(t *testing.T) {
		tempDir := t.TempDir()
		goalsFile := filepath.Join(tempDir, "goals.yml")

		// Create a goals file with existing IDs
		yamlContent := `version: "1.0.0"
goals:
  - title: "Morning Exercise"
    id: "custom_exercise_id"
    goal_type: "simple"
    field_type:
      type: "boolean"
    scoring_type: "manual"
  - title: "Daily Reading"
    id: "custom_reading_id"
    goal_type: "simple"
    field_type:
      type: "boolean"
    scoring_type: "manual"
`

		err := os.WriteFile(goalsFile, []byte(yamlContent), 0o600) //nolint:gosec // Test file in temp dir
		require.NoError(t, err)

		// Get original file modification time
		originalInfo, err := os.Stat(goalsFile)
		require.NoError(t, err)
		originalModTime := originalInfo.ModTime()

		parser := NewGoalParser()

		// Load with ID persistence enabled
		schema, err := parser.LoadFromFileWithIDPersistence(goalsFile, true)
		require.NoError(t, err)

		// Verify original IDs are preserved
		assert.Equal(t, "custom_exercise_id", schema.Goals[0].ID)
		assert.Equal(t, "custom_reading_id", schema.Goals[1].ID)

		// Check that file was not modified (allow small time differences)
		newInfo, err := os.Stat(goalsFile)
		require.NoError(t, err)
		assert.True(t, newInfo.ModTime().Equal(originalModTime) || newInfo.ModTime().Before(originalModTime.Add(100)))
	})

	t.Run("handles read-only files gracefully", func(t *testing.T) {
		tempDir := t.TempDir()
		goalsFile := filepath.Join(tempDir, "goals.yml")

		// Create a goals file without IDs
		yamlContent := `version: "1.0.0"
goals:
  - title: "Morning Exercise"
    goal_type: "simple"
    field_type:
      type: "boolean"
    scoring_type: "manual"
`

		err := os.WriteFile(goalsFile, []byte(yamlContent), 0o600) //nolint:gosec // Test file in temp dir
		require.NoError(t, err)

		// Make file read-only
		err = os.Chmod(goalsFile, 0o444) //nolint:gosec // Test needs read-only file
		require.NoError(t, err)

		parser := NewGoalParser()

		// Load should succeed despite being unable to persist IDs
		schema, err := parser.LoadFromFileWithIDPersistence(goalsFile, true)
		require.NoError(t, err)
		require.Len(t, schema.Goals, 1)

		// ID should still be generated in memory
		assert.Equal(t, "morning_exercise", schema.Goals[0].ID)

		// Restore write permissions for cleanup
		err = os.Chmod(goalsFile, 0o644) //nolint:gosec // Test cleanup
		require.NoError(t, err)
	})

	t.Run("persistence disabled works as before", func(t *testing.T) {
		tempDir := t.TempDir()
		goalsFile := filepath.Join(tempDir, "goals.yml")

		// Create a goals file without IDs
		yamlContent := `version: "1.0.0"
goals:
  - title: "Morning Exercise"
    goal_type: "simple"
    field_type:
      type: "boolean"
    scoring_type: "manual"
`

		err := os.WriteFile(goalsFile, []byte(yamlContent), 0o600) //nolint:gosec // Test file in temp dir
		require.NoError(t, err)

		originalContent, err := os.ReadFile(goalsFile) //nolint:gosec // Test file in temp dir
		require.NoError(t, err)

		parser := NewGoalParser()

		// Load with ID persistence disabled
		schema, err := parser.LoadFromFileWithIDPersistence(goalsFile, false)
		require.NoError(t, err)

		// ID should be generated in memory
		assert.Equal(t, "morning_exercise", schema.Goals[0].ID)

		// File should not be modified
		newContent, err := os.ReadFile(goalsFile) //nolint:gosec // Test file in temp dir
		require.NoError(t, err)
		assert.Equal(t, string(originalContent), string(newContent))
	})

	t.Run("mixed scenarios - some goals have IDs, some don't", func(t *testing.T) {
		tempDir := t.TempDir()
		goalsFile := filepath.Join(tempDir, "goals.yml")

		// Create goals file with mixed ID presence
		yamlContent := `version: "1.0.0"
goals:
  - title: "Morning Exercise"
    id: "existing_exercise_id"
    goal_type: "simple"
    field_type:
      type: "boolean"
    scoring_type: "manual"
  - title: "Daily Reading"
    goal_type: "simple"
    field_type:
      type: "boolean"
    scoring_type: "manual"
  - title: "Water Intake"
    goal_type: "elastic"
    field_type:
      type: "unsigned_int"
      unit: "glasses"
    scoring_type: "manual"
`

		err := os.WriteFile(goalsFile, []byte(yamlContent), 0o600) //nolint:gosec // Test file in temp dir
		require.NoError(t, err)

		parser := NewGoalParser()

		// Load with ID persistence enabled
		schema, err := parser.LoadFromFileWithIDPersistence(goalsFile, true)
		require.NoError(t, err)
		require.Len(t, schema.Goals, 3)

		// Verify existing ID is preserved and new IDs are generated
		assert.Equal(t, "existing_exercise_id", schema.Goals[0].ID)
		assert.Equal(t, "daily_reading", schema.Goals[1].ID)
		assert.Equal(t, "water_intake", schema.Goals[2].ID)

		// Reload to verify persistence
		reloadedSchema, err := parser.LoadFromFile(goalsFile)
		require.NoError(t, err)

		assert.Equal(t, "existing_exercise_id", reloadedSchema.Goals[0].ID)
		assert.Equal(t, "daily_reading", reloadedSchema.Goals[1].ID)
		assert.Equal(t, "water_intake", reloadedSchema.Goals[2].ID)
	})
}

func TestSchema_ValidateAndTrackChanges(t *testing.T) {
	t.Run("tracks when goal IDs are generated", func(t *testing.T) {
		schema := &models.Schema{
			Version: "1.0.0",
			Goals: []models.Goal{
				{
					Title:       "Test Goal",
					GoalType:    models.SimpleGoal,
					FieldType:   models.FieldType{Type: models.BooleanFieldType},
					ScoringType: models.ManualScoring,
				},
			},
		}

		wasModified, err := schema.ValidateAndTrackChanges()
		require.NoError(t, err)
		assert.True(t, wasModified, "should detect that ID was generated")
		assert.Equal(t, "test_goal", schema.Goals[0].ID)
	})

	t.Run("does not track changes when IDs already exist", func(t *testing.T) {
		schema := &models.Schema{
			Version: "1.0.0",
			Goals: []models.Goal{
				{
					Title:       "Test Goal",
					ID:          "existing_id",
					GoalType:    models.SimpleGoal,
					FieldType:   models.FieldType{Type: models.BooleanFieldType},
					ScoringType: models.ManualScoring,
				},
			},
		}

		wasModified, err := schema.ValidateAndTrackChanges()
		require.NoError(t, err)
		assert.False(t, wasModified, "should not detect changes when ID exists")
		assert.Equal(t, "existing_id", schema.Goals[0].ID)
	})

	t.Run("tracks partial modifications", func(t *testing.T) {
		schema := &models.Schema{
			Version: "1.0.0",
			Goals: []models.Goal{
				{
					Title:       "Goal 1",
					ID:          "existing_id",
					GoalType:    models.SimpleGoal,
					FieldType:   models.FieldType{Type: models.BooleanFieldType},
					ScoringType: models.ManualScoring,
				},
				{
					Title:       "Goal 2",
					GoalType:    models.SimpleGoal,
					FieldType:   models.FieldType{Type: models.BooleanFieldType},
					ScoringType: models.ManualScoring,
				},
			},
		}

		wasModified, err := schema.ValidateAndTrackChanges()
		require.NoError(t, err)
		assert.True(t, wasModified, "should detect that one ID was generated")
		assert.Equal(t, "existing_id", schema.Goals[0].ID)
		assert.Equal(t, "goal_2", schema.Goals[1].ID)
	})
}

func TestGoal_ValidateAndTrackChanges(t *testing.T) {
	t.Run("tracks ID generation", func(t *testing.T) {
		goal := &models.Goal{
			Title:       "Test Goal",
			GoalType:    models.SimpleGoal,
			FieldType:   models.FieldType{Type: models.BooleanFieldType},
			ScoringType: models.ManualScoring,
		}

		wasModified, err := goal.ValidateAndTrackChanges()
		require.NoError(t, err)
		assert.True(t, wasModified)
		assert.Equal(t, "test_goal", goal.ID)
	})

	t.Run("does not track when ID exists", func(t *testing.T) {
		goal := &models.Goal{
			Title:       "Test Goal",
			ID:          "existing_id",
			GoalType:    models.SimpleGoal,
			FieldType:   models.FieldType{Type: models.BooleanFieldType},
			ScoringType: models.ManualScoring,
		}

		wasModified, err := goal.ValidateAndTrackChanges()
		require.NoError(t, err)
		assert.False(t, wasModified)
		assert.Equal(t, "existing_id", goal.ID)
	})

	t.Run("validation errors still occur", func(t *testing.T) {
		goal := &models.Goal{
			Title: "", // Invalid - title required
		}

		wasModified, err := goal.ValidateAndTrackChanges()
		require.Error(t, err)
		assert.False(t, wasModified)
		assert.Contains(t, err.Error(), "goal title is required")
	})
}
