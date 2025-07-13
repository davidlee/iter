package entry

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"davidlee/iter/internal/models"
	"davidlee/iter/internal/scoring"
)

// AIDEV-NOTE: T010/3.3-test-suite; comprehensive testing for InformationalGoalCollectionFlow
// Features tested: headless data collection, all field types, no scoring, notes handling, value types

func TestInformationalGoalCollectionFlow_GetFlowType(t *testing.T) {
	factory := NewEntryFieldInputFactory()
	flow := NewInformationalGoalCollectionFlowForTesting(factory)

	assert.Equal(t, string(models.InformationalGoal), flow.GetFlowType())
}

func TestInformationalGoalCollectionFlow_RequiresScoring(t *testing.T) {
	factory := NewEntryFieldInputFactory()
	flow := NewInformationalGoalCollectionFlowForTesting(factory)

	assert.False(t, flow.RequiresScoring())
}

func TestInformationalGoalCollectionFlow_GetExpectedFieldTypes(t *testing.T) {
	factory := NewEntryFieldInputFactory()
	flow := NewInformationalGoalCollectionFlowForTesting(factory)

	fieldTypes := flow.GetExpectedFieldTypes()

	// Informational goals support all field types
	expectedTypes := []string{
		models.BooleanFieldType,
		models.TextFieldType,
		models.UnsignedIntFieldType,
		models.UnsignedDecimalFieldType,
		models.DecimalFieldType,
		models.TimeFieldType,
		models.DurationFieldType,
		models.ChecklistFieldType,
	}

	assert.ElementsMatch(t, expectedTypes, fieldTypes)
}

func TestInformationalGoalCollectionFlow_CollectEntryDirectly_NoScoring(t *testing.T) {
	factory := NewEntryFieldInputFactory()
	flow := NewInformationalGoalCollectionFlowForTesting(factory)

	goal := models.Goal{
		ID:          "test_info",
		Title:       "Test Informational Goal",
		GoalType:    models.InformationalGoal,
		ScoringType: models.ManualScoring,
		FieldType: models.FieldType{
			Type: models.UnsignedIntFieldType,
		},
	}

	result, err := flow.CollectEntryDirectly(goal, 42, "Test notes", nil)

	require.NoError(t, err)
	assert.Equal(t, 42, result.Value)
	assert.Nil(t, result.AchievementLevel) // Never has achievement level
	assert.Equal(t, "Test notes", result.Notes)
}

func TestInformationalGoalCollectionFlow_FieldTypes(t *testing.T) {
	factory := NewEntryFieldInputFactory()
	flow := NewInformationalGoalCollectionFlowForTesting(factory)

	testCases := []struct {
		name      string
		fieldType string
		value     interface{}
		expected  interface{}
	}{
		{
			name:      "Boolean field",
			fieldType: models.BooleanFieldType,
			value:     true,
			expected:  true,
		},
		{
			name:      "Text field",
			fieldType: models.TextFieldType,
			value:     "Sample text data",
			expected:  "Sample text data",
		},
		{
			name:      "UnsignedInt field",
			fieldType: models.UnsignedIntFieldType,
			value:     100,
			expected:  100,
		},
		{
			name:      "UnsignedDecimal field",
			fieldType: models.UnsignedDecimalFieldType,
			value:     25.5,
			expected:  25.5,
		},
		{
			name:      "Decimal field",
			fieldType: models.DecimalFieldType,
			value:     -15.75,
			expected:  -15.75,
		},
		{
			name:      "Time field",
			fieldType: models.TimeFieldType,
			value:     "14:30",
			expected:  "14:30",
		},
		{
			name:      "Duration field",
			fieldType: models.DurationFieldType,
			value:     "2h30m",
			expected:  "2h30m",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			goal := models.Goal{
				ID:          "test_" + tc.name,
				Title:       "Test " + tc.name,
				GoalType:    models.InformationalGoal,
				ScoringType: models.ManualScoring,
				FieldType: models.FieldType{
					Type: tc.fieldType,
				},
			}

			result, err := flow.CollectEntryDirectly(goal, tc.value, "", nil)

			require.NoError(t, err)
			assert.Equal(t, tc.expected, result.Value)
			assert.Nil(t, result.AchievementLevel)
			assert.Equal(t, "", result.Notes)
		})
	}
}

func TestInformationalGoalCollectionFlow_WithNotes(t *testing.T) {
	factory := NewEntryFieldInputFactory()
	flow := NewInformationalGoalCollectionFlowForTesting(factory)

	goal := models.Goal{
		ID:          "test_notes",
		Title:       "Test with Notes",
		GoalType:    models.InformationalGoal,
		ScoringType: models.ManualScoring,
		FieldType: models.FieldType{
			Type: models.TextFieldType,
		},
	}

	notes := "Detailed observations about this data point"
	result, err := flow.CollectEntryDirectly(goal, "Data value", notes, nil)

	require.NoError(t, err)
	assert.Equal(t, "Data value", result.Value)
	assert.Nil(t, result.AchievementLevel)
	assert.Equal(t, notes, result.Notes)
}

func TestInformationalGoalCollectionFlow_AutomaticScoringIgnored(t *testing.T) {
	factory := NewEntryFieldInputFactory()
	flow := NewInformationalGoalCollectionFlowForTesting(factory)

	// Create a goal with automatic scoring configuration
	goal := models.Goal{
		ID:          "test_auto_scoring",
		Title:       "Test Automatic Scoring Ignored",
		GoalType:    models.InformationalGoal,
		ScoringType: models.AutomaticScoring, // This should be ignored
		FieldType: models.FieldType{
			Type: models.UnsignedIntFieldType,
		},
		// Add criteria that would normally trigger scoring
		Criteria: &models.Criteria{
			Condition: &models.Condition{
				GreaterThanOrEqual: float64Ptr(50),
			},
		},
	}

	result, err := flow.CollectEntryDirectly(goal, 75, "Test notes", nil)

	require.NoError(t, err)
	assert.Equal(t, 75, result.Value)
	assert.Nil(t, result.AchievementLevel) // Should still be nil despite criteria
	assert.Equal(t, "Test notes", result.Notes)
}

func TestInformationalGoalCollectionFlow_ZeroValues(t *testing.T) {
	factory := NewEntryFieldInputFactory()
	flow := NewInformationalGoalCollectionFlowForTesting(factory)

	testCases := []struct {
		name      string
		fieldType string
		value     interface{}
	}{
		{
			name:      "Zero integer",
			fieldType: models.UnsignedIntFieldType,
			value:     0,
		},
		{
			name:      "Zero decimal",
			fieldType: models.UnsignedDecimalFieldType,
			value:     0.0,
		},
		{
			name:      "False boolean",
			fieldType: models.BooleanFieldType,
			value:     false,
		},
		{
			name:      "Empty string",
			fieldType: models.TextFieldType,
			value:     "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			goal := models.Goal{
				ID:          "test_zero_" + tc.name,
				Title:       "Test Zero " + tc.name,
				GoalType:    models.InformationalGoal,
				ScoringType: models.ManualScoring,
				FieldType: models.FieldType{
					Type: tc.fieldType,
				},
			}

			result, err := flow.CollectEntryDirectly(goal, tc.value, "", nil)

			require.NoError(t, err)
			assert.Equal(t, tc.value, result.Value)
			assert.Nil(t, result.AchievementLevel) // Always nil for informational goals
			assert.Equal(t, "", result.Notes)
		})
	}
}

func TestInformationalGoalCollectionFlow_WithExistingEntry(t *testing.T) {
	factory := NewEntryFieldInputFactory()
	flow := NewInformationalGoalCollectionFlowForTesting(factory)

	goal := models.Goal{
		ID:          "test_existing",
		Title:       "Test with Existing Entry",
		GoalType:    models.InformationalGoal,
		ScoringType: models.ManualScoring,
		FieldType: models.FieldType{
			Type: models.UnsignedIntFieldType,
		},
	}

	existing := &ExistingEntry{
		Value: 25,
		Notes: "Previous notes",
	}

	result, err := flow.CollectEntryDirectly(goal, 50, "Updated notes", existing)

	require.NoError(t, err)
	assert.Equal(t, 50, result.Value)              // New value
	assert.Nil(t, result.AchievementLevel)         // Still nil
	assert.Equal(t, "Updated notes", result.Notes) // New notes
}

func TestInformationalGoalCollectionFlow_WithScoringEngine(t *testing.T) {
	// Test that scoring engine presence doesn't affect informational goals
	factory := NewEntryFieldInputFactory()
	scoringEngine := scoring.NewEngine()

	// Even if we somehow created the flow with a scoring engine, it shouldn't be used
	flow := NewInformationalGoalCollectionFlowForTesting(factory)

	goal := models.Goal{
		ID:          "test_scoring_engine",
		Title:       "Test Scoring Engine Ignored",
		GoalType:    models.InformationalGoal,
		ScoringType: models.AutomaticScoring,
		FieldType: models.FieldType{
			Type: models.UnsignedIntFieldType,
		},
		Criteria: &models.Criteria{
			Condition: &models.Condition{
				GreaterThanOrEqual: float64Ptr(100),
			},
		},
	}

	// Even with scoring engine and criteria, informational goals don't score
	result, err := flow.CollectEntryDirectly(goal, 150, "", nil)

	require.NoError(t, err)
	assert.Equal(t, 150, result.Value)
	assert.Nil(t, result.AchievementLevel) // Still nil despite meeting criteria
	assert.Equal(t, "", result.Notes)

	// Ensure scoring engine is available but not used
	_ = scoringEngine
}

func TestInformationalGoalCollectionFlow_DirectionAwareness(t *testing.T) {
	factory := NewEntryFieldInputFactory()
	flow := NewInformationalGoalCollectionFlowForTesting(factory)

	testCases := []struct {
		name      string
		direction string
		value     interface{}
	}{
		{
			name:      "Higher better direction",
			direction: "higher_better",
			value:     85,
		},
		{
			name:      "Lower better direction",
			direction: "lower_better",
			value:     12.5,
		},
		{
			name:      "Neutral direction",
			direction: "neutral",
			value:     "Status update",
		},
		{
			name:      "Empty direction (defaults to neutral)",
			direction: "",
			value:     42,
		},
		{
			name:      "Unknown direction (falls back to neutral)",
			direction: "unknown_direction",
			value:     100,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			goal := models.Goal{
				ID:          "test_direction_" + strings.ReplaceAll(tc.name, " ", "_"),
				Title:       "Test " + tc.name,
				GoalType:    models.InformationalGoal,
				ScoringType: models.ManualScoring,
				Direction:   tc.direction,
				FieldType: models.FieldType{
					Type: models.UnsignedDecimalFieldType,
				},
			}

			result, err := flow.CollectEntryDirectly(goal, tc.value, "", nil)

			require.NoError(t, err)
			assert.Equal(t, tc.value, result.Value)
			assert.Nil(t, result.AchievementLevel) // Always nil for informational goals
			assert.Equal(t, "", result.Notes)
		})
	}
}
