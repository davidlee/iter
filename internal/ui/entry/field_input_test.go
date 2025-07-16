package entry

import (
	"testing"

	"davidlee/vice/internal/models"
)

// AIDEV-NOTE: entry-field-input-tests; unit tests for field input components with validation testing
// Tests core functionality, validation, and factory patterns for T010/2.1 implementation

func TestEntryFieldInputFactory(t *testing.T) {
	factory := NewEntryFieldInputFactory()

	tests := []struct {
		name      string
		fieldType models.FieldType
		wantType  string
		shouldErr bool
	}{
		{
			name: "Boolean field type",
			fieldType: models.FieldType{
				Type: models.BooleanFieldType,
			},
			wantType: models.BooleanFieldType,
		},
		{
			name: "Text field type",
			fieldType: models.FieldType{
				Type: models.TextFieldType,
			},
			wantType: models.TextFieldType,
		},
		{
			name: "Numeric field type",
			fieldType: models.FieldType{
				Type: models.UnsignedIntFieldType,
				Unit: "minutes",
			},
			wantType: models.UnsignedIntFieldType,
		},
		{
			name: "Time field type",
			fieldType: models.FieldType{
				Type: models.TimeFieldType,
			},
			wantType: models.TimeFieldType,
		},
		{
			name: "Duration field type",
			fieldType: models.FieldType{
				Type: models.DurationFieldType,
			},
			wantType: models.DurationFieldType,
		},
		{
			name: "Checklist field type",
			fieldType: models.FieldType{
				Type:        models.ChecklistFieldType,
				ChecklistID: "test-checklist",
			},
			wantType: models.ChecklistFieldType,
		},
		{
			name: "Unsupported field type",
			fieldType: models.FieldType{
				Type: "invalid",
			},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := EntryFieldInputConfig{
				Habit: models.Habit{
					Title: "Test Habit",
				},
				FieldType:     tt.fieldType,
				ExistingEntry: nil,
				ShowScoring:   false,
			}

			input, err := factory.CreateInput(config)

			if tt.shouldErr {
				if err == nil {
					t.Errorf("CreateInput() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("CreateInput() unexpected error: %v", err)
				return
			}

			if input == nil {
				t.Errorf("CreateInput() returned nil input")
				return
			}

			if input.GetFieldType() != tt.wantType {
				t.Errorf("GetFieldType() = %v, want %v", input.GetFieldType(), tt.wantType)
			}
		})
	}
}

func TestBooleanEntryInput(t *testing.T) {
	config := EntryFieldInputConfig{
		Habit: models.Habit{
			Title:  "Test Boolean Habit",
			Prompt: "Complete this task?",
		},
		FieldType: models.FieldType{
			Type: models.BooleanFieldType,
		},
		ExistingEntry: nil,
		ShowScoring:   false,
	}

	input := NewBooleanEntryInput(config)

	// Test field type
	if input.GetFieldType() != models.BooleanFieldType {
		t.Errorf("GetFieldType() = %v, want %v", input.GetFieldType(), models.BooleanFieldType)
	}

	// Test validation (should always pass for boolean)
	if err := input.Validate(); err != nil {
		t.Errorf("Validate() unexpected error: %v", err)
	}

	// Test setting existing value
	if err := input.SetExistingValue(true); err != nil {
		t.Errorf("SetExistingValue(true) unexpected error: %v", err)
	}

	if value := input.GetValue(); value != true {
		t.Errorf("GetValue() = %v, want true", value)
	}

	if strValue := input.GetStringValue(); strValue != "yes" {
		t.Errorf("GetStringValue() = %v, want 'yes'", strValue)
	}

	// Test invalid value type
	if err := input.SetExistingValue("invalid"); err == nil {
		t.Errorf("SetExistingValue('invalid') expected error, got nil")
	}
}

func TestBooleanEntryInputSkipFunctionality(t *testing.T) {
	config := EntryFieldInputConfig{
		Habit: models.Habit{
			Title:  "Test Boolean Habit",
			Prompt: "Did you complete this?",
		},
		FieldType: models.FieldType{
			Type: models.BooleanFieldType,
		},
		ExistingEntry: nil,
		ShowScoring:   false,
	}

	input := NewBooleanEntryInput(config)

	// Test skip option
	if err := input.SetExistingValue(nil); err != nil {
		t.Errorf("SetExistingValue(nil) unexpected error: %v", err)
	}

	if value := input.GetValue(); value != nil {
		t.Errorf("GetValue() for skip = %v, want nil", value)
	}

	if strValue := input.GetStringValue(); strValue != "skip" {
		t.Errorf("GetStringValue() for skip = %v, want 'skip'", strValue)
	}

	if status := input.GetStatus(); status != models.EntrySkipped {
		t.Errorf("GetStatus() for skip = %v, want EntrySkipped", status)
	}

	// Test No option
	if err := input.SetExistingValue(false); err != nil {
		t.Errorf("SetExistingValue(false) unexpected error: %v", err)
	}

	if value := input.GetValue(); value != false {
		t.Errorf("GetValue() for false = %v, want false", value)
	}

	if strValue := input.GetStringValue(); strValue != "no" {
		t.Errorf("GetStringValue() for false = %v, want 'no'", strValue)
	}

	if status := input.GetStatus(); status != models.EntryFailed {
		t.Errorf("GetStatus() for false = %v, want EntryFailed", status)
	}
}

func TestTextEntryInput(t *testing.T) {
	config := EntryFieldInputConfig{
		Habit: models.Habit{
			Title:  "Test Text Habit",
			Prompt: "Enter your thoughts",
		},
		FieldType: models.FieldType{
			Type: models.TextFieldType,
		},
		ExistingEntry: nil,
		ShowScoring:   false,
	}

	input := NewTextEntryInput(config)

	// Test field type
	if input.GetFieldType() != models.TextFieldType {
		t.Errorf("GetFieldType() = %v, want %v", input.GetFieldType(), models.TextFieldType)
	}

	// Test validation (should always pass for text)
	if err := input.Validate(); err != nil {
		t.Errorf("Validate() unexpected error: %v", err)
	}

	// Test setting existing value
	testText := "This is test text"
	if err := input.SetExistingValue(testText); err != nil {
		t.Errorf("SetExistingValue() unexpected error: %v", err)
	}

	if value := input.GetValue(); value != testText {
		t.Errorf("GetValue() = %v, want %v", value, testText)
	}

	if strValue := input.GetStringValue(); strValue != testText {
		t.Errorf("GetStringValue() = %v, want %v", strValue, testText)
	}

	// Test invalid value type
	if err := input.SetExistingValue(123); err == nil {
		t.Errorf("SetExistingValue(123) expected error, got nil")
	}
}

func TestNumericEntryInput(t *testing.T) {
	minVal := 0.0
	maxVal := 100.0
	config := EntryFieldInputConfig{
		Habit: models.Habit{
			Title:  "Test Numeric Habit",
			Prompt: "Enter a number",
		},
		FieldType: models.FieldType{
			Type: models.UnsignedIntFieldType,
			Unit: "points",
			Min:  &minVal,
			Max:  &maxVal,
		},
		ExistingEntry: nil,
		ShowScoring:   false,
	}

	input := NewNumericEntryInput(config)

	// Test field type
	if input.GetFieldType() != models.UnsignedIntFieldType {
		t.Errorf("GetFieldType() = %v, want %v", input.GetFieldType(), models.UnsignedIntFieldType)
	}

	// Test setting existing value
	testValue := 42.0
	if err := input.SetExistingValue(testValue); err != nil {
		t.Errorf("SetExistingValue() unexpected error: %v", err)
	}

	if strValue := input.GetStringValue(); strValue != "42" {
		t.Errorf("GetStringValue() = %v, want '42'", strValue)
	}
}

func TestScoringAwareInput(t *testing.T) {
	factory := NewEntryFieldInputFactory()

	config := EntryFieldInputConfig{
		Habit: models.Habit{
			Title:       "Test Scoring Habit",
			ScoringType: models.AutomaticScoring,
		},
		FieldType: models.FieldType{
			Type: models.BooleanFieldType,
		},
		ExistingEntry: nil,
		ShowScoring:   true,
	}

	scoringInput, err := factory.CreateScoringAwareInput(config)
	if err != nil {
		t.Errorf("CreateScoringAwareInput() unexpected error: %v", err)
		return
	}

	if scoringInput == nil {
		t.Errorf("CreateScoringAwareInput() returned nil")
		return
	}

	// Test that it implements the ScoringAwareInput interface
	if !scoringInput.CanShowScoring() {
		t.Errorf("CanShowScoring() expected true for automatic scoring habit")
	}

	// Test scoring display update (should not error)
	level := models.AchievementMini
	if err := scoringInput.UpdateScoringDisplay(&level); err != nil {
		t.Errorf("UpdateScoringDisplay() unexpected error: %v", err)
	}
}

func TestHabitCollectionFlowFactory(t *testing.T) {
	fieldInputFactory := NewEntryFieldInputFactory()
	factory := NewHabitCollectionFlowFactory(fieldInputFactory, nil, "checklists.yml")

	tests := []struct {
		name     string
		goalType string
		wantErr  bool
	}{
		{"Simple habit", string(models.SimpleHabit), false},
		{"Elastic habit", string(models.ElasticHabit), false},
		{"Informational habit", string(models.InformationalHabit), false},
		{"Checklist habit", string(models.ChecklistHabit), false},
		{"Invalid habit type", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flow, err := factory.CreateFlow(tt.goalType)

			if tt.wantErr {
				if err == nil {
					t.Errorf("CreateFlow() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("CreateFlow() unexpected error: %v", err)
				return
			}

			if flow == nil {
				t.Errorf("CreateFlow() returned nil flow")
				return
			}

			if flow.GetFlowType() != tt.goalType {
				t.Errorf("GetFlowType() = %v, want %v", flow.GetFlowType(), tt.goalType)
			}
		})
	}
}

func TestValidationPatterns(t *testing.T) {
	validator := NewFieldValidator()

	tests := []struct {
		name      string
		value     interface{}
		fieldType string
		wantValid bool
	}{
		{"Required string with value", "test", models.TextFieldType, true},
		{"Required string empty", "", models.TextFieldType, false},
		{"Required string nil", nil, models.TextFieldType, false},
		{"Boolean true", true, models.BooleanFieldType, true},
		{"Boolean false", false, models.BooleanFieldType, true},
		{"Empty slice", []string{}, models.ChecklistFieldType, false},
		{"Non-empty slice", []string{"item1"}, models.ChecklistFieldType, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateRequired(tt.value, tt.fieldType)

			if result.IsValid != tt.wantValid {
				t.Errorf("ValidateRequired() = %v, want %v", result.IsValid, tt.wantValid)
			}

			if !result.IsValid && result.Error == nil {
				t.Errorf("ValidateRequired() invalid but no error provided")
			}
		})
	}
}
