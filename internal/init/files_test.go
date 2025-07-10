package init

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"davidlee/iter/internal/models"
	"davidlee/iter/internal/parser"
	"davidlee/iter/internal/storage"
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
		goalsFile := filepath.Join(tempDir, "goals.yml")
		entriesFile := filepath.Join(tempDir, "entries.yml")

		initializer := NewFileInitializer()
		err := initializer.EnsureConfigFiles(goalsFile, entriesFile)
		require.NoError(t, err)

		// Verify goals file was created and is valid
		assert.FileExists(t, goalsFile)
		goalParser := parser.NewGoalParser()
		schema, err := goalParser.LoadFromFile(goalsFile)
		require.NoError(t, err)
		assert.Equal(t, "1.0.0", schema.Version)
		assert.Len(t, schema.Goals, 2)

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
		goalsFile := filepath.Join(tempDir, "goals.yml")
		entriesFile := filepath.Join(tempDir, "entries.yml")

		// Create existing files
		err := os.WriteFile(goalsFile, []byte("existing goals"), 0o600)
		require.NoError(t, err)
		err = os.WriteFile(entriesFile, []byte("existing entries"), 0o600)
		require.NoError(t, err)

		initializer := NewFileInitializer()
		err = initializer.EnsureConfigFiles(goalsFile, entriesFile)
		require.NoError(t, err)

		// Verify files weren't overwritten
		goalsContent, err := os.ReadFile(goalsFile) //nolint:gosec // Test files in temp directory
		require.NoError(t, err)
		assert.Equal(t, "existing goals", string(goalsContent))

		entriesContent, err := os.ReadFile(entriesFile) //nolint:gosec // Test files in temp directory
		require.NoError(t, err)
		assert.Equal(t, "existing entries", string(entriesContent))
	})

	t.Run("creates only missing file", func(t *testing.T) {
		tempDir := t.TempDir()
		goalsFile := filepath.Join(tempDir, "goals.yml")
		entriesFile := filepath.Join(tempDir, "entries.yml")

		// Create only goals file
		err := os.WriteFile(goalsFile, []byte("existing goals"), 0o600)
		require.NoError(t, err)

		initializer := NewFileInitializer()
		err = initializer.EnsureConfigFiles(goalsFile, entriesFile)
		require.NoError(t, err)

		// Verify goals file wasn't overwritten
		goalsContent, err := os.ReadFile(goalsFile) //nolint:gosec // Test files in temp directory
		require.NoError(t, err)
		assert.Equal(t, "existing goals", string(goalsContent))

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
		goalsFile := filepath.Join(configDir, "goals.yml")
		entriesFile := filepath.Join(configDir, "entries.yml")

		initializer := NewFileInitializer()
		err := initializer.EnsureConfigFiles(goalsFile, entriesFile)
		require.NoError(t, err)

		// Verify directory was created
		assert.DirExists(t, configDir)
		assert.FileExists(t, goalsFile)
		assert.FileExists(t, entriesFile)
	})
}

func TestFileInitializer_createSampleGoalsFile(t *testing.T) {
	tempDir := t.TempDir()
	goalsFile := filepath.Join(tempDir, "goals.yml")

	initializer := NewFileInitializer()
	err := initializer.createSampleGoalsFile(goalsFile)
	require.NoError(t, err)

	// Load and validate the created file
	goalParser := parser.NewGoalParser()
	schema, err := goalParser.LoadFromFile(goalsFile)
	require.NoError(t, err)

	assert.Equal(t, "1.0.0", schema.Version)
	assert.Len(t, schema.Goals, 2)

	// Check first goal
	goal1 := schema.Goals[0]
	assert.Equal(t, "Morning Exercise", goal1.Title)
	assert.Equal(t, "morning_exercise", goal1.ID)
	assert.Equal(t, 1, goal1.Position)
	assert.Equal(t, models.SimpleGoal, goal1.GoalType)
	assert.Equal(t, models.BooleanFieldType, goal1.FieldType.Type)
	assert.Equal(t, models.ManualScoring, goal1.ScoringType)
	assert.NotEmpty(t, goal1.Description)
	assert.NotEmpty(t, goal1.Prompt)
	assert.NotEmpty(t, goal1.HelpText)

	// Check second goal
	goal2 := schema.Goals[1]
	assert.Equal(t, "Daily Reading", goal2.Title)
	assert.Equal(t, "daily_reading", goal2.ID)
	assert.Equal(t, 2, goal2.Position)
	assert.Equal(t, models.SimpleGoal, goal2.GoalType)
	assert.Equal(t, models.BooleanFieldType, goal2.FieldType.Type)
	assert.Equal(t, models.ManualScoring, goal2.ScoringType)
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
