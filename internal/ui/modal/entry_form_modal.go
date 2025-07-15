package modal

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"davidlee/vice/internal/models"
	"davidlee/vice/internal/ui"
	"davidlee/vice/internal/ui/entry"
)

// AIDEV-NOTE: T024-debug-logging; comprehensive debug logging for modal behavior investigation
var debugLogger *log.Logger

func init() {
	debugLogger = log.New(os.Stderr, "[MODAL-DEBUG] ", log.LstdFlags|log.Lshortfile)
}

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
	debugLogger.Printf("Creating EntryFormModal for goal: %s (type: %s, field: %s)", goal.ID, goal.GoalType, goal.FieldType.Type)
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
	debugLogger.Printf("Created form for goal %s, initial state: %v", goal.ID, form.State)

	modal := &EntryFormModal{
		BaseModal:  NewBaseModal(),
		goal:       goal,
		collector:  collector,
		fieldInput: fieldInput,
		form:       form,
		width:      80,
		height:     24,
	}

	debugLogger.Printf("EntryFormModal created successfully for goal %s", goal.ID)
	return modal, nil
}

// Init initializes the entry form modal.
func (efm *EntryFormModal) Init() tea.Cmd {
	debugLogger.Printf("Initializing modal for goal %s", efm.goal.ID)
	efm.Open()
	cmd := efm.form.Init()
	debugLogger.Printf("Modal initialized, form state: %v, cmd: %v", efm.form.State, cmd != nil)
	return cmd
}

// Update handles messages for the entry form modal.
func (efm *EntryFormModal) Update(msg tea.Msg) (Modal, tea.Cmd) {
	msgType := fmt.Sprintf("%T", msg)
	debugLogger.Printf("Goal %s: Update received %s, form state: %v", efm.goal.ID, msgType, efm.form.State)
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		debugLogger.Printf("Goal %s: WindowSizeMsg %dx%d", efm.goal.ID, msg.Width, msg.Height)
		efm.width = msg.Width
		efm.height = msg.Height
		return efm, nil

	case tea.KeyMsg:
		debugLogger.Printf("Goal %s: KeyMsg %s", efm.goal.ID, msg.String())
		// Handle modal-specific keys first
		return efm.HandleKey(msg)

	default:
		// Let the form handle other messages
		// AIDEV-NOTE: form-integration; critical type assertion and state monitoring
		oldState := efm.form.State
		var cmd tea.Cmd
		formModel, cmd := efm.form.Update(msg)
		efm.form = formModel.(*huh.Form)
		
		if efm.form.State != oldState {
			debugLogger.Printf("Goal %s: Form state changed from %v to %v after %s", efm.goal.ID, oldState, efm.form.State, msgType)
		}

		// Check if form is complete
		if efm.form.State == huh.StateCompleted {
			// AIDEV-NOTE: T024-debug; form completed via non-key message, processing entry
			debugLogger.Printf("Goal %s: Form completed via non-key message (%s), processing entry", efm.goal.ID, msgType)
			efm.formComplete = true
			return efm.processEntry()
		}

		// Check if form was aborted
		if efm.form.State == huh.StateAborted {
			// AIDEV-NOTE: T024-debug; form aborted via non-key message, closing modal
			debugLogger.Printf("Goal %s: Form aborted via non-key message (%s), closing modal", efm.goal.ID, msgType)
			efm.Close()
			return efm, cmd
		}

		return efm, cmd
	}
}

// HandleKey handles keyboard input for the entry form modal.
func (efm *EntryFormModal) HandleKey(msg tea.KeyMsg) (Modal, tea.Cmd) {
	debugLogger.Printf("Goal %s: HandleKey %s, form state: %v", efm.goal.ID, msg.String(), efm.form.State)
	
	switch msg.String() {
	case "esc":
		// Close modal without saving
		debugLogger.Printf("Goal %s: ESC pressed, closing modal without saving", efm.goal.ID)
		efm.Close()
		return efm, nil

	default:
		// Let the form handle all other keys
		oldState := efm.form.State
		var cmd tea.Cmd
		formModel, cmd := efm.form.Update(msg)
		efm.form = formModel.(*huh.Form)
		
		if efm.form.State != oldState {
			debugLogger.Printf("Goal %s: Form state changed from %v to %v after key %s", efm.goal.ID, oldState, efm.form.State, msg.String())
		}

		// Check if form is complete
		if efm.form.State == huh.StateCompleted {
			// AIDEV-NOTE: T024-debug; form completed via key input, processing entry
			debugLogger.Printf("Goal %s: Form completed via key input (%s), processing entry", efm.goal.ID, msg.String())
			efm.formComplete = true
			return efm.processEntry()
		}

		// Check if form was aborted
		if efm.form.State == huh.StateAborted {
			// AIDEV-NOTE: T024-debug; form aborted via key input, closing modal
			debugLogger.Printf("Goal %s: Form aborted via key input (%s), closing modal", efm.goal.ID, msg.String())
			efm.Close()
			return efm, cmd
		}

		return efm, cmd
	}
}

// processEntry processes the goal entry and closes the modal.
// AIDEV-NOTE: entry-processing; processes form completion and creates EntryResult
func (efm *EntryFormModal) processEntry() (Modal, tea.Cmd) {
	debugLogger.Printf("Goal %s: Processing entry, validating input", efm.goal.ID)
	
	// Validate the input
	if err := efm.fieldInput.Validate(); err != nil {
		debugLogger.Printf("Goal %s: Validation failed: %v", efm.goal.ID, err)
		efm.error = fmt.Errorf("validation failed: %w", err)
		return efm, nil
	}

	// Get the collected value and status
	value := efm.fieldInput.GetValue()
	status := efm.fieldInput.GetStatus()
	debugLogger.Printf("Goal %s: Collected value: %v, status: %v", efm.goal.ID, value, status)

	// Create the entry result
	result := &entry.EntryResult{
		Value:  value,
		Status: status,
	}

	// Handle scoring if needed
	if efm.goal.ScoringType == models.AutomaticScoring {
		debugLogger.Printf("Goal %s: Automatic scoring required (TODO: not implemented)", efm.goal.ID)
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

	debugLogger.Printf("Goal %s: Entry processed successfully, modal closed", efm.goal.ID)
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
