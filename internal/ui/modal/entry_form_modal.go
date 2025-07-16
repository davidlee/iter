package modal

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"davidlee/vice/internal/debug"
	"davidlee/vice/internal/models"
	"davidlee/vice/internal/ui"
	"davidlee/vice/internal/ui/entry"
)

// EntryFormModal represents a modal for collecting goal entries.
// AIDEV-NOTE: entry-form-modal; replaces form.Run() takeover with modal overlay approach
// AIDEV-NOTE: T024-bug2-fix; eliminates edit looping by providing clean modal close → menu return
// AIDEV-NOTE: T024-experiment; temporarily replaced BaseModal with simple boolean to test lifecycle hypothesis
type EntryFormModal struct {
	// *BaseModal  // TEMPORARILY REMOVED for BaseModal experiment
	isOpen       bool        // Simple boolean flag replacing BaseModal state machine
	result       interface{} // Simple result storage
	goal         models.Goal
	collector    *ui.EntryCollector
	fieldInput   entry.EntryFieldInput
	form         *huh.Form
	entryResult  *entry.EntryResult
	error        error
	width        int
	height       int
	formComplete bool
}

// NewEntryFormModal creates a new entry form modal.
// AIDEV-NOTE: modal-factory; key factory method integrating existing entry field input system
func NewEntryFormModal(goal models.Goal, collector *ui.EntryCollector, fieldInputFactory *entry.EntryFieldInputFactory) (*EntryFormModal, error) {
	debug.Modal("Creating EntryFormModal for goal: %s (type: %s, field: %s)", goal.ID, goal.GoalType, goal.FieldType.Type)
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
	debug.Modal("Created form for goal %s, initial state: %v", goal.ID, form.State)

	modal := &EntryFormModal{
		// BaseModal:  NewBaseModal(),  // TEMPORARILY REMOVED for BaseModal experiment
		isOpen:     false, // Simple boolean flag - will be set to true in Init()
		goal:       goal,
		collector:  collector,
		fieldInput: fieldInput,
		form:       form,
		width:      80,
		height:     24,
	}

	debug.Modal("EntryFormModal created successfully for goal %s", goal.ID)
	return modal, nil
}

// Init initializes the entry form modal.
func (efm *EntryFormModal) Init() tea.Cmd {
	debug.Modal("Initializing modal for goal %s", efm.goal.ID)
	efm.Open()
	cmd := efm.form.Init()
	debug.Modal("Modal initialized, form state: %v, cmd: %v", efm.form.State, cmd != nil)
	return cmd
}

// Modal interface compliance - replaced BaseModal methods with simple boolean operations
func (efm *EntryFormModal) IsOpen() bool   { return efm.isOpen }
func (efm *EntryFormModal) IsClosed() bool { return !efm.isOpen }
func (efm *EntryFormModal) Open() {
	debug.Modal("Goal %s: Opening modal (simple boolean)", efm.goal.ID)
	efm.isOpen = true
}

func (efm *EntryFormModal) Close() {
	debug.Modal("Goal %s: Closing modal (simple boolean)", efm.goal.ID)
	efm.isOpen = false
}
func (efm *EntryFormModal) SetResult(result interface{}) { efm.result = result }
func (efm *EntryFormModal) GetResult() interface{}       { return efm.result }

// Update handles messages for the entry form modal.
func (efm *EntryFormModal) Update(msg tea.Msg) (Modal, tea.Cmd) {
	msgType := fmt.Sprintf("%T", msg)
	debug.Modal("Goal %s: Update received %s, form state: %v", efm.goal.ID, msgType, efm.form.State)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		debug.Modal("Goal %s: WindowSizeMsg %dx%d", efm.goal.ID, msg.Width, msg.Height)
		efm.width = msg.Width
		efm.height = msg.Height
		return efm, nil

	case tea.KeyMsg:
		// Handle modal-specific keys (ESC) before passing to form
		if msg.String() == "esc" {
			debug.Modal("Goal %s: ESC pressed, closing modal without saving", efm.goal.ID)
			efm.Close()
			return efm, nil
		}
		debug.Modal("Goal %s: KeyMsg %s", efm.goal.ID, msg.String())
	}

	// Process the form using canonical pattern from huh/examples/bubbletea
	// AIDEV-NOTE: T024-fix; following canonical huh+bubbletea integration pattern to fix double-processing
	oldState := efm.form.State
	var cmd tea.Cmd
	formModel, cmd := efm.form.Update(msg)
	if f, ok := formModel.(*huh.Form); ok {
		efm.form = f
	}

	if efm.form.State != oldState {
		debug.Modal("Goal %s: Form state changed from %v to %v after %s", efm.goal.ID, oldState, efm.form.State, msgType)
	}

	// Check if form is complete
	if efm.form.State == huh.StateCompleted {
		debug.Modal("Goal %s: Form completed, processing entry", efm.goal.ID)
		efm.formComplete = true
		return efm.processEntry()
	}

	// Check if form was aborted
	if efm.form.State == huh.StateAborted {
		debug.Modal("Goal %s: Form aborted, closing modal", efm.goal.ID)
		efm.Close()
		return efm, cmd
	}

	return efm, cmd
}

// AIDEV-NOTE: T024-fix; removed HandleKey method - using canonical huh+bubbletea pattern instead

// processEntry processes the goal entry and closes the modal.
// AIDEV-NOTE: entry-processing; processes form completion and creates EntryResult
func (efm *EntryFormModal) processEntry() (Modal, tea.Cmd) {
	debug.Modal("Goal %s: Processing entry, validating input", efm.goal.ID)

	// Validate the input
	if err := efm.fieldInput.Validate(); err != nil {
		debug.Modal("Goal %s: Validation failed: %v", efm.goal.ID, err)
		efm.error = fmt.Errorf("validation failed: %w", err)
		return efm, nil
	}

	// Get the collected value and status
	value := efm.fieldInput.GetValue()
	status := efm.fieldInput.GetStatus()
	debug.Modal("Goal %s: Collected value: %v, status: %v", efm.goal.ID, value, status)

	// Create the entry result
	result := &entry.EntryResult{
		Value:  value,
		Status: status,
	}

	// Handle scoring if needed
	if efm.goal.ScoringType == models.AutomaticScoring {
		debug.Modal("Goal %s: Automatic scoring required (TODO: not implemented)", efm.goal.ID)
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
	efm.entryResult = result
	efm.SetResult(result)
	efm.Close()

	debug.Modal("Goal %s: Entry processed successfully, modal closed", efm.goal.ID)
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

// renderDisabledForm renders a static message when form is disabled for testing
func (efm *EntryFormModal) renderDisabledForm() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		Align(lipgloss.Center).
		Margin(0, 0, 1, 0)

	messageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("11")).
		Align(lipgloss.Center).
		Margin(1, 0, 1, 0)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true).
		Align(lipgloss.Center).
		Margin(1, 0, 0, 0)

	title := titleStyle.Render(efm.goal.Title)
	message := messageStyle.Render("🔬 T024 DEBUG: Form processing disabled")
	help := helpStyle.Render("Press Esc to close • Modal should stay open indefinitely")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		message,
		help,
	)
}

// GetEntryResult returns the entry result if available.
func (efm *EntryFormModal) GetEntryResult() *entry.EntryResult {
	return efm.entryResult
}
