package entry

import (
	"testing"
	"time"

	"davidlee/iter/internal/models"
)

// AIDEV-NOTE: time-duration-tests; comprehensive testing for time and duration input components
// Tests parsing, validation, formatting, and edge cases for T010/2.2 implementation

func TestTimeEntryInput(t *testing.T) {
	config := EntryFieldInputConfig{
		Goal: models.Goal{
			Title:  "Wake Up Time",
			Prompt: "What time did you wake up?",
		},
		FieldType: models.FieldType{
			Type: models.TimeFieldType,
		},
		ExistingEntry: nil,
		ShowScoring:   false,
	}

	input := NewTimeEntryInput(config)

	// Test field type
	if input.GetFieldType() != models.TimeFieldType {
		t.Errorf("GetFieldType() = %v, want %v", input.GetFieldType(), models.TimeFieldType)
	}

	// Test setting valid time values
	testCases := []struct {
		name        string
		timeValue   string
		expectValid bool
	}{
		{"24-hour format", "14:30", true},
		{"morning time", "06:15", true},
		{"midnight", "00:00", true},
		{"noon", "12:00", true},
		{"single digit hour", "9:30", true},
		{"invalid format", "25:30", false},
		{"invalid minutes", "14:65", false},
		{"no colon", "1430", false},
		{"empty string", "", false},
		{"text input", "afternoon", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if err := input.SetExistingValue(tc.timeValue); err != nil {
				t.Errorf("SetExistingValue(%q) unexpected error: %v", tc.timeValue, err)
				return
			}

			err := input.Validate()
			if tc.expectValid && err != nil {
				t.Errorf("Validate(%q) expected valid, got error: %v", tc.timeValue, err)
			} else if !tc.expectValid && err == nil {
				t.Errorf("Validate(%q) expected error, got nil", tc.timeValue)
			}

			if tc.expectValid {
				value := input.GetValue()
				if value == nil {
					t.Errorf("GetValue() returned nil for valid time %q", tc.timeValue)
				}

				stringValue := input.GetStringValue()
				if stringValue != tc.timeValue {
					t.Errorf("GetStringValue() = %q, want %q", stringValue, tc.timeValue)
				}
			}
		})
	}
}

func TestTimeEntryInputExistingTime(t *testing.T) {
	// Test with time.Time value
	now := time.Now()
	expectedStr := now.Format("15:04")

	config := EntryFieldInputConfig{
		Goal: models.Goal{
			Title: "Test Time Goal",
		},
		FieldType: models.FieldType{
			Type: models.TimeFieldType,
		},
		ExistingEntry: &ExistingEntry{
			Value: now,
		},
		ShowScoring: false,
	}

	input := NewTimeEntryInput(config)

	if input.GetStringValue() != expectedStr {
		t.Errorf("GetStringValue() = %q, want %q", input.GetStringValue(), expectedStr)
	}
}

func TestTimeEntryInputScoringAwareness(t *testing.T) {
	config := EntryFieldInputConfig{
		Goal: models.Goal{
			Title:       "Morning Routine",
			ScoringType: models.AutomaticScoring,
		},
		FieldType: models.FieldType{
			Type: models.TimeFieldType,
		},
		ShowScoring: true,
	}

	input := NewTimeEntryInput(config)

	if !input.CanShowScoring() {
		t.Errorf("CanShowScoring() expected true for automatic scoring goal")
	}

	// Test scoring display update
	level := models.AchievementMidi
	if err := input.UpdateScoringDisplay(&level); err != nil {
		t.Errorf("UpdateScoringDisplay() unexpected error: %v", err)
	}
}

func TestDurationEntryInput(t *testing.T) {
	config := EntryFieldInputConfig{
		Goal: models.Goal{
			Title:  "Exercise Duration",
			Prompt: "How long did you exercise?",
		},
		FieldType: models.FieldType{
			Type: models.DurationFieldType,
		},
		ExistingEntry: nil,
		ShowScoring:   false,
	}

	input := NewDurationEntryInput(config)

	// Test field type
	if input.GetFieldType() != models.DurationFieldType {
		t.Errorf("GetFieldType() = %v, want %v", input.GetFieldType(), models.DurationFieldType)
	}

	// Test setting valid duration values
	testCases := []struct {
		name          string
		durationValue string
		expectValid   bool
	}{
		{"minutes only", "30m", true},
		{"hours and minutes", "1h30m", true},
		{"hours only", "2h", true},
		{"seconds", "45s", true},
		{"complex duration", "2h15m30s", true},
		{"decimal hours", "1.5h", true},
		{"large numbers", "90m", true},
		{"zero duration", "0s", true},
		{"invalid format", "30 minutes", false},
		{"empty string", "", false},
		{"no unit", "30", false},
		{"invalid unit", "30x", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if err := input.SetExistingValue(tc.durationValue); err != nil {
				t.Errorf("SetExistingValue(%q) unexpected error: %v", tc.durationValue, err)
				return
			}

			err := input.Validate()
			if tc.expectValid && err != nil {
				t.Errorf("Validate(%q) expected valid, got error: %v", tc.durationValue, err)
			} else if !tc.expectValid && err == nil {
				t.Errorf("Validate(%q) expected error, got nil", tc.durationValue)
			}

			if tc.expectValid {
				value := input.GetValue()
				if value == nil {
					t.Errorf("GetValue() returned nil for valid duration %q", tc.durationValue)
				}

				stringValue := input.GetStringValue()
				if stringValue != tc.durationValue {
					t.Errorf("GetStringValue() = %q, want %q", stringValue, tc.durationValue)
				}
			}
		})
	}
}

func TestDurationEntryInputExistingDuration(t *testing.T) {
	// Test with time.Duration value
	duration := 90 * time.Minute
	expectedStr := duration.String()

	config := EntryFieldInputConfig{
		Goal: models.Goal{
			Title: "Test Duration Goal",
		},
		FieldType: models.FieldType{
			Type: models.DurationFieldType,
		},
		ExistingEntry: &ExistingEntry{
			Value: duration,
		},
		ShowScoring: false,
	}

	input := NewDurationEntryInput(config)

	if input.GetStringValue() != expectedStr {
		t.Errorf("GetStringValue() = %q, want %q", input.GetStringValue(), expectedStr)
	}
}

func TestDurationEntryInputScoringAwareness(t *testing.T) {
	config := EntryFieldInputConfig{
		Goal: models.Goal{
			Title:       "Workout Duration",
			ScoringType: models.AutomaticScoring,
		},
		FieldType: models.FieldType{
			Type: models.DurationFieldType,
		},
		ShowScoring: true,
	}

	input := NewDurationEntryInput(config)

	if !input.CanShowScoring() {
		t.Errorf("CanShowScoring() expected true for automatic scoring goal")
	}

	// Test scoring display update
	level := models.AchievementMaxi
	if err := input.UpdateScoringDisplay(&level); err != nil {
		t.Errorf("UpdateScoringDisplay() unexpected error: %v", err)
	}
}

func TestTimeValidationPatterns(t *testing.T) {
	validator := NewFieldValidator()

	testCases := []struct {
		name      string
		value     interface{}
		wantValid bool
	}{
		{"valid time string", "14:30", true},
		{"valid time object", time.Now(), true},
		{"empty time string", "", false},
		{"nil time", nil, false},
		{"invalid time format", "25:30", true}, // Validation happens at parse time, not required validation
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := validator.ValidateRequired(tc.value, models.TimeFieldType)
			if result.IsValid != tc.wantValid {
				t.Errorf("ValidateRequired() = %v, want %v", result.IsValid, tc.wantValid)
			}
		})
	}
}

func TestDurationValidationPatterns(t *testing.T) {
	validator := NewFieldValidator()

	testCases := []struct {
		name      string
		value     interface{}
		wantValid bool
	}{
		{"valid duration string", "30m", true},
		{"valid duration object", 30 * time.Minute, true},
		{"empty duration string", "", false},
		{"nil duration", nil, false},
		{"zero duration", time.Duration(0), true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := validator.ValidateRequired(tc.value, models.DurationFieldType)
			if result.IsValid != tc.wantValid {
				t.Errorf("ValidateRequired() = %v, want %v", result.IsValid, tc.wantValid)
			}
		})
	}
}

func TestValidationHelperFunctions(t *testing.T) {
	testCases := []struct {
		name     string
		timeStr  string
		expected bool
	}{
		{"valid HH:MM", "14:30", true},
		{"valid H:MM", "9:30", true},
		{"empty string", "", false},
		{"no colon", "1430", false},
		{"invalid format", "25:30", true}, // Basic format check passes, validation happens elsewhere
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsValidTimeFormat(tc.timeStr)
			if result != tc.expected {
				t.Errorf("IsValidTimeFormat(%q) = %v, want %v", tc.timeStr, result, tc.expected)
			}
		})
	}

	durationCases := []struct {
		name        string
		durationStr string
		expected    bool
	}{
		{"valid duration with h", "2h", true},
		{"valid duration with m", "30m", true},
		{"valid duration with s", "45s", true},
		{"complex duration", "1h30m", true},
		{"empty string", "", false},
		{"no unit", "30", false},
	}

	for _, tc := range durationCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsValidDurationFormat(tc.durationStr)
			if result != tc.expected {
				t.Errorf("IsValidDurationFormat(%q) = %v, want %v", tc.durationStr, result, tc.expected)
			}
		})
	}
}

func TestFactoryTimeAndDurationCreation(t *testing.T) {
	factory := NewEntryFieldInputFactory()

	// Test Time field creation
	timeConfig := EntryFieldInputConfig{
		Goal: models.Goal{
			Title: "Test Time Goal",
		},
		FieldType: models.FieldType{
			Type: models.TimeFieldType,
		},
	}

	timeInput, err := factory.CreateInput(timeConfig)
	if err != nil {
		t.Errorf("CreateInput(time) unexpected error: %v", err)
	}

	if timeInput.GetFieldType() != models.TimeFieldType {
		t.Errorf("Time input GetFieldType() = %v, want %v", timeInput.GetFieldType(), models.TimeFieldType)
	}

	// Test Duration field creation
	durationConfig := EntryFieldInputConfig{
		Goal: models.Goal{
			Title: "Test Duration Goal",
		},
		FieldType: models.FieldType{
			Type: models.DurationFieldType,
		},
	}

	durationInput, err := factory.CreateInput(durationConfig)
	if err != nil {
		t.Errorf("CreateInput(duration) unexpected error: %v", err)
	}

	if durationInput.GetFieldType() != models.DurationFieldType {
		t.Errorf("Duration input GetFieldType() = %v, want %v", durationInput.GetFieldType(), models.DurationFieldType)
	}
}
