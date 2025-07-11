package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"davidlee/iter/internal/models"
)

// InformationalGoalHandler handles entry collection for informational goals.
// These goals collect data without scoring - they're for tracking information.
type InformationalGoalHandler struct{}

// NewInformationalGoalHandler creates a new informational goal handler.
func NewInformationalGoalHandler() *InformationalGoalHandler {
	return &InformationalGoalHandler{}
}

// CollectEntry collects an entry for an informational goal (data collection without scoring).
func (h *InformationalGoalHandler) CollectEntry(goal models.Goal, existing *ExistingEntry) (*EntryResult, error) {
	// Prepare the form title with goal information
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("14")). // Bright cyan for informational
		Margin(1, 0)

	_ = titleStyle.Render(fmt.Sprintf("📊 %s", goal.Title)) // Title styling available for future use

	// Prepare description if available
	var description string
	if goal.Description != "" {
		descStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")). // Gray
			Italic(true)
		description = descStyle.Render(goal.Description)
	}

	// Add informational note
	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("14")).
		Faint(true)
	infoNote := infoStyle.Render("ℹ️  This is an informational goal - for tracking data only")

	if description != "" {
		description += "\n" + infoNote
	} else {
		description = infoNote
	}
	_ = description // Description used in field type collection

	// Collect value based on field type (similar to elastic goals but without scoring)
	value, err := h.collectValueByFieldType(goal, existing)
	if err != nil {
		return nil, fmt.Errorf("failed to collect value: %w", err)
	}

	// Collect optional notes
	notes, err := h.collectOptionalNotes(goal, value, existing)
	if err != nil {
		return nil, fmt.Errorf("failed to collect notes: %w", err)
	}

	// Return result (no achievement level for informational goals)
	return &EntryResult{
		Value:            value,
		AchievementLevel: nil, // Informational goals don't have achievement levels
		Notes:            notes,
	}, nil
}

// collectValueByFieldType collects a value based on the goal's field type.
func (h *InformationalGoalHandler) collectValueByFieldType(goal models.Goal, existing *ExistingEntry) (interface{}, error) {
	switch goal.FieldType.Type {
	case models.BooleanFieldType:
		return h.collectBooleanValue(goal, existing)
	case models.UnsignedIntFieldType, models.UnsignedDecimalFieldType, models.DecimalFieldType:
		return h.collectNumericValue(goal, existing)
	case models.DurationFieldType:
		return h.collectDurationValue(goal, existing)
	case models.TimeFieldType:
		return h.collectTimeValue(goal, existing)
	case models.TextFieldType:
		return h.collectTextValue(goal, existing)
	default:
		return nil, fmt.Errorf("unsupported field type: %s", goal.FieldType.Type)
	}
}

// collectBooleanValue collects a boolean value.
func (h *InformationalGoalHandler) collectBooleanValue(goal models.Goal, existing *ExistingEntry) (bool, error) {
	var currentValue bool
	var hasExisting bool
	if existing != nil && existing.Value != nil {
		if boolVal, ok := existing.Value.(bool); ok {
			currentValue = boolVal
			hasExisting = true
		}
	}

	var value bool
	prompt := goal.Prompt
	if prompt == "" {
		prompt = fmt.Sprintf("Record yes/no for %s:", goal.Title)
	}

	if hasExisting {
		status := "No"
		if currentValue {
			status = "Yes"
		}
		prompt = fmt.Sprintf("%s (current: %s)", prompt, status)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(prompt).
				Value(&value).
				Affirmative("Yes").
				Negative("No"),
		),
	)

	if err := form.Run(); err != nil {
		return false, fmt.Errorf("boolean form failed: %w", err)
	}

	return value, nil
}

// collectNumericValue collects a numeric value.
func (h *InformationalGoalHandler) collectNumericValue(goal models.Goal, existing *ExistingEntry) (string, error) {
	var currentValue string
	if existing != nil && existing.Value != nil {
		currentValue = fmt.Sprintf("%v", existing.Value)
	}

	valueStr := currentValue
	prompt := goal.Prompt
	if prompt == "" {
		unit := goal.FieldType.Unit
		if unit != "" {
			prompt = fmt.Sprintf("Record value for %s (%s):", goal.Title, unit)
		} else {
			prompt = fmt.Sprintf("Record value for %s:", goal.Title)
		}
	}

	if currentValue != "" {
		prompt = fmt.Sprintf("%s (current: %s)", prompt, currentValue)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(prompt).
				Value(&valueStr).
				Placeholder("Enter value"),
		),
	)

	if err := form.Run(); err != nil {
		return "", fmt.Errorf("numeric form failed: %w", err)
	}

	return strings.TrimSpace(valueStr), nil
}

// collectDurationValue collects a duration value.
func (h *InformationalGoalHandler) collectDurationValue(goal models.Goal, existing *ExistingEntry) (string, error) {
	var currentValue string
	if existing != nil && existing.Value != nil {
		currentValue = fmt.Sprintf("%v", existing.Value)
	}

	valueStr := currentValue
	prompt := goal.Prompt
	if prompt == "" {
		prompt = fmt.Sprintf("Record duration for %s:", goal.Title)
	}

	if currentValue != "" {
		prompt = fmt.Sprintf("%s (current: %s)", prompt, currentValue)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(prompt).
				Description("Examples: 30 (minutes), 1h30m, 1:30:00").
				Value(&valueStr).
				Placeholder("30m"),
		),
	)

	if err := form.Run(); err != nil {
		return "", fmt.Errorf("duration form failed: %w", err)
	}

	return strings.TrimSpace(valueStr), nil
}

// collectTimeValue collects a time value.
func (h *InformationalGoalHandler) collectTimeValue(goal models.Goal, existing *ExistingEntry) (string, error) {
	var currentValue string
	if existing != nil && existing.Value != nil {
		currentValue = fmt.Sprintf("%v", existing.Value)
	}

	valueStr := currentValue
	prompt := goal.Prompt
	if prompt == "" {
		prompt = fmt.Sprintf("Record time for %s:", goal.Title)
	}

	if currentValue != "" {
		prompt = fmt.Sprintf("%s (current: %s)", prompt, currentValue)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(prompt).
				Description("Format: HH:MM (24-hour format)").
				Value(&valueStr).
				Placeholder("14:30"),
		),
	)

	if err := form.Run(); err != nil {
		return "", fmt.Errorf("time form failed: %w", err)
	}

	return strings.TrimSpace(valueStr), nil
}

// collectTextValue collects a text value.
func (h *InformationalGoalHandler) collectTextValue(goal models.Goal, existing *ExistingEntry) (string, error) {
	var currentValue string
	if existing != nil && existing.Value != nil {
		currentValue = fmt.Sprintf("%v", existing.Value)
	}

	valueStr := currentValue
	prompt := goal.Prompt
	if prompt == "" {
		prompt = fmt.Sprintf("Record text for %s:", goal.Title)
	}

	if currentValue != "" {
		prompt = fmt.Sprintf("%s (current: %s)", prompt, currentValue)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(prompt).
				Value(&valueStr).
				Placeholder("Enter your notes"),
		),
	)

	if err := form.Run(); err != nil {
		return "", fmt.Errorf("text form failed: %w", err)
	}

	return strings.TrimSpace(valueStr), nil
}

// collectOptionalNotes allows the user to optionally add notes.
func (h *InformationalGoalHandler) collectOptionalNotes(_ models.Goal, _ interface{}, existing *ExistingEntry) (string, error) {
	// Get existing notes
	var existingNotes string
	if existing != nil {
		existingNotes = existing.Notes
	}

	// Ask if user wants to add notes
	var wantNotes bool
	notesPrompt := "Add notes for this entry?"
	if existingNotes != "" {
		notesPrompt = fmt.Sprintf("Update notes? (current: %s)", existingNotes)
	}

	notesForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(notesPrompt).
				Value(&wantNotes).
				Affirmative("Yes").
				Negative("Skip"),
		),
	)

	if err := notesForm.Run(); err != nil {
		return "", fmt.Errorf("notes prompt failed: %w", err)
	}

	if !wantNotes {
		return existingNotes, nil
	}

	// Collect the notes
	var notes string
	if existingNotes != "" {
		notes = existingNotes
	}

	notesInputForm := huh.NewForm(
		huh.NewGroup(
			huh.NewText().
				Title("Notes:").
				Description("Optional notes about this entry (press Enter when done)").
				Value(&notes).
				Placeholder("Any additional observations or context?"),
		),
	)

	if err := notesInputForm.Run(); err != nil {
		return "", fmt.Errorf("notes input failed: %w", err)
	}

	return strings.TrimSpace(notes), nil
}