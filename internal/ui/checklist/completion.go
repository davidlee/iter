// Package checklist provides UI components for checklist interactions
package checklist

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/davidlee/vice/internal/models"
)

// CompletionModel represents the interactive checklist completion state.
// This is adapted from the prototype in internal/ui/checklist.go
type CompletionModel struct {
	checklist  *models.Checklist
	items      []string
	cursor     int
	selected   map[int]struct{}
	completion *models.ChecklistCompletion
	progress   progress.Model
}

// NewCompletionModel creates a new checklist completion model.
func NewCompletionModel(checklist *models.Checklist) *CompletionModel {
	prog := progress.New(progress.WithDefaultGradient())
	prog.Width = 60

	model := &CompletionModel{
		checklist: checklist,
		items:     checklist.Items,
		selected:  make(map[int]struct{}),
		completion: &models.ChecklistCompletion{
			ChecklistID:    checklist.ID,
			CompletedItems: make(map[string]bool),
		},
		progress: prog,
	}

	// Set cursor to index of first non-heading
	for model.cursor < len(model.items) && strings.HasPrefix(model.items[model.cursor], "# ") {
		model.cursor++
	}

	return model
}

// NewCompletionModelWithState creates a completion model with existing completion state.
// AIDEV-NOTE: state-ui-restore; maps completion data back to UI selection state
func NewCompletionModelWithState(checklist *models.Checklist, completion *models.ChecklistCompletion) *CompletionModel {
	model := NewCompletionModel(checklist)
	model.completion = completion

	// Restore selected state from completion data
	for i, item := range model.items {
		if !strings.HasPrefix(item, "# ") && completion.CompletedItems[item] {
			model.selected[i] = struct{}{}
		}
	}

	return model
}

// Init implements the bubbletea.Model interface.
func (m CompletionModel) Init() tea.Cmd {
	return nil
}

// Update implements the bubbletea.Model interface.
func (m CompletionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		// Exit keys
		case "ctrl+c", "q":
			return m, tea.Quit

		// Navigation - up
		case "up", "e":
			if m.cursor > 0 {
				m.cursor--
				for m.cursor > 0 && strings.HasPrefix(m.items[m.cursor], "# ") {
					m.cursor--
				}
				// Handle case where first item(s) are headings
				for m.cursor < len(m.items) && strings.HasPrefix(m.items[m.cursor], "# ") {
					m.cursor++
				}
			}

		// Navigation - down
		case "down", "a":
			if m.cursor < len(m.items)-1 {
				m.cursor++
				for m.cursor < len(m.items)-1 && strings.HasPrefix(m.items[m.cursor], "# ") {
					m.cursor++
				}
			}

		// Toggle selection
		case "enter", " ":
			if m.cursor < len(m.items) && !strings.HasPrefix(m.items[m.cursor], "# ") {
				item := m.items[m.cursor]

				_, selected := m.selected[m.cursor]
				if selected {
					delete(m.selected, m.cursor)
					m.completion.CompletedItems[item] = false
				} else {
					m.selected[m.cursor] = struct{}{}
					m.completion.CompletedItems[item] = true
				}
			}
		}
	}

	return m, nil
}

// View implements the bubbletea.Model interface.
func (m CompletionModel) View() string {
	// Styles (same as prototype)
	headerStyle := lipgloss.NewStyle().Bold(true).Underline(true).Foreground(lipgloss.Color("63"))
	headingStyle := lipgloss.NewStyle().Bold(true).Underline(true).Foreground(lipgloss.Color("202"))
	itemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	checkedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#3C3C3C"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("201"))

	// Header with checklist title
	title := m.checklist.Title
	if title == "" {
		title = m.checklist.ID
	}
	s := headerStyle.Render(fmt.Sprintf("Complete checklist: %s", title)) + "\n\n"

	// Show description if available
	if m.checklist.Description != "" {
		s += lipgloss.NewStyle().Italic(true).Render(m.checklist.Description) + "\n\n"
	}

	// Iterate over items (same logic as prototype)
	for i, item := range m.items {
		isHeading := strings.HasPrefix(item, "# ")

		// Cursor indicator
		cursor := " " // no cursor
		if m.cursor == i {
			if !isHeading {
				cursor = ">" // cursor!
			}
		}

		// Selection indicator
		checked := " " // not selected
		if _, ok := m.selected[i]; ok {
			checked = "x" // selected!
		}

		// Render the row
		if isHeading {
			if i > 0 {
				s += "\n"
			}
			s += "      "
			text := strings.TrimLeft(item, "# ")

			// Add progress indicator to heading
			// AIDEV-NOTE: heading-progress-display; injects "(completed/total)" into section headings
			completed, total := m.getSectionProgress(i)
			if total > 0 {
				text = fmt.Sprintf("%s (%d/%d)", text, completed, total)
			}

			s += headingStyle.Render(text)
			s += "\n"
		} else {
			text := fmt.Sprintf("%s [%s] %s", cursor, checked, item)
			switch {
			case cursor == ">":
				s += selectedStyle.Render(text)
			case checked == "x":
				s += checkedStyle.Render(text)
			default:
				s += itemStyle.Render(text)
			}
			s += "\n"
		}
	}

	// Progress bar and footer
	completedCount := len(m.selected)
	totalItems := m.getTotalItemCount()

	// Calculate progress percentage
	var progressPercent float64
	if totalItems > 0 {
		progressPercent = float64(completedCount) / float64(totalItems)
	}

	// AIDEV-NOTE: bubbles-progress-bar; visual gradient progress bar with percentage display (commit 04973be)
	s += "\n" + m.progress.ViewAs(progressPercent) + "\n"
	s += fmt.Sprintf("Completed: %d/%d items (%.0f%%)", completedCount, totalItems, progressPercent*100)
	s += "\nPress q to quit.\n"

	return s
}

// GetCompletion returns the current completion state.
func (m CompletionModel) GetCompletion() *models.ChecklistCompletion {
	// Update partial completion flag
	totalItems := m.getTotalItemCount()
	completedItems := len(m.selected)
	m.completion.PartialComplete = completedItems > 0 && completedItems < totalItems

	return m.completion
}

// getTotalItemCount returns the number of non-heading items.
func (m CompletionModel) getTotalItemCount() int {
	count := 0
	for _, item := range m.items {
		if !strings.HasPrefix(item, "# ") {
			count++
		}
	}
	return count
}

// getSectionProgress returns completion progress for a heading section.
// Returns (completed, total) for items between the current heading and next heading (or end).
// AIDEV-NOTE: section-progress-calc; parses checklist items into sections for heading progress indicators
func (m CompletionModel) getSectionProgress(headingIndex int) (int, int) {
	if headingIndex < 0 || headingIndex >= len(m.items) || !strings.HasPrefix(m.items[headingIndex], "# ") {
		return 0, 0
	}

	completed := 0
	total := 0

	// Find items in this section (between this heading and next heading)
	for i := headingIndex + 1; i < len(m.items); i++ {
		// Stop at next heading
		if strings.HasPrefix(m.items[i], "# ") {
			break
		}

		// Count non-heading items
		total++
		if _, ok := m.selected[i]; ok {
			completed++
		}
	}

	return completed, total
}

// RunChecklistCompletion runs the checklist completion interface and returns the completion state.
func RunChecklistCompletion(checklist *models.Checklist) (*models.ChecklistCompletion, error) {
	model := NewCompletionModel(checklist)

	program := tea.NewProgram(model)
	finalModel, err := program.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to run checklist interface: %w", err)
	}

	// Extract the completion state
	if completionModel, ok := finalModel.(CompletionModel); ok {
		return completionModel.GetCompletion(), nil
	}

	return nil, fmt.Errorf("unexpected model type returned")
}

// RunChecklistCompletionWithState runs the checklist completion interface with existing state.
func RunChecklistCompletionWithState(checklist *models.Checklist, existingCompletion *models.ChecklistCompletion) (*models.ChecklistCompletion, error) {
	model := NewCompletionModelWithState(checklist, existingCompletion)

	program := tea.NewProgram(model)
	finalModel, err := program.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to run checklist interface: %w", err)
	}

	// Extract the completion state
	if completionModel, ok := finalModel.(CompletionModel); ok {
		return completionModel.GetCompletion(), nil
	}

	return nil, fmt.Errorf("unexpected model type returned")
}
