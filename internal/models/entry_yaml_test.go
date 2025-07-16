package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// AIDEV-NOTE: T020 test coverage; comprehensive tests for human-readable YAML marshaling and permissive parsing
func TestHabitEntry_HumanReadableYAMLMarshaling(t *testing.T) {
	t.Run("marshal time field values as HH:MM", func(t *testing.T) {
		// Create a time field value (zero date with time component)
		timeValue := time.Date(0, 1, 1, 8, 30, 0, 0, time.UTC)

		entry := HabitEntry{
			HabitID:   "wake_up",
			Value:     timeValue,
			Status:    EntryCompleted,
			CreatedAt: time.Date(2025, 7, 15, 9, 11, 27, 0, time.UTC),
		}

		yamlData, err := yaml.Marshal(&entry)
		require.NoError(t, err)

		yamlStr := string(yamlData)
		assert.Contains(t, yamlStr, `value: "08:30"`)
		assert.Contains(t, yamlStr, `created_at: "2025-07-15 09:11:27"`)
		assert.NotContains(t, yamlStr, "T08:30:00Z")
		assert.NotContains(t, yamlStr, "2025-07-15T09:11:27")
	})

	t.Run("marshal timestamps in human-readable format", func(t *testing.T) {
		createdAt := time.Date(2025, 7, 15, 9, 11, 27, 0, time.UTC)
		updatedAt := time.Date(2025, 7, 15, 9, 15, 32, 0, time.UTC)

		entry := HabitEntry{
			HabitID:   "meditation",
			Value:     true,
			Status:    EntryCompleted,
			CreatedAt: createdAt,
			UpdatedAt: &updatedAt,
			Notes:     "slept well",
		}

		yamlData, err := yaml.Marshal(&entry)
		require.NoError(t, err)

		yamlStr := string(yamlData)
		assert.Contains(t, yamlStr, `created_at: "2025-07-15 09:11:27"`)
		assert.Contains(t, yamlStr, `updated_at: "2025-07-15 09:15:32"`)
		assert.NotContains(t, yamlStr, "T09:11:27")
		assert.NotContains(t, yamlStr, "T09:15:32")
	})

	t.Run("marshal non-time values unchanged", func(t *testing.T) {
		entry := HabitEntry{
			HabitID:   "steps",
			Value:     12000,
			Status:    EntryCompleted,
			CreatedAt: time.Date(2025, 7, 15, 9, 11, 27, 0, time.UTC),
		}

		yamlData, err := yaml.Marshal(&entry)
		require.NoError(t, err)

		yamlStr := string(yamlData)
		assert.Contains(t, yamlStr, `value: 12000`)
	})

	t.Run("omit optional fields when empty", func(t *testing.T) {
		entry := HabitEntry{
			HabitID:   "minimal",
			Status:    EntrySkipped,
			CreatedAt: time.Date(2025, 7, 15, 9, 11, 27, 0, time.UTC),
		}

		yamlData, err := yaml.Marshal(&entry)
		require.NoError(t, err)

		yamlStr := string(yamlData)
		assert.NotContains(t, yamlStr, "value:")
		assert.NotContains(t, yamlStr, "achievement_level:")
		assert.NotContains(t, yamlStr, "notes:")
		assert.NotContains(t, yamlStr, "updated_at:")
	})
}

func TestHabitEntry_PermissiveYAMLUnmarshaling(t *testing.T) {
	t.Run("unmarshal human-readable time formats", func(t *testing.T) {
		yamlData := `
habit_id: "wake_up"
value: "08:30"
status: "completed"
created_at: "2025-07-15 09:11:27"
updated_at: "2025-07-15 09:15:32"
`

		var entry HabitEntry
		err := yaml.Unmarshal([]byte(yamlData), &entry)
		require.NoError(t, err)

		assert.Equal(t, "wake_up", entry.HabitID)
		assert.Equal(t, EntryCompleted, entry.Status)

		// Value should be parsed as time
		timeVal, ok := entry.Value.(time.Time)
		require.True(t, ok)
		assert.Equal(t, 8, timeVal.Hour())
		assert.Equal(t, 30, timeVal.Minute())

		// Timestamps should be parsed correctly
		assert.Equal(t, 2025, entry.CreatedAt.Year())
		assert.Equal(t, 9, entry.CreatedAt.Hour())
		assert.Equal(t, 11, entry.CreatedAt.Minute())
		assert.Equal(t, 27, entry.CreatedAt.Second())

		require.NotNil(t, entry.UpdatedAt)
		assert.Equal(t, 15, entry.UpdatedAt.Minute())
		assert.Equal(t, 32, entry.UpdatedAt.Second())
	})

	t.Run("unmarshal legacy RFC3339 formats", func(t *testing.T) {
		yamlData := `
habit_id: "wake_up"
value: "0000-01-01T08:30:00Z"
status: "completed"
created_at: "2025-07-15T09:11:27.886682863+10:00"
`

		var entry HabitEntry
		err := yaml.Unmarshal([]byte(yamlData), &entry)
		require.NoError(t, err)

		assert.Equal(t, "wake_up", entry.HabitID)

		// Value should be parsed as time
		timeVal, ok := entry.Value.(time.Time)
		require.True(t, ok)
		assert.Equal(t, 8, timeVal.Hour())
		assert.Equal(t, 30, timeVal.Minute())

		// Timestamp should be parsed correctly
		assert.Equal(t, 2025, entry.CreatedAt.Year())
		assert.Equal(t, time.July, entry.CreatedAt.Month())
		assert.Equal(t, 15, entry.CreatedAt.Day())
	})

	t.Run("unmarshal various time formats", func(t *testing.T) {
		testCases := []struct {
			name     string
			value    string
			expected string
		}{
			{"HH:MM format", `"08:30"`, "08:30"},
			{"HH:MM:SS format", `"08:30:45"`, "08:30"},
			{"ISO time with Z", `"08:30:00"`, "08:30"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				yamlData := `
habit_id: "test"
value: ` + tc.value + `
status: "completed"
created_at: "2025-07-15 09:00:00"
`

				var entry HabitEntry
				err := yaml.Unmarshal([]byte(yamlData), &entry)
				require.NoError(t, err)

				timeVal, ok := entry.Value.(time.Time)
				require.True(t, ok)
				assert.Equal(t, tc.expected, timeVal.Format("15:04"))
			})
		}
	})

	t.Run("unmarshal non-time values unchanged", func(t *testing.T) {
		yamlData := `
habit_id: "steps"
value: 12000
status: "completed"
created_at: "2025-07-15 09:11:27"
`

		var entry HabitEntry
		err := yaml.Unmarshal([]byte(yamlData), &entry)
		require.NoError(t, err)

		assert.Equal(t, 12000, entry.Value)
	})

	t.Run("unmarshal boolean values", func(t *testing.T) {
		yamlData := `
habit_id: "meditation"
value: true
status: "completed"
created_at: "2025-07-15 09:11:27"
`

		var entry HabitEntry
		err := yaml.Unmarshal([]byte(yamlData), &entry)
		require.NoError(t, err)

		assert.Equal(t, true, entry.Value)
	})

	t.Run("unmarshal string values", func(t *testing.T) {
		yamlData := `
habit_id: "routine"
value: "Completed morning routine"
status: "completed"
created_at: "2025-07-15 09:11:27"
`

		var entry HabitEntry
		err := yaml.Unmarshal([]byte(yamlData), &entry)
		require.NoError(t, err)

		assert.Equal(t, "Completed morning routine", entry.Value)
	})
}

func TestHabitEntry_YAMLRoundTrip(t *testing.T) {
	t.Run("round trip with time field value", func(t *testing.T) {
		// Create original entry with time field value
		timeValue := time.Date(0, 1, 1, 8, 30, 0, 0, time.UTC)
		original := HabitEntry{
			HabitID:   "wake_up",
			Value:     timeValue,
			Status:    EntryCompleted,
			CreatedAt: time.Date(2025, 7, 15, 9, 11, 27, 0, time.UTC),
			Notes:     "slept well",
		}

		// Marshal to YAML
		yamlData, err := yaml.Marshal(&original)
		require.NoError(t, err)

		// Unmarshal back
		var unmarshaled HabitEntry
		err = yaml.Unmarshal(yamlData, &unmarshaled)
		require.NoError(t, err)

		// Verify round trip
		assert.Equal(t, original.HabitID, unmarshaled.HabitID)
		assert.Equal(t, original.Status, unmarshaled.Status)
		assert.Equal(t, original.Notes, unmarshaled.Notes)

		// Check time value
		originalTime, ok1 := original.Value.(time.Time)
		unmarshaledTime, ok2 := unmarshaled.Value.(time.Time)
		require.True(t, ok1)
		require.True(t, ok2)
		assert.Equal(t, originalTime.Hour(), unmarshaledTime.Hour())
		assert.Equal(t, originalTime.Minute(), unmarshaledTime.Minute())

		// Timestamps should be equal within second precision (no subseconds in human format)
		assert.True(t, original.CreatedAt.Truncate(time.Second).Equal(unmarshaled.CreatedAt.Truncate(time.Second)))
	})

	t.Run("round trip with boolean value", func(t *testing.T) {
		original := CreateBooleanHabitEntry("meditation", true)
		original.Notes = "Great session"

		// Marshal to YAML
		yamlData, err := yaml.Marshal(&original)
		require.NoError(t, err)

		// Unmarshal back
		var unmarshaled HabitEntry
		err = yaml.Unmarshal(yamlData, &unmarshaled)
		require.NoError(t, err)

		// Verify round trip
		assert.Equal(t, original.HabitID, unmarshaled.HabitID)
		assert.Equal(t, original.Value, unmarshaled.Value)
		assert.Equal(t, original.Status, unmarshaled.Status)
		assert.Equal(t, original.Notes, unmarshaled.Notes)
	})
}

func TestHabitEntry_BackwardCompatibility(t *testing.T) {
	t.Run("parse legacy YAML with RFC3339 timestamps", func(t *testing.T) {
		// This simulates old entries.yml format
		legacyYAML := `
habit_id: "morning_meditation"
value: true
achievement_level: mini
notes: "Great session today"
created_at: "2024-01-01T10:00:00Z"
status: "completed"
`

		var entry HabitEntry
		err := yaml.Unmarshal([]byte(legacyYAML), &entry)
		require.NoError(t, err)

		assert.Equal(t, "morning_meditation", entry.HabitID)
		assert.Equal(t, true, entry.Value)
		assert.Equal(t, EntryCompleted, entry.Status)
		assert.Equal(t, "Great session today", entry.Notes)

		level, ok := entry.GetAchievementLevel()
		assert.True(t, ok)
		assert.Equal(t, AchievementMini, level)

		assert.Equal(t, 2024, entry.CreatedAt.Year())
		assert.Equal(t, 10, entry.CreatedAt.Hour())
	})

	t.Run("parse legacy time field with RFC3339", func(t *testing.T) {
		legacyYAML := `
habit_id: "wake_up"
value: "0000-01-01T08:30:00Z"
status: "completed"
created_at: "2024-01-01T10:00:00Z"
`

		var entry HabitEntry
		err := yaml.Unmarshal([]byte(legacyYAML), &entry)
		require.NoError(t, err)

		timeVal, ok := entry.Value.(time.Time)
		require.True(t, ok)
		assert.Equal(t, 8, timeVal.Hour())
		assert.Equal(t, 30, timeVal.Minute())
	})
}

func TestEntryLog_HumanReadableFormat(t *testing.T) {
	t.Run("marshal complete entry log with human-readable format", func(t *testing.T) {
		// Create entry log with various habit types
		entryLog := CreateEmptyEntryLog()

		dayEntry := DayEntry{
			Date: "2025-07-15",
			Habits: []HabitEntry{
				{
					HabitID:   "wake_up",
					Value:     time.Date(0, 1, 1, 8, 30, 0, 0, time.UTC),
					Status:    EntryCompleted,
					CreatedAt: time.Date(2025, 7, 15, 9, 11, 27, 0, time.UTC),
				},
				CreateBooleanHabitEntry("meditation", true),
			},
		}

		err := entryLog.AddDayEntry(dayEntry)
		require.NoError(t, err)

		// Marshal to YAML
		yamlData, err := yaml.Marshal(entryLog)
		require.NoError(t, err)

		yamlStr := string(yamlData)

		// Check structure matches specification
		assert.Contains(t, yamlStr, `version: 1.0.0`)
		assert.Contains(t, yamlStr, `date: "2025-07-15"`)
		assert.Contains(t, yamlStr, `habit_id: wake_up`)
		assert.Contains(t, yamlStr, `value: "08:30"`)
		assert.Contains(t, yamlStr, `created_at: "2025-07-15 09:11:27"`)
		assert.Contains(t, yamlStr, `status: completed`)

		// Should not contain legacy RFC3339 format
		assert.NotContains(t, yamlStr, "T08:30:00Z")
		assert.NotContains(t, yamlStr, "T09:11:27")
	})
}
