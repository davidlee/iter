package integration

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	initpkg "davidlee/iter/internal/init"
	"davidlee/iter/internal/models"
	"davidlee/iter/internal/parser"
	"davidlee/iter/internal/scoring"
	"davidlee/iter/internal/storage"
	"davidlee/iter/internal/ui"
)

// TestElasticGoalsEndToEnd verifies the complete elastic goals workflow:
// 1. Define elastic goals in goals.yml
// 2. Load and parse the goals
// 3. Collect entries with scoring
// 4. Save entries with achievement levels
// 5. Load and verify the saved data
func TestElasticGoalsEndToEnd(t *testing.T) {
	tempDir := t.TempDir()
	goalsFile := filepath.Join(tempDir, "goals.yml")
	entriesFile := filepath.Join(tempDir, "entries.yml")

	// Step 1: Create sample files with elastic goals
	initializer := initpkg.NewFileInitializer()
	err := initializer.EnsureConfigFiles(goalsFile, entriesFile)
	require.NoError(t, err)

	// Step 2: Load and verify goals
	goalParser := parser.NewGoalParser()
	schema, err := goalParser.LoadFromFile(goalsFile)
	require.NoError(t, err)
	assert.Len(t, schema.Goals, 4) // 2 simple + 2 elastic goals

	// Find elastic goals
	var exerciseDurationGoal, waterIntakeGoal *models.Goal
	for i := range schema.Goals {
		if schema.Goals[i].GoalType == models.ElasticGoal {
			switch schema.Goals[i].ID {
			case "exercise_duration":
				exerciseDurationGoal = &schema.Goals[i]
			case "water_intake":
				waterIntakeGoal = &schema.Goals[i]
			}
		}
	}

	require.NotNil(t, exerciseDurationGoal, "exercise duration goal should exist")
	require.NotNil(t, waterIntakeGoal, "water intake goal should exist")

	// Step 3: Test scoring engine with elastic goals
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
			result, err := scoringEngine.ScoreElasticGoal(exerciseDurationGoal, tc.value)
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
			result, err := scoringEngine.ScoreElasticGoal(waterIntakeGoal, tc.value)
			require.NoError(t, err, "scoring should succeed for value %v", tc.value)
			assert.Equal(t, tc.expectedLevel, result.AchievementLevel,
				"expected %s for value %v", tc.expectedLevel, tc.value)
		}
	})

	// Step 4: Test entry collection and storage simulation
	t.Run("entry collection and storage", func(t *testing.T) {
		// Create entry collector
		collector := ui.NewEntryCollector("checklists.yml")

		// Load goals manually (since we can't interact with UI in tests)
		collector.SetGoalsForTesting(schema.Goals)

		// Simulate user entries for elastic goals
		testEntries := map[string]interface{}{
			"exercise_duration": "45m", // Should be midi achievement
			"water_intake":      8.0,   // Should be maxi achievement
		}

		// Simulate collecting entries with scoring
		for goalID, value := range testEntries {
			var goal models.Goal
			for _, g := range schema.Goals {
				if g.ID == goalID {
					goal = g
					break
				}
			}

			if goal.GoalType == models.ElasticGoal {
				// Test scoring
				result, err := scoringEngine.ScoreElasticGoal(&goal, value)
				require.NoError(t, err)

				// Store in collector
				collector.SetEntryForTesting(goalID, value, &result.AchievementLevel, "Test notes")
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

		// Verify elastic goal entries
		goalEntryMap := make(map[string]models.GoalEntry)
		for _, goalEntry := range dayEntry.Goals {
			goalEntryMap[goalEntry.GoalID] = goalEntry
		}

		// Check exercise duration entry
		exerciseEntry, found := goalEntryMap["exercise_duration"]
		require.True(t, found, "exercise duration entry should exist")
		assert.Equal(t, "45m", exerciseEntry.Value)
		require.NotNil(t, exerciseEntry.AchievementLevel)
		assert.Equal(t, models.AchievementMidi, *exerciseEntry.AchievementLevel)

		// Check water intake entry
		waterEntry, found := goalEntryMap["water_intake"]
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

// TestBackwardsCompatibility ensures that elastic goals work alongside simple goals.
func TestBackwardsCompatibility(t *testing.T) {
	tempDir := t.TempDir()
	goalsFile := filepath.Join(tempDir, "goals.yml")
	entriesFile := filepath.Join(tempDir, "entries.yml")

	// Create mixed goals (simple + elastic)
	initializer := initpkg.NewFileInitializer()
	err := initializer.EnsureConfigFiles(goalsFile, entriesFile)
	require.NoError(t, err)

	// Load schema
	goalParser := parser.NewGoalParser()
	schema, err := goalParser.LoadFromFile(goalsFile)
	require.NoError(t, err)

	// Verify we have both simple and elastic goals
	simpleGoals := 0
	elasticGoals := 0
	for _, goal := range schema.Goals {
		switch goal.GoalType {
		case models.SimpleGoal:
			simpleGoals++
		case models.ElasticGoal:
			elasticGoals++
		}
	}
	assert.Equal(t, 2, simpleGoals, "should have 2 simple goals")
	assert.Equal(t, 2, elasticGoals, "should have 2 elastic goals")

	// Test that all goals can be processed
	collector := ui.NewEntryCollector("checklists.yml")
	collector.SetGoalsForTesting(schema.Goals)

	// Add entries for all goal types
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
	assert.Len(t, dayEntry.Goals, 4, "all 4 goals should be saved")

	// Verify each goal type
	goalEntryMap := make(map[string]models.GoalEntry)
	for _, goalEntry := range dayEntry.Goals {
		goalEntryMap[goalEntry.GoalID] = goalEntry
	}

	// Simple boolean goals
	morningExercise := goalEntryMap["morning_exercise"]
	assert.Equal(t, true, morningExercise.Value)
	assert.Nil(t, morningExercise.AchievementLevel) // Simple goals don't have achievement levels

	dailyReading := goalEntryMap["daily_reading"]
	assert.Equal(t, false, dailyReading.Value)
	assert.Nil(t, dailyReading.AchievementLevel)

	// Elastic goals
	exerciseDuration := goalEntryMap["exercise_duration"]
	assert.Equal(t, "30m", exerciseDuration.Value)
	require.NotNil(t, exerciseDuration.AchievementLevel)
	assert.Equal(t, models.AchievementMidi, *exerciseDuration.AchievementLevel)

	waterIntake := goalEntryMap["water_intake"]
	assert.Equal(t, 8.0, waterIntake.Value)
	require.NotNil(t, waterIntake.AchievementLevel)
	assert.Equal(t, models.AchievementMaxi, *waterIntake.AchievementLevel)
}
