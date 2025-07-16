package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntryLog_Validate(t *testing.T) {
	t.Run("valid entry log", func(t *testing.T) {
		entryLog := EntryLog{
			Version: "1.0.0",
			Entries: []DayEntry{
				{
					Date: "2024-01-01",
					Habits: []HabitEntry{
						{
							HabitID:   "morning_meditation",
							Value:     true,
							Status:    EntryCompleted,
							CreatedAt: time.Now(),
						},
					},
				},
			},
		}

		err := entryLog.Validate()
		assert.NoError(t, err)
	})

	t.Run("version is required", func(t *testing.T) {
		entryLog := EntryLog{
			Entries: []DayEntry{},
		}

		err := entryLog.Validate()
		assert.EqualError(t, err, "entry log version is required")
	})

	t.Run("duplicate dates", func(t *testing.T) {
		entryLog := EntryLog{
			Version: "1.0.0",
			Entries: []DayEntry{
				{Date: "2024-01-01", Habits: []HabitEntry{}},
				{Date: "2024-01-01", Habits: []HabitEntry{}}, // Duplicate
			},
		}

		err := entryLog.Validate()
		assert.EqualError(t, err, "duplicate date: 2024-01-01")
	})

	t.Run("invalid day entry", func(t *testing.T) {
		entryLog := EntryLog{
			Version: "1.0.0",
			Entries: []DayEntry{
				{Date: ""}, // Invalid date
			},
		}

		err := entryLog.Validate()
		assert.Contains(t, err.Error(), "day entry at index 0")
		assert.Contains(t, err.Error(), "date is required")
	})
}

func TestDayEntry_Validate(t *testing.T) {
	t.Run("valid day entry", func(t *testing.T) {
		dayEntry := DayEntry{
			Date: "2024-01-01",
			Habits: []HabitEntry{
				{
					HabitID:   "meditation",
					Value:     true,
					Status:    EntryCompleted,
					CreatedAt: time.Now(),
				},
				{
					HabitID:   "exercise",
					Value:     false,
					Status:    EntryFailed,
					CreatedAt: time.Now(),
				},
			},
		}

		err := dayEntry.Validate()
		assert.NoError(t, err)
	})

	t.Run("date is required", func(t *testing.T) {
		dayEntry := DayEntry{
			Habits: []HabitEntry{},
		}

		err := dayEntry.Validate()
		assert.EqualError(t, err, "date is required")
	})

	t.Run("invalid date format", func(t *testing.T) {
		dayEntry := DayEntry{
			Date:   "invalid-date",
			Habits: []HabitEntry{},
		}

		err := dayEntry.Validate()
		assert.Contains(t, err.Error(), "invalid date format")
	})

	t.Run("duplicate habit IDs within day", func(t *testing.T) {
		dayEntry := DayEntry{
			Date: "2024-01-01",
			Habits: []HabitEntry{
				{HabitID: "meditation", Value: true, Status: EntryCompleted, CreatedAt: time.Now()},
				{HabitID: "meditation", Value: false, Status: EntryFailed, CreatedAt: time.Now()}, // Duplicate
			},
		}

		err := dayEntry.Validate()
		assert.Contains(t, err.Error(), "duplicate habit ID for date 2024-01-01: meditation")
	})

	t.Run("invalid habit entry", func(t *testing.T) {
		dayEntry := DayEntry{
			Date: "2024-01-01",
			Habits: []HabitEntry{
				{HabitID: ""}, // Invalid habit entry
			},
		}

		err := dayEntry.Validate()
		assert.Contains(t, err.Error(), "habit entry at index 0")
		assert.Contains(t, err.Error(), "habit ID is required")
	})
}

func TestHabitEntry_Validate(t *testing.T) {
	t.Run("valid boolean habit entry", func(t *testing.T) {
		goalEntry := HabitEntry{
			HabitID:   "meditation",
			Value:     true,
			Status:    EntryCompleted,
			CreatedAt: time.Now(),
		}

		err := goalEntry.Validate()
		assert.NoError(t, err)
	})

	t.Run("valid habit entry with notes", func(t *testing.T) {
		goalEntry := HabitEntry{
			HabitID:   "exercise",
			Value:     false,
			Status:    EntryFailed,
			CreatedAt: time.Now(),
			Notes:     "Was feeling sick today",
		}

		err := goalEntry.Validate()
		assert.NoError(t, err)
	})

	t.Run("habit ID is required", func(t *testing.T) {
		goalEntry := HabitEntry{
			Value:     true,
			Status:    EntryCompleted,
			CreatedAt: time.Now(),
		}

		err := goalEntry.Validate()
		assert.EqualError(t, err, "habit ID is required")
	})

	t.Run("whitespace-only habit ID is invalid", func(t *testing.T) {
		goalEntry := HabitEntry{
			HabitID:   "   ",
			Value:     true,
			Status:    EntryCompleted,
			CreatedAt: time.Now(),
		}

		err := goalEntry.Validate()
		assert.EqualError(t, err, "habit ID is required")
	})

	t.Run("value is required for completed/failed entries", func(t *testing.T) {
		goalEntry := HabitEntry{
			HabitID:   "meditation",
			Status:    EntryCompleted,
			CreatedAt: time.Now(),
			// Value is nil
		}

		err := goalEntry.Validate()
		assert.EqualError(t, err, "completed and failed entries must have values")
	})

	t.Run("skipped entries with achievement levels are allowed", func(t *testing.T) {
		achievementLevel := AchievementMini
		goalEntry := HabitEntry{
			HabitID:          "meditation",
			Status:           EntrySkipped,
			AchievementLevel: &achievementLevel,
			CreatedAt:        time.Now(),
			// Value is nil (valid for skipped entries)
		}

		err := goalEntry.Validate()
		assert.NoError(t, err, "skipped entries should allow achievement levels per ADR-001")
	})

	t.Run("skipped entries with values are still invalid", func(t *testing.T) {
		goalEntry := HabitEntry{
			HabitID:   "meditation",
			Value:     true,
			Status:    EntrySkipped,
			CreatedAt: time.Now(),
		}

		err := goalEntry.Validate()
		assert.EqualError(t, err, "skipped entries cannot have values")
	})
}

func TestEntryLog_GetDayEntry(t *testing.T) {
	entryLog := EntryLog{
		Entries: []DayEntry{
			{Date: "2024-01-01", Habits: []HabitEntry{}},
			{Date: "2024-01-02", Habits: []HabitEntry{}},
		},
	}

	t.Run("existing date", func(t *testing.T) {
		entry, found := entryLog.GetDayEntry("2024-01-01")
		assert.True(t, found)
		require.NotNil(t, entry)
		assert.Equal(t, "2024-01-01", entry.Date)
	})

	t.Run("non-existing date", func(t *testing.T) {
		entry, found := entryLog.GetDayEntry("2024-01-03")
		assert.False(t, found)
		assert.Nil(t, entry)
	})
}

func TestEntryLog_AddDayEntry(t *testing.T) {
	t.Run("add valid day entry", func(t *testing.T) {
		entryLog := CreateEmptyEntryLog()

		dayEntry := DayEntry{
			Date: "2024-01-01",
			Habits: []HabitEntry{
				{HabitID: "meditation", Value: true, Status: EntryCompleted, CreatedAt: time.Now()},
			},
		}

		err := entryLog.AddDayEntry(dayEntry)
		assert.NoError(t, err)
		assert.Len(t, entryLog.Entries, 1)
		assert.Equal(t, "2024-01-01", entryLog.Entries[0].Date)
	})

	t.Run("add duplicate date", func(t *testing.T) {
		entryLog := EntryLog{
			Version: "1.0.0",
			Entries: []DayEntry{
				{Date: "2024-01-01", Habits: []HabitEntry{}},
			},
		}

		dayEntry := DayEntry{
			Date:   "2024-01-01",
			Habits: []HabitEntry{},
		}

		err := entryLog.AddDayEntry(dayEntry)
		assert.EqualError(t, err, "entry for date 2024-01-01 already exists")
	})

	t.Run("add invalid day entry", func(t *testing.T) {
		entryLog := CreateEmptyEntryLog()

		dayEntry := DayEntry{
			Date: "", // Invalid
		}

		err := entryLog.AddDayEntry(dayEntry)
		assert.Contains(t, err.Error(), "invalid day entry")
	})
}

func TestEntryLog_UpdateDayEntry(t *testing.T) {
	t.Run("update existing entry", func(t *testing.T) {
		entryLog := EntryLog{
			Version: "1.0.0",
			Entries: []DayEntry{
				{
					Date: "2024-01-01",
					Habits: []HabitEntry{
						{HabitID: "meditation", Value: false, Status: EntryFailed, CreatedAt: time.Now()},
					},
				},
			},
		}

		updatedEntry := DayEntry{
			Date: "2024-01-01",
			Habits: []HabitEntry{
				{HabitID: "meditation", Value: true, Status: EntryCompleted, CreatedAt: time.Now()},
				{HabitID: "exercise", Value: true, Status: EntryCompleted, CreatedAt: time.Now()},
			},
		}

		err := entryLog.UpdateDayEntry(updatedEntry)
		assert.NoError(t, err)
		assert.Len(t, entryLog.Entries, 1)
		assert.Len(t, entryLog.Entries[0].Habits, 2)
	})

	t.Run("add new entry when not exists", func(t *testing.T) {
		entryLog := CreateEmptyEntryLog()

		dayEntry := DayEntry{
			Date: "2024-01-01",
			Habits: []HabitEntry{
				{HabitID: "meditation", Value: true, Status: EntryCompleted, CreatedAt: time.Now()},
			},
		}

		err := entryLog.UpdateDayEntry(dayEntry)
		assert.NoError(t, err)
		assert.Len(t, entryLog.Entries, 1)
	})
}

func TestDayEntry_GetHabitEntry(t *testing.T) {
	dayEntry := DayEntry{
		Date: "2024-01-01",
		Habits: []HabitEntry{
			{HabitID: "meditation", Value: true, Status: EntryCompleted, CreatedAt: time.Now()},
			{HabitID: "exercise", Value: false, Status: EntryFailed, CreatedAt: time.Now()},
		},
	}

	t.Run("existing habit", func(t *testing.T) {
		entry, found := dayEntry.GetHabitEntry("meditation")
		assert.True(t, found)
		require.NotNil(t, entry)
		assert.Equal(t, "meditation", entry.HabitID)
		assert.Equal(t, true, entry.Value)
	})

	t.Run("non-existing habit", func(t *testing.T) {
		entry, found := dayEntry.GetHabitEntry("reading")
		assert.False(t, found)
		assert.Nil(t, entry)
	})
}

func TestDayEntry_AddHabitEntry(t *testing.T) {
	t.Run("add valid habit entry", func(t *testing.T) {
		dayEntry := DayEntry{
			Date:   "2024-01-01",
			Habits: []HabitEntry{},
		}

		goalEntry := HabitEntry{
			HabitID:   "meditation",
			Value:     true,
			Status:    EntryCompleted,
			CreatedAt: time.Now(),
		}

		err := dayEntry.AddHabitEntry(goalEntry)
		assert.NoError(t, err)
		assert.Len(t, dayEntry.Habits, 1)
		assert.Equal(t, "meditation", dayEntry.Habits[0].HabitID)
	})

	t.Run("add duplicate habit", func(t *testing.T) {
		dayEntry := DayEntry{
			Date: "2024-01-01",
			Habits: []HabitEntry{
				{HabitID: "meditation", Value: true, Status: EntryCompleted, CreatedAt: time.Now()},
			},
		}

		goalEntry := HabitEntry{
			HabitID:   "meditation",
			Value:     false,
			Status:    EntryFailed,
			CreatedAt: time.Now(),
		}

		err := dayEntry.AddHabitEntry(goalEntry)
		assert.EqualError(t, err, "entry for habit meditation already exists on date 2024-01-01")
	})
}

func TestDayEntry_UpdateHabitEntry(t *testing.T) {
	t.Run("update existing habit entry", func(t *testing.T) {
		dayEntry := DayEntry{
			Date: "2024-01-01",
			Habits: []HabitEntry{
				{HabitID: "meditation", Value: false, Status: EntryFailed, CreatedAt: time.Now()},
			},
		}

		updatedHabit := HabitEntry{
			HabitID:   "meditation",
			Value:     true,
			Notes:     "Had a great session",
			Status:    EntryCompleted,
			CreatedAt: time.Now(),
		}

		err := dayEntry.UpdateHabitEntry(updatedHabit)
		assert.NoError(t, err)
		assert.Len(t, dayEntry.Habits, 1)
		assert.Equal(t, true, dayEntry.Habits[0].Value)
		assert.Equal(t, "Had a great session", dayEntry.Habits[0].Notes)
	})

	t.Run("add new habit when not exists", func(t *testing.T) {
		dayEntry := DayEntry{
			Date:   "2024-01-01",
			Habits: []HabitEntry{},
		}

		goalEntry := HabitEntry{
			HabitID:   "meditation",
			Value:     true,
			Status:    EntryCompleted,
			CreatedAt: time.Now(),
		}

		err := dayEntry.UpdateHabitEntry(goalEntry)
		assert.NoError(t, err)
		assert.Len(t, dayEntry.Habits, 1)
	})
}

func TestHabitEntry_BooleanValue(t *testing.T) {
	t.Run("get boolean value", func(t *testing.T) {
		goalEntry := HabitEntry{
			HabitID:   "meditation",
			Value:     true,
			Status:    EntryCompleted,
			CreatedAt: time.Now(),
		}

		value, ok := goalEntry.GetBooleanValue()
		assert.True(t, ok)
		assert.True(t, value)
	})

	t.Run("non-boolean value", func(t *testing.T) {
		goalEntry := HabitEntry{
			HabitID: "steps",
			Value:   12345,
		}

		value, ok := goalEntry.GetBooleanValue()
		assert.False(t, ok)
		assert.False(t, value)
	})

	t.Run("set boolean value", func(t *testing.T) {
		goalEntry := HabitEntry{
			HabitID:   "meditation",
			Value:     false,
			Status:    EntryFailed,
			CreatedAt: time.Now(),
		}

		goalEntry.SetBooleanValue(true)
		assert.Equal(t, true, goalEntry.Value)
	})
}

func TestCreateTodayEntry(t *testing.T) {
	entry := CreateTodayEntry()
	today := time.Now().Format("2006-01-02")

	assert.Equal(t, today, entry.Date)
	assert.Empty(t, entry.Habits)
	assert.True(t, entry.IsToday())
}

func TestCreateBooleanHabitEntry(t *testing.T) {
	entry := CreateBooleanHabitEntry("meditation", true)

	assert.Equal(t, "meditation", entry.HabitID)
	assert.Equal(t, true, entry.Value)

	value, ok := entry.GetBooleanValue()
	assert.True(t, ok)
	assert.True(t, value)
}

func TestDayEntry_IsToday(t *testing.T) {
	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	t.Run("today's entry", func(t *testing.T) {
		entry := DayEntry{Date: today}
		assert.True(t, entry.IsToday())
	})

	t.Run("yesterday's entry", func(t *testing.T) {
		entry := DayEntry{Date: yesterday}
		assert.False(t, entry.IsToday())
	})
}

func TestDayEntry_GetDate(t *testing.T) {
	t.Run("valid date", func(t *testing.T) {
		entry := DayEntry{Date: "2024-01-01"}
		date, err := entry.GetDate()

		assert.NoError(t, err)
		assert.Equal(t, 2024, date.Year())
		assert.Equal(t, time.January, date.Month())
		assert.Equal(t, 1, date.Day())
	})

	t.Run("invalid date", func(t *testing.T) {
		entry := DayEntry{Date: "invalid-date"}
		_, err := entry.GetDate()

		assert.Error(t, err)
	})
}

func TestCreateEmptyEntryLog(t *testing.T) {
	entryLog := CreateEmptyEntryLog()

	assert.Equal(t, "1.0.0", entryLog.Version)
	assert.Empty(t, entryLog.Entries)

	err := entryLog.Validate()
	assert.NoError(t, err)
}

func TestEntryLog_GetEntriesForDateRange(t *testing.T) {
	entryLog := EntryLog{
		Version: "1.0.0",
		Entries: []DayEntry{
			{Date: "2024-01-01", Habits: []HabitEntry{}},
			{Date: "2024-01-03", Habits: []HabitEntry{}},
			{Date: "2024-01-05", Habits: []HabitEntry{}},
			{Date: "2024-01-07", Habits: []HabitEntry{}},
		},
	}

	t.Run("range includes multiple entries", func(t *testing.T) {
		entries, err := entryLog.GetEntriesForDateRange("2024-01-02", "2024-01-06")
		assert.NoError(t, err)
		assert.Len(t, entries, 2)
		assert.Equal(t, "2024-01-03", entries[0].Date)
		assert.Equal(t, "2024-01-05", entries[1].Date)
	})

	t.Run("range includes boundary dates", func(t *testing.T) {
		entries, err := entryLog.GetEntriesForDateRange("2024-01-01", "2024-01-05")
		assert.NoError(t, err)
		assert.Len(t, entries, 3)
		assert.Equal(t, "2024-01-01", entries[0].Date)
		assert.Equal(t, "2024-01-05", entries[2].Date)
	})

	t.Run("no entries in range", func(t *testing.T) {
		entries, err := entryLog.GetEntriesForDateRange("2024-02-01", "2024-02-28")
		assert.NoError(t, err)
		assert.Empty(t, entries)
	})

	t.Run("invalid start date", func(t *testing.T) {
		_, err := entryLog.GetEntriesForDateRange("invalid-date", "2024-01-05")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid start date format")
	})

	t.Run("invalid end date", func(t *testing.T) {
		_, err := entryLog.GetEntriesForDateRange("2024-01-01", "invalid-date")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid end date format")
	})

	t.Run("start date after end date", func(t *testing.T) {
		_, err := entryLog.GetEntriesForDateRange("2024-01-10", "2024-01-05")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "start date 2024-01-10 is after end date 2024-01-05")
	})
}

func TestHabitEntry_AchievementLevel(t *testing.T) {
	t.Run("GetAchievementLevel with no level set", func(t *testing.T) {
		entry := HabitEntry{
			HabitID:   "test_goal",
			Value:     30,
			Status:    EntryCompleted,
			CreatedAt: time.Now(),
		}

		level, ok := entry.GetAchievementLevel()
		assert.False(t, ok)
		assert.Equal(t, AchievementNone, level)
		assert.False(t, entry.HasAchievementLevel())
	})

	t.Run("SetAchievementLevel and GetAchievementLevel", func(t *testing.T) {
		entry := HabitEntry{
			HabitID:   "test_goal",
			Value:     30,
			Status:    EntryCompleted,
			CreatedAt: time.Now(),
		}

		entry.SetAchievementLevel(AchievementMidi)

		level, ok := entry.GetAchievementLevel()
		assert.True(t, ok)
		assert.Equal(t, AchievementMidi, level)
		assert.True(t, entry.HasAchievementLevel())
	})

	t.Run("ClearAchievementLevel", func(t *testing.T) {
		level := AchievementMidi
		entry := HabitEntry{
			HabitID:          "test_goal",
			Value:            30,
			AchievementLevel: &level,
			Status:           EntryCompleted,
			CreatedAt:        time.Now(),
		}

		assert.True(t, entry.HasAchievementLevel())

		entry.ClearAchievementLevel()

		retrievedLevel, ok := entry.GetAchievementLevel()
		assert.False(t, ok)
		assert.Equal(t, AchievementNone, retrievedLevel)
		assert.False(t, entry.HasAchievementLevel())
	})

	t.Run("validate valid achievement levels", func(t *testing.T) {
		validLevels := []AchievementLevel{
			AchievementNone,
			AchievementMini,
			AchievementMidi,
			AchievementMaxi,
		}

		for _, levelValue := range validLevels {
			level := levelValue // Create a copy for pointer
			entry := HabitEntry{
				HabitID:          "test_goal",
				Value:            30,
				AchievementLevel: &level,
				Status:           EntryCompleted,
				CreatedAt:        time.Now(),
			}

			err := entry.Validate()
			assert.NoError(t, err, "Level %s should be valid", level)
		}
	})

	t.Run("validate invalid achievement level", func(t *testing.T) {
		invalidLevel := AchievementLevel("invalid")
		entry := HabitEntry{
			HabitID:          "test_goal",
			Value:            30,
			AchievementLevel: &invalidLevel,
			Status:           EntryCompleted,
			CreatedAt:        time.Now(),
		}

		err := entry.Validate()
		assert.EqualError(t, err, "invalid achievement level: invalid")
	})
}

func TestCreateElasticHabitEntry(t *testing.T) {
	t.Run("create elastic habit entry", func(t *testing.T) {
		entry := CreateElasticHabitEntry("exercise", 45, AchievementMidi)

		assert.Equal(t, "exercise", entry.HabitID)
		assert.Equal(t, 45, entry.Value)

		level, ok := entry.GetAchievementLevel()
		assert.True(t, ok)
		assert.Equal(t, AchievementMidi, level)
		assert.True(t, entry.HasAchievementLevel())
	})
}

func TestCreateValueOnlyHabitEntry(t *testing.T) {
	t.Run("create value-only habit entry", func(t *testing.T) {
		entry := CreateValueOnlyHabitEntry("reading", "30 minutes")

		assert.Equal(t, "reading", entry.HabitID)
		assert.Equal(t, "30 minutes", entry.Value)
		assert.False(t, entry.HasAchievementLevel())
	})
}

func TestCreateSkippedHabitEntry(t *testing.T) {
	t.Run("create skipped habit entry", func(t *testing.T) {
		entry := CreateSkippedHabitEntry("meditation")

		assert.Equal(t, "meditation", entry.HabitID)
		assert.Nil(t, entry.Value)
		assert.Nil(t, entry.AchievementLevel)
		assert.Equal(t, EntrySkipped, entry.Status)
		assert.False(t, entry.CreatedAt.IsZero())
		assert.True(t, entry.IsSkipped())
		assert.False(t, entry.RequiresValue())

		// Validate the skipped entry
		err := entry.Validate()
		assert.NoError(t, err)
	})
}

func TestHabitEntry_StatusHelperMethods(t *testing.T) {
	t.Run("completed entry", func(t *testing.T) {
		entry := CreateBooleanHabitEntry("meditation", true)

		assert.True(t, entry.IsCompleted())
		assert.False(t, entry.IsSkipped())
		assert.False(t, entry.HasFailure())
		assert.True(t, entry.IsFinalized())
		assert.True(t, entry.RequiresValue())
	})

	t.Run("failed entry", func(t *testing.T) {
		entry := CreateBooleanHabitEntry("exercise", false)

		assert.False(t, entry.IsCompleted())
		assert.False(t, entry.IsSkipped())
		assert.True(t, entry.HasFailure())
		assert.True(t, entry.IsFinalized())
		assert.True(t, entry.RequiresValue())
	})

	t.Run("skipped entry", func(t *testing.T) {
		entry := CreateSkippedHabitEntry("reading")

		assert.False(t, entry.IsCompleted())
		assert.True(t, entry.IsSkipped())
		assert.False(t, entry.HasFailure())
		assert.True(t, entry.IsFinalized())
		assert.False(t, entry.RequiresValue())
	})
}

func TestHabitEntry_TimestampMethods(t *testing.T) {
	t.Run("mark created and updated", func(t *testing.T) {
		entry := HabitEntry{
			HabitID: "test",
			Value:   true,
			Status:  EntryCompleted,
		}

		// Test MarkCreated
		entry.MarkCreated()
		assert.False(t, entry.CreatedAt.IsZero())
		assert.Nil(t, entry.UpdatedAt)

		originalCreated := entry.CreatedAt
		lastModified := entry.GetLastModified()
		assert.Equal(t, originalCreated, lastModified)

		// Test MarkUpdated
		time.Sleep(1 * time.Millisecond) // Ensure different timestamp
		entry.MarkUpdated()
		assert.NotNil(t, entry.UpdatedAt)
		assert.True(t, entry.UpdatedAt.After(originalCreated))

		// GetLastModified should now return UpdatedAt
		lastModified = entry.GetLastModified()
		assert.Equal(t, *entry.UpdatedAt, lastModified)
	})
}

func TestAchievementLevelValidation(t *testing.T) {
	t.Run("IsValidAchievementLevel with valid levels", func(t *testing.T) {
		validLevels := []string{"none", "mini", "midi", "maxi"}

		for _, level := range validLevels {
			assert.True(t, IsValidAchievementLevel(level), "Level %s should be valid", level)
		}
	})

	t.Run("IsValidAchievementLevel with invalid levels", func(t *testing.T) {
		invalidLevels := []string{"", "invalid", "MINI", "maximum", "minimum"}

		for _, level := range invalidLevels {
			assert.False(t, IsValidAchievementLevel(level), "Level %s should be invalid", level)
		}
	})

	t.Run("isValidAchievementLevel with constants", func(t *testing.T) {
		assert.True(t, isValidAchievementLevel(AchievementNone))
		assert.True(t, isValidAchievementLevel(AchievementMini))
		assert.True(t, isValidAchievementLevel(AchievementMidi))
		assert.True(t, isValidAchievementLevel(AchievementMaxi))
		assert.False(t, isValidAchievementLevel(AchievementLevel("invalid")))
	})
}
