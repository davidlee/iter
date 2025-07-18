package scoring

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/davidlee/vice/internal/models"
)

func TestEngine_ScoreSimpleHabit(t *testing.T) {
	engine := NewEngine()

	t.Run("numeric simple habit with criteria", func(t *testing.T) {
		habit := createTestSimpleHabit(models.UnsignedIntFieldType, 10)

		testCases := []struct {
			value         interface{}
			expectedLevel models.AchievementLevel
			expectedMini  bool
		}{
			{5, models.AchievementNone, false}, // Below threshold
			{10, models.AchievementMini, true}, // At threshold
			{15, models.AchievementMini, true}, // Above threshold
		}

		for _, tc := range testCases {
			result, err := engine.ScoreSimpleHabit(&habit, tc.value)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedLevel, result.AchievementLevel)
			assert.Equal(t, tc.expectedMini, result.MetMini)
			assert.False(t, result.MetMidi) // Simple habits don't have midi
			assert.False(t, result.MetMaxi) // Simple habits don't have maxi
		}
	})

	t.Run("boolean simple habit", func(t *testing.T) {
		habit := createTestSimpleBooleanHabit()

		// Test true value
		result, err := engine.ScoreSimpleHabit(&habit, true)
		require.NoError(t, err)
		assert.Equal(t, models.AchievementMini, result.AchievementLevel)
		assert.True(t, result.MetMini)

		// Test false value
		result, err = engine.ScoreSimpleHabit(&habit, false)
		require.NoError(t, err)
		assert.Equal(t, models.AchievementNone, result.AchievementLevel)
		assert.False(t, result.MetMini)
	})

	t.Run("error cases", func(t *testing.T) {
		// Test nil habit
		_, err := engine.ScoreSimpleHabit(nil, 5)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "habit cannot be nil")

		// Test non-simple habit
		elasticHabit := createTestElasticHabit(models.UnsignedIntFieldType, 5, 10, 15)
		_, err = engine.ScoreSimpleHabit(elasticHabit, 5)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "is not a simple habit")

		// Test manual scoring habit
		manualHabit := createTestSimpleHabit(models.UnsignedIntFieldType, 10)
		manualHabit.ScoringType = models.ManualScoring
		_, err = engine.ScoreSimpleHabit(&manualHabit, 5)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not require automatic scoring")

		// Test habit without criteria
		noCriteriaHabit := createTestSimpleHabit(models.UnsignedIntFieldType, 10)
		noCriteriaHabit.Criteria = nil
		_, err = engine.ScoreSimpleHabit(&noCriteriaHabit, 5)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "has no criteria for automatic scoring")
	})
}

func TestEngine_ScoreElasticHabit(t *testing.T) {
	engine := NewEngine()

	t.Run("numeric habit with all achievement levels", func(t *testing.T) {
		habit := createTestElasticHabit(models.UnsignedIntFieldType, 5000, 10000, 15000)

		testCases := []struct {
			value         interface{}
			expectedLevel models.AchievementLevel
			expectedMini  bool
			expectedMidi  bool
			expectedMaxi  bool
		}{
			{3000, models.AchievementNone, false, false, false},
			{5000, models.AchievementMini, true, false, false},
			{7500, models.AchievementMini, true, false, false},
			{10000, models.AchievementMidi, true, true, false},
			{12500, models.AchievementMidi, true, true, false},
			{15000, models.AchievementMaxi, true, true, true},
			{20000, models.AchievementMaxi, true, true, true},
		}

		for _, tc := range testCases {
			result, err := engine.ScoreElasticHabit(habit, tc.value)
			require.NoError(t, err)
			require.NotNil(t, result)

			assert.Equal(t, tc.expectedLevel, result.AchievementLevel, "Value: %v", tc.value)
			assert.Equal(t, tc.expectedMini, result.MetMini, "Value: %v", tc.value)
			assert.Equal(t, tc.expectedMidi, result.MetMidi, "Value: %v", tc.value)
			assert.Equal(t, tc.expectedMaxi, result.MetMaxi, "Value: %v", tc.value)
		}
	})

	t.Run("duration habit with string and numeric values", func(t *testing.T) {
		habit := createTestElasticHabit(models.DurationFieldType, 15, 30, 60) // 15, 30, 60 minutes

		testCases := []struct {
			value         interface{}
			expectedLevel models.AchievementLevel
		}{
			{10, models.AchievementNone},
			{15.0, models.AchievementMini},
			{"20", models.AchievementMini},
			{30, models.AchievementMidi},
			{"45", models.AchievementMidi},
			{60.0, models.AchievementMaxi},
			{"1h30m", models.AchievementMaxi},   // 90 minutes
			{"1:30:00", models.AchievementMaxi}, // 90 minutes
		}

		for _, tc := range testCases {
			result, err := engine.ScoreElasticHabit(habit, tc.value)
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tc.expectedLevel, result.AchievementLevel, "Value: %v", tc.value)
		}
	})

	t.Run("boolean habit", func(t *testing.T) {
		habit := createTestBooleanElasticHabit()

		testCases := []struct {
			value         interface{}
			expectedLevel models.AchievementLevel
		}{
			{false, models.AchievementNone},
			{true, models.AchievementMini},
			{"false", models.AchievementNone},
			{"true", models.AchievementMini},
		}

		for _, tc := range testCases {
			result, err := engine.ScoreElasticHabit(habit, tc.value)
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tc.expectedLevel, result.AchievementLevel, "Value: %v", tc.value)
		}
	})

	t.Run("text habit with length-based criteria", func(t *testing.T) {
		habit := createTestTextElasticHabit()

		testCases := []struct {
			value         interface{}
			expectedLevel models.AchievementLevel
		}{
			{"", models.AchievementNone},
			{"short", models.AchievementMini},                                                       // 5 chars >= 5
			{"medium length text", models.AchievementMidi},                                          // 18 chars >= 15
			{"this is a very long text that exceeds the maximum threshold", models.AchievementMaxi}, // 63 chars >= 30
		}

		for _, tc := range testCases {
			result, err := engine.ScoreElasticHabit(habit, tc.value)
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tc.expectedLevel, result.AchievementLevel, "Value: %v", tc.value)
		}
	})

	t.Run("error cases", func(t *testing.T) {
		habit := createTestElasticHabit(models.UnsignedIntFieldType, 5000, 10000, 15000)

		// Nil habit
		_, err := engine.ScoreElasticHabit(nil, 1000)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "habit cannot be nil")

		// Non-elastic habit
		simpleHabit := &models.Habit{
			ID:        "simple_habit",
			HabitType: models.SimpleHabit,
		}
		_, err = engine.ScoreElasticHabit(simpleHabit, 1000)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "is not an elastic habit")

		// Manual scoring habit
		manualHabit := &models.Habit{
			ID:          "manual_habit",
			HabitType:   models.ElasticHabit,
			ScoringType: models.ManualScoring,
		}
		_, err = engine.ScoreElasticHabit(manualHabit, 1000)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not require automatic scoring")

		// Nil value
		_, err = engine.ScoreElasticHabit(habit, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "value cannot be nil")
	})
}

func TestEngine_ConvertValueForEvaluation(t *testing.T) {
	engine := NewEngine()

	t.Run("numeric conversions", func(t *testing.T) {
		testCases := []struct {
			value     interface{}
			expected  float64
			fieldType string
		}{
			{42, 42.0, models.UnsignedIntFieldType},
			{42.5, 42.5, models.DecimalFieldType},
			{"123", 123.0, models.UnsignedDecimalFieldType},
			{uint64(999), 999.0, models.UnsignedIntFieldType},
		}

		for _, tc := range testCases {
			result, err := engine.convertValueForEvaluation(tc.value, tc.fieldType)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		}
	})

	t.Run("duration conversions", func(t *testing.T) {
		testCases := []struct {
			value    interface{}
			expected float64
		}{
			{30, 30.0},         // 30 minutes
			{"45", 45.0},       // 45 minutes
			{"1h", 60.0},       // 1 hour = 60 minutes
			{"1h30m", 90.0},    // 1.5 hours = 90 minutes
			{"2:30:00", 150.0}, // 2.5 hours = 150 minutes
		}

		for _, tc := range testCases {
			result, err := engine.convertValueForEvaluation(tc.value, models.DurationFieldType)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		}
	})

	t.Run("time conversions", func(t *testing.T) {
		testCases := []struct {
			value    interface{}
			expected float64
		}{
			{"09:00", 540.0},  // 9 AM = 540 minutes since midnight
			{"14:30", 870.0},  // 2:30 PM = 870 minutes since midnight
			{"23:59", 1439.0}, // 11:59 PM = 1439 minutes since midnight
		}

		for _, tc := range testCases {
			result, err := engine.convertValueForEvaluation(tc.value, models.TimeFieldType)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		}
	})

	t.Run("boolean conversions", func(t *testing.T) {
		testCases := []struct {
			value    interface{}
			expected bool
		}{
			{true, true},
			{false, false},
			{"true", true},
			{"false", false},
			{"1", true},
			{"0", false},
		}

		for _, tc := range testCases {
			result, err := engine.convertValueForEvaluation(tc.value, models.BooleanFieldType)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		}
	})

	t.Run("text conversions", func(t *testing.T) {
		testCases := []struct {
			value    interface{}
			expected string
		}{
			{"hello", "hello"},
			{123, "123"},
			{true, "true"},
		}

		for _, tc := range testCases {
			result, err := engine.convertValueForEvaluation(tc.value, models.TextFieldType)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		}
	})
}

func TestEngine_EvaluateNumericCondition(t *testing.T) {
	engine := NewEngine()

	t.Run("greater than", func(t *testing.T) {
		threshold := 10.0
		condition := &models.Condition{GreaterThan: &threshold}

		assert.True(t, mustEvaluateNumeric(t, engine, 15.0, condition))
		assert.False(t, mustEvaluateNumeric(t, engine, 10.0, condition))
		assert.False(t, mustEvaluateNumeric(t, engine, 5.0, condition))
	})

	t.Run("greater than or equal", func(t *testing.T) {
		threshold := 10.0
		condition := &models.Condition{GreaterThanOrEqual: &threshold}

		assert.True(t, mustEvaluateNumeric(t, engine, 15.0, condition))
		assert.True(t, mustEvaluateNumeric(t, engine, 10.0, condition))
		assert.False(t, mustEvaluateNumeric(t, engine, 5.0, condition))
	})

	t.Run("less than", func(t *testing.T) {
		threshold := 10.0
		condition := &models.Condition{LessThan: &threshold}

		assert.False(t, mustEvaluateNumeric(t, engine, 15.0, condition))
		assert.False(t, mustEvaluateNumeric(t, engine, 10.0, condition))
		assert.True(t, mustEvaluateNumeric(t, engine, 5.0, condition))
	})

	t.Run("range condition", func(t *testing.T) {
		rangeCondition := &models.RangeCondition{
			Min: 5.0,
			Max: 15.0,
		}
		condition := &models.Condition{Range: rangeCondition}

		assert.False(t, mustEvaluateNumeric(t, engine, 3.0, condition))
		assert.True(t, mustEvaluateNumeric(t, engine, 5.0, condition))
		assert.True(t, mustEvaluateNumeric(t, engine, 10.0, condition))
		assert.True(t, mustEvaluateNumeric(t, engine, 15.0, condition))
		assert.False(t, mustEvaluateNumeric(t, engine, 20.0, condition))
	})
}

func TestEngine_EvaluateTimeCondition(t *testing.T) {
	engine := NewEngine()

	t.Run("before time", func(t *testing.T) {
		condition := &models.Condition{Before: "12:00"}

		// 11:00 AM (660 minutes) should be before 12:00 PM (720 minutes)
		assert.True(t, mustEvaluateTime(t, engine, 660.0, condition))
		// 13:00 PM (780 minutes) should not be before 12:00 PM
		assert.False(t, mustEvaluateTime(t, engine, 780.0, condition))
	})

	t.Run("after time", func(t *testing.T) {
		condition := &models.Condition{After: "12:00"}

		// 11:00 AM should not be after 12:00 PM
		assert.False(t, mustEvaluateTime(t, engine, 660.0, condition))
		// 13:00 PM should be after 12:00 PM
		assert.True(t, mustEvaluateTime(t, engine, 780.0, condition))
	})
}

func TestEngine_ParseDurationToMinutes(t *testing.T) {
	engine := NewEngine()

	testCases := []struct {
		input    string
		expected float64
	}{
		{"30", 30.0},
		{"1h", 60.0},
		{"1h30m", 90.0},
		{"2h15m30s", 135.5},
		{"0:45:00", 45.0},
		{"1:30:30", 90.5},
	}

	for _, tc := range testCases {
		result, err := engine.parseDurationToMinutes(tc.input)
		require.NoError(t, err, "Input: %s", tc.input)
		assert.Equal(t, tc.expected, result, "Input: %s", tc.input)
	}
}

func TestEngine_ParseTimeToMinutes(t *testing.T) {
	engine := NewEngine()

	testCases := []struct {
		input    string
		expected float64
	}{
		{"00:00", 0.0},
		{"09:00", 540.0},
		{"12:30", 750.0},
		{"23:59", 1439.0},
	}

	for _, tc := range testCases {
		result, err := engine.parseTimeToMinutes(tc.input)
		require.NoError(t, err, "Input: %s", tc.input)
		assert.Equal(t, tc.expected, result, "Input: %s", tc.input)
	}

	// Test invalid formats
	invalidInputs := []string{"25:00", "12:60", "invalid", "12:30:45"}
	for _, input := range invalidInputs {
		_, err := engine.parseTimeToMinutes(input)
		assert.Error(t, err, "Input: %s", input)
	}
}

// Helper functions for testing

func createTestElasticHabit(fieldType string, mini, midi, maxi float64) *models.Habit {
	return &models.Habit{
		ID:        "test_elastic_habit",
		HabitType: models.ElasticHabit,
		FieldType: models.FieldType{
			Type: fieldType,
		},
		ScoringType: models.AutomaticScoring,
		MiniCriteria: &models.Criteria{
			Condition: &models.Condition{
				GreaterThanOrEqual: &mini,
			},
		},
		MidiCriteria: &models.Criteria{
			Condition: &models.Condition{
				GreaterThanOrEqual: &midi,
			},
		},
		MaxiCriteria: &models.Criteria{
			Condition: &models.Condition{
				GreaterThanOrEqual: &maxi,
			},
		},
	}
}

func createTestBooleanElasticHabit() *models.Habit {
	trueValue := true
	return &models.Habit{
		ID:        "test_boolean_habit",
		HabitType: models.ElasticHabit,
		FieldType: models.FieldType{
			Type: models.BooleanFieldType,
		},
		ScoringType: models.AutomaticScoring,
		MiniCriteria: &models.Criteria{
			Condition: &models.Condition{
				Equals: &trueValue,
			},
		},
	}
}

func createTestTextElasticHabit() *models.Habit {
	miniLength := 5.0
	midiLength := 15.0
	maxiLength := 30.0
	return &models.Habit{
		ID:        "test_text_habit",
		HabitType: models.ElasticHabit,
		FieldType: models.FieldType{
			Type: models.TextFieldType,
		},
		ScoringType: models.AutomaticScoring,
		MiniCriteria: &models.Criteria{
			Condition: &models.Condition{
				GreaterThanOrEqual: &miniLength,
			},
		},
		MidiCriteria: &models.Criteria{
			Condition: &models.Condition{
				GreaterThanOrEqual: &midiLength,
			},
		},
		MaxiCriteria: &models.Criteria{
			Condition: &models.Condition{
				GreaterThanOrEqual: &maxiLength,
			},
		},
	}
}

func createTestSimpleHabit(fieldType string, threshold float64) models.Habit {
	return models.Habit{
		ID:        "test_simple_habit",
		HabitType: models.SimpleHabit,
		FieldType: models.FieldType{
			Type: fieldType,
		},
		ScoringType: models.AutomaticScoring,
		Criteria: &models.Criteria{
			Condition: &models.Condition{
				GreaterThanOrEqual: &threshold,
			},
		},
	}
}

func createTestSimpleBooleanHabit() models.Habit {
	trueValue := true
	return models.Habit{
		ID:        "test_simple_boolean_habit",
		HabitType: models.SimpleHabit,
		FieldType: models.FieldType{
			Type: models.BooleanFieldType,
		},
		ScoringType: models.AutomaticScoring,
		Criteria: &models.Criteria{
			Condition: &models.Condition{
				Equals: &trueValue,
			},
		},
	}
}

func mustEvaluateNumeric(t *testing.T, engine *Engine, value float64, condition *models.Condition) bool {
	result, err := engine.evaluateNumericCondition(value, condition)
	require.NoError(t, err)
	return result
}

func mustEvaluateTime(t *testing.T, engine *Engine, value float64, condition *models.Condition) bool {
	result, err := engine.evaluateTimeCondition(value, condition)
	require.NoError(t, err)
	return result
}
