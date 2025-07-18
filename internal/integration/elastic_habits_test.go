package integration

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	initpkg "github.com/davidlee/vice/internal/init"
	"github.com/davidlee/vice/internal/models"
	"github.com/davidlee/vice/internal/parser"
	"github.com/davidlee/vice/internal/scoring"
	"github.com/davidlee/vice/internal/storage"
	"github.com/davidlee/vice/internal/ui"
)

// TestElasticHabitsEndToEnd verifies the complete elastic habits workflow:
// 1. Define elastic habits in habits.yml
// 2. Load and parse the habits
// 3. Collect entries with scoring
// 4. Save entries with achievement levels
// 5. Load and verify the saved data
func TestElasticHabitsEndToEnd(t *testing.T) {
	tempDir := t.TempDir()
	habitsFile := filepath.Join(tempDir, "habits.yml")
	entriesFile := filepath.Join(tempDir, "entries.yml")

	// Step 1: Create sample files with elastic habits
	initializer := initpkg.NewFileInitializer()
	err := initializer.EnsureConfigFiles(habitsFile, entriesFile)
	require.NoError(t, err)

	// Step 2: Load and verify habits
	habitParser := parser.NewHabitParser()
	schema, err := habitParser.LoadFromFile(habitsFile)
	require.NoError(t, err)
	assert.Len(t, schema.Habits, 4) // 2 simple + 2 elastic habits

	// Find elastic habits
	var exerciseDurationHabit, waterIntakeHabit *models.Habit
	for i := range schema.Habits {
		if schema.Habits[i].HabitType == models.ElasticHabit {
			switch schema.Habits[i].ID {
			case "exercise_duration":
				exerciseDurationHabit = &schema.Habits[i]
			case "water_intake":
				waterIntakeHabit = &schema.Habits[i]
			}
		}
	}

	require.NotNil(t, exerciseDurationHabit, "exercise duration habit should exist")
	require.NotNil(t, waterIntakeHabit, "water intake habit should exist")

	// Step 3: Test scoring engine with elastic habits
	scoringEngine := scoring.NewEngine()

	// Test exercise duration scoring (duration field)
	t.Run("exercise duration scoring", func(t *testing.T) {
		// Test different duration values
		testCases := []struct {
			value         interface{}
			expectedLevel models.AchievementLevel
		}{
			{"10m", models.AchievementNone},   // Below mini (15m)
			{"20m", models.AchievementMini},   // Between mini (15m) and midi (30m)
			{"45m", models.AchievementMidi},   // Between midi (30m) and maxi (60m)
			{"1h30m", models.AchievementMaxi}, // Above maxi (60m)
		}

		for _, tc := range testCases {
			result, err := scoringEngine.ScoreElasticHabit(exerciseDurationHabit, tc.value)
			require.NoError(t, err, "scoring should succeed for value %v", tc.value)
			assert.Equal(t, tc.expectedLevel, result.AchievementLevel,
				"expected %s for value %v", tc.expectedLevel, tc.value)
		}
	})

	// Test water intake scoring (numeric field with units)
	t.Run("water intake scoring", func(t *testing.T) {
		testCases := []struct {
			value         interface{}
			expectedLevel models.AchievementLevel
		}{
			{2.0, models.AchievementNone},  // Below mini (4 glasses)
			{5.0, models.AchievementMini},  // Between mini (4) and midi (6)
			{7.0, models.AchievementMidi},  // Between midi (6) and maxi (8)
			{10.0, models.AchievementMaxi}, // Above maxi (8)
		}

		for _, tc := range testCases {
			result, err := scoringEngine.ScoreElasticHabit(waterIntakeHabit, tc.value)
			require.NoError(t, err, "scoring should succeed for value %v", tc.value)
			assert.Equal(t, tc.expectedLevel, result.AchievementLevel,
				"expected %s for value %v", tc.expectedLevel, tc.value)
		}
	})

	// Step 4: Test entry collection and storage simulation
	t.Run("entry collection and storage", func(t *testing.T) {
		// Create entry collector
		collector := ui.NewEntryCollector("checklists.yml")

		// Load habits manually (since we can't interact with UI in tests)
		collector.SetHabitsForTesting(schema.Habits)

		// Simulate user entries for elastic habits
		testEntries := map[string]interface{}{
			"exercise_duration": "45m", // Should be midi achievement
			"water_intake":      8.0,   // Should be maxi achievement
		}

		// Simulate collecting entries with scoring
		for habitID, value := range testEntries {
			var habit models.Habit
			for _, g := range schema.Habits {
				if g.ID == habitID {
					habit = g
					break
				}
			}

			if habit.HabitType == models.ElasticHabit {
				// Test scoring
				result, err := scoringEngine.ScoreElasticHabit(&habit, value)
				require.NoError(t, err)

				// Store in collector
				collector.SetEntryForTesting(habitID, value, &result.AchievementLevel, "Test notes")
			}
		}

		// Save entries
		err := collector.SaveEntriesForTesting(entriesFile)
		require.NoError(t, err)

		// Step 5: Verify saved data
		entryStorage := storage.NewEntryStorage()
		entryLog, err := entryStorage.LoadFromFile(entriesFile)
		require.NoError(t, err)

		today := time.Now().Format("2006-01-02")
		dayEntry, found := entryLog.GetDayEntry(today)
		require.True(t, found, "today's entry should exist")

		// Verify elastic habit entries
		habitEntryMap := make(map[string]models.HabitEntry)
		for _, habitEntry := range dayEntry.Habits {
			habitEntryMap[habitEntry.HabitID] = habitEntry
		}

		// Check exercise duration entry
		exerciseEntry, found := habitEntryMap["exercise_duration"]
		require.True(t, found, "exercise duration entry should exist")
		assert.Equal(t, "45m", exerciseEntry.Value)
		require.NotNil(t, exerciseEntry.AchievementLevel)
		assert.Equal(t, models.AchievementMidi, *exerciseEntry.AchievementLevel)

		// Check water intake entry
		waterEntry, found := habitEntryMap["water_intake"]
		require.True(t, found, "water intake entry should exist")
		assert.Equal(t, 8.0, waterEntry.Value)
		require.NotNil(t, waterEntry.AchievementLevel)
		assert.Equal(t, models.AchievementMaxi, *waterEntry.AchievementLevel)
	})

	// Step 6: Test loading existing entries and updating
	t.Run("load and update existing entries", func(t *testing.T) {
		collector := ui.NewEntryCollector("checklists.yml")

		// Load existing entries
		err := collector.LoadExistingEntriesForTesting(entriesFile)
		require.NoError(t, err)

		// Verify loaded values
		entries := collector.GetEntriesForTesting()
		assert.Equal(t, "45m", entries["exercise_duration"])
		assert.Equal(t, 8.0, entries["water_intake"])

		// Verify loaded achievement levels
		achievements := collector.GetAchievementsForTesting()
		require.NotNil(t, achievements["exercise_duration"])
		assert.Equal(t, models.AchievementMidi, *achievements["exercise_duration"])
		require.NotNil(t, achievements["water_intake"])
		assert.Equal(t, models.AchievementMaxi, *achievements["water_intake"])
	})
}

// TestBackwardsCompatibility ensures that elastic habits work alongside simple habits.
func TestBackwardsCompatibility(t *testing.T) {
	tempDir := t.TempDir()
	habitsFile := filepath.Join(tempDir, "habits.yml")
	entriesFile := filepath.Join(tempDir, "entries.yml")

	// Create mixed habits (simple + elastic)
	initializer := initpkg.NewFileInitializer()
	err := initializer.EnsureConfigFiles(habitsFile, entriesFile)
	require.NoError(t, err)

	// Load schema
	habitParser := parser.NewHabitParser()
	schema, err := habitParser.LoadFromFile(habitsFile)
	require.NoError(t, err)

	// Verify we have both simple and elastic habits
	simpleHabits := 0
	elasticHabits := 0
	for _, habit := range schema.Habits {
		switch habit.HabitType {
		case models.SimpleHabit:
			simpleHabits++
		case models.ElasticHabit:
			elasticHabits++
		}
	}
	assert.Equal(t, 2, simpleHabits, "should have 2 simple habits")
	assert.Equal(t, 2, elasticHabits, "should have 2 elastic habits")

	// Test that all habits can be processed
	collector := ui.NewEntryCollector("checklists.yml")
	collector.SetHabitsForTesting(schema.Habits)

	// Add entries for all habit types
	collector.SetEntryForTesting("morning_exercise", true, nil, "Great workout")
	collector.SetEntryForTesting("daily_reading", false, nil, "Too busy today")

	exerciseLevelPtr := models.AchievementMidi
	collector.SetEntryForTesting("exercise_duration", "30m", &exerciseLevelPtr, "Perfect timing")

	waterLevelPtr := models.AchievementMaxi
	collector.SetEntryForTesting("water_intake", 8.0, &waterLevelPtr, "Stayed hydrated")

	// Save and verify
	err = collector.SaveEntriesForTesting(entriesFile)
	require.NoError(t, err)

	// Load and verify all entries were saved correctly
	entryStorage := storage.NewEntryStorage()
	entryLog, err := entryStorage.LoadFromFile(entriesFile)
	require.NoError(t, err)

	today := time.Now().Format("2006-01-02")
	dayEntry, found := entryLog.GetDayEntry(today)
	require.True(t, found)
	assert.Len(t, dayEntry.Habits, 4, "all 4 habits should be saved")

	// Verify each habit type
	habitEntryMap := make(map[string]models.HabitEntry)
	for _, habitEntry := range dayEntry.Habits {
		habitEntryMap[habitEntry.HabitID] = habitEntry
	}

	// Simple boolean habits
	morningExercise := habitEntryMap["morning_exercise"]
	assert.Equal(t, true, morningExercise.Value)
	assert.Nil(t, morningExercise.AchievementLevel) // Simple habits don't have achievement levels

	dailyReading := habitEntryMap["daily_reading"]
	assert.Equal(t, false, dailyReading.Value)
	assert.Nil(t, dailyReading.AchievementLevel)

	// Elastic habits
	exerciseDuration := habitEntryMap["exercise_duration"]
	assert.Equal(t, "30m", exerciseDuration.Value)
	require.NotNil(t, exerciseDuration.AchievementLevel)
	assert.Equal(t, models.AchievementMidi, *exerciseDuration.AchievementLevel)

	waterIntake := habitEntryMap["water_intake"]
	assert.Equal(t, 8.0, waterIntake.Value)
	require.NotNil(t, waterIntake.AchievementLevel)
	assert.Equal(t, models.AchievementMaxi, *waterIntake.AchievementLevel)
}
