// Package entry provides field-type-aware input collection for goal entry recording.
package entry

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"davidlee/vice/internal/models"
)

// AIDEV-NOTE: entry-boolean-input; implements EntryFieldInput for Boolean fields with three-option skip support
// T012/2.1-complete: Three-way selection (Yes/No/Skip) with EntryStatus integration for skip functionality

// BooleanOption represents the three possible boolean entry options
type BooleanOption string

// Boolean entry options for three-way selection
const (
	BooleanYes  BooleanOption = "yes"
	BooleanNo   BooleanOption = "no"
	BooleanSkip BooleanOption = "skip"
)

// BooleanEntryInput handles boolean field value input for entry collection
type BooleanEntryInput struct {
	option        BooleanOption
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
		option:        BooleanYes, // Default to Yes
	}

	// Set existing value if available
	if config.ExistingEntry != nil && config.ExistingEntry.Value != nil {
		if boolVal, ok := config.ExistingEntry.Value.(bool); ok {
			if boolVal {
				input.option = BooleanYes
			} else {
				input.option = BooleanNo
			}
		}
	}

	return input
}

// CreateInputForm creates a three-option select form (Yes/No/Skip)
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
		var status string
		switch bi.option {
		case BooleanYes:
			status = "✅ Yes"
		case BooleanNo:
			status = "❌ No"
		case BooleanSkip:
			status = "⏭️ Skip"
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

	// Create the form with three-option select
	// AIDEV-NOTE: T024-modal-fix; simplified to single select field for now
	bi.form = huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[BooleanOption]().
				Title(prompt).
				Description(description).
				Options(
					huh.NewOption("✅ Yes - Completed", BooleanYes),
					huh.NewOption("❌ No - Not completed", BooleanNo),
					huh.NewOption("⏭️ Skip - Unable to complete", BooleanSkip),
				).
				Value(&bi.option),
		).Title(title),
	)

	// Add help text if available
	if goal.HelpText != "" {
		bi.form = bi.form.WithShowHelp(true)
	}

	return bi.form
}

// GetValue returns the boolean value (nil for skip)
func (bi *BooleanEntryInput) GetValue() interface{} {
	switch bi.option {
	case BooleanYes:
		return true
	case BooleanNo:
		return false
	case BooleanSkip:
		return nil // Skip has no value
	default:
		return false
	}
}

// GetStringValue returns the option as a string
func (bi *BooleanEntryInput) GetStringValue() string {
	switch bi.option {
	case BooleanYes:
		return "yes"
	case BooleanNo:
		return "no"
	case BooleanSkip:
		return "skip"
	default:
		return "no"
	}
}

// GetStatus returns the EntryStatus based on the selected option
func (bi *BooleanEntryInput) GetStatus() models.EntryStatus {
	switch bi.option {
	case BooleanYes:
		return models.EntryCompleted
	case BooleanNo:
		return models.EntryFailed
	case BooleanSkip:
		return models.EntrySkipped
	default:
		return models.EntryFailed
	}
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
	if value == nil {
		bi.option = BooleanSkip
		return nil
	}
	if boolVal, ok := value.(bool); ok {
		if boolVal {
			bi.option = BooleanYes
		} else {
			bi.option = BooleanNo
		}
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
		switch bi.option {
		case BooleanYes:
			if *level == models.AchievementMini {
				feedback = "✅ Goal Completed!"
			} else {
				feedback = fmt.Sprintf("✅ Achievement: %v", *level)
			}
		case BooleanNo:
			feedback = "❌ Goal Not Completed"
		case BooleanSkip:
			feedback = "⏭️ Goal Skipped"
		}

		// Update form with achievement feedback
		// Implementation details will depend on final huh API patterns
		_ = achievementStyle.Render(feedback)
	}

	return nil
}
