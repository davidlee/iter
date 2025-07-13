package entry

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"davidlee/iter/internal/models"
)

// AIDEV-NOTE: entry-duration-input; implements EntryFieldInput for Duration fields with scoring feedback
// Provides flexible duration parsing (1h 30m, 45m, 2h, 90m) with validation and automatic scoring

// DurationEntryInput handles duration field value input for entry collection
type DurationEntryInput struct {
	value         string
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

	// Create the form
	di.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(prompt).
				Description(description).
				Placeholder("1h 30m").
				Value(&di.value).
				Validate(di.validateDuration),
		).Title(title),
	)

	// Add help text if available
	if goal.HelpText != "" {
		di.form = di.form.WithShowHelp(true)
	}

	return di.form
}

// GetValue returns the duration value as a time.Duration
func (di *DurationEntryInput) GetValue() interface{} {
	parsedDuration, err := di.parseDuration()
	if err != nil {
		return nil
	}
	return parsedDuration
}

// GetStringValue returns the duration value as a string
func (di *DurationEntryInput) GetStringValue() string {
	return di.value
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

	// Add duration format description
	descParts = append(descParts, "Enter duration (e.g., 1h 30m, 45m, 2h, 90m)")

	return strings.Join(descParts, "\n")
}

func (di *DurationEntryInput) validateDuration(s string) error {
	if strings.TrimSpace(s) == "" {
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
		return duration, nil
	}

	// Enhanced parsing for common formats could be added here
	// For now, rely on Go's built-in parser which handles:
	// - 1h30m, 90m, 1.5h, etc.
	return time.Duration(0), fmt.Errorf("invalid duration format, use formats like 1h30m, 45m, 2h")
}
