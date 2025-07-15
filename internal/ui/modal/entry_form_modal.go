package modal

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"davidlee/vice/internal/models"
	"davidlee/vice/internal/ui"
	"davidlee/vice/internal/ui/entry"
)

// EntryFormModal represents a modal for collecting goal entries.
// AIDEV-NOTE: entry-form-modal; replaces form.Run() takeover with modal overlay approach
// AIDEV-NOTE: T024-bug2-fix; eliminates edit looping by providing clean modal close → menu return
type EntryFormModal struct {
	*BaseModal
	goal         models.Goal
	collector    *ui.EntryCollector
	fieldInput   entry.EntryFieldInput
	form         *huh.Form
	result       *entry.EntryResult
	error        error
	width        int
	height       int
	formComplete bool
}

// NewEntryFormModal creates a new entry form modal.
// AIDEV-NOTE: modal-factory; key factory method integrating existing entry field input system
func NewEntryFormModal(goal models.Goal, collector *ui.EntryCollector, fieldInputFactory *entry.EntryFieldInputFactory) (*EntryFormModal, error) {
	// Create existing entry data from collector
	var existing *entry.ExistingEntry
	if collector != nil {
		value, notes, achievement, _, hasEntry := collector.GetGoalEntry(goal.ID)
		if hasEntry {
			existing = &entry.ExistingEntry{
				Value:            value,
				Notes:            notes,
				AchievementLevel: achievement,
			}
		}
	}

	// Create field input configuration
	config := entry.EntryFieldInputConfig{
		Goal:          goal,
		FieldType:     goal.FieldType,
		ExistingEntry: existing,
		ShowScoring:   goal.ScoringType == models.AutomaticScoring,
	}

	// Create field input component
	fieldInput, err := fieldInputFactory.CreateInput(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create field input: %w", err)
	}

	// Create the form
	form := fieldInput.CreateInputForm(goal)

	modal := &EntryFormModal{
		BaseModal:  NewBaseModal(),
		goal:       goal,
		collector:  collector,
		fieldInput: fieldInput,
		form:       form,
		width:      80,
		height:     24,
	}

	return modal, nil
}

// Init initializes the entry form modal.
func (efm *EntryFormModal) Init() tea.Cmd {
	efm.Open()
	return efm.form.Init()
}

// Update handles messages for the entry form modal.
func (efm *EntryFormModal) Update(msg tea.Msg) (Modal, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		efm.width = msg.Width
		efm.height = msg.Height
		return efm, nil

	case tea.KeyMsg:
		// Handle modal-specific keys first
		return efm.HandleKey(msg)

	default:
		// Let the form handle other messages
		// AIDEV-NOTE: form-integration; critical type assertion and state monitoring
		var cmd tea.Cmd
		formModel, cmd := efm.form.Update(msg)
		efm.form = formModel.(*huh.Form)

		// Check if form is complete
		if efm.form.State == huh.StateCompleted {
			// AIDEV-NOTE: T024-debug; form completed via non-key message, processing entry
			efm.formComplete = true
			return efm.processEntry()
		}

		// Check if form was aborted
		if efm.form.State == huh.StateAborted {
			// AIDEV-NOTE: T024-debug; form aborted via non-key message, closing modal
			efm.Close()
			return efm, cmd
		}

		return efm, cmd
	}
}

// HandleKey handles keyboard input for the entry form modal.
func (efm *EntryFormModal) HandleKey(msg tea.KeyMsg) (Modal, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Close modal without saving
		efm.Close()
		return efm, nil

	default:
		// Let the form handle all other keys
		var cmd tea.Cmd
		formModel, cmd := efm.form.Update(msg)
		efm.form = formModel.(*huh.Form)

		// Check if form is complete
		if efm.form.State == huh.StateCompleted {
			// AIDEV-NOTE: T024-debug; form completed via key input, processing entry
			efm.formComplete = true
			return efm.processEntry()
		}

		// Check if form was aborted
		if efm.form.State == huh.StateAborted {
			// AIDEV-NOTE: T024-debug; form aborted via key input, closing modal
			efm.Close()
			return efm, cmd
		}

		return efm, cmd
	}
}

// processEntry processes the goal entry and closes the modal.
// AIDEV-NOTE: entry-processing; processes form completion and creates EntryResult
func (efm *EntryFormModal) processEntry() (Modal, tea.Cmd) {
	// Validate the input
	if err := efm.fieldInput.Validate(); err != nil {
		efm.error = fmt.Errorf("validation failed: %w", err)
		return efm, nil
	}

	// Get the collected value and status
	value := efm.fieldInput.GetValue()
	status := efm.fieldInput.GetStatus()

	// Create the entry result
	result := &entry.EntryResult{
		Value:  value,
		Status: status,
	}

	// Handle scoring if needed
	if efm.goal.ScoringType == models.AutomaticScoring {
		// TODO: Implement scoring integration
		// AIDEV-NOTE: scoring-todo; needs integration with existing scoring engine
		// For now, just set achievement level to nil
		result.AchievementLevel = nil
	}

	// TODO: Collect notes if needed
	// AIDEV-NOTE: notes-todo; needs integration with collectStandardOptionalNotes pattern
	// For now, just set empty notes
	result.Notes = ""

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

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true).
		Align(lipgloss.Center).
		Margin(1, 0, 0, 0)

	title := titleStyle.Render(fmt.Sprintf("Entry: %s", efm.goal.Title))

	// Render the form
	formContent := efm.form.View()

	help := helpStyle.Render("Press Esc to cancel")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		formContent,
		"",
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
