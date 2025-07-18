package modal

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"github.com/davidlee/vice/internal/debug"
	"github.com/davidlee/vice/internal/models"
	"github.com/davidlee/vice/internal/ui"
	"github.com/davidlee/vice/internal/ui/entry"
)

// EntryFormModal represents a modal for collecting habit entries.
// AIDEV-NOTE: entry-form-modal; replaces form.Run() takeover with modal overlay approach
// AIDEV-NOTE: T024-bug2-fix; eliminates edit looping by providing clean modal close → menu return
// AIDEV-NOTE: T024-experiment; temporarily replaced BaseModal with simple boolean to test lifecycle hypothesis
type EntryFormModal struct {
	// *BaseModal  // TEMPORARILY REMOVED for BaseModal experiment
	isOpen       bool        // Simple boolean flag replacing BaseModal state machine
	result       interface{} // Simple result storage
	habit        models.Habit
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
func NewEntryFormModal(habit models.Habit, collector *ui.EntryCollector, fieldInputFactory *entry.EntryFieldInputFactory) (*EntryFormModal, error) {
	debug.Modal("Creating EntryFormModal for habit: %s (type: %s, field: %s)", habit.ID, habit.HabitType, habit.FieldType.Type)
	// Create existing entry data from collector
	var existing *entry.ExistingEntry
	if collector != nil {
		value, notes, achievement, _, hasEntry := collector.GetHabitEntry(habit.ID)
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
		Habit:         habit,
		FieldType:     habit.FieldType,
		ExistingEntry: existing,
		ShowScoring:   habit.ScoringType == models.AutomaticScoring,
	}

	// Create field input component
	fieldInput, err := fieldInputFactory.CreateInput(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create field input: %w", err)
	}

	// Create the form
	form := fieldInput.CreateInputForm(habit)
	debug.Modal("Created form for habit %s, initial state: %v", habit.ID, form.State)

	modal := &EntryFormModal{
		// BaseModal:  NewBaseModal(),  // TEMPORARILY REMOVED for BaseModal experiment
		isOpen:     false, // Simple boolean flag - will be set to true in Init()
		habit:      habit,
		collector:  collector,
		fieldInput: fieldInput,
		form:       form,
		width:      80,
		height:     24,
	}

	debug.Modal("EntryFormModal created successfully for habit %s", habit.ID)
	return modal, nil
}

// Init initializes the entry form modal.
func (efm *EntryFormModal) Init() tea.Cmd {
	debug.Modal("Initializing modal for habit %s", efm.habit.ID)
	efm.Open()
	cmd := efm.form.Init()
	debug.Modal("Modal initialized, form state: %v, cmd: %v", efm.form.State, cmd != nil)
	return cmd
}

// IsOpen returns true if the modal is currently open
func (efm *EntryFormModal) IsOpen() bool { return efm.isOpen }

// IsClosed returns true if the modal is closed
func (efm *EntryFormModal) IsClosed() bool { return !efm.isOpen }

// Open sets the modal to open state
func (efm *EntryFormModal) Open() {
	debug.Modal("Habit %s: Opening modal (simple boolean)", efm.habit.ID)
	efm.isOpen = true
}

// Close sets the modal to closed state
func (efm *EntryFormModal) Close() {
	debug.Modal("Habit %s: Closing modal (simple boolean)", efm.habit.ID)
	efm.isOpen = false
}

// SetResult stores the modal result
func (efm *EntryFormModal) SetResult(result interface{}) { efm.result = result }

// GetResult returns the stored modal result
func (efm *EntryFormModal) GetResult() interface{} { return efm.result }

// Update handles messages for the entry form modal.
func (efm *EntryFormModal) Update(msg tea.Msg) (Modal, tea.Cmd) {
	msgType := fmt.Sprintf("%T", msg)
	debug.Modal("Habit %s: Update received %s, form state: %v", efm.habit.ID, msgType, efm.form.State)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		debug.Modal("Habit %s: WindowSizeMsg %dx%d", efm.habit.ID, msg.Width, msg.Height)
		efm.width = msg.Width
		efm.height = msg.Height
		return efm, nil

	case tea.KeyMsg:
		// Handle modal-specific keys (ESC) before passing to form
		if msg.String() == "esc" {
			debug.Modal("Habit %s: ESC pressed, closing modal without saving", efm.habit.ID)
			efm.Close()
			return efm, nil
		}
		debug.Modal("Habit %s: KeyMsg %s", efm.habit.ID, msg.String())
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
		debug.Modal("Habit %s: Form state changed from %v to %v after %s", efm.habit.ID, oldState, efm.form.State, msgType)
	}

	// Check if form is complete
	if efm.form.State == huh.StateCompleted {
		debug.Modal("Habit %s: Form completed, processing entry", efm.habit.ID)
		efm.formComplete = true
		return efm.processEntry()
	}

	// Check if form was aborted
	if efm.form.State == huh.StateAborted {
		debug.Modal("Habit %s: Form aborted, closing modal", efm.habit.ID)
		efm.Close()
		return efm, cmd
	}

	return efm, cmd
}

// AIDEV-NOTE: T024-fix; removed HandleKey method - using canonical huh+bubbletea pattern instead

// processEntry processes the habit entry and closes the modal.
// AIDEV-NOTE: entry-processing; processes form completion and creates EntryResult
func (efm *EntryFormModal) processEntry() (Modal, tea.Cmd) {
	debug.Modal("Habit %s: Processing entry, validating input", efm.habit.ID)

	// Validate the input
	if err := efm.fieldInput.Validate(); err != nil {
		debug.Modal("Habit %s: Validation failed: %v", efm.habit.ID, err)
		efm.error = fmt.Errorf("validation failed: %w", err)
		return efm, nil
	}

	// Get the collected value and status
	value := efm.fieldInput.GetValue()
	status := efm.fieldInput.GetStatus()
	debug.Modal("Habit %s: Collected value: %v, status: %v", efm.habit.ID, value, status)

	// Create the entry result
	result := &entry.EntryResult{
		Value:  value,
		Status: status,
	}

	// Handle scoring if needed
	if efm.habit.ScoringType == models.AutomaticScoring {
		debug.Modal("Habit %s: Automatic scoring required (TODO: not implemented)", efm.habit.ID)
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

	debug.Modal("Habit %s: Entry processed successfully, modal closed", efm.habit.ID)
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

	title := titleStyle.Render(fmt.Sprintf("Entry: %s", efm.habit.Title))

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
	return efm.entryResult
}
