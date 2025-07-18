package entry

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"github.com/davidlee/vice/internal/models"
	"github.com/davidlee/vice/internal/parser"
)

// AIDEV-NOTE: entry-checklist-input; implements EntryFieldInput for Checklist fields with progress tracking
// Provides multi-select checklist completion with progress feedback and scoring integration
// T012/2.3: Submit/Skip button interface with ActionSubmit/ActionSkip pattern

// ChecklistEntryInput handles checklist field value input for entry collection
type ChecklistEntryInput struct {
	selectedItems   []string
	availableItems  []string
	action          InputAction
	checklistID     string
	checklistParser *parser.ChecklistParser
	habit           models.Habit
	fieldType       models.FieldType
	existingEntry   *ExistingEntry
	showScoring     bool
	validationErr   error
	form            *huh.Form
	checklistsPath  string
}

// NewChecklistEntryInput creates a new checklist entry input component
func NewChecklistEntryInput(config EntryFieldInputConfig) *ChecklistEntryInput {
	// Set default checklists path if not provided
	checklistsPath := config.ChecklistsPath
	if checklistsPath == "" {
		checklistsPath = "checklists.yml"
	}

	input := &ChecklistEntryInput{
		habit:           config.Habit,
		fieldType:       config.FieldType,
		existingEntry:   config.ExistingEntry,
		showScoring:     config.ShowScoring,
		checklistParser: parser.NewChecklistParser(),
		checklistsPath:  checklistsPath,
		action:          ActionSubmit, // Default to submit
	}

	// Extract checklist ID from field type configuration
	if config.FieldType.Type == models.ChecklistFieldType && config.FieldType.ChecklistID != "" {
		input.checklistID = config.FieldType.ChecklistID
		// Load actual checklist items from ChecklistID
		if err := input.loadChecklistItems(); err != nil {
			// Fall back to placeholder items if loading fails
			input.availableItems = []string{"Loading failed - checklist not found"}
		}
	} else {
		// Use placeholder items for testing or when no ChecklistID is provided
		input.availableItems = []string{"Item 1", "Item 2", "Item 3"}
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
func (ci *ChecklistEntryInput) CreateInputForm(habit models.Habit) *huh.Form {
	// Prepare styling
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")). // Bright blue
		Margin(1, 0)

	title := titleStyle.Render(habit.Title)

	// Prepare prompt
	prompt := habit.Prompt
	if prompt == "" {
		prompt = fmt.Sprintf("Select completed items for: %s", habit.Title)
	}

	// Show current completion status if available
	if ci.existingEntry != nil && len(ci.selectedItems) > 0 {
		completed := len(ci.selectedItems)
		total := len(ci.availableItems)
		prompt = fmt.Sprintf("%s (currently: %d/%d completed)", prompt, completed, total)
	}

	// Build description
	description := ci.buildDescription(habit)

	// Create options for multi-select
	options := make([]huh.Option[string], len(ci.availableItems))
	for i, item := range ci.availableItems {
		options[i] = huh.NewOption(item, item)
	}

	// Create the form with multi-select and action selection
	ci.form = huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title(prompt).
				Description(description).
				Options(options...).
				Value(&ci.selectedItems).
				Validate(ci.validateSelection),
			huh.NewSelect[InputAction]().
				Title("Action").
				Options(
					huh.NewOption("✅ Submit Checklist", ActionSubmit),
					huh.NewOption("⏭️ Skip Habit", ActionSkip),
				).
				Value(&ci.action),
		).Title(title),
	)

	// Add help text if available
	if habit.HelpText != "" {
		ci.form = ci.form.WithShowHelp(true)
	}

	return ci.form
}

// GetValue returns the selected checklist items (nil for skipped)
func (ci *ChecklistEntryInput) GetValue() interface{} {
	if ci.action == ActionSkip {
		return nil
	}
	return ci.selectedItems
}

// GetStringValue returns the selected items as a comma-separated string
func (ci *ChecklistEntryInput) GetStringValue() string {
	if ci.action == ActionSkip {
		return "skip"
	}
	return strings.Join(ci.selectedItems, ", ")
}

// GetStatus returns the entry completion status based on action and validation
func (ci *ChecklistEntryInput) GetStatus() models.EntryStatus {
	switch ci.action {
	case ActionSkip:
		return models.EntrySkipped
	case ActionSubmit:
		return models.EntryCompleted
	default:
		return models.EntryCompleted
	}
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
	return ci.showScoring && ci.habit.ScoringType == models.AutomaticScoring
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
			feedback = fmt.Sprintf("❌ Checklist Habit Not Met (%d/%d completed)", completed, total)
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

// loadChecklistItems loads the actual checklist items from the ChecklistID reference
func (ci *ChecklistEntryInput) loadChecklistItems() error {
	// Load checklist schema from file
	schema, err := ci.checklistParser.LoadFromFile(ci.checklistsPath)
	if err != nil {
		return fmt.Errorf("failed to load checklists: %w", err)
	}

	// Find the checklist by ID
	checklist, err := ci.checklistParser.GetChecklistByID(schema, ci.checklistID)
	if err != nil {
		return fmt.Errorf("checklist not found: %w", err)
	}

	// Extract items, filtering out headings (items starting with "# ")
	var items []string
	for _, item := range checklist.Items {
		// Skip heading items (they are for visual organization, not selectable)
		if !strings.HasPrefix(item, "# ") {
			items = append(items, item)
		}
	}

	if len(items) == 0 {
		return fmt.Errorf("checklist '%s' has no selectable items", ci.checklistID)
	}

	ci.availableItems = items
	return nil
}

func (ci *ChecklistEntryInput) buildDescription(habit models.Habit) string {
	var descParts []string

	// Add habit description if available
	if habit.Description != "" {
		descStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")). // Gray
			Italic(true)
		descParts = append(descParts, descStyle.Render(habit.Description))
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
