package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"davidlee/vice/internal/models"
)

// SimpleGoalHandler handles entry collection for simple boolean goals.
// This maintains the exact same UI behavior as the original collectGoalEntry method.
type SimpleGoalHandler struct{}

// NewSimpleGoalHandler creates a new simple goal handler.
func NewSimpleGoalHandler() *SimpleGoalHandler {
	return &SimpleGoalHandler{}
}

// CollectEntry collects a boolean entry for a simple goal.
// This preserves the exact UI flow from the original implementation.
func (h *SimpleGoalHandler) CollectEntry(goal models.Goal, existing *ExistingEntry) (*EntryResult, error) {
	// Prepare the form title with goal information
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")). // Bright blue
		Margin(1, 0)

	title := titleStyle.Render(goal.Title)

	// Prepare description if available
	var description string
	if goal.Description != "" {
		descStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")). // Gray
			Italic(true)
		description = descStyle.Render(goal.Description)
	}

	// Prepare help text if available
	var help string
	if goal.HelpText != "" {
		helpStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")). // Bright green
			Faint(true)
		help = helpStyle.Render("💡 " + goal.HelpText)
	}

	// Get current boolean value (if any)
	var currentValue bool
	var hasExisting bool
	if existing != nil && existing.Value != nil {
		if boolVal, ok := existing.Value.(bool); ok {
			currentValue = boolVal
			hasExisting = true
		}
	}

	// Create the completion question - initialize with existing value
	completed := currentValue
	prompt := goal.Prompt
	if prompt == "" {
		prompt = fmt.Sprintf("Did you complete: %s?", goal.Title)
	}

	// Show existing value in prompt if available
	if hasExisting {
		status := "❌ No"
		if currentValue {
			status = "✅ Yes"
		}
		prompt = fmt.Sprintf("%s (currently: %s)", prompt, status)
	}

	// Create the form
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(prompt).
				Description(description).
				Value(&completed).
				Affirmative("Yes").
				Negative("No"),
		).Title(title),
	)

	// Add help text as a note if available
	if help != "" {
		form = form.WithShowHelp(true)
	}

	// Run the form
	if err := form.Run(); err != nil {
		return nil, fmt.Errorf("form execution failed: %w", err)
	}

	// Collect optional notes
	notes, err := h.collectOptionalNotes(goal, completed, existing)
	if err != nil {
		return nil, fmt.Errorf("failed to collect notes: %w", err)
	}

	// Return the result (no achievement level for simple goals)
	return &EntryResult{
		Value:            completed,
		AchievementLevel: nil, // Simple goals don't have achievement levels
		Notes:            notes,
	}, nil
}

// collectOptionalNotes allows the user to optionally add notes for a simple goal.
// This preserves the exact behavior from the original implementation.
func (h *SimpleGoalHandler) collectOptionalNotes(_ models.Goal, _ bool, existing *ExistingEntry) (string, error) {
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
		return existingNotes, nil // Return existing notes unchanged
	}

	// Collect the notes
	var notes string
	if existingNotes != "" {
		notes = existingNotes // Pre-populate with existing notes
	}

	notesInputForm := huh.NewForm(
		huh.NewGroup(
			huh.NewText().
				Title("Notes:").
				Description("Optional notes about this goal (press Enter when done)").
				Value(&notes).
				Placeholder("Why did you succeed/fail? How did you feel?"),
		),
	)

	if err := notesInputForm.Run(); err != nil {
		return "", fmt.Errorf("notes input failed: %w", err)
	}

	// Return the notes (trimmed)
	return strings.TrimSpace(notes), nil
}
