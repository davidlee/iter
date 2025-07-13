package entry

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"davidlee/iter/internal/models"
)

// AIDEV-NOTE: entry-time-input; implements EntryFieldInput for Time fields with scoring feedback
// Provides HH:MM time input with validation and automatic scoring integration
// T012/2.2: Submit/Skip button interface with hybrid shortcut support ("s" key)

// TimeEntryInput handles time field value input for entry collection
type TimeEntryInput struct {
	value         string
	action        InputAction
	goal          models.Goal
	fieldType     models.FieldType
	existingEntry *ExistingEntry
	showScoring   bool
	validationErr error
	form          *huh.Form
}

// NewTimeEntryInput creates a new time entry input component
func NewTimeEntryInput(config EntryFieldInputConfig) *TimeEntryInput {
	input := &TimeEntryInput{
		goal:          config.Goal,
		fieldType:     config.FieldType,
		existingEntry: config.ExistingEntry,
		showScoring:   config.ShowScoring,
		action:        ActionSubmit, // Default to submit
	}

	// Set existing value if available
	if config.ExistingEntry != nil && config.ExistingEntry.Value != nil {
		// Convert time value to string for editing
		if timeVal, ok := config.ExistingEntry.Value.(time.Time); ok {
			input.value = timeVal.Format("15:04")
		} else if strVal, ok := config.ExistingEntry.Value.(string); ok {
			input.value = strVal
		}
	}

	return input
}

// CreateInputForm creates a time input form with formatted input
func (ti *TimeEntryInput) CreateInputForm(goal models.Goal) *huh.Form {
	// Prepare styling
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")). // Bright blue
		Margin(1, 0)

	title := titleStyle.Render(goal.Title)

	// Prepare prompt
	prompt := goal.Prompt
	if prompt == "" {
		prompt = fmt.Sprintf("Enter time for: %s", goal.Title)
	}

	// Show existing value in prompt if available
	if ti.existingEntry != nil && ti.existingEntry.Value != nil && ti.value != "" {
		prompt = fmt.Sprintf("%s (current: %s)", prompt, ti.value)
	}

	// Build description with time format examples
	description := ti.buildDescription(goal)

	// Create the form with input and action selection
	ti.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(prompt + " (or press 's' to skip)").
				Description(description).
				Placeholder("14:30").
				Value(&ti.value).
				Validate(ti.validateTime),
			huh.NewSelect[InputAction]().
				Title("Action").
				Options(
					huh.NewOption("✅ Submit Value", ActionSubmit),
					huh.NewOption("⏭️ Skip Goal", ActionSkip),
				).
				Value(&ti.action),
		).Title(title),
	)

	// Add help text if available
	if goal.HelpText != "" {
		ti.form = ti.form.WithShowHelp(true)
	}

	return ti.form
}

// GetValue returns the time value as a parsed time (nil for skipped)
func (ti *TimeEntryInput) GetValue() interface{} {
	if ti.action == ActionSkip {
		return nil
	}
	parsedTime, err := ti.parseTime()
	if err != nil {
		return nil
	}
	return parsedTime
}

// GetStringValue returns the time value as a string
func (ti *TimeEntryInput) GetStringValue() string {
	if ti.action == ActionSkip {
		return "skip"
	}
	return ti.value
}

// GetStatus returns the entry completion status based on action and validation
func (ti *TimeEntryInput) GetStatus() models.EntryStatus {
	switch ti.action {
	case ActionSkip:
		return models.EntrySkipped
	case ActionSubmit:
		if ti.GetValidationError() != nil {
			return models.EntryFailed
		}
		return models.EntryCompleted
	default:
		return models.EntryCompleted
	}
}

// Validate validates the time value
func (ti *TimeEntryInput) Validate() error {
	ti.validationErr = ti.validateTime(ti.value)
	return ti.validationErr
}

// GetFieldType returns the field type
func (ti *TimeEntryInput) GetFieldType() string {
	return models.TimeFieldType
}

// SetExistingValue sets an existing value for editing scenarios
func (ti *TimeEntryInput) SetExistingValue(value interface{}) error {
	if timeVal, ok := value.(time.Time); ok {
		ti.value = timeVal.Format("15:04")
		return nil
	}
	if strVal, ok := value.(string); ok {
		ti.value = strVal
		return nil
	}
	return fmt.Errorf("invalid time value type: %T", value)
}

// GetValidationError returns the current validation error state
func (ti *TimeEntryInput) GetValidationError() error {
	return ti.validationErr
}

// CanShowScoring returns true for time inputs with automatic scoring
func (ti *TimeEntryInput) CanShowScoring() bool {
	return ti.showScoring && ti.goal.ScoringType == models.AutomaticScoring
}

// UpdateScoringDisplay updates the form to show scoring feedback
func (ti *TimeEntryInput) UpdateScoringDisplay(level *models.AchievementLevel) error {
	if !ti.CanShowScoring() || ti.form == nil {
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
			feedback = "🥉 Mini Time Achievement!"
		case models.AchievementMidi:
			feedback = "🥈 Midi Time Achievement!"
		case models.AchievementMaxi:
			feedback = "🥇 Maxi Time Achievement!"
		case models.AchievementNone:
			feedback = "❌ Time Goal Not Met"
		default:
			feedback = fmt.Sprintf("Achievement: %v", *level)
		}

		// Update form with achievement feedback
		_ = achievementStyle.Render(feedback)
	}

	return nil
}

// Private methods

func (ti *TimeEntryInput) buildDescription(goal models.Goal) string {
	var descParts []string

	// Add goal description if available
	if goal.Description != "" {
		descStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")). // Gray
			Italic(true)
		descParts = append(descParts, descStyle.Render(goal.Description))
	}

	// Add time format description with examples
	formatDesc := "Enter time in HH:MM format"

	// Add format-specific guidance
	if ti.fieldType.Format != "" {
		formatDesc += fmt.Sprintf(" (%s)", ti.fieldType.Format)
	} else {
		formatDesc += " (e.g., 14:30, 09:15, 6:00)"
	}

	descParts = append(descParts, formatDesc)

	return strings.Join(descParts, "\n")
}

func (ti *TimeEntryInput) validateTime(s string) error {
	trimmed := strings.TrimSpace(s)
	
	// Fast-path shortcut detection for skip
	if trimmed == "s" || trimmed == "S" {
		ti.action = ActionSkip
		ti.value = ""
		return nil // Allow form completion with skip action
	}

	if trimmed == "" {
		return fmt.Errorf("time value is required")
	}

	_, err := ti.parseTime()
	if err != nil {
		return err
	}

	return nil
}

func (ti *TimeEntryInput) parseTime() (time.Time, error) {
	trimmed := strings.TrimSpace(ti.value)

	// Validate basic format before parsing
	if !strings.Contains(trimmed, ":") {
		return time.Time{}, fmt.Errorf("invalid time format: missing colon, use HH:MM (e.g., 14:30)")
	}

	parts := strings.Split(trimmed, ":")
	if len(parts) != 2 {
		return time.Time{}, fmt.Errorf("invalid time format: use HH:MM (e.g., 14:30)")
	}

	// Try parsing as HH:MM (24-hour format)
	parsedTime, err := time.Parse("15:04", trimmed)
	if err != nil {
		// Try parsing as H:MM (single digit hour)
		parsedTime, err = time.Parse("3:04", trimmed)
		if err != nil {
			// Provide specific error messages for common mistakes
			if len(parts) == 2 {
				hour, minute := parts[0], parts[1]
				if len(hour) == 0 || len(minute) != 2 {
					return time.Time{}, fmt.Errorf("invalid time format: hour and minute must be numbers (e.g., 14:30)")
				}
				return time.Time{}, fmt.Errorf("invalid time: hours must be 0-23, minutes must be 0-59")
			}
			return time.Time{}, fmt.Errorf("invalid time format, use HH:MM (e.g., 14:30)")
		}
	}

	return parsedTime, nil
}
