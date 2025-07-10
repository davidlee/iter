package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"davidlee/iter/internal/models"
)

func TestEntryStorage_ParseYAML(t *testing.T) {
	storage := NewEntryStorage()

	t.Run("valid entry log YAML", func(t *testing.T) {
		yamlData := `
version: "1.0.0"
entries:
  - date: "2024-01-01"
    goals:
      - goal_id: "morning_meditation"
        value: true
        notes: "Great session today"
      - goal_id: "daily_exercise"
        value: false
`

		entryLog, err := storage.ParseYAML([]byte(yamlData))
		require.NoError(t, err)
		require.NotNil(t, entryLog)

		assert.Equal(t, "1.0.0", entryLog.Version)
		assert.Len(t, entryLog.Entries, 1)

		dayEntry := entryLog.Entries[0]
		assert.Equal(t, "2024-01-01", dayEntry.Date)
		assert.Len(t, dayEntry.Goals, 2)

		goal1 := dayEntry.Goals[0]
		assert.Equal(t, "morning_meditation", goal1.GoalID)
		assert.Equal(t, true, goal1.Value)
		assert.Equal(t, "Great session today", goal1.Notes)

		goal2 := dayEntry.Goals[1]
		assert.Equal(t, "daily_exercise", goal2.GoalID)
		assert.Equal(t, false, goal2.Value)
	})

	t.Run("empty entry log", func(t *testing.T) {
		yamlData := `
version: "1.0.0"
entries: []
`

		entryLog, err := storage.ParseYAML([]byte(yamlData))
		require.NoError(t, err)
		require.NotNil(t, entryLog)

		assert.Equal(t, "1.0.0", entryLog.Version)
		assert.Empty(t, entryLog.Entries)
	})

	t.Run("invalid YAML syntax", func(t *testing.T) {
		yamlData := `
version: "1.0.0"
entries:
  - invalid_yaml: [unclosed
`

		_, err := storage.ParseYAML([]byte(yamlData))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse YAML")
	})

	t.Run("validation failure", func(t *testing.T) {
		yamlData := `
version: ""
entries: []
`

		_, err := storage.ParseYAML([]byte(yamlData))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "entry log validation failed")
	})

	t.Run("unknown field in strict mode", func(t *testing.T) {
		yamlData := `
version: "1.0.0"
unknown_field: "should cause error"
entries: []
`

		_, err := storage.ParseYAML([]byte(yamlData))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse YAML")
	})
}

func TestEntryStorage_LoadFromFile(t *testing.T) {
	storage := NewEntryStorage()

	t.Run("load existing file", func(t *testing.T) {
		tempDir := t.TempDir()
		entriesFile := filepath.Join(tempDir, "entries.yml")

		yamlContent := `
version: "1.0.0"
entries:
  - date: "2024-01-01"
    goals:
      - goal_id: "meditation"
        value: true
`

		err := os.WriteFile(entriesFile, []byte(yamlContent), 0o600)
		require.NoError(t, err)

		entryLog, err := storage.LoadFromFile(entriesFile)
		require.NoError(t, err)
		require.NotNil(t, entryLog)

		assert.Equal(t, "1.0.0", entryLog.Version)
		assert.Len(t, entryLog.Entries, 1)
	})

	t.Run("load non-existing file returns empty log", func(t *testing.T) {
		entryLog, err := storage.LoadFromFile("/nonexistent/entries.yml")
		require.NoError(t, err)
		require.NotNil(t, entryLog)

		assert.Equal(t, "1.0.0", entryLog.Version)
		assert.Empty(t, entryLog.Entries)
	})

	t.Run("file read permission error", func(t *testing.T) {
		tempDir := t.TempDir()
		entriesFile := filepath.Join(tempDir, "unreadable.yml")

		// Create file and remove read permission
		err := os.WriteFile(entriesFile, []byte("test"), 0o000)
		require.NoError(t, err)

		_, err = storage.LoadFromFile(entriesFile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read entries file")
	})
}

func TestEntryStorage_SaveToFile(t *testing.T) {
	storage := NewEntryStorage()

	t.Run("save valid entry log", func(t *testing.T) {
		tempDir := t.TempDir()
		entriesFile := filepath.Join(tempDir, "entries.yml")

		entryLog := models.CreateEmptyEntryLog()
		dayEntry := models.DayEntry{
			Date: "2024-01-01",
			Goals: []models.GoalEntry{
				{
					GoalID: "meditation",
					Value:  true,
				},
			},
		}
		err := entryLog.AddDayEntry(dayEntry)
		require.NoError(t, err)

		err = storage.SaveToFile(entryLog, entriesFile)
		require.NoError(t, err)

		// Verify file was created and can be loaded back
		loadedLog, err := storage.LoadFromFile(entriesFile)
		require.NoError(t, err)

		assert.Equal(t, entryLog.Version, loadedLog.Version)
		assert.Len(t, loadedLog.Entries, 1)
		assert.Equal(t, "2024-01-01", loadedLog.Entries[0].Date)
	})

	t.Run("save invalid entry log", func(t *testing.T) {
		tempDir := t.TempDir()
		entriesFile := filepath.Join(tempDir, "entries.yml")

		// Invalid entry log (missing version)
		entryLog := &models.EntryLog{
			Entries: []models.DayEntry{},
		}

		err := storage.SaveToFile(entryLog, entriesFile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot save invalid entry log")
	})

	t.Run("atomic write with temporary file", func(t *testing.T) {
		tempDir := t.TempDir()
		entriesFile := filepath.Join(tempDir, "entries.yml")

		entryLog := models.CreateEmptyEntryLog()

		err := storage.SaveToFile(entryLog, entriesFile)
		require.NoError(t, err)

		// Verify no temporary file exists after successful write
		tempFile := entriesFile + ".tmp"
		_, err = os.Stat(tempFile)
		assert.True(t, os.IsNotExist(err))

		// Verify final file exists
		_, err = os.Stat(entriesFile)
		assert.NoError(t, err)
	})

	t.Run("create directory if not exists", func(t *testing.T) {
		tempDir := t.TempDir()
		nestedDir := filepath.Join(tempDir, "nested", "dir")
		entriesFile := filepath.Join(nestedDir, "entries.yml")

		entryLog := models.CreateEmptyEntryLog()

		err := storage.SaveToFile(entryLog, entriesFile)
		require.NoError(t, err)

		// Verify directory was created
		info, err := os.Stat(nestedDir)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
	})
}

func TestEntryStorage_AddDayEntry(t *testing.T) {
	storage := NewEntryStorage()

	t.Run("add to new file", func(t *testing.T) {
		tempDir := t.TempDir()
		entriesFile := filepath.Join(tempDir, "entries.yml")

		dayEntry := models.DayEntry{
			Date: "2024-01-01",
			Goals: []models.GoalEntry{
				{GoalID: "meditation", Value: true},
			},
		}

		err := storage.AddDayEntry(entriesFile, dayEntry)
		require.NoError(t, err)

		// Verify entry was added
		entryLog, err := storage.LoadFromFile(entriesFile)
		require.NoError(t, err)
		assert.Len(t, entryLog.Entries, 1)
		assert.Equal(t, "2024-01-01", entryLog.Entries[0].Date)
	})

	t.Run("add to existing file", func(t *testing.T) {
		tempDir := t.TempDir()
		entriesFile := filepath.Join(tempDir, "entries.yml")

		// Create initial entry
		initialEntry := models.DayEntry{
			Date: "2024-01-01",
			Goals: []models.GoalEntry{
				{GoalID: "meditation", Value: true},
			},
		}

		err := storage.AddDayEntry(entriesFile, initialEntry)
		require.NoError(t, err)

		// Add another entry
		secondEntry := models.DayEntry{
			Date: "2024-01-02",
			Goals: []models.GoalEntry{
				{GoalID: "exercise", Value: false},
			},
		}

		err = storage.AddDayEntry(entriesFile, secondEntry)
		require.NoError(t, err)

		// Verify both entries exist
		entryLog, err := storage.LoadFromFile(entriesFile)
		require.NoError(t, err)
		assert.Len(t, entryLog.Entries, 2)
	})

	t.Run("add duplicate date fails", func(t *testing.T) {
		tempDir := t.TempDir()
		entriesFile := filepath.Join(tempDir, "entries.yml")

		dayEntry := models.DayEntry{
			Date: "2024-01-01",
			Goals: []models.GoalEntry{
				{GoalID: "meditation", Value: true},
			},
		}

		// Add first time
		err := storage.AddDayEntry(entriesFile, dayEntry)
		require.NoError(t, err)

		// Try to add same date again
		err = storage.AddDayEntry(entriesFile, dayEntry)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "entry for date 2024-01-01 already exists")
	})
}

func TestEntryStorage_UpdateDayEntry(t *testing.T) {
	storage := NewEntryStorage()

	t.Run("update existing entry", func(t *testing.T) {
		tempDir := t.TempDir()
		entriesFile := filepath.Join(tempDir, "entries.yml")

		// Create initial entry
		initialEntry := models.DayEntry{
			Date: "2024-01-01",
			Goals: []models.GoalEntry{
				{GoalID: "meditation", Value: false},
			},
		}

		err := storage.AddDayEntry(entriesFile, initialEntry)
		require.NoError(t, err)

		// Update entry
		updatedEntry := models.DayEntry{
			Date: "2024-01-01",
			Goals: []models.GoalEntry{
				{GoalID: "meditation", Value: true},
				{GoalID: "exercise", Value: true},
			},
		}

		err = storage.UpdateDayEntry(entriesFile, updatedEntry)
		require.NoError(t, err)

		// Verify update
		entryLog, err := storage.LoadFromFile(entriesFile)
		require.NoError(t, err)
		assert.Len(t, entryLog.Entries, 1)
		assert.Len(t, entryLog.Entries[0].Goals, 2)
	})

	t.Run("create new entry when not exists", func(t *testing.T) {
		tempDir := t.TempDir()
		entriesFile := filepath.Join(tempDir, "entries.yml")

		dayEntry := models.DayEntry{
			Date: "2024-01-01",
			Goals: []models.GoalEntry{
				{GoalID: "meditation", Value: true},
			},
		}

		err := storage.UpdateDayEntry(entriesFile, dayEntry)
		require.NoError(t, err)

		// Verify entry was created
		entryLog, err := storage.LoadFromFile(entriesFile)
		require.NoError(t, err)
		assert.Len(t, entryLog.Entries, 1)
	})
}

func TestEntryStorage_GetDayEntry(t *testing.T) {
	storage := NewEntryStorage()

	t.Run("get existing day entry", func(t *testing.T) {
		tempDir := t.TempDir()
		entriesFile := filepath.Join(tempDir, "entries.yml")

		// Create entry
		dayEntry := models.DayEntry{
			Date: "2024-01-01",
			Goals: []models.GoalEntry{
				{GoalID: "meditation", Value: true},
			},
		}

		err := storage.AddDayEntry(entriesFile, dayEntry)
		require.NoError(t, err)

		// Get entry
		retrievedEntry, err := storage.GetDayEntry(entriesFile, "2024-01-01")
		require.NoError(t, err)
		require.NotNil(t, retrievedEntry)

		assert.Equal(t, "2024-01-01", retrievedEntry.Date)
		assert.Len(t, retrievedEntry.Goals, 1)
	})

	t.Run("get non-existing day entry", func(t *testing.T) {
		tempDir := t.TempDir()
		entriesFile := filepath.Join(tempDir, "entries.yml")

		// Create empty file
		err := storage.SaveToFile(models.CreateEmptyEntryLog(), entriesFile)
		require.NoError(t, err)

		// Try to get non-existing entry
		_, err = storage.GetDayEntry(entriesFile, "2024-01-01")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no entry found for date 2024-01-01")
	})
}

func TestEntryStorage_GetTodayEntry(t *testing.T) {
	storage := NewEntryStorage()

	t.Run("get today's entry", func(t *testing.T) {
		tempDir := t.TempDir()
		entriesFile := filepath.Join(tempDir, "entries.yml")

		today := time.Now().Format("2006-01-02")
		dayEntry := models.DayEntry{
			Date: today,
			Goals: []models.GoalEntry{
				{GoalID: "meditation", Value: true},
			},
		}

		err := storage.AddDayEntry(entriesFile, dayEntry)
		require.NoError(t, err)

		// Get today's entry
		retrievedEntry, err := storage.GetTodayEntry(entriesFile)
		require.NoError(t, err)
		require.NotNil(t, retrievedEntry)

		assert.Equal(t, today, retrievedEntry.Date)
	})
}

func TestEntryStorage_GoalEntry(t *testing.T) {
	storage := NewEntryStorage()

	t.Run("add goal entry to existing day", func(t *testing.T) {
		tempDir := t.TempDir()
		entriesFile := filepath.Join(tempDir, "entries.yml")

		// Create day entry
		dayEntry := models.DayEntry{
			Date:  "2024-01-01",
			Goals: []models.GoalEntry{},
		}

		err := storage.AddDayEntry(entriesFile, dayEntry)
		require.NoError(t, err)

		// Add goal entry
		goalEntry := models.GoalEntry{
			GoalID: "meditation",
			Value:  true,
		}

		err = storage.AddGoalEntry(entriesFile, "2024-01-01", goalEntry)
		require.NoError(t, err)

		// Verify goal was added
		entryLog, err := storage.LoadFromFile(entriesFile)
		require.NoError(t, err)
		assert.Len(t, entryLog.Entries[0].Goals, 1)
		assert.Equal(t, "meditation", entryLog.Entries[0].Goals[0].GoalID)
	})

	t.Run("add goal entry creates new day", func(t *testing.T) {
		tempDir := t.TempDir()
		entriesFile := filepath.Join(tempDir, "entries.yml")

		goalEntry := models.GoalEntry{
			GoalID: "meditation",
			Value:  true,
		}

		err := storage.AddGoalEntry(entriesFile, "2024-01-01", goalEntry)
		require.NoError(t, err)

		// Verify day and goal were created
		entryLog, err := storage.LoadFromFile(entriesFile)
		require.NoError(t, err)
		assert.Len(t, entryLog.Entries, 1)
		assert.Equal(t, "2024-01-01", entryLog.Entries[0].Date)
		assert.Len(t, entryLog.Entries[0].Goals, 1)
	})

	t.Run("update goal entry", func(t *testing.T) {
		tempDir := t.TempDir()
		entriesFile := filepath.Join(tempDir, "entries.yml")

		// Add initial goal entry
		goalEntry := models.GoalEntry{
			GoalID: "meditation",
			Value:  false,
		}

		err := storage.AddGoalEntry(entriesFile, "2024-01-01", goalEntry)
		require.NoError(t, err)

		// Update goal entry
		updatedGoal := models.GoalEntry{
			GoalID: "meditation",
			Value:  true,
			Notes:  "Great session!",
		}

		err = storage.UpdateGoalEntry(entriesFile, "2024-01-01", updatedGoal)
		require.NoError(t, err)

		// Verify update
		entryLog, err := storage.LoadFromFile(entriesFile)
		require.NoError(t, err)
		goal := entryLog.Entries[0].Goals[0]
		assert.Equal(t, true, goal.Value)
		assert.Equal(t, "Great session!", goal.Notes)
	})

	t.Run("update today goal entry", func(t *testing.T) {
		tempDir := t.TempDir()
		entriesFile := filepath.Join(tempDir, "entries.yml")

		goalEntry := models.GoalEntry{
			GoalID: "meditation",
			Value:  true,
		}

		err := storage.UpdateTodayGoalEntry(entriesFile, goalEntry)
		require.NoError(t, err)

		// Verify today's entry was created
		today := time.Now().Format("2006-01-02")
		retrievedEntry, err := storage.GetDayEntry(entriesFile, today)
		require.NoError(t, err)
		assert.Len(t, retrievedEntry.Goals, 1)
		assert.Equal(t, "meditation", retrievedEntry.Goals[0].GoalID)
	})
}

func TestEntryStorage_GetEntriesForDateRange(t *testing.T) {
	storage := NewEntryStorage()

	t.Run("get entries in range", func(t *testing.T) {
		tempDir := t.TempDir()
		entriesFile := filepath.Join(tempDir, "entries.yml")

		// Create multiple entries
		dates := []string{"2024-01-01", "2024-01-03", "2024-01-05", "2024-01-07"}
		for _, date := range dates {
			dayEntry := models.DayEntry{
				Date: date,
				Goals: []models.GoalEntry{
					{GoalID: "meditation", Value: true},
				},
			}
			err := storage.AddDayEntry(entriesFile, dayEntry)
			require.NoError(t, err)
		}

		// Get entries in range
		entries, err := storage.GetEntriesForDateRange(entriesFile, "2024-01-02", "2024-01-06")
		require.NoError(t, err)

		assert.Len(t, entries, 2)
		assert.Equal(t, "2024-01-03", entries[0].Date)
		assert.Equal(t, "2024-01-05", entries[1].Date)
	})
}

func TestEntryStorage_ValidateFile(t *testing.T) {
	storage := NewEntryStorage()

	t.Run("valid file", func(t *testing.T) {
		tempDir := t.TempDir()
		entriesFile := filepath.Join(tempDir, "entries.yml")

		err := storage.SaveToFile(models.CreateEmptyEntryLog(), entriesFile)
		require.NoError(t, err)

		err = storage.ValidateFile(entriesFile)
		assert.NoError(t, err)
	})

	t.Run("invalid file", func(t *testing.T) {
		tempDir := t.TempDir()
		entriesFile := filepath.Join(tempDir, "entries.yml")

		invalidYAML := `
version: ""
entries: []
`

		err := os.WriteFile(entriesFile, []byte(invalidYAML), 0o600)
		require.NoError(t, err)

		err = storage.ValidateFile(entriesFile)
		assert.Error(t, err)
	})
}

func TestEntryStorage_BackupFile(t *testing.T) {
	storage := NewEntryStorage()

	t.Run("backup existing file", func(t *testing.T) {
		tempDir := t.TempDir()
		entriesFile := filepath.Join(tempDir, "entries.yml")
		backupFile := entriesFile + ".backup"

		// Create original file
		err := storage.SaveToFile(models.CreateEmptyEntryLog(), entriesFile)
		require.NoError(t, err)

		// Create backup
		err = storage.BackupFile(entriesFile)
		require.NoError(t, err)

		// Verify backup exists
		_, err = os.Stat(backupFile)
		assert.NoError(t, err)

		// Verify backup content matches original
		// #nosec G304 - test files in temp directory
		originalData, err := os.ReadFile(entriesFile)
		require.NoError(t, err)

		// #nosec G304 - test files in temp directory
		backupData, err := os.ReadFile(backupFile)
		require.NoError(t, err)

		assert.Equal(t, originalData, backupData)
	})

	t.Run("backup non-existing file fails", func(t *testing.T) {
		storage := NewEntryStorage()

		err := storage.BackupFile("/nonexistent/entries.yml")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "entries file not found")
	})
}

func TestEntryStorage_CreateSampleEntryLog(t *testing.T) {
	storage := NewEntryStorage()

	entryLog := storage.CreateSampleEntryLog()
	require.NotNil(t, entryLog)

	// Validate the sample entry log
	err := entryLog.Validate()
	assert.NoError(t, err)

	// Check basic properties
	assert.Equal(t, "1.0.0", entryLog.Version)
	assert.Len(t, entryLog.Entries, 1)

	dayEntry := entryLog.Entries[0]
	assert.Equal(t, "2024-01-01", dayEntry.Date)
	assert.Len(t, dayEntry.Goals, 2)

	// Check goal entries
	assert.Equal(t, "morning_meditation", dayEntry.Goals[0].GoalID)
	assert.Equal(t, true, dayEntry.Goals[0].Value)
	assert.NotEmpty(t, dayEntry.Goals[0].Notes)

	assert.Equal(t, "daily_exercise", dayEntry.Goals[1].GoalID)
	assert.Equal(t, false, dayEntry.Goals[1].Value)
	assert.NotEmpty(t, dayEntry.Goals[1].Notes)
}
