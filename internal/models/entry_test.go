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
					Goals: []GoalEntry{
						{
							GoalID: "morning_meditation",
							Value:  true,
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
				{Date: "2024-01-01", Goals: []GoalEntry{}},
				{Date: "2024-01-01", Goals: []GoalEntry{}}, // Duplicate
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
			Goals: []GoalEntry{
				{
					GoalID: "meditation",
					Value:  true,
				},
				{
					GoalID: "exercise",
					Value:  false,
				},
			},
		}
		
		err := dayEntry.Validate()
		assert.NoError(t, err)
	})

	t.Run("date is required", func(t *testing.T) {
		dayEntry := DayEntry{
			Goals: []GoalEntry{},
		}
		
		err := dayEntry.Validate()
		assert.EqualError(t, err, "date is required")
	})

	t.Run("invalid date format", func(t *testing.T) {
		dayEntry := DayEntry{
			Date:  "invalid-date",
			Goals: []GoalEntry{},
		}
		
		err := dayEntry.Validate()
		assert.Contains(t, err.Error(), "invalid date format")
	})

	t.Run("duplicate goal IDs within day", func(t *testing.T) {
		dayEntry := DayEntry{
			Date: "2024-01-01",
			Goals: []GoalEntry{
				{GoalID: "meditation", Value: true},
				{GoalID: "meditation", Value: false}, // Duplicate
			},
		}
		
		err := dayEntry.Validate()
		assert.Contains(t, err.Error(), "duplicate goal ID for date 2024-01-01: meditation")
	})

	t.Run("invalid goal entry", func(t *testing.T) {
		dayEntry := DayEntry{
			Date: "2024-01-01",
			Goals: []GoalEntry{
				{GoalID: ""}, // Invalid goal entry
			},
		}
		
		err := dayEntry.Validate()
		assert.Contains(t, err.Error(), "goal entry at index 0")
		assert.Contains(t, err.Error(), "goal ID is required")
	})
}

func TestGoalEntry_Validate(t *testing.T) {
	t.Run("valid boolean goal entry", func(t *testing.T) {
		goalEntry := GoalEntry{
			GoalID: "meditation",
			Value:  true,
		}
		
		err := goalEntry.Validate()
		assert.NoError(t, err)
	})

	t.Run("valid goal entry with notes", func(t *testing.T) {
		goalEntry := GoalEntry{
			GoalID: "exercise",
			Value:  false,
			Notes:  "Was feeling sick today",
		}
		
		err := goalEntry.Validate()
		assert.NoError(t, err)
	})

	t.Run("goal ID is required", func(t *testing.T) {
		goalEntry := GoalEntry{
			Value: true,
		}
		
		err := goalEntry.Validate()
		assert.EqualError(t, err, "goal ID is required")
	})

	t.Run("whitespace-only goal ID is invalid", func(t *testing.T) {
		goalEntry := GoalEntry{
			GoalID: "   ",
			Value:  true,
		}
		
		err := goalEntry.Validate()
		assert.EqualError(t, err, "goal ID is required")
	})

	t.Run("value is required", func(t *testing.T) {
		goalEntry := GoalEntry{
			GoalID: "meditation",
			// Value is nil
		}
		
		err := goalEntry.Validate()
		assert.EqualError(t, err, "goal value is required")
	})
}

func TestEntryLog_GetDayEntry(t *testing.T) {
	entryLog := EntryLog{
		Entries: []DayEntry{
			{Date: "2024-01-01", Goals: []GoalEntry{}},
			{Date: "2024-01-02", Goals: []GoalEntry{}},
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
			Goals: []GoalEntry{
				{GoalID: "meditation", Value: true},
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
				{Date: "2024-01-01", Goals: []GoalEntry{}},
			},
		}
		
		dayEntry := DayEntry{
			Date:  "2024-01-01",
			Goals: []GoalEntry{},
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
					Goals: []GoalEntry{
						{GoalID: "meditation", Value: false},
					},
				},
			},
		}
		
		updatedEntry := DayEntry{
			Date: "2024-01-01",
			Goals: []GoalEntry{
				{GoalID: "meditation", Value: true},
				{GoalID: "exercise", Value: true},
			},
		}
		
		err := entryLog.UpdateDayEntry(updatedEntry)
		assert.NoError(t, err)
		assert.Len(t, entryLog.Entries, 1)
		assert.Len(t, entryLog.Entries[0].Goals, 2)
	})

	t.Run("add new entry when not exists", func(t *testing.T) {
		entryLog := CreateEmptyEntryLog()
		
		dayEntry := DayEntry{
			Date: "2024-01-01",
			Goals: []GoalEntry{
				{GoalID: "meditation", Value: true},
			},
		}
		
		err := entryLog.UpdateDayEntry(dayEntry)
		assert.NoError(t, err)
		assert.Len(t, entryLog.Entries, 1)
	})
}

func TestDayEntry_GetGoalEntry(t *testing.T) {
	dayEntry := DayEntry{
		Date: "2024-01-01",
		Goals: []GoalEntry{
			{GoalID: "meditation", Value: true},
			{GoalID: "exercise", Value: false},
		},
	}

	t.Run("existing goal", func(t *testing.T) {
		entry, found := dayEntry.GetGoalEntry("meditation")
		assert.True(t, found)
		require.NotNil(t, entry)
		assert.Equal(t, "meditation", entry.GoalID)
		assert.Equal(t, true, entry.Value)
	})

	t.Run("non-existing goal", func(t *testing.T) {
		entry, found := dayEntry.GetGoalEntry("reading")
		assert.False(t, found)
		assert.Nil(t, entry)
	})
}

func TestDayEntry_AddGoalEntry(t *testing.T) {
	t.Run("add valid goal entry", func(t *testing.T) {
		dayEntry := DayEntry{
			Date:  "2024-01-01",
			Goals: []GoalEntry{},
		}
		
		goalEntry := GoalEntry{
			GoalID: "meditation",
			Value:  true,
		}
		
		err := dayEntry.AddGoalEntry(goalEntry)
		assert.NoError(t, err)
		assert.Len(t, dayEntry.Goals, 1)
		assert.Equal(t, "meditation", dayEntry.Goals[0].GoalID)
	})

	t.Run("add duplicate goal", func(t *testing.T) {
		dayEntry := DayEntry{
			Date: "2024-01-01",
			Goals: []GoalEntry{
				{GoalID: "meditation", Value: true},
			},
		}
		
		goalEntry := GoalEntry{
			GoalID: "meditation",
			Value:  false,
		}
		
		err := dayEntry.AddGoalEntry(goalEntry)
		assert.EqualError(t, err, "entry for goal meditation already exists on date 2024-01-01")
	})
}

func TestDayEntry_UpdateGoalEntry(t *testing.T) {
	t.Run("update existing goal entry", func(t *testing.T) {
		dayEntry := DayEntry{
			Date: "2024-01-01",
			Goals: []GoalEntry{
				{GoalID: "meditation", Value: false},
			},
		}
		
		updatedGoal := GoalEntry{
			GoalID: "meditation",
			Value:  true,
			Notes:  "Had a great session",
		}
		
		err := dayEntry.UpdateGoalEntry(updatedGoal)
		assert.NoError(t, err)
		assert.Len(t, dayEntry.Goals, 1)
		assert.Equal(t, true, dayEntry.Goals[0].Value)
		assert.Equal(t, "Had a great session", dayEntry.Goals[0].Notes)
	})

	t.Run("add new goal when not exists", func(t *testing.T) {
		dayEntry := DayEntry{
			Date:  "2024-01-01",
			Goals: []GoalEntry{},
		}
		
		goalEntry := GoalEntry{
			GoalID: "meditation",
			Value:  true,
		}
		
		err := dayEntry.UpdateGoalEntry(goalEntry)
		assert.NoError(t, err)
		assert.Len(t, dayEntry.Goals, 1)
	})
}

func TestGoalEntry_BooleanValue(t *testing.T) {
	t.Run("get boolean value", func(t *testing.T) {
		goalEntry := GoalEntry{
			GoalID: "meditation",
			Value:  true,
		}
		
		value, ok := goalEntry.GetBooleanValue()
		assert.True(t, ok)
		assert.True(t, value)
	})

	t.Run("non-boolean value", func(t *testing.T) {
		goalEntry := GoalEntry{
			GoalID: "steps",
			Value:  12345,
		}
		
		value, ok := goalEntry.GetBooleanValue()
		assert.False(t, ok)
		assert.False(t, value)
	})

	t.Run("set boolean value", func(t *testing.T) {
		goalEntry := GoalEntry{
			GoalID: "meditation",
			Value:  false,
		}
		
		goalEntry.SetBooleanValue(true)
		assert.Equal(t, true, goalEntry.Value)
	})
}

func TestCreateTodayEntry(t *testing.T) {
	entry := CreateTodayEntry()
	today := time.Now().Format("2006-01-02")
	
	assert.Equal(t, today, entry.Date)
	assert.Empty(t, entry.Goals)
	assert.True(t, entry.IsToday())
}

func TestCreateBooleanGoalEntry(t *testing.T) {
	entry := CreateBooleanGoalEntry("meditation", true)
	
	assert.Equal(t, "meditation", entry.GoalID)
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
			{Date: "2024-01-01", Goals: []GoalEntry{}},
			{Date: "2024-01-03", Goals: []GoalEntry{}},
			{Date: "2024-01-05", Goals: []GoalEntry{}},
			{Date: "2024-01-07", Goals: []GoalEntry{}},
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