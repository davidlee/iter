package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/davidlee/vice/internal/models"
	"github.com/davidlee/vice/internal/scoring"
	"github.com/davidlee/vice/internal/ui/entry"
)

// AIDEV-NOTE: T016-integration-tests; critical for validating habit configuration change resilience
// TestHabitConfigurationChanges tests scenarios where users modify habit configurations
// and ensure the system handles these changes gracefully without breaking.
// These tests cover the exact user scenario from T016 and prevent similar regressions.
func TestHabitConfigurationChanges(t *testing.T) {
	t.Run("boolean_to_numeric_automatic_scoring", func(t *testing.T) {
		// This tests the exact scenario reported by the user:
		// Habit was originally: simple boolean with manual scoring
		// User changed to: simple numeric with automatic scoring

		// Create the "after" habit configuration (what user has now)
		habit := models.Habit{
			Title:     "Do 10 push-ups",
			ID:        "do_10_push_ups",
			HabitType: models.SimpleHabit,
			FieldType: models.FieldType{
				Type: models.UnsignedIntFieldType,
				Unit: "push-ups",
			},
			ScoringType: models.AutomaticScoring,
			Criteria: &models.Criteria{
				Description: "Habit achieved when value > 10.0 push-ups",
				Condition: &models.Condition{
					GreaterThan: func() *float64 { v := 10.0; return &v }(),
				},
			},
			Prompt: "How many push-ups did you do?",
		}

		// Test that entry collection works with the new configuration
		factory := entry.NewEntryFieldInputFactory()
		scoringEngine := scoring.NewEngine()
		flow := entry.NewSimpleHabitCollectionFlow(factory, scoringEngine)

		// Test values that should pass and fail
		testCases := []struct {
			value               interface{}
			expectedAchievement models.AchievementLevel
			description         string
		}{
			{5, models.AchievementNone, "below threshold"},
			{10, models.AchievementNone, "at threshold (= 10, but criteria is > 10)"},
			{15, models.AchievementMini, "above threshold"},
		}

		for _, tc := range testCases {
			t.Run(tc.description, func(t *testing.T) {
				result, err := flow.CollectEntryDirectly(habit, tc.value, "test notes", nil)
				require.NoError(t, err, "Entry collection should not fail for numeric simple habit with automatic scoring")
				require.NotNil(t, result, "Result should not be nil")
				assert.Equal(t, tc.expectedAchievement, *result.AchievementLevel, "Achievement level should match expected")
			})
		}
	})

	t.Run("manual_to_automatic_scoring_conversion", func(t *testing.T) {
		// Test converting from manual to automatic scoring (different field types)

		testCases := []struct {
			name      string
			fieldType string
			threshold float64
			testValue interface{}
			expected  models.AchievementLevel
		}{
			{
				name:      "numeric_manual_to_automatic",
				fieldType: models.UnsignedIntFieldType,
				threshold: 100,
				testValue: 150,
				expected:  models.AchievementMini,
			},
			{
				name:      "boolean_manual_to_automatic",
				fieldType: models.BooleanFieldType,
				threshold: 1, // Not used for boolean, but criteria needs a value
				testValue: true,
				expected:  models.AchievementMini,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				var criteria *models.Criteria
				if tc.fieldType == models.BooleanFieldType {
					trueVal := true
					criteria = &models.Criteria{
						Condition: &models.Condition{
							Equals: &trueVal,
						},
					}
				} else {
					criteria = &models.Criteria{
						Condition: &models.Condition{
							GreaterThanOrEqual: &tc.threshold,
						},
					}
				}

				habit := models.Habit{
					Title:       "Test Habit",
					ID:          "test_habit",
					HabitType:   models.SimpleHabit,
					FieldType:   models.FieldType{Type: tc.fieldType},
					ScoringType: models.AutomaticScoring,
					Criteria:    criteria,
				}

				factory := entry.NewEntryFieldInputFactory()
				scoringEngine := scoring.NewEngine()
				flow := entry.NewSimpleHabitCollectionFlow(factory, scoringEngine)

				result, err := flow.CollectEntryDirectly(habit, tc.testValue, "", nil)
				require.NoError(t, err, "Conversion from manual to automatic scoring should work")
				require.NotNil(t, result)
				assert.Equal(t, tc.expected, *result.AchievementLevel)
			})
		}
	})

	t.Run("different_habit_types_with_automatic_scoring", func(t *testing.T) {
		// Ensure that all habit types can use automatic scoring appropriately

		factory := entry.NewEntryFieldInputFactory()
		scoringEngine := scoring.NewEngine()

		t.Run("simple_habit_automatic", func(t *testing.T) {
			habit := models.Habit{
				ID:          "simple_auto",
				HabitType:   models.SimpleHabit,
				FieldType:   models.FieldType{Type: models.UnsignedIntFieldType},
				ScoringType: models.AutomaticScoring,
				Criteria: &models.Criteria{
					Condition: &models.Condition{
						GreaterThanOrEqual: func() *float64 { v := 10.0; return &v }(),
					},
				},
			}

			flow := entry.NewSimpleHabitCollectionFlow(factory, scoringEngine)
			result, err := flow.CollectEntryDirectly(habit, 15, "", nil)
			require.NoError(t, err, "Simple habit with automatic scoring should work")
			assert.Equal(t, models.AchievementMini, *result.AchievementLevel)
		})

		t.Run("elastic_habit_automatic", func(t *testing.T) {
			habit := models.Habit{
				ID:          "elastic_auto",
				HabitType:   models.ElasticHabit,
				FieldType:   models.FieldType{Type: models.UnsignedIntFieldType},
				ScoringType: models.AutomaticScoring,
				MiniCriteria: &models.Criteria{
					Condition: &models.Condition{
						GreaterThanOrEqual: func() *float64 { v := 5.0; return &v }(),
					},
				},
				MidiCriteria: &models.Criteria{
					Condition: &models.Condition{
						GreaterThanOrEqual: func() *float64 { v := 10.0; return &v }(),
					},
				},
				MaxiCriteria: &models.Criteria{
					Condition: &models.Condition{
						GreaterThanOrEqual: func() *float64 { v := 15.0; return &v }(),
					},
				},
			}

			flow := entry.NewElasticHabitCollectionFlow(factory, scoringEngine)
			result, err := flow.CollectEntryDirectly(habit, 12, "", nil)
			require.NoError(t, err, "Elastic habit with automatic scoring should work")
			assert.Equal(t, models.AchievementMidi, *result.AchievementLevel)
		})
	})
}
