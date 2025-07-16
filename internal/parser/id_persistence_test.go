package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"davidlee/vice/internal/models"
)

func TestHabitParser_LoadFromFileWithIDPersistence(t *testing.T) {
	t.Run("generates and persists missing habit IDs", func(t *testing.T) {
		tempDir := t.TempDir()
		habitsFile := filepath.Join(tempDir, "habits.yml")

		// Create a habits file without IDs
		yamlContent := `version: "1.0.0"
habits:
  - title: "Morning Exercise"
    habit_type: "simple"
    field_type:
      type: "boolean"
    scoring_type: "manual"
    prompt: "Did you exercise this morning?"
  - title: "Daily Reading"
    habit_type: "simple"
    field_type:
      type: "boolean"
    scoring_type: "manual"
    prompt: "Did you read today?"
`

		err := os.WriteFile(habitsFile, []byte(yamlContent), 0o600) //nolint:gosec // Test file in temp dir
		require.NoError(t, err)

		parser := NewHabitParser()

		// Load with ID persistence enabled
		schema, err := parser.LoadFromFileWithIDPersistence(habitsFile, true)
		require.NoError(t, err)
		require.Len(t, schema.Habits, 2)

		// Verify IDs were generated
		assert.Equal(t, "morning_exercise", schema.Habits[0].ID)
		assert.Equal(t, "daily_reading", schema.Habits[1].ID)

		// Reload the file to verify IDs were persisted
		reloadedSchema, err := parser.LoadFromFile(habitsFile)
		require.NoError(t, err)
		require.Len(t, reloadedSchema.Habits, 2)

		// Verify IDs are now present in the file
		assert.Equal(t, "morning_exercise", reloadedSchema.Habits[0].ID)
		assert.Equal(t, "daily_reading", reloadedSchema.Habits[1].ID)

		// Check file contents directly
		savedContent, err := os.ReadFile(habitsFile) //nolint:gosec // Test file in temp dir
		require.NoError(t, err)
		savedYAML := string(savedContent)
		assert.Contains(t, savedYAML, "id: morning_exercise")
		assert.Contains(t, savedYAML, "id: daily_reading")
	})

	t.Run("does not modify file when IDs already exist", func(t *testing.T) {
		tempDir := t.TempDir()
		habitsFile := filepath.Join(tempDir, "habits.yml")

		// Create a habits file with existing IDs
		yamlContent := `version: "1.0.0"
habits:
  - title: "Morning Exercise"
    id: "custom_exercise_id"
    habit_type: "simple"
    field_type:
      type: "boolean"
    scoring_type: "manual"
  - title: "Daily Reading"
    id: "custom_reading_id"
    habit_type: "simple"
    field_type:
      type: "boolean"
    scoring_type: "manual"
`

		err := os.WriteFile(habitsFile, []byte(yamlContent), 0o600) //nolint:gosec // Test file in temp dir
		require.NoError(t, err)

		// Get original file modification time
		originalInfo, err := os.Stat(habitsFile)
		require.NoError(t, err)
		originalModTime := originalInfo.ModTime()

		parser := NewHabitParser()

		// Load with ID persistence enabled
		schema, err := parser.LoadFromFileWithIDPersistence(habitsFile, true)
		require.NoError(t, err)

		// Verify original IDs are preserved
		assert.Equal(t, "custom_exercise_id", schema.Habits[0].ID)
		assert.Equal(t, "custom_reading_id", schema.Habits[1].ID)

		// Check that file was not modified (allow small time differences)
		newInfo, err := os.Stat(habitsFile)
		require.NoError(t, err)
		assert.True(t, newInfo.ModTime().Equal(originalModTime) || newInfo.ModTime().Before(originalModTime.Add(100)))
	})

	t.Run("handles read-only files gracefully", func(t *testing.T) {
		tempDir := t.TempDir()
		habitsFile := filepath.Join(tempDir, "habits.yml")

		// Create a habits file without IDs
		yamlContent := `version: "1.0.0"
habits:
  - title: "Morning Exercise"
    habit_type: "simple"
    field_type:
      type: "boolean"
    scoring_type: "manual"
`

		err := os.WriteFile(habitsFile, []byte(yamlContent), 0o600) //nolint:gosec // Test file in temp dir
		require.NoError(t, err)

		// Make file read-only
		err = os.Chmod(habitsFile, 0o444) //nolint:gosec // Test needs read-only file
		require.NoError(t, err)

		parser := NewHabitParser()

		// Load should succeed despite being unable to persist IDs
		schema, err := parser.LoadFromFileWithIDPersistence(habitsFile, true)
		require.NoError(t, err)
		require.Len(t, schema.Habits, 1)

		// ID should still be generated in memory
		assert.Equal(t, "morning_exercise", schema.Habits[0].ID)

		// Restore write permissions for cleanup
		err = os.Chmod(habitsFile, 0o644) //nolint:gosec // Test cleanup
		require.NoError(t, err)
	})

	t.Run("persistence disabled works as before", func(t *testing.T) {
		tempDir := t.TempDir()
		habitsFile := filepath.Join(tempDir, "habits.yml")

		// Create a habits file without IDs
		yamlContent := `version: "1.0.0"
habits:
  - title: "Morning Exercise"
    habit_type: "simple"
    field_type:
      type: "boolean"
    scoring_type: "manual"
`

		err := os.WriteFile(habitsFile, []byte(yamlContent), 0o600) //nolint:gosec // Test file in temp dir
		require.NoError(t, err)

		originalContent, err := os.ReadFile(habitsFile) //nolint:gosec // Test file in temp dir
		require.NoError(t, err)

		parser := NewHabitParser()

		// Load with ID persistence disabled
		schema, err := parser.LoadFromFileWithIDPersistence(habitsFile, false)
		require.NoError(t, err)

		// ID should be generated in memory
		assert.Equal(t, "morning_exercise", schema.Habits[0].ID)

		// File should not be modified
		newContent, err := os.ReadFile(habitsFile) //nolint:gosec // Test file in temp dir
		require.NoError(t, err)
		assert.Equal(t, string(originalContent), string(newContent))
	})

	t.Run("mixed scenarios - some habits have IDs, some don't", func(t *testing.T) {
		tempDir := t.TempDir()
		habitsFile := filepath.Join(tempDir, "habits.yml")

		// Create habits file with mixed ID presence
		yamlContent := `version: "1.0.0"
habits:
  - title: "Morning Exercise"
    id: "existing_exercise_id"
    habit_type: "simple"
    field_type:
      type: "boolean"
    scoring_type: "manual"
  - title: "Daily Reading"
    habit_type: "simple"
    field_type:
      type: "boolean"
    scoring_type: "manual"
  - title: "Water Intake"
    habit_type: "elastic"
    field_type:
      type: "unsigned_int"
      unit: "glasses"
    scoring_type: "manual"
`

		err := os.WriteFile(habitsFile, []byte(yamlContent), 0o600) //nolint:gosec // Test file in temp dir
		require.NoError(t, err)

		parser := NewHabitParser()

		// Load with ID persistence enabled
		schema, err := parser.LoadFromFileWithIDPersistence(habitsFile, true)
		require.NoError(t, err)
		require.Len(t, schema.Habits, 3)

		// Verify existing ID is preserved and new IDs are generated
		assert.Equal(t, "existing_exercise_id", schema.Habits[0].ID)
		assert.Equal(t, "daily_reading", schema.Habits[1].ID)
		assert.Equal(t, "water_intake", schema.Habits[2].ID)

		// Reload to verify persistence
		reloadedSchema, err := parser.LoadFromFile(habitsFile)
		require.NoError(t, err)

		assert.Equal(t, "existing_exercise_id", reloadedSchema.Habits[0].ID)
		assert.Equal(t, "daily_reading", reloadedSchema.Habits[1].ID)
		assert.Equal(t, "water_intake", reloadedSchema.Habits[2].ID)
	})
}

func TestSchema_ValidateAndTrackChanges(t *testing.T) {
	t.Run("tracks when habit IDs are generated", func(t *testing.T) {
		schema := &models.Schema{
			Version: "1.0.0",
			Habits: []models.Habit{
				{
					Title:       "Test Habit",
					HabitType:   models.SimpleHabit,
					FieldType:   models.FieldType{Type: models.BooleanFieldType},
					ScoringType: models.ManualScoring,
				},
			},
		}

		wasModified, err := schema.ValidateAndTrackChanges()
		require.NoError(t, err)
		assert.True(t, wasModified, "should detect that ID was generated")
		assert.Equal(t, "test_habit", schema.Habits[0].ID)
	})

	t.Run("does not track changes when IDs already exist", func(t *testing.T) {
		schema := &models.Schema{
			Version: "1.0.0",
			Habits: []models.Habit{
				{
					Title:       "Test Habit",
					ID:          "existing_id",
					HabitType:   models.SimpleHabit,
					FieldType:   models.FieldType{Type: models.BooleanFieldType},
					ScoringType: models.ManualScoring,
				},
			},
		}

		wasModified, err := schema.ValidateAndTrackChanges()
		require.NoError(t, err)
		assert.False(t, wasModified, "should not detect changes when ID exists")
		assert.Equal(t, "existing_id", schema.Habits[0].ID)
	})

	t.Run("tracks partial modifications", func(t *testing.T) {
		schema := &models.Schema{
			Version: "1.0.0",
			Habits: []models.Habit{
				{
					Title:       "Habit 1",
					ID:          "existing_id",
					HabitType:   models.SimpleHabit,
					FieldType:   models.FieldType{Type: models.BooleanFieldType},
					ScoringType: models.ManualScoring,
				},
				{
					Title:       "Habit 2",
					HabitType:   models.SimpleHabit,
					FieldType:   models.FieldType{Type: models.BooleanFieldType},
					ScoringType: models.ManualScoring,
				},
			},
		}

		wasModified, err := schema.ValidateAndTrackChanges()
		require.NoError(t, err)
		assert.True(t, wasModified, "should detect that one ID was generated")
		assert.Equal(t, "existing_id", schema.Habits[0].ID)
		assert.Equal(t, "habit_2", schema.Habits[1].ID)
	})
}

func TestHabit_ValidateAndTrackChanges(t *testing.T) {
	t.Run("tracks ID generation", func(t *testing.T) {
		habit := &models.Habit{
			Title:       "Test Habit",
			HabitType:   models.SimpleHabit,
			FieldType:   models.FieldType{Type: models.BooleanFieldType},
			ScoringType: models.ManualScoring,
		}

		wasModified, err := habit.ValidateAndTrackChanges()
		require.NoError(t, err)
		assert.True(t, wasModified)
		assert.Equal(t, "test_habit", habit.ID)
	})

	t.Run("does not track when ID exists", func(t *testing.T) {
		habit := &models.Habit{
			Title:       "Test Habit",
			ID:          "existing_id",
			HabitType:   models.SimpleHabit,
			FieldType:   models.FieldType{Type: models.BooleanFieldType},
			ScoringType: models.ManualScoring,
		}

		wasModified, err := habit.ValidateAndTrackChanges()
		require.NoError(t, err)
		assert.False(t, wasModified)
		assert.Equal(t, "existing_id", habit.ID)
	})

	t.Run("validation errors still occur", func(t *testing.T) {
		habit := &models.Habit{
			Title: "", // Invalid - title required
		}

		wasModified, err := habit.ValidateAndTrackChanges()
		require.Error(t, err)
		assert.False(t, wasModified)
		assert.Contains(t, err.Error(), "habit title is required")
	})
}
