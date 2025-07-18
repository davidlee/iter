package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"github.com/davidlee/vice/internal/models"
)

// SimpleHabitHandler handles entry collection for simple boolean habits.
// This maintains the exact same UI behavior as the original collectHabitEntry method.
type SimpleHabitHandler struct{}

// NewSimpleHabitHandler creates a new simple habit handler.
func NewSimpleHabitHandler() *SimpleHabitHandler {
	return &SimpleHabitHandler{}
}

// CollectEntry collects a boolean entry for a simple habit.
// This preserves the exact UI flow from the original implementation.
func (h *SimpleHabitHandler) CollectEntry(habit models.Habit, existing *ExistingEntry) (*EntryResult, error) {
	// Prepare the form title with habit information
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")). // Bright blue
		Margin(1, 0)

	title := titleStyle.Render(habit.Title)

	// Prepare description if available
	var description string
	if habit.Description != "" {
		descStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")). // Gray
			Italic(true)
		description = descStyle.Render(habit.Description)
	}

	// Prepare help text if available
	var help string
	if habit.HelpText != "" {
		helpStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")). // Bright green
			Faint(true)
		help = helpStyle.Render("💡 " + habit.HelpText)
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
	prompt := habit.Prompt
	if prompt == "" {
		prompt = fmt.Sprintf("Did you complete: %s?", habit.Title)
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
	notes, err := h.collectOptionalNotes(habit, completed, existing)
	if err != nil {
		return nil, fmt.Errorf("failed to collect notes: %w", err)
	}

	// Return the result (no achievement level for simple habits)
	return &EntryResult{
		Value:            completed,
		AchievementLevel: nil, // Simple habits don't have achievement levels
		Notes:            notes,
	}, nil
}

// collectOptionalNotes allows the user to optionally add notes for a simple habit.
// This preserves the exact behavior from the original implementation.
func (h *SimpleHabitHandler) collectOptionalNotes(_ models.Habit, _ bool, existing *ExistingEntry) (string, error) {
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
				Description("Optional notes about this habit (press Enter when done)").
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
