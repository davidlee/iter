package entry

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"davidlee/iter/internal/models"
)

// AIDEV-NOTE: entry-checklist-input; implements EntryFieldInput for Checklist fields with progress tracking
// Provides multi-select checklist completion with progress feedback and scoring integration

// ChecklistEntryInput handles checklist field value input for entry collection
type ChecklistEntryInput struct {
	selectedItems  []string
	availableItems []string
	goal           models.Goal
	fieldType      models.FieldType
	existingEntry  *ExistingEntry
	showScoring    bool
	validationErr  error
	form           *huh.Form
}

// NewChecklistEntryInput creates a new checklist entry input component
func NewChecklistEntryInput(config EntryFieldInputConfig) *ChecklistEntryInput {
	input := &ChecklistEntryInput{
		goal:          config.Goal,
		fieldType:     config.FieldType,
		existingEntry: config.ExistingEntry,
		showScoring:   config.ShowScoring,
	}

	// Extract checklist items from field type configuration
	// ChecklistFieldType uses ChecklistID to reference external checklist definitions
	// For now, use placeholder items until checklist system integration
	if config.FieldType.Type == models.ChecklistFieldType {
		// TODO: Load actual checklist items from ChecklistID
		input.availableItems = []string{"Item 1", "Item 2", "Item 3"} // Placeholder
	}

	// Set existing selected items if available
	if config.ExistingEntry != nil && config.ExistingEntry.Value != nil {
		if selectedList, ok := config.ExistingEntry.Value.([]string); ok {
			input.selectedItems = selectedList
		}
	}

	return input
}

// CreateInputForm creates a checklist input form with multi-select interface
func (ci *ChecklistEntryInput) CreateInputForm(goal models.Goal) *huh.Form {
	// Prepare styling
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")). // Bright blue
		Margin(1, 0)

	title := titleStyle.Render(goal.Title)

	// Prepare prompt
	prompt := goal.Prompt
	if prompt == "" {
		prompt = fmt.Sprintf("Select completed items for: %s", goal.Title)
	}

	// Show current completion status if available
	if ci.existingEntry != nil && len(ci.selectedItems) > 0 {
		completed := len(ci.selectedItems)
		total := len(ci.availableItems)
		prompt = fmt.Sprintf("%s (currently: %d/%d completed)", prompt, completed, total)
	}

	// Build description
	description := ci.buildDescription(goal)

	// Create options for multi-select
	options := make([]huh.Option[string], len(ci.availableItems))
	for i, item := range ci.availableItems {
		options[i] = huh.NewOption(item, item)
	}

	// Create the form
	ci.form = huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title(prompt).
				Description(description).
				Options(options...).
				Value(&ci.selectedItems).
				Validate(ci.validateSelection),
		).Title(title),
	)

	// Add help text if available
	if goal.HelpText != "" {
		ci.form = ci.form.WithShowHelp(true)
	}

	return ci.form
}

// GetValue returns the selected checklist items
func (ci *ChecklistEntryInput) GetValue() interface{} {
	return ci.selectedItems
}

// GetStringValue returns the selected items as a comma-separated string
func (ci *ChecklistEntryInput) GetStringValue() string {
	return strings.Join(ci.selectedItems, ", ")
}

// Validate validates the checklist selection
func (ci *ChecklistEntryInput) Validate() error {
	ci.validationErr = ci.validateSelection(ci.selectedItems)
	return ci.validationErr
}

// GetFieldType returns the field type
func (ci *ChecklistEntryInput) GetFieldType() string {
	return models.ChecklistFieldType
}

// SetExistingValue sets an existing value for editing scenarios
func (ci *ChecklistEntryInput) SetExistingValue(value interface{}) error {
	if selectedList, ok := value.([]string); ok {
		ci.selectedItems = selectedList
		return nil
	}
	return fmt.Errorf("invalid checklist value type: %T", value)
}

// GetValidationError returns the current validation error state
func (ci *ChecklistEntryInput) GetValidationError() error {
	return ci.validationErr
}

// CanShowScoring returns true for checklist inputs with automatic scoring
func (ci *ChecklistEntryInput) CanShowScoring() bool {
	return ci.showScoring && ci.goal.ScoringType == models.AutomaticScoring
}

// UpdateScoringDisplay updates the form to show scoring feedback
func (ci *ChecklistEntryInput) UpdateScoringDisplay(level *models.AchievementLevel) error {
	if !ci.CanShowScoring() || ci.form == nil {
		return nil
	}

	// Add achievement feedback to the form display
	if level != nil {
		achievementStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")). // Bright green
			Bold(true)

		feedback := ""
		completed := len(ci.selectedItems)
		total := len(ci.availableItems)

		switch *level {
		case models.AchievementMini:
			feedback = fmt.Sprintf("🥉 Mini Checklist Achievement! (%d/%d completed)", completed, total)
		case models.AchievementMidi:
			feedback = fmt.Sprintf("🥈 Midi Checklist Achievement! (%d/%d completed)", completed, total)
		case models.AchievementMaxi:
			feedback = fmt.Sprintf("🥇 Maxi Checklist Achievement! (%d/%d completed)", completed, total)
		case models.AchievementNone:
			feedback = fmt.Sprintf("❌ Checklist Goal Not Met (%d/%d completed)", completed, total)
		default:
			feedback = fmt.Sprintf("Achievement: %v (%d/%d completed)", *level, completed, total)
		}

		// Update form with achievement feedback
		_ = achievementStyle.Render(feedback)
	}

	return nil
}

// GetCompletionProgress returns the current completion progress
func (ci *ChecklistEntryInput) GetCompletionProgress() (completed, total int) {
	return len(ci.selectedItems), len(ci.availableItems)
}

// Private methods

func (ci *ChecklistEntryInput) buildDescription(goal models.Goal) string {
	var descParts []string

	// Add goal description if available
	if goal.Description != "" {
		descStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")). // Gray
			Italic(true)
		descParts = append(descParts, descStyle.Render(goal.Description))
	}

	// Add progress information
	total := len(ci.availableItems)
	if total > 0 {
		descParts = append(descParts, fmt.Sprintf("Select completed items (%d items available)", total))
	} else {
		descParts = append(descParts, "Select completed items")
	}

	return strings.Join(descParts, "\n")
}

func (ci *ChecklistEntryInput) validateSelection(selected []string) error {
	// Basic validation - ensure all selected items are valid options
	for _, item := range selected {
		found := false
		for _, available := range ci.availableItems {
			if item == available {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("invalid checklist item: %s", item)
		}
	}

	// Additional validation could be added here:
	// - Minimum required items
	// - Maximum allowed items
	// - Specific item combinations

	return nil
}
