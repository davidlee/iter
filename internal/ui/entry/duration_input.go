package entry

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"davidlee/vice/internal/models"
)

// AIDEV-NOTE: entry-duration-input; implements EntryFieldInput for Duration fields with scoring feedback
// Provides flexible duration parsing (1h 30m, 45m, 2h, 90m) with validation and automatic scoring
// T012/2.2: Submit/Skip button interface with hybrid shortcut support ("s" key)

// DurationEntryInput handles duration field value input for entry collection
type DurationEntryInput struct {
	value         string
	action        InputAction
	goal          models.Goal
	fieldType     models.FieldType
	existingEntry *ExistingEntry
	showScoring   bool
	validationErr error
	form          *huh.Form
}

// NewDurationEntryInput creates a new duration entry input component
func NewDurationEntryInput(config EntryFieldInputConfig) *DurationEntryInput {
	input := &DurationEntryInput{
		goal:          config.Goal,
		fieldType:     config.FieldType,
		existingEntry: config.ExistingEntry,
		showScoring:   config.ShowScoring,
		action:        ActionSubmit, // Default to submit
	}

	// Set existing value if available
	if config.ExistingEntry != nil && config.ExistingEntry.Value != nil {
		// Convert duration value to string for editing
		if durVal, ok := config.ExistingEntry.Value.(time.Duration); ok {
			input.value = durVal.String()
		} else if strVal, ok := config.ExistingEntry.Value.(string); ok {
			input.value = strVal
		}
	}

	return input
}

// CreateInputForm creates a duration input form with flexible format support
func (di *DurationEntryInput) CreateInputForm(goal models.Goal) *huh.Form {
	// Prepare styling
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")). // Bright blue
		Margin(1, 0)

	title := titleStyle.Render(goal.Title)

	// Prepare prompt
	prompt := goal.Prompt
	if prompt == "" {
		prompt = fmt.Sprintf("Enter duration for: %s", goal.Title)
	}

	// Show existing value in prompt if available
	if di.existingEntry != nil && di.existingEntry.Value != nil && di.value != "" {
		prompt = fmt.Sprintf("%s (current: %s)", prompt, di.value)
	}

	// Build description
	description := di.buildDescription(goal)

	// Create the form with input and action selection
	di.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(prompt+" (or press 's' to skip)").
				Description(description).
				Placeholder("1h 30m").
				Value(&di.value).
				Validate(di.validateDuration),
			huh.NewSelect[InputAction]().
				Title("Action").
				Options(
					huh.NewOption("✅ Submit Value", ActionSubmit),
					huh.NewOption("⏭️ Skip Goal", ActionSkip),
				).
				Value(&di.action),
		).Title(title),
	)

	// Add help text if available
	if goal.HelpText != "" {
		di.form = di.form.WithShowHelp(true)
	}

	return di.form
}

// GetValue returns the duration value as a time.Duration (nil for skipped)
func (di *DurationEntryInput) GetValue() interface{} {
	if di.action == ActionSkip {
		return nil
	}
	parsedDuration, err := di.parseDuration()
	if err != nil {
		return nil
	}
	return parsedDuration
}

// GetStringValue returns the duration value as a string
func (di *DurationEntryInput) GetStringValue() string {
	if di.action == ActionSkip {
		return "skip"
	}
	return di.value
}

// GetStatus returns the entry completion status based on action and validation
func (di *DurationEntryInput) GetStatus() models.EntryStatus {
	switch di.action {
	case ActionSkip:
		return models.EntrySkipped
	case ActionSubmit:
		if di.GetValidationError() != nil {
			return models.EntryFailed
		}
		return models.EntryCompleted
	default:
		return models.EntryCompleted
	}
}

// Validate validates the duration value
func (di *DurationEntryInput) Validate() error {
	di.validationErr = di.validateDuration(di.value)
	return di.validationErr
}

// GetFieldType returns the field type
func (di *DurationEntryInput) GetFieldType() string {
	return models.DurationFieldType
}

// SetExistingValue sets an existing value for editing scenarios
func (di *DurationEntryInput) SetExistingValue(value interface{}) error {
	if durVal, ok := value.(time.Duration); ok {
		di.value = durVal.String()
		return nil
	}
	if strVal, ok := value.(string); ok {
		di.value = strVal
		return nil
	}
	return fmt.Errorf("invalid duration value type: %T", value)
}

// GetValidationError returns the current validation error state
func (di *DurationEntryInput) GetValidationError() error {
	return di.validationErr
}

// CanShowScoring returns true for duration inputs with automatic scoring
func (di *DurationEntryInput) CanShowScoring() bool {
	return di.showScoring && di.goal.ScoringType == models.AutomaticScoring
}

// UpdateScoringDisplay updates the form to show scoring feedback
func (di *DurationEntryInput) UpdateScoringDisplay(level *models.AchievementLevel) error {
	if !di.CanShowScoring() || di.form == nil {
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
			feedback = "🥉 Mini Duration Achievement!"
		case models.AchievementMidi:
			feedback = "🥈 Midi Duration Achievement!"
		case models.AchievementMaxi:
			feedback = "🥇 Maxi Duration Achievement!"
		case models.AchievementNone:
			feedback = "❌ Duration Goal Not Met"
		default:
			feedback = fmt.Sprintf("Achievement: %v", *level)
		}

		// Update form with achievement feedback
		_ = achievementStyle.Render(feedback)
	}

	return nil
}

// Private methods

func (di *DurationEntryInput) buildDescription(goal models.Goal) string {
	var descParts []string

	// Add goal description if available
	if goal.Description != "" {
		descStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")). // Gray
			Italic(true)
		descParts = append(descParts, descStyle.Render(goal.Description))
	}

	// Add duration format description with comprehensive examples
	formatDesc := "Enter duration"

	// Add format-specific guidance if available
	if di.fieldType.Format != "" {
		formatDesc += fmt.Sprintf(" (%s)", di.fieldType.Format)
	} else {
		formatDesc += " (e.g., 1h30m, 45m, 2h, 90m, 30s)"
	}

	descParts = append(descParts, formatDesc)

	// Add additional helpful format examples
	descParts = append(descParts, "Supported units: h (hours), m (minutes), s (seconds)")

	return strings.Join(descParts, "\n")
}

func (di *DurationEntryInput) validateDuration(s string) error {
	trimmed := strings.TrimSpace(s)

	// Fast-path shortcut detection for skip
	if trimmed == "s" || trimmed == "S" {
		di.action = ActionSkip
		di.value = ""
		return nil // Allow form completion with skip action
	}

	if trimmed == "" {
		return fmt.Errorf("duration value is required")
	}

	_, err := di.parseDuration()
	if err != nil {
		return err
	}

	return nil
}

func (di *DurationEntryInput) parseDuration() (time.Duration, error) {
	trimmed := strings.TrimSpace(di.value)

	// Try parsing as Go duration format first
	duration, err := time.ParseDuration(trimmed)
	if err == nil {
		if duration < 0 {
			return time.Duration(0), fmt.Errorf("duration cannot be negative")
		}
		return duration, nil
	}

	// Provide helpful error messages for common mistakes
	originalErr := err.Error()
	switch {
	case strings.Contains(trimmed, " "):
		return time.Duration(0), fmt.Errorf("invalid duration format: remove spaces, use formats like 1h30m, 45m, 2h")
	case !strings.ContainsAny(trimmed, "hms"):
		return time.Duration(0), fmt.Errorf("invalid duration format: missing unit, use h (hours), m (minutes), s (seconds)")
	case strings.Contains(originalErr, "unknown unit"):
		return time.Duration(0), fmt.Errorf("invalid duration unit: use h (hours), m (minutes), s (seconds)")
	case strings.Contains(originalErr, "invalid syntax"):
		return time.Duration(0), fmt.Errorf("invalid duration syntax: use formats like 1h30m, 45m, 2h")
	default:
		return time.Duration(0), fmt.Errorf("invalid duration format: %s. Use formats like 1h30m, 45m, 2h", originalErr)
	}
}
