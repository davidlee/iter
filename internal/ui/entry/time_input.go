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

// TimeEntryInput handles time field value input for entry collection
type TimeEntryInput struct {
	value         string
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

	// Build description
	description := ti.buildDescription(goal)

	// Create the form
	ti.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(prompt).
				Description(description).
				Placeholder("14:30").
				Value(&ti.value).
				Validate(ti.validateTime),
		).Title(title),
	)

	// Add help text if available
	if goal.HelpText != "" {
		ti.form = ti.form.WithShowHelp(true)
	}

	return ti.form
}

// GetValue returns the time value as a parsed time
func (ti *TimeEntryInput) GetValue() interface{} {
	parsedTime, err := ti.parseTime()
	if err != nil {
		return nil
	}
	return parsedTime
}

// GetStringValue returns the time value as a string
func (ti *TimeEntryInput) GetStringValue() string {
	return ti.value
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

	// Add time format description
	descParts = append(descParts, "Enter time in HH:MM format (e.g., 14:30, 09:15)")

	return strings.Join(descParts, "\n")
}

func (ti *TimeEntryInput) validateTime(s string) error {
	if strings.TrimSpace(s) == "" {
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

	// Try parsing as HH:MM (24-hour format)
	parsedTime, err := time.Parse("15:04", trimmed)
	if err != nil {
		// Try parsing as H:MM
		parsedTime, err = time.Parse("3:04", trimmed)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid time format, use HH:MM (e.g., 14:30)")
		}
	}

	return parsedTime, nil
}
