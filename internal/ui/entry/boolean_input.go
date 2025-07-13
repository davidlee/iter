package entry

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"davidlee/iter/internal/models"
)

// AIDEV-NOTE: entry-boolean-input; implements EntryFieldInput for Boolean fields with scoring feedback
// Provides yes/no confirmation with clear completion indication and optional achievement display

// BooleanEntryInput handles boolean field value input for entry collection
type BooleanEntryInput struct {
	value         bool
	goal          models.Goal
	fieldType     models.FieldType
	existingEntry *ExistingEntry
	showScoring   bool
	validationErr error
	form          *huh.Form
}

// NewBooleanEntryInput creates a new boolean entry input component
func NewBooleanEntryInput(config EntryFieldInputConfig) *BooleanEntryInput {
	input := &BooleanEntryInput{
		goal:          config.Goal,
		fieldType:     config.FieldType,
		existingEntry: config.ExistingEntry,
		showScoring:   config.ShowScoring,
	}

	// Set existing value if available
	if config.ExistingEntry != nil && config.ExistingEntry.Value != nil {
		if boolVal, ok := config.ExistingEntry.Value.(bool); ok {
			input.value = boolVal
		}
	}

	return input
}

// CreateInputForm creates a boolean input form with clear yes/no display
func (bi *BooleanEntryInput) CreateInputForm(goal models.Goal) *huh.Form {
	// Prepare styling
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")). // Bright blue
		Margin(1, 0)

	title := titleStyle.Render(goal.Title)

	// Prepare prompt
	prompt := goal.Prompt
	if prompt == "" {
		prompt = fmt.Sprintf("Did you complete: %s?", goal.Title)
	}

	// Show existing value in prompt if available
	if bi.existingEntry != nil && bi.existingEntry.Value != nil {
		status := "❌ No"
		if bi.value {
			status = "✅ Yes"
		}
		prompt = fmt.Sprintf("%s (currently: %s)", prompt, status)
	}

	// Prepare description
	var description string
	if goal.Description != "" {
		descStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")). // Gray
			Italic(true)
		description = descStyle.Render(goal.Description)
	}

	// Create the form
	bi.form = huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(prompt).
				Description(description).
				Value(&bi.value).
				Affirmative("Yes").
				Negative("No"),
		).Title(title),
	)

	// Add help text if available
	if goal.HelpText != "" {
		bi.form = bi.form.WithShowHelp(true)
	}

	return bi.form
}

// GetValue returns the boolean value
func (bi *BooleanEntryInput) GetValue() interface{} {
	return bi.value
}

// GetStringValue returns the boolean as a string
func (bi *BooleanEntryInput) GetStringValue() string {
	if bi.value {
		return "true"
	}
	return "false"
}

// Validate validates the boolean value (always valid)
func (bi *BooleanEntryInput) Validate() error {
	bi.validationErr = nil // Boolean values are always valid
	return nil
}

// GetFieldType returns the field type
func (bi *BooleanEntryInput) GetFieldType() string {
	return models.BooleanFieldType
}

// SetExistingValue sets an existing value for editing scenarios
func (bi *BooleanEntryInput) SetExistingValue(value interface{}) error {
	if boolVal, ok := value.(bool); ok {
		bi.value = boolVal
		return nil
	}
	return fmt.Errorf("invalid boolean value type: %T", value)
}

// GetValidationError returns the current validation error state
func (bi *BooleanEntryInput) GetValidationError() error {
	return bi.validationErr
}

// CanShowScoring returns true for boolean inputs with automatic scoring
func (bi *BooleanEntryInput) CanShowScoring() bool {
	return bi.showScoring && bi.goal.ScoringType == models.AutomaticScoring
}

// UpdateScoringDisplay updates the form to show scoring feedback
func (bi *BooleanEntryInput) UpdateScoringDisplay(level *models.AchievementLevel) error {
	if !bi.CanShowScoring() || bi.form == nil {
		return nil
	}

	// Add achievement feedback to the form display
	// This will be enhanced when scoring integration is implemented
	if level != nil {
		achievementStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")). // Bright green
			Bold(true)

		feedback := ""
		switch *level {
		case models.AchievementMini:
			feedback = "✅ Goal Completed!"
		case models.AchievementNone:
			feedback = "❌ Goal Not Completed"
		default:
			feedback = fmt.Sprintf("Achievement: %v", *level)
		}

		// Update form with achievement feedback
		// Implementation details will depend on final huh API patterns
		_ = achievementStyle.Render(feedback)
	}

	return nil
}
