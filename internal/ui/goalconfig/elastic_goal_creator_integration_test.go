package goalconfig

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"davidlee/vice/internal/models"
)

// TestElasticGoalCreator_Integration_AllCombinations tests all field type + scoring type combinations for elastic goals
func TestElasticGoalCreator_Integration_AllCombinations(t *testing.T) {
	tests := []struct {
		name              string
		fieldType         string
		scoringType       models.ScoringType
		testData          TestElasticGoalData
		expectCriteria    bool
		expectedFieldType string
	}{
		// Text field combinations (elastic goals can use text for subjective tracking)
		{
			name:        "Text + Manual (multiline)",
			fieldType:   models.TextFieldType,
			scoringType: models.ManualScoring,
			testData: TestElasticGoalData{
				FieldType:     models.TextFieldType,
				ScoringType:   models.ManualScoring,
				MultilineText: true,
				Prompt:        "How was your exercise intensity today?",
			},
			expectCriteria:    false,
			expectedFieldType: models.TextFieldType,
		},
		{
			name:        "Text + Manual (single line)",
			fieldType:   models.TextFieldType,
			scoringType: models.ManualScoring,
			testData: TestElasticGoalData{
				FieldType:     models.TextFieldType,
				ScoringType:   models.ManualScoring,
				MultilineText: false,
				Prompt:        "Rate your energy level",
			},
			expectCriteria:    false,
			expectedFieldType: models.TextFieldType,
		},

		// Numeric field combinations with three-tier criteria
		{
			name:        "UnsignedInt + Manual",
			fieldType:   "numeric",
			scoringType: models.ManualScoring,
			testData: TestElasticGoalData{
				FieldType:      "numeric",
				NumericSubtype: models.UnsignedIntFieldType,
				Unit:           "minutes",
				ScoringType:    models.ManualScoring,
				Prompt:         "How many minutes did you exercise?",
			},
			expectCriteria:    false,
			expectedFieldType: models.UnsignedIntFieldType,
		},
		{
			name:        "UnsignedInt + Automatic (three-tier)",
			fieldType:   "numeric",
			scoringType: models.AutomaticScoring,
			testData: TestElasticGoalData{
				FieldType:         "numeric",
				NumericSubtype:    models.UnsignedIntFieldType,
				Unit:              "minutes",
				ScoringType:       models.AutomaticScoring,
				Prompt:            "How many minutes did you exercise?",
				MiniCriteriaValue: "15",
				MidiCriteriaValue: "30",
				MaxiCriteriaValue: "60",
			},
			expectCriteria:    true,
			expectedFieldType: models.UnsignedIntFieldType,
		},
		{
			name:        "UnsignedDecimal + Automatic (three-tier)",
			fieldType:   "numeric",
			scoringType: models.AutomaticScoring,
			testData: TestElasticGoalData{
				FieldType:         "numeric",
				NumericSubtype:    models.UnsignedDecimalFieldType,
				Unit:              "km",
				ScoringType:       models.AutomaticScoring,
				Prompt:            "How far did you run?",
				MiniCriteriaValue: "2.0",
				MidiCriteriaValue: "5.0",
				MaxiCriteriaValue: "10.0",
			},
			expectCriteria:    true,
			expectedFieldType: models.UnsignedDecimalFieldType,
		},
		{
			name:        "Decimal + Automatic (three-tier)",
			fieldType:   "numeric",
			scoringType: models.AutomaticScoring,
			testData: TestElasticGoalData{
				FieldType:         "numeric",
				NumericSubtype:    models.DecimalFieldType,
				Unit:              "kg",
				ScoringType:       models.AutomaticScoring,
				Prompt:            "Weight change progress",
				MiniCriteriaValue: "0.5",
				MidiCriteriaValue: "1.0",
				MaxiCriteriaValue: "2.0",
			},
			expectCriteria:    true,
			expectedFieldType: models.DecimalFieldType,
		},

		// Numeric with constraints
		{
			name:        "UnsignedInt + Manual (with min/max)",
			fieldType:   "numeric",
			scoringType: models.ManualScoring,
			testData: TestElasticGoalData{
				FieldType:      "numeric",
				NumericSubtype: models.UnsignedIntFieldType,
				Unit:           "reps",
				ScoringType:    models.ManualScoring,
				HasMinMax:      true,
				MinValue:       "0",
				MaxValue:       "200",
				Prompt:         "How many push-ups did you do?",
			},
			expectCriteria:    false,
			expectedFieldType: models.UnsignedIntFieldType,
		},

		// Time field combinations with three-tier criteria
		{
			name:        "Time + Manual",
			fieldType:   models.TimeFieldType,
			scoringType: models.ManualScoring,
			testData: TestElasticGoalData{
				FieldType:   models.TimeFieldType,
				ScoringType: models.ManualScoring,
				Prompt:      "What time did you wake up?",
			},
			expectCriteria:    false,
			expectedFieldType: models.TimeFieldType,
		},
		{
			name:        "Time + Automatic (three-tier wake-up)",
			fieldType:   models.TimeFieldType,
			scoringType: models.AutomaticScoring,
			testData: TestElasticGoalData{
				FieldType:             models.TimeFieldType,
				ScoringType:           models.AutomaticScoring,
				Prompt:                "What time did you wake up?",
				MiniCriteriaTimeValue: "08:00",
				MidiCriteriaTimeValue: "07:00",
				MaxiCriteriaTimeValue: "06:00",
			},
			expectCriteria:    true,
			expectedFieldType: models.TimeFieldType,
		},
		{
			name:        "Time + Automatic (three-tier bedtime)",
			fieldType:   models.TimeFieldType,
			scoringType: models.AutomaticScoring,
			testData: TestElasticGoalData{
				FieldType:             models.TimeFieldType,
				ScoringType:           models.AutomaticScoring,
				Prompt:                "What time did you go to bed?",
				MiniCriteriaTimeValue: "23:00",
				MidiCriteriaTimeValue: "22:30",
				MaxiCriteriaTimeValue: "22:00",
			},
			expectCriteria:    true,
			expectedFieldType: models.TimeFieldType,
		},

		// Duration field combinations with three-tier criteria
		{
			name:        "Duration + Manual",
			fieldType:   models.DurationFieldType,
			scoringType: models.ManualScoring,
			testData: TestElasticGoalData{
				FieldType:   models.DurationFieldType,
				ScoringType: models.ManualScoring,
				Prompt:      "How long did you meditate?",
			},
			expectCriteria:    false,
			expectedFieldType: models.DurationFieldType,
		},
		{
			name:        "Duration + Automatic (three-tier meditation)",
			fieldType:   models.DurationFieldType,
			scoringType: models.AutomaticScoring,
			testData: TestElasticGoalData{
				FieldType:         models.DurationFieldType,
				ScoringType:       models.AutomaticScoring,
				Prompt:            "How long did you meditate?",
				MiniCriteriaValue: "10m",
				MidiCriteriaValue: "20m",
				MaxiCriteriaValue: "30m",
			},
			expectCriteria:    true,
			expectedFieldType: models.DurationFieldType,
		},
		{
			name:        "Duration + Automatic (three-tier exercise)",
			fieldType:   models.DurationFieldType,
			scoringType: models.AutomaticScoring,
			testData: TestElasticGoalData{
				FieldType:         models.DurationFieldType,
				ScoringType:       models.AutomaticScoring,
				Prompt:            "How long did you exercise?",
				MiniCriteriaValue: "20m",
				MidiCriteriaValue: "45m",
				MaxiCriteriaValue: "90m",
			},
			expectCriteria:    true,
			expectedFieldType: models.DurationFieldType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create goal using headless testing
			creator := NewElasticGoalCreatorForTesting("Test Elastic Goal", "Test Description", models.ElasticGoal, tt.testData)

			// Create goal directly without UI
			goal, err := creator.CreateGoalDirectly()
			require.NoError(t, err, "Goal creation should not fail")
			require.NotNil(t, goal, "Goal should be created")

			// Validate basic goal properties
			assert.Equal(t, "Test Elastic Goal", goal.Title)
			assert.Equal(t, "Test Description", goal.Description)
			assert.Equal(t, models.ElasticGoal, goal.GoalType)
			assert.Equal(t, tt.expectedFieldType, goal.FieldType.Type)
			assert.Equal(t, tt.scoringType, goal.ScoringType)

			// Validate criteria presence
			if tt.expectCriteria {
				assert.NotNil(t, goal.MiniCriteria, "Elastic goal should have mini criteria for automatic scoring")
				assert.NotNil(t, goal.MidiCriteria, "Elastic goal should have midi criteria for automatic scoring")
				assert.NotNil(t, goal.MaxiCriteria, "Elastic goal should have maxi criteria for automatic scoring")

				assert.NotNil(t, goal.MiniCriteria.Condition, "Mini criteria should have condition")
				assert.NotNil(t, goal.MidiCriteria.Condition, "Midi criteria should have condition")
				assert.NotNil(t, goal.MaxiCriteria.Condition, "Maxi criteria should have condition")

				assert.NotEmpty(t, goal.MiniCriteria.Description, "Mini criteria should have description")
				assert.NotEmpty(t, goal.MidiCriteria.Description, "Midi criteria should have description")
				assert.NotEmpty(t, goal.MaxiCriteria.Description, "Maxi criteria should have description")
			} else {
				assert.Nil(t, goal.MiniCriteria, "Goal should not have mini criteria for manual scoring")
				assert.Nil(t, goal.MidiCriteria, "Goal should not have midi criteria for manual scoring")
				assert.Nil(t, goal.MaxiCriteria, "Goal should not have maxi criteria for manual scoring")
			}

			// Validate goal against schema
			err = goal.Validate()
			assert.NoError(t, err, "Created elastic goal should pass validation")
		})
	}
}

// TestElasticGoalCreator_Integration_FieldTypeSpecificValidation tests field type specific configurations
func TestElasticGoalCreator_Integration_FieldTypeSpecificValidation(t *testing.T) {
	t.Run("Text field with multiline configuration", func(t *testing.T) {
		creator := NewElasticGoalCreatorForTesting("Exercise Log", "Daily exercise notes", models.ElasticGoal, TestElasticGoalData{
			FieldType:     models.TextFieldType,
			ScoringType:   models.ManualScoring,
			MultilineText: true,
			Prompt:        "How was your exercise intensity today?",
		})

		goal, err := creator.CreateGoalDirectly()
		require.NoError(t, err)

		assert.Equal(t, models.TextFieldType, goal.FieldType.Type)
		assert.NotNil(t, goal.FieldType.Multiline)
		assert.True(t, *goal.FieldType.Multiline)
	})

	t.Run("Numeric field with unit and constraints", func(t *testing.T) {
		creator := NewElasticGoalCreatorForTesting("Exercise", "Daily exercise", models.ElasticGoal, TestElasticGoalData{
			FieldType:      "numeric",
			NumericSubtype: models.UnsignedIntFieldType,
			Unit:           "minutes",
			ScoringType:    models.ManualScoring,
			HasMinMax:      true,
			MinValue:       "0",
			MaxValue:       "300",
			Prompt:         "How long did you exercise?",
		})

		goal, err := creator.CreateGoalDirectly()
		require.NoError(t, err)

		assert.Equal(t, models.UnsignedIntFieldType, goal.FieldType.Type)
		assert.Equal(t, "minutes", goal.FieldType.Unit)
		assert.NotNil(t, goal.FieldType.Min)
		assert.Equal(t, 0.0, *goal.FieldType.Min)
		assert.NotNil(t, goal.FieldType.Max)
		assert.Equal(t, 300.0, *goal.FieldType.Max)
	})
}

// TestElasticGoalCreator_Integration_ThreeTierCriteriaValidation tests all three-tier criteria types
func TestElasticGoalCreator_Integration_ThreeTierCriteriaValidation(t *testing.T) {
	tests := []struct {
		name             string
		testData         TestElasticGoalData
		validateCriteria func(t *testing.T, goal *models.Goal)
	}{
		{
			name: "Numeric three-tier criteria",
			testData: TestElasticGoalData{
				FieldType:         "numeric",
				NumericSubtype:    models.UnsignedIntFieldType,
				Unit:              "minutes",
				ScoringType:       models.AutomaticScoring,
				Prompt:            "How many minutes did you exercise?",
				MiniCriteriaValue: "15",
				MidiCriteriaValue: "30",
				MaxiCriteriaValue: "60",
			},
			validateCriteria: func(t *testing.T, goal *models.Goal) {
				// Validate mini criteria
				require.NotNil(t, goal.MiniCriteria.Condition.GreaterThanOrEqual)
				assert.Equal(t, 15.0, *goal.MiniCriteria.Condition.GreaterThanOrEqual)
				assert.Contains(t, goal.MiniCriteria.Description, "Mini achievement when value >= 15.0 minutes")

				// Validate midi criteria
				require.NotNil(t, goal.MidiCriteria.Condition.GreaterThanOrEqual)
				assert.Equal(t, 30.0, *goal.MidiCriteria.Condition.GreaterThanOrEqual)
				assert.Contains(t, goal.MidiCriteria.Description, "Midi achievement when value >= 30.0 minutes")

				// Validate maxi criteria
				require.NotNil(t, goal.MaxiCriteria.Condition.GreaterThanOrEqual)
				assert.Equal(t, 60.0, *goal.MaxiCriteria.Condition.GreaterThanOrEqual)
				assert.Contains(t, goal.MaxiCriteria.Description, "Maxi achievement when value >= 60.0 minutes")
			},
		},
		{
			name: "Time three-tier criteria (wake-up times)",
			testData: TestElasticGoalData{
				FieldType:             models.TimeFieldType,
				ScoringType:           models.AutomaticScoring,
				Prompt:                "What time did you wake up?",
				MiniCriteriaTimeValue: "08:00",
				MidiCriteriaTimeValue: "07:00",
				MaxiCriteriaTimeValue: "06:00",
			},
			validateCriteria: func(t *testing.T, goal *models.Goal) {
				// Validate mini criteria (before 8am)
				assert.Equal(t, "08:00", goal.MiniCriteria.Condition.Before)
				assert.Contains(t, goal.MiniCriteria.Description, "Mini achievement when time is before 08:00")

				// Validate midi criteria (before 7am)
				assert.Equal(t, "07:00", goal.MidiCriteria.Condition.Before)
				assert.Contains(t, goal.MidiCriteria.Description, "Midi achievement when time is before 07:00")

				// Validate maxi criteria (before 6am)
				assert.Equal(t, "06:00", goal.MaxiCriteria.Condition.Before)
				assert.Contains(t, goal.MaxiCriteria.Description, "Maxi achievement when time is before 06:00")
			},
		},
		{
			name: "Duration three-tier criteria",
			testData: TestElasticGoalData{
				FieldType:         models.DurationFieldType,
				ScoringType:       models.AutomaticScoring,
				Prompt:            "How long did you meditate?",
				MiniCriteriaValue: "10m",
				MidiCriteriaValue: "20m",
				MaxiCriteriaValue: "30m",
			},
			validateCriteria: func(t *testing.T, goal *models.Goal) {
				// Validate mini criteria (10+ minutes)
				assert.Equal(t, "10m", goal.MiniCriteria.Condition.After)
				assert.Contains(t, goal.MiniCriteria.Description, "Mini achievement when duration >= 10m")

				// Validate midi criteria (20+ minutes)
				assert.Equal(t, "20m", goal.MidiCriteria.Condition.After)
				assert.Contains(t, goal.MidiCriteria.Description, "Midi achievement when duration >= 20m")

				// Validate maxi criteria (30+ minutes)
				assert.Equal(t, "30m", goal.MaxiCriteria.Condition.After)
				assert.Contains(t, goal.MaxiCriteria.Description, "Maxi achievement when duration >= 30m")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creator := NewElasticGoalCreatorForTesting("Test Goal", "Test Description", models.ElasticGoal, tt.testData)

			goal, err := creator.CreateGoalDirectly()
			require.NoError(t, err)
			require.NotNil(t, goal.MiniCriteria, "Goal should have mini criteria")
			require.NotNil(t, goal.MidiCriteria, "Goal should have midi criteria")
			require.NotNil(t, goal.MaxiCriteria, "Goal should have maxi criteria")

			tt.validateCriteria(t, goal)
		})
	}
}

// TestElasticGoalCreator_Integration_CriteriaOrdering tests mini ≤ midi ≤ maxi validation
func TestElasticGoalCreator_Integration_CriteriaOrdering(t *testing.T) {
	t.Run("Valid ordering - numeric criteria", func(t *testing.T) {
		creator := NewElasticGoalCreatorForTesting("Exercise", "Daily exercise", models.ElasticGoal, TestElasticGoalData{
			FieldType:         "numeric",
			NumericSubtype:    models.UnsignedIntFieldType,
			Unit:              "minutes",
			ScoringType:       models.AutomaticScoring,
			Prompt:            "How many minutes did you exercise?",
			MiniCriteriaValue: "15", // 15 ≤ 30 ≤ 60 (valid)
			MidiCriteriaValue: "30",
			MaxiCriteriaValue: "60",
		})

		goal, err := creator.CreateGoalDirectly()
		require.NoError(t, err)

		// Should pass validation including criteria ordering
		err = goal.Validate()
		assert.NoError(t, err, "Goal with valid criteria ordering should pass validation")
	})

	t.Run("Invalid ordering - will be caught by model validation", func(t *testing.T) {
		creator := NewElasticGoalCreatorForTesting("Exercise", "Daily exercise", models.ElasticGoal, TestElasticGoalData{
			FieldType:         "numeric",
			NumericSubtype:    models.UnsignedIntFieldType,
			Unit:              "minutes",
			ScoringType:       models.AutomaticScoring,
			Prompt:            "How many minutes did you exercise?",
			MiniCriteriaValue: "60", // 60 > 30 > 15 (invalid ordering)
			MidiCriteriaValue: "30",
			MaxiCriteriaValue: "15",
		})

		goal, err := creator.CreateGoalDirectly()
		require.NoError(t, err, "Goal creation should succeed")

		// Should fail validation due to criteria ordering
		err = goal.Validate()
		assert.Error(t, err, "Goal with invalid criteria ordering should fail validation")
		assert.Contains(t, err.Error(), "mini criteria value")
	})
}

// TestElasticGoalCreator_Integration_YAMLValidation tests YAML generation and parsing
func TestElasticGoalCreator_Integration_YAMLValidation(t *testing.T) {
	// Test that all elastic goal combinations produce valid YAML that can be parsed back
	testCases := []TestElasticGoalData{
		// Text + Manual
		{
			FieldType:     models.TextFieldType,
			ScoringType:   models.ManualScoring,
			MultilineText: true,
			Prompt:        "How was your exercise intensity?",
		},
		// Numeric + Automatic with three-tier criteria
		{
			FieldType:         "numeric",
			NumericSubtype:    models.UnsignedDecimalFieldType,
			Unit:              "km",
			ScoringType:       models.AutomaticScoring,
			MiniCriteriaValue: "2.0",
			MidiCriteriaValue: "5.0",
			MaxiCriteriaValue: "10.0",
			Prompt:            "How far did you run?",
		},
		// Time + Automatic with three-tier criteria
		{
			FieldType:             models.TimeFieldType,
			ScoringType:           models.AutomaticScoring,
			MiniCriteriaTimeValue: "08:00",
			MidiCriteriaTimeValue: "07:00",
			MaxiCriteriaTimeValue: "06:00",
			Prompt:                "What time did you wake up?",
		},
		// Duration + Automatic with three-tier criteria
		{
			FieldType:         models.DurationFieldType,
			ScoringType:       models.AutomaticScoring,
			MiniCriteriaValue: "15m",
			MidiCriteriaValue: "30m",
			MaxiCriteriaValue: "60m",
			Prompt:            "How long did you meditate?",
		},
	}

	for i, testData := range testCases {
		t.Run(fmt.Sprintf("YAML_validation_case_%d", i+1), func(t *testing.T) {
			creator := NewElasticGoalCreatorForTesting("Test Elastic Goal", "Test Description", models.ElasticGoal, testData)

			goal, err := creator.CreateGoalDirectly()
			require.NoError(t, err, "Goal creation should succeed")

			// Validate the goal against the schema
			err = goal.Validate()
			assert.NoError(t, err, "Generated elastic goal should pass schema validation")

			// Additional validation: ensure all required fields are present
			assert.NotEmpty(t, goal.Title, "Goal should have title")
			assert.NotEmpty(t, goal.FieldType.Type, "Goal should have field type")
			assert.NotEmpty(t, goal.ScoringType, "Goal should have scoring type")
			assert.NotEmpty(t, goal.Prompt, "Goal should have prompt")

			if goal.ScoringType == models.AutomaticScoring {
				assert.NotNil(t, goal.MiniCriteria, "Automatic scoring elastic goals should have mini criteria")
				assert.NotNil(t, goal.MidiCriteria, "Automatic scoring elastic goals should have midi criteria")
				assert.NotNil(t, goal.MaxiCriteria, "Automatic scoring elastic goals should have maxi criteria")

				assert.NotNil(t, goal.MiniCriteria.Condition, "Mini criteria should have condition")
				assert.NotNil(t, goal.MidiCriteria.Condition, "Midi criteria should have condition")
				assert.NotNil(t, goal.MaxiCriteria.Condition, "Maxi criteria should have condition")
			}
		})
	}
}
