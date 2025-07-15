package entry

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"davidlee/vice/internal/debug"
	"davidlee/vice/internal/models"
)

// AIDEV-NOTE: entry-numeric-input; implements EntryFieldInput for Numeric fields with scoring feedback
// Supports all numeric types (UnsignedInt, UnsignedDecimal, Decimal) with unit display and validation
// T012/2.2: Submit/Skip button interface with hybrid shortcut support ("s" key)

// InputAction represents the user's choice to submit or skip an input field
type InputAction string

// Input action options for Submit/Skip pattern
const (
	ActionSubmit InputAction = "submit"
	ActionSkip   InputAction = "skip"
)

// NumericEntryInput handles numeric field value input for entry collection
type NumericEntryInput struct {
	value         string
	action        InputAction
	goal          models.Goal
	fieldType     models.FieldType
	existingEntry *ExistingEntry
	showScoring   bool
	validationErr error
	form          *huh.Form
}

// NewNumericEntryInput creates a new numeric entry input component
func NewNumericEntryInput(config EntryFieldInputConfig) *NumericEntryInput {
	input := &NumericEntryInput{
		goal:          config.Goal,
		fieldType:     config.FieldType,
		existingEntry: config.ExistingEntry,
		showScoring:   config.ShowScoring,
		action:        ActionSubmit, // Default to submit
	}

	// Set existing value if available
	if config.ExistingEntry != nil && config.ExistingEntry.Value != nil {
		// Convert numeric value to string for editing
		input.value = fmt.Sprintf("%v", config.ExistingEntry.Value)
	}

	return input
}

// CreateInputForm creates a numeric input form with unit display and validation
func (ni *NumericEntryInput) CreateInputForm(goal models.Goal) *huh.Form {
	debug.Field("Creating huh.Form for numeric goal %s, current value: %v", goal.ID, ni.value)
	// Prepare styling
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")). // Bright blue
		Margin(1, 0)

	title := titleStyle.Render(goal.Title)

	// Prepare prompt with unit information
	prompt := goal.Prompt
	if prompt == "" {
		if ni.fieldType.Unit != "" && ni.fieldType.Unit != "times" {
			prompt = fmt.Sprintf("Enter value for %s (in %s)", goal.Title, ni.fieldType.Unit)
		} else {
			prompt = fmt.Sprintf("Enter numeric value for: %s", goal.Title)
		}
	}

	// Show existing value in prompt if available
	if ni.existingEntry != nil && ni.existingEntry.Value != nil && ni.value != "" {
		unitDisplay := ""
		if ni.fieldType.Unit != "" && ni.fieldType.Unit != "times" {
			unitDisplay = " " + ni.fieldType.Unit
		}
		prompt = fmt.Sprintf("%s (current: %s%s)", prompt, ni.value, unitDisplay)
	}

	// Build comprehensive description
	description := ni.buildDescription(goal)

	// Create the form with input and action selection
	ni.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Key("numeric_value").
				Title(prompt+" (or press 's' to skip)").
				Description(description).
				Value(&ni.value).
				Validate(ni.validateInput),
			huh.NewSelect[InputAction]().
				Key("action").
				Title("Action").
				Options(
					huh.NewOption("✅ Submit Value", ActionSubmit),
					huh.NewOption("⏭️ Skip Goal", ActionSkip),
				).
				Value(&ni.action),
		).Title(title),
	)

	// Add help text if available
	if goal.HelpText != "" {
		ni.form = ni.form.WithShowHelp(true)
	}

	return ni.form
}

// GetValue returns the numeric value as the appropriate type (nil for skipped)
func (ni *NumericEntryInput) GetValue() interface{} {
	if ni.action == ActionSkip {
		return nil
	}
	val, err := ni.parseValue()
	if err != nil {
		return nil
	}
	return val
}

// GetStringValue returns the numeric value as a string
func (ni *NumericEntryInput) GetStringValue() string {
	if ni.action == ActionSkip {
		return "skip"
	}
	return ni.value
}

// GetStatus returns the entry completion status based on action and validation
func (ni *NumericEntryInput) GetStatus() models.EntryStatus {
	switch ni.action {
	case ActionSkip:
		return models.EntrySkipped
	case ActionSubmit:
		if ni.GetValidationError() != nil {
			return models.EntryFailed
		}
		return models.EntryCompleted
	default:
		return models.EntryCompleted
	}
}

// Validate validates the numeric value
func (ni *NumericEntryInput) Validate() error {
	ni.validationErr = ni.validateInput(ni.value)
	return ni.validationErr
}

// GetFieldType returns the field type
func (ni *NumericEntryInput) GetFieldType() string {
	return ni.fieldType.Type
}

// SetExistingValue sets an existing value for editing scenarios
func (ni *NumericEntryInput) SetExistingValue(value interface{}) error {
	ni.value = fmt.Sprintf("%v", value)
	return nil
}

// GetValidationError returns the current validation error state
func (ni *NumericEntryInput) GetValidationError() error {
	return ni.validationErr
}

// CanShowScoring returns true for numeric inputs with automatic scoring
func (ni *NumericEntryInput) CanShowScoring() bool {
	return ni.showScoring && ni.goal.ScoringType == models.AutomaticScoring
}

// UpdateScoringDisplay updates the form to show scoring feedback
func (ni *NumericEntryInput) UpdateScoringDisplay(level *models.AchievementLevel) error {
	if !ni.CanShowScoring() || ni.form == nil {
		return nil
	}

	// Add achievement feedback to the form display
	if level != nil {
		achievementStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")). // Bright green
			Bold(true)

		feedback := ""
		switch *level {
		case models.AchievementMini:
			feedback = "🥉 Mini Achievement!"
		case models.AchievementMidi:
			feedback = "🥈 Midi Achievement!"
		case models.AchievementMaxi:
			feedback = "🥇 Maxi Achievement!"
		case models.AchievementNone:
			feedback = "❌ Goal Not Met"
		default:
			feedback = fmt.Sprintf("Achievement: %v", *level)
		}

		// Update form with achievement feedback
		_ = achievementStyle.Render(feedback)
	}

	return nil
}

// Private methods

func (ni *NumericEntryInput) buildDescription(goal models.Goal) string {
	var descParts []string

	// Add goal description if available
	if goal.Description != "" {
		descStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")). // Gray
			Italic(true)
		descParts = append(descParts, descStyle.Render(goal.Description))
	}

	// Add numeric type description
	typeDesc := ni.getNumericTypeDescription()
	if ni.fieldType.Unit != "" && ni.fieldType.Unit != "times" {
		typeDesc += fmt.Sprintf(" (in %s)", ni.fieldType.Unit)
	}
	descParts = append(descParts, typeDesc)

	// Add constraints description
	if constraints := ni.getConstraintsDescription(); constraints != "" {
		descParts = append(descParts, constraints)
	}

	return strings.Join(descParts, "\n")
}

func (ni *NumericEntryInput) validateInput(s string) error {
	trimmed := strings.TrimSpace(s)

	// Fast-path shortcut detection for skip
	if trimmed == "s" || trimmed == "S" {
		ni.action = ActionSkip
		ni.value = ""
		return nil // Allow form completion with skip action
	}

	if trimmed == "" {
		return fmt.Errorf("numeric value is required")
	}

	val, err := ni.parseValue()
	if err != nil {
		return err
	}

	floatVal := val.(float64)

	if ni.fieldType.Min != nil && floatVal < *ni.fieldType.Min {
		return fmt.Errorf("value must be at least %g", *ni.fieldType.Min)
	}

	if ni.fieldType.Max != nil && floatVal > *ni.fieldType.Max {
		return fmt.Errorf("value must be at most %g", *ni.fieldType.Max)
	}

	return nil
}

func (ni *NumericEntryInput) parseValue() (interface{}, error) {
	trimmed := strings.TrimSpace(ni.value)

	switch ni.fieldType.Type {
	case models.UnsignedIntFieldType:
		val, err := strconv.ParseUint(trimmed, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid unsigned integer: %w", err)
		}
		return float64(val), nil

	case models.UnsignedDecimalFieldType:
		val, err := strconv.ParseFloat(trimmed, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid unsigned decimal: %w", err)
		}
		if val < 0 {
			return nil, fmt.Errorf("value must be positive")
		}
		return val, nil

	case models.DecimalFieldType:
		val, err := strconv.ParseFloat(trimmed, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid decimal: %w", err)
		}
		return val, nil

	default:
		return nil, fmt.Errorf("unknown numeric type: %s", ni.fieldType.Type)
	}
}

func (ni *NumericEntryInput) getNumericTypeDescription() string {
	switch ni.fieldType.Type {
	case models.UnsignedIntFieldType:
		return "Enter whole number (0, 1, 2, 3...)"
	case models.UnsignedDecimalFieldType:
		return "Enter positive decimal (0.5, 1.25, 2.7...)"
	case models.DecimalFieldType:
		return "Enter decimal number (including negative)"
	default:
		return "Enter numeric value"
	}
}

func (ni *NumericEntryInput) getConstraintsDescription() string {
	switch {
	case ni.fieldType.Min != nil && ni.fieldType.Max != nil:
		return fmt.Sprintf("Value must be between %g and %g", *ni.fieldType.Min, *ni.fieldType.Max)
	case ni.fieldType.Min != nil:
		return fmt.Sprintf("Value must be at least %g", *ni.fieldType.Min)
	case ni.fieldType.Max != nil:
		return fmt.Sprintf("Value must be at most %g", *ni.fieldType.Max)
	default:
		return ""
	}
}
