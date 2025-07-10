// Package ui provides interactive user interface components for the iter application.
package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"davidlee/iter/internal/models"
	"davidlee/iter/internal/parser"
	"davidlee/iter/internal/storage"
)

// EntryCollector handles the interactive collection of today's habit entries.
type EntryCollector struct {
	goalParser   *parser.GoalParser
	entryStorage *storage.EntryStorage
	goals        []models.Goal
	entries      map[string]bool
	notes        map[string]string
}

// NewEntryCollector creates a new entry collector instance.
func NewEntryCollector() *EntryCollector {
	return &EntryCollector{
		goalParser:   parser.NewGoalParser(),
		entryStorage: storage.NewEntryStorage(),
		entries:      make(map[string]bool),
		notes:        make(map[string]string),
	}
}

// CollectTodayEntries runs the interactive UI to collect today's habit entries.
func (ec *EntryCollector) CollectTodayEntries(goalsFile, entriesFile string) error {
	// Load goal schema
	schema, err := ec.goalParser.LoadFromFile(goalsFile)
	if err != nil {
		return fmt.Errorf("failed to load goals: %w", err)
	}

	// Get simple boolean goals for MVP
	ec.goals = parser.GetSimpleBooleanGoals(schema)
	if len(ec.goals) == 0 {
		return fmt.Errorf("no simple boolean goals found in %s", goalsFile)
	}

	// Load existing entries for today (if any)
	if err := ec.loadExistingEntries(entriesFile); err != nil {
		return fmt.Errorf("failed to load existing entries: %w", err)
	}

	// Display welcome message
	ec.displayWelcome()

	// Collect entries for each goal
	for _, goal := range ec.goals {
		if err := ec.collectGoalEntry(goal); err != nil {
			return fmt.Errorf("failed to collect entry for goal %s: %w", goal.ID, err)
		}
	}

	// Save entries
	if err := ec.saveEntries(entriesFile); err != nil {
		return fmt.Errorf("failed to save entries: %w", err)
	}

	// Display completion message
	ec.displayCompletion()

	return nil
}

// loadExistingEntries loads any existing entries for today.
func (ec *EntryCollector) loadExistingEntries(entriesFile string) error {
	today := time.Now().Format("2006-01-02")

	dayEntry, err := ec.entryStorage.GetDayEntry(entriesFile, today)
	if err != nil {
		// No existing entries for today, which is fine
		return nil //nolint:nilerr // No entries for today is expected
	}

	// Load existing entries into our maps
	for _, goalEntry := range dayEntry.Goals {
		if boolVal, ok := goalEntry.GetBooleanValue(); ok {
			ec.entries[goalEntry.GoalID] = boolVal
			ec.notes[goalEntry.GoalID] = goalEntry.Notes
		}
	}

	return nil
}

// collectGoalEntry collects the entry for a single goal using interactive UI.
func (ec *EntryCollector) collectGoalEntry(goal models.Goal) error {
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

	// Get current value (if any)
	currentValue, hasExisting := ec.entries[goal.ID]

	// Create the completion question
	var completed bool
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
		return fmt.Errorf("form execution failed: %w", err)
	}

	// Store the result
	ec.entries[goal.ID] = completed

	// Optionally collect notes if the user wants to add them
	if err := ec.collectOptionalNotes(goal, completed); err != nil {
		return fmt.Errorf("failed to collect notes: %w", err)
	}

	return nil
}

// collectOptionalNotes allows the user to optionally add notes for a goal.
func (ec *EntryCollector) collectOptionalNotes(goal models.Goal, _ bool) error {
	// Get existing notes
	existingNotes := ec.notes[goal.ID]

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
		return fmt.Errorf("notes prompt failed: %w", err)
	}

	if !wantNotes {
		return nil
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
		return fmt.Errorf("notes input failed: %w", err)
	}

	// Store the notes (even if empty, to clear existing notes)
	ec.notes[goal.ID] = strings.TrimSpace(notes)

	return nil
}

// saveEntries saves all collected entries to the entries file.
func (ec *EntryCollector) saveEntries(entriesFile string) error {
	today := time.Now().Format("2006-01-02")

	// Create goal entries from collected data
	var goalEntries []models.GoalEntry
	for _, goal := range ec.goals {
		completed, exists := ec.entries[goal.ID]
		if !exists {
			continue // Skip goals that weren't processed
		}

		goalEntry := models.GoalEntry{
			GoalID:      goal.ID,
			Value:       completed,
			Notes:       ec.notes[goal.ID],
			CompletedAt: timePtr(time.Now()),
		}

		goalEntries = append(goalEntries, goalEntry)
	}

	// Create day entry
	dayEntry := models.DayEntry{
		Date:  today,
		Goals: goalEntries,
	}

	// Save to storage
	if err := ec.entryStorage.UpdateDayEntry(entriesFile, dayEntry); err != nil {
		return fmt.Errorf("failed to update day entry: %w", err)
	}

	return nil
}

// displayWelcome shows a welcome message with today's date.
func (ec *EntryCollector) displayWelcome() {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("14")). // Bright cyan
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2).
		Margin(1, 0)

	today := time.Now().Format("Monday, January 2, 2006")
	welcome := fmt.Sprintf("🎯 Habit Tracker - %s", today)

	goalCountStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")). // Gray
		Margin(0, 0, 1, 0)

	goalCount := goalCountStyle.Render(fmt.Sprintf("Ready to track %d goals for today!", len(ec.goals)))

	fmt.Println(headerStyle.Render(welcome))
	fmt.Println(goalCount)
}

// displayCompletion shows a completion message with summary.
func (ec *EntryCollector) displayCompletion() {
	completedCount := 0
	totalCount := len(ec.goals)

	for _, completed := range ec.entries {
		if completed {
			completedCount++
		}
	}

	// Choose appropriate styling based on completion rate
	var completionStyle lipgloss.Style
	var emoji string

	completionRate := float64(completedCount) / float64(totalCount)
	switch {
	case completionRate == 1.0:
		completionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")) // Bright green
		emoji = "🎉"
	case completionRate >= 0.7:
		completionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11")) // Bright yellow
		emoji = "💪"
	case completionRate >= 0.5:
		completionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("12")) // Bright blue
		emoji = "👍"
	default:
		completionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8")) // Gray
		emoji = "🤗"
	}

	summaryStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2).
		Margin(1, 0)

	summary := fmt.Sprintf("%s Completed %d out of %d goals today!", emoji, completedCount, totalCount)

	// Add motivational message
	var message string
	switch {
	case completionRate == 1.0:
		message = "Perfect day! You're building amazing habits! ✨"
	case completionRate >= 0.7:
		message = "Great job! You're making excellent progress! 🚀"
	case completionRate >= 0.5:
		message = "Good work! Every step counts towards your goals! 📈"
	default:
		message = "Tomorrow is a new opportunity to build your habits! 🌅"
	}

	messageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Italic(true).
		Margin(1, 0, 0, 0)

	fmt.Println(summaryStyle.Render(completionStyle.Render(summary)))
	fmt.Println(messageStyle.Render(message))

	// Show saved location
	savedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Faint(true)

	fmt.Println(savedStyle.Render("✅ Entries saved successfully!"))
}

// timePtr creates a pointer to a time.Time value.
func timePtr(t time.Time) *time.Time {
	return &t
}
