package entry

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"davidlee/iter/internal/models"
)

// AIDEV-NOTE: entry-text-input; implements EntryFieldInput for Text fields with multiline support
// Provides single-line and multiline text input with validation and optional comment support
// T012/2.2: Submit/Skip button interface with hybrid shortcut support ("s" key)

// TextEntryInput handles text field value input for entry collection
type TextEntryInput struct {
	value         string
	action        InputAction
	goal          models.Goal
	fieldType     models.FieldType
	existingEntry *ExistingEntry
	showScoring   bool
	validationErr error
	form          *huh.Form
	multiline     bool
}

// NewTextEntryInput creates a new text entry input component
func NewTextEntryInput(config EntryFieldInputConfig) *TextEntryInput {
	input := &TextEntryInput{
		goal:          config.Goal,
		fieldType:     config.FieldType,
		existingEntry: config.ExistingEntry,
		showScoring:   config.ShowScoring,
		multiline:     config.FieldType.Multiline != nil && *config.FieldType.Multiline,
		action:        ActionSubmit, // Default to submit
	}

	// Set existing value if available
	if config.ExistingEntry != nil && config.ExistingEntry.Value != nil {
		if textVal, ok := config.ExistingEntry.Value.(string); ok {
			input.value = textVal
		}
	}

	return input
}

// CreateInputForm creates a text input form with validation
func (ti *TextEntryInput) CreateInputForm(goal models.Goal) *huh.Form {
	// Prepare styling
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")). // Bright blue
		Margin(1, 0)

	title := titleStyle.Render(goal.Title)

	// Prepare prompt
	prompt := goal.Prompt
	if prompt == "" {
		if ti.multiline {
			prompt = fmt.Sprintf("Enter text for: %s", goal.Title)
		} else {
			prompt = fmt.Sprintf("Enter value for: %s", goal.Title)
		}
	}

	// Show existing value in prompt if available
	if ti.existingEntry != nil && ti.existingEntry.Value != nil && ti.value != "" {
		truncated := ti.value
		if len(truncated) > 30 {
			truncated = truncated[:27] + "..."
		}
		prompt = fmt.Sprintf("%s (current: %s)", prompt, truncated)
	}

	// Prepare description
	var description string
	if goal.Description != "" {
		descStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")). // Gray
			Italic(true)
		description = descStyle.Render(goal.Description)
	}

	// Add field type specific description
	if ti.multiline {
		if description != "" {
			description += "\n"
		}
		description += "Enter multiple lines of text (press Enter for new lines)"
	} else {
		if description != "" {
			description += "\n"
		}
		description += "Enter text value"
	}

	// Create the appropriate form based on multiline setting with action selection
	var inputField huh.Field
	if ti.multiline {
		inputField = huh.NewText().
			Title(prompt + " (or press 's' to skip)").
			Description(description).
			Value(&ti.value).
			Validate(ti.validateInput)
	} else {
		inputField = huh.NewInput().
			Title(prompt + " (or press 's' to skip)").
			Description(description).
			Value(&ti.value).
			Validate(ti.validateInput)
	}

	ti.form = huh.NewForm(
		huh.NewGroup(
			inputField,
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

// GetValue returns the text value (nil for skipped)
func (ti *TextEntryInput) GetValue() interface{} {
	if ti.action == ActionSkip {
		return nil
	}
	return ti.value
}

// GetStringValue returns the text value
func (ti *TextEntryInput) GetStringValue() string {
	if ti.action == ActionSkip {
		return "skip"
	}
	return ti.value
}

// GetStatus returns the entry completion status based on action and validation
func (ti *TextEntryInput) GetStatus() models.EntryStatus {
	switch ti.action {
	case ActionSkip:
		return models.EntrySkipped
	case ActionSubmit:
		// Text inputs are generally always successful if not skipped
		return models.EntryCompleted
	default:
		return models.EntryCompleted
	}
}

// Validate validates the text value
func (ti *TextEntryInput) Validate() error {
	ti.validationErr = ti.validateInput(ti.value)
	return ti.validationErr
}

// GetFieldType returns the field type
func (ti *TextEntryInput) GetFieldType() string {
	return models.TextFieldType
}

// SetExistingValue sets an existing value for editing scenarios
func (ti *TextEntryInput) SetExistingValue(value interface{}) error {
	if textVal, ok := value.(string); ok {
		ti.value = textVal
		return nil
	}
	return fmt.Errorf("invalid text value type: %T", value)
}

// GetValidationError returns the current validation error state
func (ti *TextEntryInput) GetValidationError() error {
	return ti.validationErr
}

// CanShowScoring returns false for text inputs (manual scoring only)
func (ti *TextEntryInput) CanShowScoring() bool {
	// Text fields are restricted to manual scoring only per T009 design decisions
	return false
}

// UpdateScoringDisplay is a no-op for text inputs (manual scoring only)
func (ti *TextEntryInput) UpdateScoringDisplay(_ *models.AchievementLevel) error {
	// Text fields don't support automatic scoring
	return nil
}

// Private validation method
func (ti *TextEntryInput) validateInput(s string) error {
	trimmed := strings.TrimSpace(s)

	// Fast-path shortcut detection for skip
	if trimmed == "s" || trimmed == "S" {
		ti.action = ActionSkip
		ti.value = ""
		return nil // Allow form completion with skip action
	}

	// Text validation is generally permissive
	// Could add length constraints if needed in the future

	// Check if field is required (basic validation)
	if trimmed == "" {
		// For now, allow empty text values
		// This could be enhanced with required field configuration
		return nil
	}

	return nil
}
