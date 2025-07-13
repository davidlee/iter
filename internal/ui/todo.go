package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"

	"davidlee/iter/internal/models"
	"davidlee/iter/internal/parser"
	"davidlee/iter/internal/storage"
	"davidlee/iter/internal/config"
)

// TodoDashboard displays today's habit status in a table format
type TodoDashboard struct {
	paths *config.Paths
}

// NewTodoDashboard creates a new todo dashboard instance
func NewTodoDashboard(paths *config.Paths) *TodoDashboard {
	return &TodoDashboard{
		paths: paths,
	}
}

// HabitStatus represents the status of a single habit for today
type HabitStatus struct {
	Goal   models.Goal
	Status models.EntryStatus
	Value  interface{}
	Notes  string
}

// Display shows the todo dashboard with bubbles table (non-interactive)
func (td *TodoDashboard) Display() error {
	// Load today's habit statuses
	statuses, err := td.loadTodayStatuses()
	if err != nil {
		return fmt.Errorf("failed to load habit statuses: %w", err)
	}

	return td.displayBubblesTable(statuses)
}

// DisplayASCII shows a plain ASCII table
func (td *TodoDashboard) DisplayASCII() error {
	// Load today's habit statuses
	statuses, err := td.loadTodayStatuses()
	if err != nil {
		return fmt.Errorf("failed to load habit statuses: %w", err)
	}

	return td.displaySimpleTable(statuses)
}

// DisplayMarkdown shows the todo dashboard as markdown checklist
func (td *TodoDashboard) DisplayMarkdown() error {
	// Load today's habit statuses
	statuses, err := td.loadTodayStatuses()
	if err != nil {
		return fmt.Errorf("failed to load habit statuses: %w", err)
	}

	// Output markdown format
	return td.displayMarkdownList(statuses)
}

// displayBubblesTable shows a non-interactive bubbles table
func (td *TodoDashboard) displayBubblesTable(statuses []HabitStatus) error {
	columns := []table.Column{
		{Title: "Status", Width: 6},
		{Title: "Habit", Width: 30},
		{Title: "Value", Width: 20},
		{Title: "Notes", Width: 30},
	}

	rows := make([]table.Row, len(statuses))
	for i, status := range statuses {
		symbol := td.getStatusSymbol(status.Status)
		value := td.formatValue(status.Value)
		notes := td.truncateString(status.Notes, 30)
		
		rows[i] = table.Row{
			symbol,
			td.truncateString(status.Goal.Title, 30),
			value,
			notes,
		}
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(false),
		table.WithHeight(len(rows)),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	// Create a simple model that just renders once
	model := simpleTableModel{table: t, statuses: statuses}
	
	// Just render the view directly (no interaction needed)
	fmt.Print(model.View())
	td.displaySummary(statuses)
	
	return nil
}

// simpleTableModel wraps the bubbles table for static rendering
type simpleTableModel struct {
	table    table.Model
	statuses []HabitStatus
}

func (m simpleTableModel) View() string {
	baseStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))
	
	return baseStyle.Render(m.table.View()) + "\n\n"
}

// displayMarkdownList shows a markdown todo list
func (td *TodoDashboard) displayMarkdownList(statuses []HabitStatus) error {
	fmt.Println("# Today's Habits")
	fmt.Println()

	for _, status := range statuses {
		checkbox := td.getMarkdownCheckbox(status.Status)
		fmt.Printf("%s %s\n", checkbox, status.Goal.Title)
		
		// Add notes aligned with habit text (no bullet, indented to align with habit title)
		if status.Notes != "" {
			fmt.Printf("      %s\n", status.Notes)
		}
		
		// Add value aligned with habit text if present and not boolean
		if status.Value != nil {
			valueStr := td.formatValue(status.Value)
			if valueStr != "" && valueStr != "true" && valueStr != "false" {
				fmt.Printf("      Value: %s\n", valueStr)
			}
		}
	}

	fmt.Println()
	td.displaySummary(statuses)
	return nil
}

// getMarkdownCheckbox returns the markdown checkbox for a given status
func (td *TodoDashboard) getMarkdownCheckbox(status models.EntryStatus) string {
	switch status {
	case models.EntryCompleted:
		return "- [x]"
	case models.EntrySkipped:
		return "- [-]"
	case models.EntryFailed:
		return "- [ ]" // Failed shown as unchecked
	case "pending":
		return "- [ ]"
	default:
		return "- [ ]"
	}
}


// displaySimpleTable shows a basic text table
func (td *TodoDashboard) displaySimpleTable(statuses []HabitStatus) error {
	fmt.Println("Today's Habits:")
	fmt.Println("Status | Habit                         | Value               | Notes")
	fmt.Println("-------|-------------------------------|---------------------|------------------------------")
	
	for _, status := range statuses {
		symbol := td.getStatusSymbol(status.Status)
		habit := td.truncateString(status.Goal.Title, 29)
		value := td.truncateString(td.formatValue(status.Value), 19)
		notes := td.truncateString(status.Notes, 30)
		
		fmt.Printf("%-6s | %-29s | %-19s | %-30s\n", symbol, habit, value, notes)
	}
	
	td.displaySummary(statuses)
	return nil
}

// loadTodayStatuses loads all goals and today's entries to determine status
func (td *TodoDashboard) loadTodayStatuses() ([]HabitStatus, error) {
	// Load goals
	goalParser := parser.NewGoalParser()
	schema, err := goalParser.LoadFromFile(td.paths.GoalsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load goals: %w", err)
	}

	// Load today's entries
	entryStorage := storage.NewEntryStorage()
	entryLog, err := entryStorage.LoadFromFile(td.paths.EntriesFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load entries: %w", err)
	}

	// Get today's date
	today := time.Now().Format("2006-01-02")
	
	// Find today's entry
	var todayEntry *models.DayEntry
	for _, dayEntry := range entryLog.Entries {
		if dayEntry.Date == today {
			todayEntry = &dayEntry
			break
		}
	}

	// Build status list
	var statuses []HabitStatus
	for _, goal := range schema.Goals {
		status := HabitStatus{
			Goal:   goal,
			Status: "pending", // Default to pending (no EntryPending constant)
		}

		// Check if we have an entry for this goal today
		if todayEntry != nil {
			for _, goalEntry := range todayEntry.Goals {
				if goalEntry.GoalID == goal.ID {
					status.Status = goalEntry.Status
					status.Value = goalEntry.Value
					status.Notes = goalEntry.Notes
					break
				}
			}
		}

		statuses = append(statuses, status)
	}

	return statuses, nil
}

// displaySummary shows completion statistics
func (td *TodoDashboard) displaySummary(statuses []HabitStatus) {
	completed := 0
	skipped := 0
	failed := 0
	total := len(statuses)

	for _, status := range statuses {
		switch status.Status {
		case models.EntryCompleted:
			completed++
		case models.EntrySkipped:
			skipped++
		case models.EntryFailed:
			failed++
		}
	}

	pending := total - completed - skipped - failed
	
	fmt.Printf("\nSummary: %d/%d completed", completed, total)
	if skipped > 0 {
		fmt.Printf(", %d skipped", skipped)
	}
	if failed > 0 {
		fmt.Printf(", %d failed", failed)
	}
	if pending > 0 {
		fmt.Printf(", %d pending", pending)
	}
	fmt.Println()
}

// formatValue converts a value to a display string
func (td *TodoDashboard) formatValue(value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%v", value)
}

// truncateString truncates a string to the specified length
func (td *TodoDashboard) truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// getStatusSymbol returns the Unicode symbol for a given status
func (td *TodoDashboard) getStatusSymbol(status models.EntryStatus) string {
	switch status {
	case models.EntryCompleted:
		return "✓"
	case models.EntrySkipped:
		return "⤫"
	case models.EntryFailed:
		return "✗"
	case "pending":
		return "○"
	default:
		return "?"
	}
}