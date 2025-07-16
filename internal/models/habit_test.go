package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHabit_Validate(t *testing.T) {
	t.Run("valid simple boolean habit with manual scoring", func(t *testing.T) {
		habit := Habit{
			Title:     "Morning Meditation",
			Position:  1,
			HabitType: SimpleHabit,
			FieldType: FieldType{
				Type: BooleanFieldType,
			},
			ScoringType: ManualScoring,
		}

		err := habit.Validate()
		require.NoError(t, err)

		// ID should be auto-generated
		assert.Equal(t, "morning_meditation", habit.ID)
	})

	t.Run("valid simple boolean habit with automatic scoring", func(t *testing.T) {
		habit := Habit{
			Title:     "Daily Exercise",
			Position:  1,
			HabitType: SimpleHabit,
			FieldType: FieldType{
				Type: BooleanFieldType,
			},
			ScoringType: AutomaticScoring,
			Criteria: &Criteria{
				Description: "Exercise completed",
				Condition: &Condition{
					Equals: boolPtr(true),
				},
			},
		}

		err := habit.Validate()
		require.NoError(t, err)
		assert.Equal(t, "daily_exercise", habit.ID)
	})

	t.Run("custom ID is preserved", func(t *testing.T) {
		habit := Habit{
			Title:     "Morning Meditation",
			ID:        "custom_meditation_id",
			Position:  1,
			HabitType: SimpleHabit,
			FieldType: FieldType{
				Type: BooleanFieldType,
			},
			ScoringType: ManualScoring,
		}

		err := habit.Validate()
		require.NoError(t, err)
		assert.Equal(t, "custom_meditation_id", habit.ID)
	})

	t.Run("title is required", func(t *testing.T) {
		habit := Habit{
			Position:  1,
			HabitType: SimpleHabit,
			FieldType: FieldType{
				Type: BooleanFieldType,
			},
			ScoringType: ManualScoring,
		}

		err := habit.Validate()
		assert.EqualError(t, err, "habit title is required")
	})

	t.Run("habit type is required", func(t *testing.T) {
		habit := Habit{
			Title:    "Test Habit",
			Position: 1,
			FieldType: FieldType{
				Type: BooleanFieldType,
			},
			ScoringType: ManualScoring,
		}

		err := habit.Validate()
		assert.EqualError(t, err, "habit_type is required")
	})

	t.Run("invalid habit type", func(t *testing.T) {
		habit := Habit{
			Title:     "Test Habit",
			Position:  1,
			HabitType: "invalid_type",
			FieldType: FieldType{
				Type: BooleanFieldType,
			},
			ScoringType: ManualScoring,
		}

		err := habit.Validate()
		assert.EqualError(t, err, "invalid habit_type: invalid_type")
	})

	t.Run("scoring type required for simple habits", func(t *testing.T) {
		habit := Habit{
			Title:     "Test Habit",
			Position:  1,
			HabitType: SimpleHabit,
			FieldType: FieldType{
				Type: BooleanFieldType,
			},
		}

		err := habit.Validate()
		assert.EqualError(t, err, "scoring_type is required for simple habits")
	})

	t.Run("criteria required for automatic scoring", func(t *testing.T) {
		habit := Habit{
			Title:       "Test Habit",
			Position:    1,
			HabitType:   SimpleHabit,
			FieldType:   FieldType{Type: BooleanFieldType},
			ScoringType: AutomaticScoring,
		}

		err := habit.Validate()
		assert.EqualError(t, err, "criteria is required for automatic scoring")
	})

	t.Run("invalid ID characters", func(t *testing.T) {
		habit := Habit{
			Title:     "Test Habit",
			ID:        "invalid-id-with-dashes",
			Position:  1,
			HabitType: SimpleHabit,
			FieldType: FieldType{
				Type: BooleanFieldType,
			},
			ScoringType: ManualScoring,
		}

		err := habit.Validate()
		assert.Contains(t, err.Error(), "habit ID 'invalid-id-with-dashes' is invalid")
	})
}

func TestFieldType_Validate(t *testing.T) {
	t.Run("valid boolean field", func(t *testing.T) {
		ft := FieldType{Type: BooleanFieldType}
		err := ft.Validate()
		assert.NoError(t, err)
	})

	t.Run("valid text field", func(t *testing.T) {
		ft := FieldType{
			Type:      TextFieldType,
			Multiline: boolPtr(true),
		}
		err := ft.Validate()
		assert.NoError(t, err)
	})

	t.Run("valid unsigned int field with constraints", func(t *testing.T) {
		ft := FieldType{
			Type: UnsignedIntFieldType,
			Unit: "count",
			Min:  float64Ptr(0),
			Max:  float64Ptr(100),
		}
		err := ft.Validate()
		assert.NoError(t, err)
	})

	t.Run("field type is required", func(t *testing.T) {
		ft := FieldType{}
		err := ft.Validate()
		assert.EqualError(t, err, "field type is required")
	})

	t.Run("unsigned field cannot have negative min", func(t *testing.T) {
		ft := FieldType{
			Type: UnsignedIntFieldType,
			Min:  float64Ptr(-1),
		}
		err := ft.Validate()
		assert.EqualError(t, err, "unsigned fields cannot have negative min value")
	})

	t.Run("min cannot be greater than max", func(t *testing.T) {
		ft := FieldType{
			Type: DecimalFieldType,
			Min:  float64Ptr(10),
			Max:  float64Ptr(5),
		}
		err := ft.Validate()
		assert.EqualError(t, err, "min value (10) cannot be greater than max value (5)")
	})

	t.Run("invalid field type", func(t *testing.T) {
		ft := FieldType{Type: "invalid_type"}
		err := ft.Validate()
		assert.EqualError(t, err, "unknown field type: invalid_type")
	})

	t.Run("time field with invalid format", func(t *testing.T) {
		ft := FieldType{
			Type:   TimeFieldType,
			Format: "invalid_format",
		}
		err := ft.Validate()
		assert.EqualError(t, err, "time fields only support HH:MM format")
	})

	t.Run("duration field with invalid format", func(t *testing.T) {
		ft := FieldType{
			Type:   DurationFieldType,
			Format: "invalid_format",
		}
		err := ft.Validate()
		assert.Contains(t, err.Error(), "duration format must be one of")
	})
}

func TestSchema_Validate(t *testing.T) {
	t.Run("valid schema with simple boolean habit", func(t *testing.T) {
		schema := Schema{
			Version:     "1.0.0",
			CreatedDate: "2024-01-01",
			Habits: []Habit{
				{
					Title:     "Morning Meditation",
					Position:  1,
					HabitType: SimpleHabit,
					FieldType: FieldType{
						Type: BooleanFieldType,
					},
					ScoringType: ManualScoring,
				},
			},
		}

		err := schema.Validate()
		assert.NoError(t, err)
	})

	t.Run("version is required", func(t *testing.T) {
		schema := Schema{
			Habits: []Habit{},
		}

		err := schema.Validate()
		assert.EqualError(t, err, "schema version is required")
	})

	t.Run("invalid created date format", func(t *testing.T) {
		schema := Schema{
			Version:     "1.0.0",
			CreatedDate: "invalid-date",
			Habits:      []Habit{},
		}

		err := schema.Validate()
		assert.Contains(t, err.Error(), "invalid created_date format")
	})

	t.Run("duplicate habit IDs", func(t *testing.T) {
		schema := Schema{
			Version: "1.0.0",
			Habits: []Habit{
				{
					Title:       "Habit 1",
					ID:          "duplicate_id",
					Position:    1,
					HabitType:   SimpleHabit,
					FieldType:   FieldType{Type: BooleanFieldType},
					ScoringType: ManualScoring,
				},
				{
					Title:       "Habit 2",
					ID:          "duplicate_id",
					Position:    2,
					HabitType:   SimpleHabit,
					FieldType:   FieldType{Type: BooleanFieldType},
					ScoringType: ManualScoring,
				},
			},
		}

		err := schema.Validate()
		assert.EqualError(t, err, "duplicate habit ID: duplicate_id")
	})

	t.Run("positions are auto-assigned from array order", func(t *testing.T) {
		schema := Schema{
			Version: "1.0.0",
			Habits: []Habit{
				{
					Title:       "Habit 1",
					Position:    0, // Will be auto-assigned to 1
					HabitType:   SimpleHabit,
					FieldType:   FieldType{Type: BooleanFieldType},
					ScoringType: ManualScoring,
				},
				{
					Title:       "Habit 2",
					Position:    0, // Will be auto-assigned to 2
					HabitType:   SimpleHabit,
					FieldType:   FieldType{Type: BooleanFieldType},
					ScoringType: ManualScoring,
				},
			},
		}

		err := schema.Validate()
		assert.NoError(t, err)

		// Check that positions were auto-assigned
		assert.Equal(t, 1, schema.Habits[0].Position)
		assert.Equal(t, 2, schema.Habits[1].Position)
	})
}

func TestGenerateIDFromTitle(t *testing.T) {
	tests := []struct {
		title    string
		expected string
	}{
		{"Morning Meditation", "morning_meditation"},
		{"Daily Exercise!", "daily_exercise"},
		{"Sleep Quality (1-10)", "sleep_quality_1_10"},
		{"   Spaced   Out   ", "spaced_out"},
		{"Special@Characters#Here", "special_characters_here"},
		{"", "unnamed_habit"},
		{"123 Numbers", "123_numbers"},
		{"___underscores___", "underscores"},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			result := generateIDFromTitle(tt.title)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsValidID(t *testing.T) {
	tests := []struct {
		id       string
		expected bool
	}{
		{"valid_id", true},
		{"valid123", true},
		{"123valid", true},
		{"valid_id_123", true},
		{"", false},
		{"invalid-id", false},
		{"invalid.id", false},
		{"invalid id", false},
		{"Invalid_ID", false}, // uppercase not allowed
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			result := isValidID(tt.id)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHabit_ElasticValidation(t *testing.T) {
	t.Run("valid elastic habit with automatic scoring", func(t *testing.T) {
		habit := Habit{
			Title:     "Daily Exercise",
			Position:  1,
			HabitType: ElasticHabit,
			FieldType: FieldType{
				Type:   DurationFieldType,
				Format: "minutes",
			},
			ScoringType: AutomaticScoring,
			MiniCriteria: &Criteria{
				Description: "Minimum 15 minutes",
				Condition: &Condition{
					GreaterThanOrEqual: float64Ptr(15),
				},
			},
			MidiCriteria: &Criteria{
				Description: "Target 30 minutes",
				Condition: &Condition{
					GreaterThanOrEqual: float64Ptr(30),
				},
			},
			MaxiCriteria: &Criteria{
				Description: "Excellent 60+ minutes",
				Condition: &Condition{
					GreaterThanOrEqual: float64Ptr(60),
				},
			},
		}

		err := habit.Validate()
		require.NoError(t, err)
		assert.Equal(t, "daily_exercise", habit.ID)
	})

	t.Run("valid elastic habit with manual scoring", func(t *testing.T) {
		habit := Habit{
			Title:     "Reading Time",
			Position:  1,
			HabitType: ElasticHabit,
			FieldType: FieldType{
				Type:   DurationFieldType,
				Format: "minutes",
			},
			ScoringType: ManualScoring,
		}

		err := habit.Validate()
		require.NoError(t, err)
		assert.Equal(t, "reading_time", habit.ID)
	})

	t.Run("elastic habit requires scoring type", func(t *testing.T) {
		habit := Habit{
			Title:     "Exercise",
			Position:  1,
			HabitType: ElasticHabit,
			FieldType: FieldType{
				Type:   DurationFieldType,
				Format: "minutes",
			},
		}

		err := habit.Validate()
		assert.EqualError(t, err, "scoring_type is required for elastic habits")
	})

	t.Run("elastic habit with automatic scoring requires mini criteria", func(t *testing.T) {
		habit := Habit{
			Title:     "Exercise",
			Position:  1,
			HabitType: ElasticHabit,
			FieldType: FieldType{
				Type:   DurationFieldType,
				Format: "minutes",
			},
			ScoringType: AutomaticScoring,
		}

		err := habit.Validate()
		assert.EqualError(t, err, "mini_criteria is required for automatic scoring of elastic habits")
	})

	t.Run("elastic habit with automatic scoring requires midi criteria", func(t *testing.T) {
		habit := Habit{
			Title:     "Exercise",
			Position:  1,
			HabitType: ElasticHabit,
			FieldType: FieldType{
				Type:   DurationFieldType,
				Format: "minutes",
			},
			ScoringType: AutomaticScoring,
			MiniCriteria: &Criteria{
				Condition: &Condition{GreaterThanOrEqual: float64Ptr(15)},
			},
		}

		err := habit.Validate()
		assert.EqualError(t, err, "midi_criteria is required for automatic scoring of elastic habits")
	})

	t.Run("elastic habit with automatic scoring requires maxi criteria", func(t *testing.T) {
		habit := Habit{
			Title:     "Exercise",
			Position:  1,
			HabitType: ElasticHabit,
			FieldType: FieldType{
				Type:   DurationFieldType,
				Format: "minutes",
			},
			ScoringType: AutomaticScoring,
			MiniCriteria: &Criteria{
				Condition: &Condition{GreaterThanOrEqual: float64Ptr(15)},
			},
			MidiCriteria: &Criteria{
				Condition: &Condition{GreaterThanOrEqual: float64Ptr(30)},
			},
		}

		err := habit.Validate()
		assert.EqualError(t, err, "maxi_criteria is required for automatic scoring of elastic habits")
	})
}

func TestHabit_HelperMethods(t *testing.T) {
	t.Run("IsElastic", func(t *testing.T) {
		tests := []struct {
			habitType HabitType
			expected  bool
		}{
			{ElasticHabit, true},
			{SimpleHabit, false},
			{InformationalHabit, false},
		}

		for _, tt := range tests {
			habit := Habit{HabitType: tt.habitType}
			assert.Equal(t, tt.expected, habit.IsElastic())
		}
	})

	t.Run("IsSimple", func(t *testing.T) {
		tests := []struct {
			habitType HabitType
			expected  bool
		}{
			{SimpleHabit, true},
			{ElasticHabit, false},
			{InformationalHabit, false},
		}

		for _, tt := range tests {
			habit := Habit{HabitType: tt.habitType}
			assert.Equal(t, tt.expected, habit.IsSimple())
		}
	})

	t.Run("IsInformational", func(t *testing.T) {
		tests := []struct {
			habitType HabitType
			expected  bool
		}{
			{InformationalHabit, true},
			{SimpleHabit, false},
			{ElasticHabit, false},
		}

		for _, tt := range tests {
			habit := Habit{HabitType: tt.habitType}
			assert.Equal(t, tt.expected, habit.IsInformational())
		}
	})

	t.Run("RequiresAutomaticScoring", func(t *testing.T) {
		tests := []struct {
			scoringType ScoringType
			expected    bool
		}{
			{AutomaticScoring, true},
			{ManualScoring, false},
		}

		for _, tt := range tests {
			habit := Habit{ScoringType: tt.scoringType}
			assert.Equal(t, tt.expected, habit.RequiresAutomaticScoring())
		}
	})

	t.Run("RequiresManualScoring", func(t *testing.T) {
		tests := []struct {
			scoringType ScoringType
			expected    bool
		}{
			{ManualScoring, true},
			{AutomaticScoring, false},
		}

		for _, tt := range tests {
			habit := Habit{ScoringType: tt.scoringType}
			assert.Equal(t, tt.expected, habit.RequiresManualScoring())
		}
	})
}

// Helper functions for creating pointers
func boolPtr(b bool) *bool {
	return &b
}

func float64Ptr(f float64) *float64 {
	return &f
}
