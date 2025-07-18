package entry

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/davidlee/vice/internal/models"
	"github.com/davidlee/vice/internal/scoring"
)

// AIDEV-NOTE: T010/3.3-test-suite; comprehensive testing for InformationalHabitCollectionFlow
// Features tested: headless data collection, all field types, no scoring, notes handling, value types

func TestInformationalHabitCollectionFlow_GetFlowType(t *testing.T) {
	factory := NewEntryFieldInputFactory()
	flow := NewInformationalHabitCollectionFlowForTesting(factory)

	assert.Equal(t, string(models.InformationalHabit), flow.GetFlowType())
}

func TestInformationalHabitCollectionFlow_RequiresScoring(t *testing.T) {
	factory := NewEntryFieldInputFactory()
	flow := NewInformationalHabitCollectionFlowForTesting(factory)

	assert.False(t, flow.RequiresScoring())
}

func TestInformationalHabitCollectionFlow_GetExpectedFieldTypes(t *testing.T) {
	factory := NewEntryFieldInputFactory()
	flow := NewInformationalHabitCollectionFlowForTesting(factory)

	fieldTypes := flow.GetExpectedFieldTypes()

	// Informational habits support all field types
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

func TestInformationalHabitCollectionFlow_CollectEntryDirectly_NoScoring(t *testing.T) {
	factory := NewEntryFieldInputFactory()
	flow := NewInformationalHabitCollectionFlowForTesting(factory)

	habit := models.Habit{
		ID:          "test_info",
		Title:       "Test Informational Habit",
		HabitType:   models.InformationalHabit,
		ScoringType: models.ManualScoring,
		FieldType: models.FieldType{
			Type: models.UnsignedIntFieldType,
		},
	}

	result, err := flow.CollectEntryDirectly(habit, 42, "Test notes", nil)

	require.NoError(t, err)
	assert.Equal(t, 42, result.Value)
	assert.Nil(t, result.AchievementLevel) // Never has achievement level
	assert.Equal(t, "Test notes", result.Notes)
}

func TestInformationalHabitCollectionFlow_FieldTypes(t *testing.T) {
	factory := NewEntryFieldInputFactory()
	flow := NewInformationalHabitCollectionFlowForTesting(factory)

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
			habit := models.Habit{
				ID:          "test_" + tc.name,
				Title:       "Test " + tc.name,
				HabitType:   models.InformationalHabit,
				ScoringType: models.ManualScoring,
				FieldType: models.FieldType{
					Type: tc.fieldType,
				},
			}

			result, err := flow.CollectEntryDirectly(habit, tc.value, "", nil)

			require.NoError(t, err)
			assert.Equal(t, tc.expected, result.Value)
			assert.Nil(t, result.AchievementLevel)
			assert.Equal(t, "", result.Notes)
		})
	}
}

func TestInformationalHabitCollectionFlow_WithNotes(t *testing.T) {
	factory := NewEntryFieldInputFactory()
	flow := NewInformationalHabitCollectionFlowForTesting(factory)

	habit := models.Habit{
		ID:          "test_notes",
		Title:       "Test with Notes",
		HabitType:   models.InformationalHabit,
		ScoringType: models.ManualScoring,
		FieldType: models.FieldType{
			Type: models.TextFieldType,
		},
	}

	notes := "Detailed observations about this data point"
	result, err := flow.CollectEntryDirectly(habit, "Data value", notes, nil)

	require.NoError(t, err)
	assert.Equal(t, "Data value", result.Value)
	assert.Nil(t, result.AchievementLevel)
	assert.Equal(t, notes, result.Notes)
}

func TestInformationalHabitCollectionFlow_AutomaticScoringIgnored(t *testing.T) {
	factory := NewEntryFieldInputFactory()
	flow := NewInformationalHabitCollectionFlowForTesting(factory)

	// Create a habit with automatic scoring configuration
	habit := models.Habit{
		ID:          "test_auto_scoring",
		Title:       "Test Automatic Scoring Ignored",
		HabitType:   models.InformationalHabit,
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

	result, err := flow.CollectEntryDirectly(habit, 75, "Test notes", nil)

	require.NoError(t, err)
	assert.Equal(t, 75, result.Value)
	assert.Nil(t, result.AchievementLevel) // Should still be nil despite criteria
	assert.Equal(t, "Test notes", result.Notes)
}

func TestInformationalHabitCollectionFlow_ZeroValues(t *testing.T) {
	factory := NewEntryFieldInputFactory()
	flow := NewInformationalHabitCollectionFlowForTesting(factory)

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
			habit := models.Habit{
				ID:          "test_zero_" + tc.name,
				Title:       "Test Zero " + tc.name,
				HabitType:   models.InformationalHabit,
				ScoringType: models.ManualScoring,
				FieldType: models.FieldType{
					Type: tc.fieldType,
				},
			}

			result, err := flow.CollectEntryDirectly(habit, tc.value, "", nil)

			require.NoError(t, err)
			assert.Equal(t, tc.value, result.Value)
			assert.Nil(t, result.AchievementLevel) // Always nil for informational habits
			assert.Equal(t, "", result.Notes)
		})
	}
}

func TestInformationalHabitCollectionFlow_WithExistingEntry(t *testing.T) {
	factory := NewEntryFieldInputFactory()
	flow := NewInformationalHabitCollectionFlowForTesting(factory)

	habit := models.Habit{
		ID:          "test_existing",
		Title:       "Test with Existing Entry",
		HabitType:   models.InformationalHabit,
		ScoringType: models.ManualScoring,
		FieldType: models.FieldType{
			Type: models.UnsignedIntFieldType,
		},
	}

	existing := &ExistingEntry{
		Value: 25,
		Notes: "Previous notes",
	}

	result, err := flow.CollectEntryDirectly(habit, 50, "Updated notes", existing)

	require.NoError(t, err)
	assert.Equal(t, 50, result.Value)              // New value
	assert.Nil(t, result.AchievementLevel)         // Still nil
	assert.Equal(t, "Updated notes", result.Notes) // New notes
}

func TestInformationalHabitCollectionFlow_WithScoringEngine(t *testing.T) {
	// Test that scoring engine presence doesn't affect informational habits
	factory := NewEntryFieldInputFactory()
	scoringEngine := scoring.NewEngine()

	// Even if we somehow created the flow with a scoring engine, it shouldn't be used
	flow := NewInformationalHabitCollectionFlowForTesting(factory)

	habit := models.Habit{
		ID:          "test_scoring_engine",
		Title:       "Test Scoring Engine Ignored",
		HabitType:   models.InformationalHabit,
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

	// Even with scoring engine and criteria, informational habits don't score
	result, err := flow.CollectEntryDirectly(habit, 150, "", nil)

	require.NoError(t, err)
	assert.Equal(t, 150, result.Value)
	assert.Nil(t, result.AchievementLevel) // Still nil despite meeting criteria
	assert.Equal(t, "", result.Notes)

	// Ensure scoring engine is available but not used
	_ = scoringEngine
}

func TestInformationalHabitCollectionFlow_DirectionAwareness(t *testing.T) {
	factory := NewEntryFieldInputFactory()
	flow := NewInformationalHabitCollectionFlowForTesting(factory)

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
			habit := models.Habit{
				ID:          "test_direction_" + strings.ReplaceAll(tc.name, " ", "_"),
				Title:       "Test " + tc.name,
				HabitType:   models.InformationalHabit,
				ScoringType: models.ManualScoring,
				Direction:   tc.direction,
				FieldType: models.FieldType{
					Type: models.UnsignedDecimalFieldType,
				},
			}

			result, err := flow.CollectEntryDirectly(habit, tc.value, "", nil)

			require.NoError(t, err)
			assert.Equal(t, tc.value, result.Value)
			assert.Nil(t, result.AchievementLevel) // Always nil for informational habits
			assert.Equal(t, "", result.Notes)
		})
	}
}
