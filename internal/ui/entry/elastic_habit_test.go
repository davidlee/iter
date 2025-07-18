package entry

import (
	"fmt"
	"testing"

	"github.com/davidlee/vice/internal/models"
	"github.com/davidlee/vice/internal/scoring"
)

// AIDEV-NOTE: elastic-habit-tests; comprehensive testing for elastic habit collection flow
// Tests three-tier achievement system (Mini/Midi/Maxi), automatic/manual scoring, and field type integration for T010/3.2
// AIDEV-NOTE: testing-patterns; uses NewElasticHabitCollectionFlowForTesting() and CollectEntryDirectly() for headless testing
// All major scenarios covered: numeric thresholds, boolean achievement, text fields, automatic scoring with criteria

func TestElasticHabitCollectionFlow(t *testing.T) {
	factory := NewEntryFieldInputFactory()
	scoringEngine := &scoring.Engine{} // Mock or real scoring engine

	flow := NewElasticHabitCollectionFlow(factory, scoringEngine)

	// Test flow type identification
	if flow.GetFlowType() != string(models.ElasticHabit) {
		t.Errorf("GetFlowType() = %v, want %v", flow.GetFlowType(), string(models.ElasticHabit))
	}

	// Test scoring requirement
	if !flow.RequiresScoring() {
		t.Errorf("RequiresScoring() expected true for elastic habits")
	}

	// Test supported field types (elastic habits support all field types)
	expectedFieldTypes := []string{
		models.BooleanFieldType,
		models.TextFieldType,
		models.UnsignedIntFieldType,
		models.UnsignedDecimalFieldType,
		models.DecimalFieldType,
		models.TimeFieldType,
		models.DurationFieldType,
		models.ChecklistFieldType,
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
}

func TestElasticHabitTestingAchievementLevels(t *testing.T) {
	factory := NewEntryFieldInputFactory()
	flow := NewElasticHabitCollectionFlow(factory, nil) // No scoring engine for manual tests

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
			name:      "Integer Maxi level (≥100)",
			fieldType: models.FieldType{Type: models.UnsignedIntFieldType, Unit: "steps"},
			value:     150,
			expected:  models.AchievementMaxi,
		},
		{
			name:      "Integer Midi level (≥50, <100)",
			fieldType: models.FieldType{Type: models.UnsignedIntFieldType, Unit: "steps"},
			value:     75,
			expected:  models.AchievementMidi,
		},
		{
			name:      "Integer Mini level (>0, <50)",
			fieldType: models.FieldType{Type: models.UnsignedIntFieldType, Unit: "steps"},
			value:     25,
			expected:  models.AchievementMini,
		},
		{
			name:      "Integer zero",
			fieldType: models.FieldType{Type: models.UnsignedIntFieldType, Unit: "steps"},
			value:     0,
			expected:  models.AchievementNone,
		},
		{
			name:      "Float Maxi level",
			fieldType: models.FieldType{Type: models.DecimalFieldType, Unit: "kg"},
			value:     120.5,
			expected:  models.AchievementMaxi,
		},
		{
			name:      "Float Midi level",
			fieldType: models.FieldType{Type: models.DecimalFieldType, Unit: "kg"},
			value:     65.0,
			expected:  models.AchievementMidi,
		},
		{
			name:      "Float Mini level",
			fieldType: models.FieldType{Type: models.DecimalFieldType, Unit: "kg"},
			value:     10.5,
			expected:  models.AchievementMini,
		},
		{
			name:      "Text with content",
			fieldType: models.FieldType{Type: models.TextFieldType},
			value:     "Some reflection text",
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
			value:     "45m",
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
				Title:     "Test Elastic Habit",
				HabitType: models.ElasticHabit,
				FieldType: tc.fieldType,
			}

			result := flow.determineTestingAchievementLevel(habit, tc.value)
			if result == nil {
				t.Errorf("determineTestingAchievementLevel() returned nil")
				return
			}

			if *result != tc.expected {
				t.Errorf("determineTestingAchievementLevel() = %v, want %v", *result, tc.expected)
			}
		})
	}
}

func TestElasticHabitAutomaticScoringWithRealEngine(t *testing.T) {
	factory := NewEntryFieldInputFactory()
	scoringEngine := scoring.NewEngine() // Use real scoring engine
	flow := NewElasticHabitCollectionFlow(factory, scoringEngine)

	// Test with a elastic habit that has three-tier criteria
	habit := models.Habit{
		Title:       "Daily Steps",
		HabitType:   models.ElasticHabit,
		ScoringType: models.AutomaticScoring,
		FieldType:   models.FieldType{Type: models.UnsignedIntFieldType, Unit: "steps"},
		MiniCriteria: &models.Criteria{
			Description: "At least 5000 steps",
			Condition: &models.Condition{
				GreaterThanOrEqual: float64Ptr(5000),
			},
		},
		MidiCriteria: &models.Criteria{
			Description: "At least 8000 steps",
			Condition: &models.Condition{
				GreaterThanOrEqual: float64Ptr(8000),
			},
		},
		MaxiCriteria: &models.Criteria{
			Description: "At least 12000 steps",
			Condition: &models.Condition{
				GreaterThanOrEqual: float64Ptr(12000),
			},
		},
	}

	testCases := []struct {
		name     string
		value    interface{}
		expected models.AchievementLevel
	}{
		{
			name:     "Maxi achievement (≥12000)",
			value:    float64(15000),
			expected: models.AchievementMaxi,
		},
		{
			name:     "Midi achievement (≥8000, <12000)",
			value:    float64(10000),
			expected: models.AchievementMidi,
		},
		{
			name:     "Mini achievement (≥5000, <8000)",
			value:    float64(6000),
			expected: models.AchievementMini,
		},
		{
			name:     "No achievement (<5000)",
			value:    float64(3000),
			expected: models.AchievementNone,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := flow.performElasticScoring(habit, tc.value)
			if err != nil {
				t.Errorf("performElasticScoring() unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Errorf("performElasticScoring() returned nil result")
				return
			}

			if *result != tc.expected {
				t.Errorf("performElasticScoring() = %v, want %v", *result, tc.expected)
			}
		})
	}
}

func TestElasticHabitAutomaticScoringNoEngine(t *testing.T) {
	factory := NewEntryFieldInputFactory()
	flow := NewElasticHabitCollectionFlow(factory, nil) // No scoring engine

	habit := models.Habit{
		Title:       "Test Habit",
		HabitType:   models.ElasticHabit,
		ScoringType: models.AutomaticScoring,
		FieldType:   models.FieldType{Type: models.UnsignedIntFieldType},
	}

	_, err := flow.performElasticScoring(habit, 100)
	if err == nil {
		t.Errorf("performElasticScoring() expected error when no scoring engine provided")
	}
}

func TestElasticHabitFieldTypeSupport(t *testing.T) {
	factory := NewEntryFieldInputFactory()
	flow := NewElasticHabitCollectionFlow(factory, nil)

	// Test that elastic habits support all field types including checklist
	supportedTypes := flow.GetExpectedFieldTypes()

	// Should support all types including checklist
	expectedSupported := []string{
		models.BooleanFieldType,
		models.TextFieldType,
		models.UnsignedIntFieldType,
		models.UnsignedDecimalFieldType,
		models.DecimalFieldType,
		models.TimeFieldType,
		models.DurationFieldType,
		models.ChecklistFieldType,
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
			t.Errorf("Elastic habits should support field type %v", expectedType)
		}
	}
}

func TestElasticHabitFlowInterface(t *testing.T) {
	factory := NewEntryFieldInputFactory()
	scoringEngine := &scoring.Engine{}
	flow := NewElasticHabitCollectionFlow(factory, scoringEngine)

	// Verify that ElasticHabitCollectionFlow implements HabitCollectionFlow interface
	var _ HabitCollectionFlow = flow

	// Test interface methods
	if flow.GetFlowType() != string(models.ElasticHabit) {
		t.Errorf("GetFlowType() = %v, want %v", flow.GetFlowType(), string(models.ElasticHabit))
	}

	if !flow.RequiresScoring() {
		t.Errorf("RequiresScoring() = false, want true")
	}

	fieldTypes := flow.GetExpectedFieldTypes()
	if len(fieldTypes) == 0 {
		t.Errorf("GetExpectedFieldTypes() returned empty slice")
	}
}

func TestElasticHabitCollectEntryDirectly(t *testing.T) {
	factory := NewEntryFieldInputFactory()
	scoringEngine := scoring.NewEngine()
	flow := NewElasticHabitCollectionFlowForTesting(factory, scoringEngine)

	testCases := []struct {
		name                string
		habit               models.Habit
		value               interface{}
		notes               string
		expectedAchievement models.AchievementLevel
	}{
		{
			name: "Manual scoring numeric Maxi",
			habit: models.Habit{
				Title:       "Push-ups",
				HabitType:   models.ElasticHabit,
				ScoringType: models.ManualScoring,
				FieldType:   models.FieldType{Type: models.UnsignedIntFieldType, Unit: "reps"},
			},
			value:               150,
			notes:               "Great workout session",
			expectedAchievement: models.AchievementMaxi,
		},
		{
			name: "Manual scoring numeric Midi",
			habit: models.Habit{
				Title:       "Push-ups",
				HabitType:   models.ElasticHabit,
				ScoringType: models.ManualScoring,
				FieldType:   models.FieldType{Type: models.UnsignedIntFieldType, Unit: "reps"},
			},
			value:               75,
			notes:               "Good effort",
			expectedAchievement: models.AchievementMidi,
		},
		{
			name: "Manual scoring numeric Mini",
			habit: models.Habit{
				Title:       "Push-ups",
				HabitType:   models.ElasticHabit,
				ScoringType: models.ManualScoring,
				FieldType:   models.FieldType{Type: models.UnsignedIntFieldType, Unit: "reps"},
			},
			value:               25,
			notes:               "Started well",
			expectedAchievement: models.AchievementMini,
		},
		{
			name: "Manual scoring numeric None",
			habit: models.Habit{
				Title:       "Push-ups",
				HabitType:   models.ElasticHabit,
				ScoringType: models.ManualScoring,
				FieldType:   models.FieldType{Type: models.UnsignedIntFieldType, Unit: "reps"},
			},
			value:               0,
			notes:               "Rest day",
			expectedAchievement: models.AchievementNone,
		},
		{
			name: "Boolean field true",
			habit: models.Habit{
				Title:       "Meditation",
				HabitType:   models.ElasticHabit,
				ScoringType: models.ManualScoring,
				FieldType:   models.FieldType{Type: models.BooleanFieldType},
			},
			value:               true,
			notes:               "20 minutes of meditation",
			expectedAchievement: models.AchievementMini,
		},
		{
			name: "Text field with content",
			habit: models.Habit{
				Title:       "Reflection",
				HabitType:   models.ElasticHabit,
				ScoringType: models.ManualScoring,
				FieldType:   models.FieldType{Type: models.TextFieldType},
			},
			value:               "Today I learned about mindfulness and practiced breathing exercises",
			notes:               "",
			expectedAchievement: models.AchievementMini,
		},
		{
			name: "Automatic scoring with three-tier criteria",
			habit: models.Habit{
				Title:       "Daily Steps",
				HabitType:   models.ElasticHabit,
				ScoringType: models.AutomaticScoring,
				FieldType:   models.FieldType{Type: models.UnsignedIntFieldType, Unit: "steps"},
				MiniCriteria: &models.Criteria{
					Description: "At least 5000 steps",
					Condition: &models.Condition{
						GreaterThanOrEqual: float64Ptr(5000),
					},
				},
				MidiCriteria: &models.Criteria{
					Description: "At least 8000 steps",
					Condition: &models.Condition{
						GreaterThanOrEqual: float64Ptr(8000),
					},
				},
				MaxiCriteria: &models.Criteria{
					Description: "At least 12000 steps",
					Condition: &models.Condition{
						GreaterThanOrEqual: float64Ptr(12000),
					},
				},
			},
			value:               15000,
			notes:               "Long hiking day",
			expectedAchievement: models.AchievementMaxi,
		},
		{
			name: "Automatic scoring Midi level",
			habit: models.Habit{
				Title:       "Daily Steps",
				HabitType:   models.ElasticHabit,
				ScoringType: models.AutomaticScoring,
				FieldType:   models.FieldType{Type: models.UnsignedIntFieldType, Unit: "steps"},
				MiniCriteria: &models.Criteria{
					Description: "At least 5000 steps",
					Condition: &models.Condition{
						GreaterThanOrEqual: float64Ptr(5000),
					},
				},
				MidiCriteria: &models.Criteria{
					Description: "At least 8000 steps",
					Condition: &models.Condition{
						GreaterThanOrEqual: float64Ptr(8000),
					},
				},
				MaxiCriteria: &models.Criteria{
					Description: "At least 12000 steps",
					Condition: &models.Condition{
						GreaterThanOrEqual: float64Ptr(12000),
					},
				},
			},
			value:               9500,
			notes:               "Active day",
			expectedAchievement: models.AchievementMidi,
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

func TestElasticHabitCollectEntryWithExisting(t *testing.T) {
	factory := NewEntryFieldInputFactory()
	flow := NewElasticHabitCollectionFlowForTesting(factory, nil)

	habit := models.Habit{
		Title:       "Daily Reading",
		HabitType:   models.ElasticHabit,
		ScoringType: models.ManualScoring,
		FieldType:   models.FieldType{Type: models.UnsignedIntFieldType, Unit: "pages"},
	}

	existingLevel := models.AchievementMidi
	existing := &ExistingEntry{
		Value:            50,
		Notes:            "Previous reading session",
		AchievementLevel: &existingLevel,
	}

	result, err := flow.CollectEntryDirectly(habit, 120, "Updated reading progress", existing)
	if err != nil {
		t.Errorf("CollectEntryDirectly() unexpected error: %v", err)
		return
	}

	if result == nil {
		t.Errorf("CollectEntryDirectly() returned nil result")
		return
	}

	// Check that new value and notes are used, not existing ones
	if result.Value != 120 {
		t.Errorf("Result value = %v, want 120", result.Value)
	}

	if result.Notes != "Updated reading progress" {
		t.Errorf("Result notes = %v, want 'Updated reading progress'", result.Notes)
	}

	// Achievement level should be recalculated based on new value (120 >= 100 = Maxi)
	if result.AchievementLevel == nil {
		t.Errorf("Result AchievementLevel is nil")
		return
	}

	if *result.AchievementLevel != models.AchievementMaxi {
		t.Errorf("Result AchievementLevel = %v, want %v", *result.AchievementLevel, models.AchievementMaxi)
	}
}

func TestElasticHabitThreeTierAchievementLogic(t *testing.T) {
	factory := NewEntryFieldInputFactory()
	flow := NewElasticHabitCollectionFlowForTesting(factory, nil)

	habit := models.Habit{
		Title:       "Workout Intensity",
		HabitType:   models.ElasticHabit,
		ScoringType: models.ManualScoring,
		FieldType:   models.FieldType{Type: models.DecimalFieldType, Unit: "intensity"},
	}

	// Test the three-tier achievement system with decimal values
	testCases := []struct {
		value    float64
		expected models.AchievementLevel
	}{
		{150.0, models.AchievementMaxi}, // ≥100
		{100.0, models.AchievementMaxi}, // ≥100
		{99.9, models.AchievementMidi},  // ≥50, <100
		{75.0, models.AchievementMidi},  // ≥50, <100
		{50.0, models.AchievementMidi},  // ≥50, <100
		{49.9, models.AchievementMini},  // >0, <50
		{25.0, models.AchievementMini},  // >0, <50
		{0.1, models.AchievementMini},   // >0, <50
		{0.0, models.AchievementNone},   // =0
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Value_%.1f", tc.value), func(t *testing.T) {
			result, err := flow.CollectEntryDirectly(habit, tc.value, "", nil)
			if err != nil {
				t.Errorf("CollectEntryDirectly() unexpected error: %v", err)
				return
			}

			if result.AchievementLevel == nil {
				t.Errorf("AchievementLevel is nil for value %.1f", tc.value)
				return
			}

			if *result.AchievementLevel != tc.expected {
				t.Errorf("Value %.1f: got %v, want %v", tc.value, *result.AchievementLevel, tc.expected)
			}
		})
	}
}
