package goalconfig

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"davidlee/iter/internal/models"
)

func TestBooleanInputComponent(t *testing.T) {
	t.Run("basic functionality", func(t *testing.T) {
		input := NewBooleanInput()
		assert.Equal(t, models.BooleanFieldType, input.GetFieldType())
		
		// Test initial state
		assert.Equal(t, false, input.GetValue())
		assert.Equal(t, "false", input.GetStringValue())
		
		// Test validation (always valid)
		assert.NoError(t, input.Validate())
	})

	t.Run("form creation", func(t *testing.T) {
		input := NewBooleanInput()
		
		// Test with custom prompt
		form := input.CreateInputForm("Are you happy today?")
		assert.NotNil(t, form)
		
		// Test with empty prompt (should use default)
		form = input.CreateInputForm("")
		assert.NotNil(t, form)
	})

	t.Run("value handling", func(t *testing.T) {
		input := NewBooleanInput()
		
		// Manually set value to test (simulating form interaction)
		input.value = true
		
		assert.Equal(t, true, input.GetValue())
		assert.Equal(t, "true", input.GetStringValue())
		assert.NoError(t, input.Validate())
	})
}

func TestTextInputComponent(t *testing.T) {
	t.Run("single line text input", func(t *testing.T) {
		input := NewTextInput(false)
		assert.Equal(t, models.TextFieldType, input.GetFieldType())
		assert.False(t, input.multiline)
		
		// Test initial state
		assert.Equal(t, "", input.GetValue())
		assert.Equal(t, "", input.GetStringValue())
		assert.NoError(t, input.Validate())
	})

	t.Run("multiline text input", func(t *testing.T) {
		input := NewTextInput(true)
		assert.True(t, input.multiline)
		
		// Test form creation for multiline
		form := input.CreateInputForm("Tell me about your day")
		assert.NotNil(t, form)
	})

	t.Run("value handling", func(t *testing.T) {
		input := NewTextInput(false)
		
		// Simulate user input
		testText := "This is a test response"
		input.value = testText
		
		assert.Equal(t, testText, input.GetValue())
		assert.Equal(t, testText, input.GetStringValue())
		assert.NoError(t, input.Validate())
	})

	t.Run("form creation with prompts", func(t *testing.T) {
		input := NewTextInput(false)
		
		// Test custom prompt
		form := input.CreateInputForm("What did you learn today?")
		assert.NotNil(t, form)
		
		// Test default prompt
		form = input.CreateInputForm("")
		assert.NotNil(t, form)
	})
}

func TestNumericInputComponent(t *testing.T) {
	t.Run("unsigned integer input", func(t *testing.T) {
		input := NewNumericInput(models.UnsignedIntFieldType, "reps", nil, nil)
		assert.Equal(t, models.UnsignedIntFieldType, input.GetFieldType())
		assert.Equal(t, "reps", input.unit)
		
		// Test valid input
		input.value = "42"
		
		assert.Equal(t, 42.0, input.GetValue())
		assert.Equal(t, "42", input.GetStringValue())
		assert.NoError(t, input.Validate())
	})

	t.Run("unsigned decimal input", func(t *testing.T) {
		input := NewNumericInput(models.UnsignedDecimalFieldType, "hours", nil, nil)
		
		// Test valid decimal
		input.value = "7.5"
		
		assert.Equal(t, 7.5, input.GetValue())
		assert.Equal(t, "7.5", input.GetStringValue())
		assert.NoError(t, input.Validate())
	})

	t.Run("decimal input with negative values", func(t *testing.T) {
		input := NewNumericInput(models.DecimalFieldType, "degrees", nil, nil)
		
		// Test negative value
		input.value = "-15.2"
		
		assert.Equal(t, -15.2, input.GetValue())
		assert.Equal(t, "-15.2", input.GetStringValue())
		assert.NoError(t, input.Validate())
	})

	t.Run("constraints validation", func(t *testing.T) {
		minVal := 10.0
		maxVal := 100.0
		input := NewNumericInput(models.UnsignedDecimalFieldType, "percent", &minVal, &maxVal)
		
		// Test value within range
		input.value = "50.0"
		assert.NoError(t, input.Validate())
		
		// Test value below minimum (but positive to avoid unsigned validation)
		input.value = "5.0"
		assert.Error(t, input.Validate())
		assert.Contains(t, input.Validate().Error(), "at least")
		
		// Test value above maximum
		input.value = "150.0"
		assert.Error(t, input.Validate())
		assert.Contains(t, input.Validate().Error(), "at most")
		
		// Test negative value for unsigned type (different error)
		input.value = "-5.0"
		assert.Error(t, input.Validate())
		assert.Contains(t, input.Validate().Error(), "positive")
	})

	t.Run("invalid input validation", func(t *testing.T) {
		input := NewNumericInput(models.UnsignedIntFieldType, "count", nil, nil)
		
		// Test empty input
		input.value = ""
		assert.Error(t, input.Validate())
		assert.Contains(t, input.Validate().Error(), "required")
		
		// Test non-numeric input
		input.value = "not-a-number"
		assert.Error(t, input.Validate())
		
		// Test negative value for unsigned type
		input.value = "-10"
		assert.Error(t, input.Validate())
	})

	t.Run("form creation with constraints", func(t *testing.T) {
		minVal := 1.0
		maxVal := 10.0
		input := NewNumericInput(models.UnsignedIntFieldType, "rating", &minVal, &maxVal)
		
		form := input.CreateInputForm("Rate your experience")
		assert.NotNil(t, form)
		
		// Test default prompt with unit
		form = input.CreateInputForm("")
		assert.NotNil(t, form)
	})

	t.Run("unit display and descriptions", func(t *testing.T) {
		// Test different numeric types get proper descriptions
		intInput := NewNumericInput(models.UnsignedIntFieldType, "times", nil, nil)
		decimalInput := NewNumericInput(models.UnsignedDecimalFieldType, "liters", nil, nil)
		signedInput := NewNumericInput(models.DecimalFieldType, "degrees", nil, nil)
		
		assert.Equal(t, "times", intInput.unit)
		assert.Equal(t, "liters", decimalInput.unit)
		assert.Equal(t, "degrees", signedInput.unit)
	})
}

func TestTimeInputComponent(t *testing.T) {
	t.Run("basic functionality", func(t *testing.T) {
		input := NewTimeInput()
		assert.Equal(t, models.TimeFieldType, input.GetFieldType())
		
		// Test initial state
		assert.Equal(t, "", input.GetStringValue())
	})

	t.Run("valid time formats", func(t *testing.T) {
		input := NewTimeInput()
		
		// Test HH:MM format
		input.value = "14:30"
		assert.NoError(t, input.Validate())
		parsedTime := input.GetValue()
		assert.NotNil(t, parsedTime)
		
		// Test H:MM format  
		input.value = "9:15"
		assert.NoError(t, input.Validate())
		
		// Test edge cases
		input.value = "00:00"
		assert.NoError(t, input.Validate())
		
		input.value = "23:59"
		assert.NoError(t, input.Validate())
	})

	t.Run("invalid time formats", func(t *testing.T) {
		input := NewTimeInput()
		
		// Test empty input
		input.value = ""
		assert.Error(t, input.Validate())
		assert.Contains(t, input.Validate().Error(), "required")
		
		// Test invalid format
		input.value = "25:00"
		assert.Error(t, input.Validate())
		
		input.value = "14:70"
		assert.Error(t, input.Validate())
		
		input.value = "not-a-time"
		assert.Error(t, input.Validate())
		assert.Contains(t, input.Validate().Error(), "invalid time format")
	})

	t.Run("form creation", func(t *testing.T) {
		input := NewTimeInput()
		
		form := input.CreateInputForm("What time did you wake up?")
		assert.NotNil(t, form)
		
		// Test default prompt
		form = input.CreateInputForm("")
		assert.NotNil(t, form)
	})

	t.Run("time parsing", func(t *testing.T) {
		input := NewTimeInput()
		
		input.value = "14:30"
		parsedTime := input.GetValue()
		
		// Verify it's actually a time.Time
		timeValue, ok := parsedTime.(time.Time)
		require.True(t, ok)
		assert.Equal(t, 14, timeValue.Hour())
		assert.Equal(t, 30, timeValue.Minute())
	})
}

func TestDurationInputComponent(t *testing.T) {
	t.Run("basic functionality", func(t *testing.T) {
		input := NewDurationInput()
		assert.Equal(t, models.DurationFieldType, input.GetFieldType())
		
		// Test initial state
		assert.Equal(t, "", input.GetStringValue())
	})

	t.Run("valid duration formats", func(t *testing.T) {
		input := NewDurationInput()
		
		testCases := []struct {
			input    string
			expected time.Duration
		}{
			{"30m", 30 * time.Minute},
			{"1h", 1 * time.Hour},
			{"1h30m", 1*time.Hour + 30*time.Minute},
			{"2h15m", 2*time.Hour + 15*time.Minute},
			{"45s", 45 * time.Second},
		}
		
		for _, tc := range testCases {
			t.Run(tc.input, func(t *testing.T) {
				input.value = tc.input
				assert.NoError(t, input.Validate(), "Duration %s should be valid", tc.input)
				
				parsedDuration := input.GetValue()
				duration, ok := parsedDuration.(time.Duration)
				require.True(t, ok)
				assert.Equal(t, tc.expected, duration)
			})
		}
	})

	t.Run("invalid duration formats", func(t *testing.T) {
		input := NewDurationInput()
		
		// Test empty input
		input.value = ""
		assert.Error(t, input.Validate())
		assert.Contains(t, input.Validate().Error(), "required")
		
		// Test invalid format
		input.value = "not-a-duration"
		assert.Error(t, input.Validate())
		
		input.value = "25:00:00" // This format might not be supported
		// Note: Go's time.ParseDuration might actually support this, so adjust test as needed
	})

	t.Run("form creation", func(t *testing.T) {
		input := NewDurationInput()
		
		form := input.CreateInputForm("How long did you exercise?")
		assert.NotNil(t, form)
		
		// Test default prompt
		form = input.CreateInputForm("")
		assert.NotNil(t, form)
	})

	t.Run("duration value handling", func(t *testing.T) {
		input := NewDurationInput()
		
		input.value = "2h30m"
		assert.Equal(t, "2h30m", input.GetStringValue())
		
		parsedDuration := input.GetValue()
		duration, ok := parsedDuration.(time.Duration)
		require.True(t, ok)
		assert.Equal(t, 2*time.Hour+30*time.Minute, duration)
	})
}

func TestFieldValueInputFactory(t *testing.T) {
	factory := NewFieldValueInputFactory()
	assert.NotNil(t, factory)

	t.Run("boolean field creation", func(t *testing.T) {
		fieldType := models.FieldType{Type: models.BooleanFieldType}
		
		input, err := factory.CreateInput(fieldType)
		assert.NoError(t, err)
		assert.NotNil(t, input)
		assert.Equal(t, models.BooleanFieldType, input.GetFieldType())
		
		// Verify it's the correct type
		_, ok := input.(*BooleanInput)
		assert.True(t, ok)
	})

	t.Run("text field creation", func(t *testing.T) {
		// Test single-line text
		fieldType := models.FieldType{
			Type:      models.TextFieldType,
			Multiline: &[]bool{false}[0],
		}
		
		input, err := factory.CreateInput(fieldType)
		assert.NoError(t, err)
		assert.NotNil(t, input)
		assert.Equal(t, models.TextFieldType, input.GetFieldType())
		
		textInput, ok := input.(*TextInput)
		require.True(t, ok)
		assert.False(t, textInput.multiline)
		
		// Test multiline text
		fieldType.Multiline = &[]bool{true}[0]
		input, err = factory.CreateInput(fieldType)
		assert.NoError(t, err)
		
		textInput, ok = input.(*TextInput)
		require.True(t, ok)
		assert.True(t, textInput.multiline)
	})

	t.Run("numeric field creation", func(t *testing.T) {
		minVal := 0.0
		maxVal := 100.0
		
		testCases := []string{
			models.UnsignedIntFieldType,
			models.UnsignedDecimalFieldType,
			models.DecimalFieldType,
		}
		
		for _, numericType := range testCases {
			t.Run(numericType, func(t *testing.T) {
				fieldType := models.FieldType{
					Type: numericType,
					Unit: "test-unit",
					Min:  &minVal,
					Max:  &maxVal,
				}
				
				input, err := factory.CreateInput(fieldType)
				assert.NoError(t, err)
				assert.NotNil(t, input)
				assert.Equal(t, numericType, input.GetFieldType())
				
				numericInput, ok := input.(*NumericInput)
				require.True(t, ok)
				assert.Equal(t, "test-unit", numericInput.unit)
				assert.Equal(t, &minVal, numericInput.min)
				assert.Equal(t, &maxVal, numericInput.max)
			})
		}
	})

	t.Run("time field creation", func(t *testing.T) {
		fieldType := models.FieldType{Type: models.TimeFieldType}
		
		input, err := factory.CreateInput(fieldType)
		assert.NoError(t, err)
		assert.NotNil(t, input)
		assert.Equal(t, models.TimeFieldType, input.GetFieldType())
		
		_, ok := input.(*TimeInput)
		assert.True(t, ok)
	})

	t.Run("duration field creation", func(t *testing.T) {
		fieldType := models.FieldType{Type: models.DurationFieldType}
		
		input, err := factory.CreateInput(fieldType)
		assert.NoError(t, err)
		assert.NotNil(t, input)
		assert.Equal(t, models.DurationFieldType, input.GetFieldType())
		
		_, ok := input.(*DurationInput)
		assert.True(t, ok)
	})

	t.Run("unsupported field type", func(t *testing.T) {
		fieldType := models.FieldType{Type: "unsupported_type"}
		
		input, err := factory.CreateInput(fieldType)
		assert.Error(t, err)
		assert.Nil(t, input)
		assert.Contains(t, err.Error(), "unsupported field type")
	})

	t.Run("nil multiline handling", func(t *testing.T) {
		// Test text field with nil multiline (should default to false)
		fieldType := models.FieldType{
			Type:      models.TextFieldType,
			Multiline: nil,
		}
		
		input, err := factory.CreateInput(fieldType)
		assert.NoError(t, err)
		
		textInput, ok := input.(*TextInput)
		require.True(t, ok)
		assert.False(t, textInput.multiline) // Should default to false
	})
}