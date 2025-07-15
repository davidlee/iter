package modal

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"davidlee/vice/internal/models"
	"davidlee/vice/internal/ui"
	"davidlee/vice/internal/ui/entry"
)

// EntryFormModal represents a modal for collecting goal entries.
// AIDEV-NOTE: entry-form-modal; replaces form.Run() takeover with modal overlay approach
type EntryFormModal struct {
	*BaseModal
	goal      models.Goal
	collector *ui.EntryCollector
	flow      entry.GoalCollectionFlow
	input     entry.EntryFieldInput
	result    *entry.EntryResult
	error     error
	width     int
	height    int
}

// NewEntryFormModal creates a new entry form modal.
func NewEntryFormModal(goal models.Goal, collector *ui.EntryCollector, flowFactory *entry.GoalCollectionFlowFactory) (*EntryFormModal, error) {
	// Create the appropriate collection flow for this goal type
	flow, err := flowFactory.CreateFlow(string(goal.GoalType))
	if err != nil {
		return nil, fmt.Errorf("failed to create collection flow: %w", err)
	}

	modal := &EntryFormModal{
		BaseModal: NewBaseModal(),
		goal:      goal,
		collector: collector,
		flow:      flow,
		width:     80,
		height:    24,
	}

	return modal, nil
}

// Init initializes the entry form modal.
func (efm *EntryFormModal) Init() tea.Cmd {
	efm.Open()
	return nil
}

// Update handles messages for the entry form modal.
func (efm *EntryFormModal) Update(msg tea.Msg) (Modal, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		efm.width = msg.Width
		efm.height = msg.Height
		return efm, nil

	case tea.KeyMsg:
		// Handle modal-specific keys
		return efm.HandleKey(msg)

	default:
		return efm, nil
	}
}

// HandleKey handles keyboard input for the entry form modal.
func (efm *EntryFormModal) HandleKey(msg tea.KeyMsg) (Modal, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Close modal without saving
		efm.Close()
		return efm, nil

	case "enter":
		// Process the entry
		return efm.processEntry()

	default:
		// For now, just close on any other key
		// TODO: Implement proper form input handling
		return efm, nil
	}
}

// processEntry processes the goal entry and closes the modal.
func (efm *EntryFormModal) processEntry() (Modal, tea.Cmd) {
	// Create existing entry data from collector
	var existing *entry.ExistingEntry
	if efm.collector != nil {
		value, notes, achievement, _, hasEntry := efm.collector.GetGoalEntry(efm.goal.ID)
		if hasEntry {
			existing = &entry.ExistingEntry{
				Value:            value,
				Notes:            notes,
				AchievementLevel: achievement,
			}
		}
	}

	// Use the flow to collect the entry
	// TODO: Replace this with proper modal form handling
	result, err := efm.flow.CollectEntry(efm.goal, existing)
	if err != nil {
		efm.error = err
		return efm, nil
	}

	// Store the result and close modal
	efm.result = result
	efm.SetResult(result)
	efm.Close()

	return efm, nil
}

// View renders the entry form modal.
func (efm *EntryFormModal) View() string {
	if efm.error != nil {
		return efm.renderError()
	}

	return efm.renderForm()
}

// renderForm renders the main form content.
func (efm *EntryFormModal) renderForm() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		Align(lipgloss.Center).
		Margin(0, 0, 1, 0)

	promptStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("7")).
		Margin(0, 0, 1, 0)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true).
		Align(lipgloss.Center).
		Margin(1, 0, 0, 0)

	title := titleStyle.Render(efm.goal.Title)
	
	prompt := efm.goal.Prompt
	if prompt == "" {
		prompt = "Enter your progress for this goal:"
	}
	promptText := promptStyle.Render(prompt)

	// TODO: Implement proper form input based on goal type
	formContent := "Form input will be implemented here"

	help := helpStyle.Render("Press Enter to save • Press Esc to cancel")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		promptText,
		formContent,
		help,
	)
}

// renderError renders an error message.
func (efm *EntryFormModal) renderError() string {
	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true).
		Align(lipgloss.Center).
		Margin(1, 0)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true).
		Align(lipgloss.Center).
		Margin(1, 0, 0, 0)

	errorText := errorStyle.Render(fmt.Sprintf("Error: %s", efm.error.Error()))
	help := helpStyle.Render("Press Esc to close")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		errorText,
		help,
	)
}

// GetEntryResult returns the entry result if available.
func (efm *EntryFormModal) GetEntryResult() *entry.EntryResult {
	return efm.result
}