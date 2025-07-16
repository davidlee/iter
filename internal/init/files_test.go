package init

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"davidlee/vice/internal/models"
	"davidlee/vice/internal/parser"
	"davidlee/vice/internal/storage"
)

func TestNewFileInitializer(t *testing.T) {
	initializer := NewFileInitializer()
	require.NotNil(t, initializer)
	assert.NotNil(t, initializer.goalParser)
	assert.NotNil(t, initializer.entryStorage)
}

func TestFileInitializer_EnsureConfigFiles(t *testing.T) {
	t.Run("creates both files when missing", func(t *testing.T) {
		tempDir := t.TempDir()
		habitsFile := filepath.Join(tempDir, "habits.yml")
		entriesFile := filepath.Join(tempDir, "entries.yml")

		initializer := NewFileInitializer()
		err := initializer.EnsureConfigFiles(habitsFile, entriesFile)
		require.NoError(t, err)

		// Verify habits file was created and is valid
		assert.FileExists(t, habitsFile)
		goalParser := parser.NewHabitParser()
		schema, err := goalParser.LoadFromFile(habitsFile)
		require.NoError(t, err)
		assert.Equal(t, "1.0.0", schema.Version)
		assert.Len(t, schema.Habits, 4)

		// Verify entries file was created and is valid
		assert.FileExists(t, entriesFile)
		entryStorage := storage.NewEntryStorage()
		entryLog, err := entryStorage.LoadFromFile(entriesFile)
		require.NoError(t, err)
		assert.Equal(t, "1.0.0", entryLog.Version)
		assert.Empty(t, entryLog.Entries)
	})

	t.Run("skips creation when files exist", func(t *testing.T) {
		tempDir := t.TempDir()
		habitsFile := filepath.Join(tempDir, "habits.yml")
		entriesFile := filepath.Join(tempDir, "entries.yml")

		// Create existing files
		err := os.WriteFile(habitsFile, []byte("existing habits"), 0o600)
		require.NoError(t, err)
		err = os.WriteFile(entriesFile, []byte("existing entries"), 0o600)
		require.NoError(t, err)

		initializer := NewFileInitializer()
		err = initializer.EnsureConfigFiles(habitsFile, entriesFile)
		require.NoError(t, err)

		// Verify files weren't overwritten
		habitsContent, err := os.ReadFile(habitsFile) //nolint:gosec // Test files in temp directory
		require.NoError(t, err)
		assert.Equal(t, "existing habits", string(habitsContent))

		entriesContent, err := os.ReadFile(entriesFile) //nolint:gosec // Test files in temp directory
		require.NoError(t, err)
		assert.Equal(t, "existing entries", string(entriesContent))
	})

	t.Run("creates only missing file", func(t *testing.T) {
		tempDir := t.TempDir()
		habitsFile := filepath.Join(tempDir, "habits.yml")
		entriesFile := filepath.Join(tempDir, "entries.yml")

		// Create only habits file
		err := os.WriteFile(habitsFile, []byte("existing habits"), 0o600)
		require.NoError(t, err)

		initializer := NewFileInitializer()
		err = initializer.EnsureConfigFiles(habitsFile, entriesFile)
		require.NoError(t, err)

		// Verify habits file wasn't overwritten
		habitsContent, err := os.ReadFile(habitsFile) //nolint:gosec // Test files in temp directory
		require.NoError(t, err)
		assert.Equal(t, "existing habits", string(habitsContent))

		// Verify entries file was created
		assert.FileExists(t, entriesFile)
		entryStorage := storage.NewEntryStorage()
		entryLog, err := entryStorage.LoadFromFile(entriesFile)
		require.NoError(t, err)
		assert.Equal(t, "1.0.0", entryLog.Version)
	})

	t.Run("creates config directory if missing", func(t *testing.T) {
		tempDir := t.TempDir()
		configDir := filepath.Join(tempDir, "nested", "config")
		habitsFile := filepath.Join(configDir, "habits.yml")
		entriesFile := filepath.Join(configDir, "entries.yml")

		initializer := NewFileInitializer()
		err := initializer.EnsureConfigFiles(habitsFile, entriesFile)
		require.NoError(t, err)

		// Verify directory was created
		assert.DirExists(t, configDir)
		assert.FileExists(t, habitsFile)
		assert.FileExists(t, entriesFile)
	})
}

func TestFileInitializer_createSampleHabitsFile(t *testing.T) {
	tempDir := t.TempDir()
	habitsFile := filepath.Join(tempDir, "habits.yml")

	initializer := NewFileInitializer()
	err := initializer.createSampleHabitsFile(habitsFile)
	require.NoError(t, err)

	// Load and validate the created file
	goalParser := parser.NewHabitParser()
	schema, err := goalParser.LoadFromFile(habitsFile)
	require.NoError(t, err)

	assert.Equal(t, "1.0.0", schema.Version)
	assert.Len(t, schema.Habits, 4)

	// Check first habit (simple boolean)
	goal1 := schema.Habits[0]
	assert.Equal(t, "Morning Exercise", goal1.Title)
	assert.Equal(t, "morning_exercise", goal1.ID)
	assert.Equal(t, 1, goal1.Position)
	assert.Equal(t, models.SimpleHabit, goal1.HabitType)
	assert.Equal(t, models.BooleanFieldType, goal1.FieldType.Type)
	assert.Equal(t, models.ManualScoring, goal1.ScoringType)
	assert.NotEmpty(t, goal1.Description)
	assert.NotEmpty(t, goal1.Prompt)
	assert.NotEmpty(t, goal1.HelpText)

	// Check second habit (simple boolean)
	goal2 := schema.Habits[1]
	assert.Equal(t, "Daily Reading", goal2.Title)
	assert.Equal(t, "daily_reading", goal2.ID)
	assert.Equal(t, 2, goal2.Position)
	assert.Equal(t, models.SimpleHabit, goal2.HabitType)
	assert.Equal(t, models.BooleanFieldType, goal2.FieldType.Type)
	assert.Equal(t, models.ManualScoring, goal2.ScoringType)

	// Check third habit (elastic duration)
	goal3 := schema.Habits[2]
	assert.Equal(t, "Exercise Duration", goal3.Title)
	assert.Equal(t, "exercise_duration", goal3.ID)
	assert.Equal(t, 3, goal3.Position)
	assert.Equal(t, models.ElasticHabit, goal3.HabitType)
	assert.Equal(t, models.DurationFieldType, goal3.FieldType.Type)
	assert.Equal(t, models.AutomaticScoring, goal3.ScoringType)
	assert.NotNil(t, goal3.MiniCriteria)
	assert.NotNil(t, goal3.MidiCriteria)
	assert.NotNil(t, goal3.MaxiCriteria)

	// Check fourth habit (elastic numeric with units)
	goal4 := schema.Habits[3]
	assert.Equal(t, "Water Intake", goal4.Title)
	assert.Equal(t, "water_intake", goal4.ID)
	assert.Equal(t, 4, goal4.Position)
	assert.Equal(t, models.ElasticHabit, goal4.HabitType)
	assert.Equal(t, models.UnsignedIntFieldType, goal4.FieldType.Type)
	assert.Equal(t, "glasses", goal4.FieldType.Unit)
	assert.Equal(t, models.AutomaticScoring, goal4.ScoringType)
	assert.NotNil(t, goal4.MiniCriteria)
	assert.NotNil(t, goal4.MidiCriteria)
	assert.NotNil(t, goal4.MaxiCriteria)
}

func TestFileInitializer_createEmptyEntriesFile(t *testing.T) {
	tempDir := t.TempDir()
	entriesFile := filepath.Join(tempDir, "entries.yml")

	initializer := NewFileInitializer()
	err := initializer.createEmptyEntriesFile(entriesFile)
	require.NoError(t, err)

	// Load and validate the created file
	entryStorage := storage.NewEntryStorage()
	entryLog, err := entryStorage.LoadFromFile(entriesFile)
	require.NoError(t, err)

	assert.Equal(t, "1.0.0", entryLog.Version)
	assert.Empty(t, entryLog.Entries)

	// Validate the file structure
	err = entryLog.Validate()
	assert.NoError(t, err)
}

func TestFileExists(t *testing.T) {
	t.Run("existing file", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "test.txt")

		err := os.WriteFile(testFile, []byte("test"), 0o600)
		require.NoError(t, err)

		assert.True(t, fileExists(testFile))
	})

	t.Run("non-existing file", func(t *testing.T) {
		assert.False(t, fileExists("/nonexistent/file.txt"))
	})

	t.Run("directory instead of file", func(t *testing.T) {
		tempDir := t.TempDir()
		assert.False(t, fileExists(tempDir))
	})
}
