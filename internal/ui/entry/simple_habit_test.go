package entry

import (
	"testing"

	"github.com/davidlee/vice/internal/models"
	"github.com/davidlee/vice/internal/scoring"
)

// AIDEV-NOTE: simple-habit-tests; comprehensive testing for simple habit collection flow
// Tests pass/fail collection, automatic/manual scoring, and field type integration for T010/3.1
// AIDEV-NOTE: testing-patterns; uses NewSimpleHabitCollectionFlowForTesting() and CollectEntryDirectly() for headless testing
// All major scenarios covered: Boolean true/false, text content, numeric values, automatic scoring with criteria

func TestSimpleHabitCollectionFlow(t *testing.T) {
	factory := NewEntryFieldInputFactory()
	scoringEngine := &scoring.Engine{} // Mock or real scoring engine

	flow := NewSimpleHabitCollectionFlow(factory, scoringEngine)

	// Test flow type identification
	if flow.GetFlowType() != string(models.SimpleHabit) {
		t.Errorf("GetFlowType() = %v, want %v", flow.GetFlowType(), string(models.SimpleHabit))
	}

	// Test scoring requirement
	if !flow.RequiresScoring() {
		t.Errorf("RequiresScoring() expected true for simple habits")
	}

	// Test supported field types
	expectedFieldTypes := []string{
		models.BooleanFieldType,
		models.TextFieldType,
		models.UnsignedIntFieldType,
		models.UnsignedDecimalFieldType,
		models.DecimalFieldType,
		models.TimeFieldType,
		models.DurationFieldType,
	}

	supportedTypes := flow.GetExpectedFieldTypes()
	if len(supportedTypes) != len(expectedFieldTypes) {
		t.Errorf("GetExpectedFieldTypes() length = %v, want %v", len(supportedTypes), len(expectedFieldTypes))
	}

	for _, expectedType := range expectedFieldTypes {
		found := false
		for _, supportedType := range supportedTypes {
			if supportedType == expectedType {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected field type %v not found in supported types", expectedType)
		}
	}

	// Ensure checklist field type is NOT supported for simple habits
	for _, supportedType := range supportedTypes {
		if supportedType == models.ChecklistFieldType {
			t.Errorf("ChecklistFieldType should not be supported for simple habits")
		}
	}
}

func TestSimpleHabitManualAchievement(t *testing.T) {
	factory := NewEntryFieldInputFactory()
	flow := NewSimpleHabitCollectionFlow(factory, nil) // No scoring engine for manual tests

	testCases := []struct {
		name      string
		fieldType models.FieldType
		value     interface{}
		expected  models.AchievementLevel
	}{
		{
			name:      "Boolean true",
			fieldType: models.FieldType{Type: models.BooleanFieldType},
			value:     true,
			expected:  models.AchievementMini,
		},
		{
			name:      "Boolean false",
			fieldType: models.FieldType{Type: models.BooleanFieldType},
			value:     false,
			expected:  models.AchievementNone,
		},
		{
			name:      "Text with content",
			fieldType: models.FieldType{Type: models.TextFieldType},
			value:     "Some text content",
			expected:  models.AchievementMini,
		},
		{
			name:      "Text empty",
			fieldType: models.FieldType{Type: models.TextFieldType},
			value:     "",
			expected:  models.AchievementNone,
		},
		{
			name:      "Text whitespace only",
			fieldType: models.FieldType{Type: models.TextFieldType},
			value:     "   ",
			expected:  models.AchievementNone,
		},
		{
			name:      "Numeric with value",
			fieldType: models.FieldType{Type: models.UnsignedIntFieldType},
			value:     42,
			expected:  models.AchievementMini,
		},
		{
			name:      "Time field",
			fieldType: models.FieldType{Type: models.TimeFieldType},
			value:     "14:30",
			expected:  models.AchievementMini,
		},
		{
			name:      "Duration field",
			fieldType: models.FieldType{Type: models.DurationFieldType},
			value:     "30m",
			expected:  models.AchievementMini,
		},
		{
			name:      "Nil value",
			fieldType: models.FieldType{Type: models.TextFieldType},
			value:     nil,
			expected:  models.AchievementNone,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			habit := models.Habit{
				Title:     "Test Habit",
				FieldType: tc.fieldType,
			}

			result := flow.determineManualAchievement(habit, tc.value)
			if result == nil {
				t.Errorf("determineManualAchievement() returned nil")
				return
			}

			if *result != tc.expected {
				t.Errorf("determineManualAchievement() = %v, want %v", *result, tc.expected)
			}
		})
	}
}

func TestSimpleHabitAutomaticScoringWithRealEngine(t *testing.T) {
	factory := NewEntryFieldInputFactory()
	scoringEngine := scoring.NewEngine() // Use real scoring engine
	flow := NewSimpleHabitCollectionFlow(factory, scoringEngine)

	// Test with a simple habit that has criteria for automatic scoring
	habit := models.Habit{
		Title:       "Test Automatic Habit",
		ID:          "test_habit",
		HabitType:   models.SimpleHabit, // Simple habit with automatic scoring
		ScoringType: models.AutomaticScoring,
		FieldType:   models.FieldType{Type: models.UnsignedIntFieldType},
		Criteria: &models.Criteria{
			Description: "At least 1",
			Condition: &models.Condition{
				GreaterThanOrEqual: float64Ptr(1),
			},
		},
	}

	// Test value that should pass mini criteria
	result, err := flow.performAutomaticScoring(habit, float64(5))
	if err != nil {
		t.Errorf("performAutomaticScoring() unexpected error: %v", err)
	}

	if result == nil {
		t.Errorf("performAutomaticScoring() returned nil result")
		return
	}

	if *result != models.AchievementMini {
		t.Errorf("performAutomaticScoring() = %v, want %v", *result, models.AchievementMini)
	}
}

func TestSimpleHabitAutomaticScoringNoEngine(t *testing.T) {
	factory := NewEntryFieldInputFactory()
	flow := NewSimpleHabitCollectionFlow(factory, nil) // No scoring engine

	habit := models.Habit{
		Title:       "Test Habit",
		ScoringType: models.AutomaticScoring,
		FieldType:   models.FieldType{Type: models.BooleanFieldType},
	}

	_, err := flow.performAutomaticScoring(habit, true)
	if err == nil {
		t.Errorf("performAutomaticScoring() expected error when no scoring engine provided")
	}
}

func TestSimpleHabitFieldTypeSupport(t *testing.T) {
	factory := NewEntryFieldInputFactory()
	flow := NewSimpleHabitCollectionFlow(factory, nil)

	// Test that simple habits support all field types except checklist
	supportedTypes := flow.GetExpectedFieldTypes()

	// Should support these types
	expectedSupported := []string{
		models.BooleanFieldType,
		models.TextFieldType,
		models.UnsignedIntFieldType,
		models.UnsignedDecimalFieldType,
		models.DecimalFieldType,
		models.TimeFieldType,
		models.DurationFieldType,
	}

	for _, expectedType := range expectedSupported {
		found := false
		for _, supportedType := range supportedTypes {
			if supportedType == expectedType {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Simple habits should support field type %v", expectedType)
		}
	}

	// Should NOT support checklist type
	for _, supportedType := range supportedTypes {
		if supportedType == models.ChecklistFieldType {
			t.Errorf("Simple habits should NOT support ChecklistFieldType")
		}
	}
}

func TestSimpleHabitFlowInterface(t *testing.T) {
	factory := NewEntryFieldInputFactory()
	scoringEngine := &scoring.Engine{}
	flow := NewSimpleHabitCollectionFlow(factory, scoringEngine)

	// Verify that SimpleHabitCollectionFlow implements HabitCollectionFlow interface
	var _ HabitCollectionFlow = flow

	// Test interface methods
	if flow.GetFlowType() != string(models.SimpleHabit) {
		t.Errorf("GetFlowType() = %v, want %v", flow.GetFlowType(), string(models.SimpleHabit))
	}

	if !flow.RequiresScoring() {
		t.Errorf("RequiresScoring() = false, want true")
	}

	fieldTypes := flow.GetExpectedFieldTypes()
	if len(fieldTypes) == 0 {
		t.Errorf("GetExpectedFieldTypes() returned empty slice")
	}
}

func TestSimpleHabitCollectEntryDirectly(t *testing.T) {
	factory := NewEntryFieldInputFactory()
	scoringEngine := scoring.NewEngine()
	flow := NewSimpleHabitCollectionFlowForTesting(factory, scoringEngine)

	testCases := []struct {
		name                string
		habit               models.Habit
		value               interface{}
		notes               string
		expectedAchievement models.AchievementLevel
	}{
		{
			name: "Boolean true manual scoring",
			habit: models.Habit{
				Title:       "Daily Exercise",
				HabitType:   models.SimpleHabit,
				ScoringType: models.ManualScoring,
				FieldType:   models.FieldType{Type: models.BooleanFieldType},
			},
			value:               true,
			notes:               "Did 30 minutes of cardio",
			expectedAchievement: models.AchievementMini,
		},
		{
			name: "Boolean false manual scoring",
			habit: models.Habit{
				Title:       "Daily Exercise",
				HabitType:   models.SimpleHabit,
				ScoringType: models.ManualScoring,
				FieldType:   models.FieldType{Type: models.BooleanFieldType},
			},
			value:               false,
			notes:               "Skipped today",
			expectedAchievement: models.AchievementNone,
		},
		{
			name: "Text field with content",
			habit: models.Habit{
				Title:       "Reflection",
				HabitType:   models.SimpleHabit,
				ScoringType: models.ManualScoring,
				FieldType:   models.FieldType{Type: models.TextFieldType},
			},
			value:               "Today was productive and I learned something new",
			notes:               "",
			expectedAchievement: models.AchievementMini,
		},
		{
			name: "Numeric field with value",
			habit: models.Habit{
				Title:       "Water Intake",
				HabitType:   models.SimpleHabit,
				ScoringType: models.ManualScoring,
				FieldType:   models.FieldType{Type: models.UnsignedIntFieldType, Unit: "glasses"},
			},
			value:               8,
			notes:               "Stayed hydrated",
			expectedAchievement: models.AchievementMini,
		},
		{
			name: "Automatic scoring with simple criteria",
			habit: models.Habit{
				Title:       "Steps Habit",
				ID:          "steps_habit",
				HabitType:   models.SimpleHabit, // Simple habit with automatic scoring
				ScoringType: models.AutomaticScoring,
				FieldType:   models.FieldType{Type: models.UnsignedIntFieldType, Unit: "steps"},
				Criteria: &models.Criteria{
					Description: "At least 5000 steps",
					Condition: &models.Condition{
						GreaterThanOrEqual: float64Ptr(5000),
					},
				},
			},
			value:               7500,
			notes:               "Good walking day",
			expectedAchievement: models.AchievementMini,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := flow.CollectEntryDirectly(tc.habit, tc.value, tc.notes, nil)
			if err != nil {
				t.Errorf("CollectEntryDirectly() unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Errorf("CollectEntryDirectly() returned nil result")
				return
			}

			// Check value
			if result.Value != tc.value {
				t.Errorf("Result value = %v, want %v", result.Value, tc.value)
			}

			// Check notes
			if result.Notes != tc.notes {
				t.Errorf("Result notes = %v, want %v", result.Notes, tc.notes)
			}

			// Check achievement level
			if result.AchievementLevel == nil {
				t.Errorf("Result AchievementLevel is nil")
				return
			}

			if *result.AchievementLevel != tc.expectedAchievement {
				t.Errorf("Result AchievementLevel = %v, want %v", *result.AchievementLevel, tc.expectedAchievement)
			}
		})
	}
}

func TestSimpleHabitCollectEntryWithExisting(t *testing.T) {
	factory := NewEntryFieldInputFactory()
	flow := NewSimpleHabitCollectionFlowForTesting(factory, nil)

	habit := models.Habit{
		Title:       "Daily Meditation",
		HabitType:   models.SimpleHabit,
		ScoringType: models.ManualScoring,
		FieldType:   models.FieldType{Type: models.BooleanFieldType},
	}

	existingLevel := models.AchievementMini
	existing := &ExistingEntry{
		Value:            true,
		Notes:            "Previous meditation session",
		AchievementLevel: &existingLevel,
	}

	result, err := flow.CollectEntryDirectly(habit, false, "Updated entry", existing)
	if err != nil {
		t.Errorf("CollectEntryDirectly() unexpected error: %v", err)
		return
	}

	if result == nil {
		t.Errorf("CollectEntryDirectly() returned nil result")
		return
	}

	// Check that new value and notes are used, not existing ones
	if result.Value != false {
		t.Errorf("Result value = %v, want false", result.Value)
	}

	if result.Notes != "Updated entry" {
		t.Errorf("Result notes = %v, want 'Updated entry'", result.Notes)
	}

	// Achievement level should be recalculated based on new value
	if result.AchievementLevel == nil {
		t.Errorf("Result AchievementLevel is nil")
		return
	}

	if *result.AchievementLevel != models.AchievementNone {
		t.Errorf("Result AchievementLevel = %v, want %v", *result.AchievementLevel, models.AchievementNone)
	}
}

// Helper functions for testing

func float64Ptr(f float64) *float64 {
	return &f
}
