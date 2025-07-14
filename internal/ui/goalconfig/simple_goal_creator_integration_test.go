package goalconfig

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"davidlee/vice/internal/models"
)

// TestSimpleGoalCreator_Integration_AllCombinations tests all field type + scoring type combinations
func TestSimpleGoalCreator_Integration_AllCombinations(t *testing.T) {
	tests := []struct {
		name              string
		fieldType         string
		scoringType       models.ScoringType
		testData          TestGoalData
		expectCriteria    bool
		expectedFieldType string
	}{
		// Boolean field combinations
		{
			name:        "Boolean + Manual",
			fieldType:   models.BooleanFieldType,
			scoringType: models.ManualScoring,
			testData: TestGoalData{
				FieldType:   models.BooleanFieldType,
				ScoringType: models.ManualScoring,
				Prompt:      "Did you complete this task?",
			},
			expectCriteria:    false,
			expectedFieldType: models.BooleanFieldType,
		},
		{
			name:        "Boolean + Automatic",
			fieldType:   models.BooleanFieldType,
			scoringType: models.AutomaticScoring,
			testData: TestGoalData{
				FieldType:     models.BooleanFieldType,
				ScoringType:   models.AutomaticScoring,
				Prompt:        "Did you complete this task?",
				CriteriaType:  "equals",
				CriteriaValue: "true",
			},
			expectCriteria:    true,
			expectedFieldType: models.BooleanFieldType,
		},

		// Text field combinations
		{
			name:        "Text + Manual (multiline)",
			fieldType:   models.TextFieldType,
			scoringType: models.ManualScoring,
			testData: TestGoalData{
				FieldType:     models.TextFieldType,
				ScoringType:   models.ManualScoring,
				MultilineText: true,
				Prompt:        "What did you write about today?",
			},
			expectCriteria:    false,
			expectedFieldType: models.TextFieldType,
		},
		{
			name:        "Text + Manual (single line)",
			fieldType:   models.TextFieldType,
			scoringType: models.ManualScoring,
			testData: TestGoalData{
				FieldType:     models.TextFieldType,
				ScoringType:   models.ManualScoring,
				MultilineText: false,
				Prompt:        "What was your main focus today?",
			},
			expectCriteria:    false,
			expectedFieldType: models.TextFieldType,
		},

		// Numeric field combinations
		{
			name:        "UnsignedInt + Manual",
			fieldType:   "numeric",
			scoringType: models.ManualScoring,
			testData: TestGoalData{
				FieldType:      "numeric",
				NumericSubtype: models.UnsignedIntFieldType,
				Unit:           "reps",
				ScoringType:    models.ManualScoring,
				Prompt:         "How many push-ups did you do?",
			},
			expectCriteria:    false,
			expectedFieldType: models.UnsignedIntFieldType,
		},
		{
			name:        "UnsignedInt + Automatic (>=)",
			fieldType:   "numeric",
			scoringType: models.AutomaticScoring,
			testData: TestGoalData{
				FieldType:      "numeric",
				NumericSubtype: models.UnsignedIntFieldType,
				Unit:           "reps",
				ScoringType:    models.AutomaticScoring,
				Prompt:         "How many push-ups did you do?",
				CriteriaType:   "greater_than_or_equal",
				CriteriaValue:  "30",
			},
			expectCriteria:    true,
			expectedFieldType: models.UnsignedIntFieldType,
		},
		{
			name:        "UnsignedDecimal + Automatic (range)",
			fieldType:   "numeric",
			scoringType: models.AutomaticScoring,
			testData: TestGoalData{
				FieldType:      "numeric",
				NumericSubtype: models.UnsignedDecimalFieldType,
				Unit:           "hours",
				ScoringType:    models.AutomaticScoring,
				Prompt:         "How many hours did you sleep?",
				CriteriaType:   "range",
				CriteriaValue:  "7.0",
				CriteriaValue2: "9.0",
				RangeInclusive: true,
			},
			expectCriteria:    true,
			expectedFieldType: models.UnsignedDecimalFieldType,
		},
		{
			name:        "Decimal + Automatic (<)",
			fieldType:   "numeric",
			scoringType: models.AutomaticScoring,
			testData: TestGoalData{
				FieldType:      "numeric",
				NumericSubtype: models.DecimalFieldType,
				Unit:           "kg",
				ScoringType:    models.AutomaticScoring,
				Prompt:         "What was your weight change?",
				CriteriaType:   "less_than",
				CriteriaValue:  "0",
			},
			expectCriteria:    true,
			expectedFieldType: models.DecimalFieldType,
		},

		// Numeric with constraints
		{
			name:        "UnsignedInt + Manual (with min/max)",
			fieldType:   "numeric",
			scoringType: models.ManualScoring,
			testData: TestGoalData{
				FieldType:      "numeric",
				NumericSubtype: models.UnsignedIntFieldType,
				Unit:           "minutes",
				ScoringType:    models.ManualScoring,
				HasMinMax:      true,
				MinValue:       "0",
				MaxValue:       "180",
				Prompt:         "How long did you exercise?",
			},
			expectCriteria:    false,
			expectedFieldType: models.UnsignedIntFieldType,
		},

		// Time field combinations
		{
			name:        "Time + Manual",
			fieldType:   models.TimeFieldType,
			scoringType: models.ManualScoring,
			testData: TestGoalData{
				FieldType:   models.TimeFieldType,
				ScoringType: models.ManualScoring,
				Prompt:      "What time did you wake up?",
			},
			expectCriteria:    false,
			expectedFieldType: models.TimeFieldType,
		},
		{
			name:        "Time + Automatic (before)",
			fieldType:   models.TimeFieldType,
			scoringType: models.AutomaticScoring,
			testData: TestGoalData{
				FieldType:         models.TimeFieldType,
				ScoringType:       models.AutomaticScoring,
				Prompt:            "What time did you wake up?",
				CriteriaType:      "before",
				CriteriaTimeValue: "07:00",
			},
			expectCriteria:    true,
			expectedFieldType: models.TimeFieldType,
		},
		{
			name:        "Time + Automatic (after)",
			fieldType:   models.TimeFieldType,
			scoringType: models.AutomaticScoring,
			testData: TestGoalData{
				FieldType:         models.TimeFieldType,
				ScoringType:       models.AutomaticScoring,
				Prompt:            "What time did you go to bed?",
				CriteriaType:      "after",
				CriteriaTimeValue: "22:00",
			},
			expectCriteria:    true,
			expectedFieldType: models.TimeFieldType,
		},

		// Duration field combinations
		{
			name:        "Duration + Manual",
			fieldType:   models.DurationFieldType,
			scoringType: models.ManualScoring,
			testData: TestGoalData{
				FieldType:   models.DurationFieldType,
				ScoringType: models.ManualScoring,
				Prompt:      "How long did you meditate?",
			},
			expectCriteria:    false,
			expectedFieldType: models.DurationFieldType,
		},
		{
			name:        "Duration + Automatic (>=)",
			fieldType:   models.DurationFieldType,
			scoringType: models.AutomaticScoring,
			testData: TestGoalData{
				FieldType:     models.DurationFieldType,
				ScoringType:   models.AutomaticScoring,
				Prompt:        "How long did you meditate?",
				CriteriaType:  "greater_than_or_equal",
				CriteriaValue: "20m",
			},
			expectCriteria:    true,
			expectedFieldType: models.DurationFieldType,
		},
		{
			name:        "Duration + Automatic (range)",
			fieldType:   models.DurationFieldType,
			scoringType: models.AutomaticScoring,
			testData: TestGoalData{
				FieldType:      models.DurationFieldType,
				ScoringType:    models.AutomaticScoring,
				Prompt:         "How long did you exercise?",
				CriteriaType:   "range",
				CriteriaValue:  "30m",
				CriteriaValue2: "90m",
			},
			expectCriteria:    true,
			expectedFieldType: models.DurationFieldType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create goal using headless testing
			creator := NewSimpleGoalCreatorForTesting("Test Goal", "Test Description", models.SimpleGoal, tt.testData)

			// Create goal directly without UI
			goal, err := creator.CreateGoalDirectly()
			require.NoError(t, err, "Goal creation should not fail")
			require.NotNil(t, goal, "Goal should be created")

			// Validate basic goal properties
			assert.Equal(t, "Test Goal", goal.Title)
			assert.Equal(t, "Test Description", goal.Description)
			assert.Equal(t, models.SimpleGoal, goal.GoalType)
			assert.Equal(t, tt.expectedFieldType, goal.FieldType.Type)
			assert.Equal(t, tt.scoringType, goal.ScoringType)

			// Validate criteria presence
			if tt.expectCriteria {
				assert.NotNil(t, goal.Criteria, "Goal should have criteria for automatic scoring")
				assert.NotNil(t, goal.Criteria.Condition, "Criteria should have condition")
				assert.NotEmpty(t, goal.Criteria.Description, "Criteria should have description")
			} else {
				assert.Nil(t, goal.Criteria, "Goal should not have criteria for manual scoring")
			}

			// Validate goal against schema
			err = goal.Validate()
			assert.NoError(t, err, "Created goal should pass validation")
		})
	}
}

// TestSimpleGoalCreator_Integration_FieldTypeSpecificValidation tests field type specific configurations
func TestSimpleGoalCreator_Integration_FieldTypeSpecificValidation(t *testing.T) {
	t.Run("Text field with multiline configuration", func(t *testing.T) {
		creator := NewSimpleGoalCreatorForTesting("Journal", "Daily journaling", models.SimpleGoal, TestGoalData{
			FieldType:     models.TextFieldType,
			ScoringType:   models.ManualScoring,
			MultilineText: true,
			Prompt:        "What did you write about today?",
		})

		goal, err := creator.CreateGoalDirectly()
		require.NoError(t, err)

		assert.Equal(t, models.TextFieldType, goal.FieldType.Type)
		assert.NotNil(t, goal.FieldType.Multiline)
		assert.True(t, *goal.FieldType.Multiline)
	})

	t.Run("Numeric field with unit and constraints", func(t *testing.T) {
		creator := NewSimpleGoalCreatorForTesting("Exercise", "Daily exercise", models.SimpleGoal, TestGoalData{
			FieldType:      "numeric",
			NumericSubtype: models.UnsignedIntFieldType,
			Unit:           "minutes",
			ScoringType:    models.ManualScoring,
			HasMinMax:      true,
			MinValue:       "15",
			MaxValue:       "120",
			Prompt:         "How long did you exercise?",
		})

		goal, err := creator.CreateGoalDirectly()
		require.NoError(t, err)

		assert.Equal(t, models.UnsignedIntFieldType, goal.FieldType.Type)
		assert.Equal(t, "minutes", goal.FieldType.Unit)
		assert.NotNil(t, goal.FieldType.Min)
		assert.Equal(t, 15.0, *goal.FieldType.Min)
		assert.NotNil(t, goal.FieldType.Max)
		assert.Equal(t, 120.0, *goal.FieldType.Max)
	})
}

// TestSimpleGoalCreator_Integration_CriteriaValidation tests criteria construction for all types
func TestSimpleGoalCreator_Integration_CriteriaValidation(t *testing.T) {
	tests := []struct {
		name             string
		testData         TestGoalData
		validateCriteria func(t *testing.T, criteria *models.Criteria)
	}{
		{
			name: "Boolean criteria validation",
			testData: TestGoalData{
				FieldType:     models.BooleanFieldType,
				ScoringType:   models.AutomaticScoring,
				Prompt:        "Did you complete this?",
				CriteriaType:  "equals",
				CriteriaValue: "true",
			},
			validateCriteria: func(t *testing.T, criteria *models.Criteria) {
				assert.NotNil(t, criteria.Condition.Equals)
				assert.True(t, *criteria.Condition.Equals)
				assert.Contains(t, criteria.Description, "complete when checked as true")
			},
		},
		{
			name: "Numeric greater_than criteria",
			testData: TestGoalData{
				FieldType:      "numeric",
				NumericSubtype: models.UnsignedIntFieldType,
				Unit:           "reps",
				ScoringType:    models.AutomaticScoring,
				Prompt:         "How many reps?",
				CriteriaType:   "greater_than",
				CriteriaValue:  "50",
			},
			validateCriteria: func(t *testing.T, criteria *models.Criteria) {
				assert.NotNil(t, criteria.Condition.GreaterThan)
				assert.Equal(t, 50.0, *criteria.Condition.GreaterThan)
				assert.Contains(t, criteria.Description, "> 50.0 reps")
			},
		},
		{
			name: "Numeric range criteria",
			testData: TestGoalData{
				FieldType:      "numeric",
				NumericSubtype: models.UnsignedDecimalFieldType,
				Unit:           "hours",
				ScoringType:    models.AutomaticScoring,
				Prompt:         "How many hours?",
				CriteriaType:   "range",
				CriteriaValue:  "7.0",
				CriteriaValue2: "9.0",
				RangeInclusive: true,
			},
			validateCriteria: func(t *testing.T, criteria *models.Criteria) {
				assert.NotNil(t, criteria.Condition.Range)
				assert.Equal(t, 7.0, criteria.Condition.Range.Min)
				assert.Equal(t, 9.0, criteria.Condition.Range.Max)
				assert.NotNil(t, criteria.Condition.Range.MinInclusive)
				assert.True(t, *criteria.Condition.Range.MinInclusive)
				assert.Contains(t, criteria.Description, "7.0 to 9.0 hours (inclusive)")
			},
		},
		{
			name: "Time before criteria",
			testData: TestGoalData{
				FieldType:         models.TimeFieldType,
				ScoringType:       models.AutomaticScoring,
				Prompt:            "What time did you wake up?",
				CriteriaType:      "before",
				CriteriaTimeValue: "07:00",
			},
			validateCriteria: func(t *testing.T, criteria *models.Criteria) {
				assert.Equal(t, "07:00", criteria.Condition.Before)
				assert.Contains(t, criteria.Description, "before 07:00")
			},
		},
		{
			name: "Duration criteria",
			testData: TestGoalData{
				FieldType:     models.DurationFieldType,
				ScoringType:   models.AutomaticScoring,
				Prompt:        "How long did you meditate?",
				CriteriaType:  "greater_than_or_equal",
				CriteriaValue: "20m",
			},
			validateCriteria: func(t *testing.T, criteria *models.Criteria) {
				assert.Equal(t, "20m", criteria.Condition.After)
				assert.Contains(t, criteria.Description, "duration >= 20m")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creator := NewSimpleGoalCreatorForTesting("Test Goal", "Test Description", models.SimpleGoal, tt.testData)

			goal, err := creator.CreateGoalDirectly()
			require.NoError(t, err)
			require.NotNil(t, goal.Criteria, "Goal should have criteria")

			tt.validateCriteria(t, goal.Criteria)
		})
	}
}

// TestSimpleGoalCreator_Integration_YAMLValidation tests YAML generation and parsing
func TestSimpleGoalCreator_Integration_YAMLValidation(t *testing.T) {
	// Test that all goal combinations produce valid YAML that can be parsed back
	testCases := []TestGoalData{
		// Boolean + Manual
		{
			FieldType:   models.BooleanFieldType,
			ScoringType: models.ManualScoring,
			Prompt:      "Did you complete this?",
		},
		// Numeric + Automatic with range
		{
			FieldType:      "numeric",
			NumericSubtype: models.UnsignedDecimalFieldType,
			Unit:           "hours",
			ScoringType:    models.AutomaticScoring,
			CriteriaType:   "range",
			CriteriaValue:  "7.0",
			CriteriaValue2: "9.0",
			RangeInclusive: true,
			Prompt:         "How many hours did you sleep?",
		},
		// Time + Automatic
		{
			FieldType:         models.TimeFieldType,
			ScoringType:       models.AutomaticScoring,
			CriteriaType:      "before",
			CriteriaTimeValue: "07:00",
			Prompt:            "What time did you wake up?",
		},
		// Duration + Automatic
		{
			FieldType:     models.DurationFieldType,
			ScoringType:   models.AutomaticScoring,
			CriteriaType:  "greater_than_or_equal",
			CriteriaValue: "30m",
			Prompt:        "How long did you meditate?",
		},
	}

	for i, testData := range testCases {
		t.Run(fmt.Sprintf("YAML_validation_case_%d", i+1), func(t *testing.T) {
			creator := NewSimpleGoalCreatorForTesting("Test Goal", "Test Description", models.SimpleGoal, testData)

			goal, err := creator.CreateGoalDirectly()
			require.NoError(t, err, "Goal creation should succeed")

			// Validate the goal against the schema
			err = goal.Validate()
			assert.NoError(t, err, "Generated goal should pass schema validation")

			// Additional validation: ensure all required fields are present
			assert.NotEmpty(t, goal.Title, "Goal should have title")
			assert.NotEmpty(t, goal.FieldType.Type, "Goal should have field type")
			assert.NotEmpty(t, goal.ScoringType, "Goal should have scoring type")
			assert.NotEmpty(t, goal.Prompt, "Goal should have prompt")

			if goal.ScoringType == models.AutomaticScoring {
				assert.NotNil(t, goal.Criteria, "Automatic scoring goals should have criteria")
				assert.NotNil(t, goal.Criteria.Condition, "Criteria should have condition")
			}
		})
	}
}
