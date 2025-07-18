// Package main provides a T024 modal investigation prototype for debugging modal behavior.
package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"github.com/davidlee/vice/internal/debug"
	"github.com/davidlee/vice/internal/models"
	"github.com/davidlee/vice/internal/ui"
	"github.com/davidlee/vice/internal/ui/entry"
)

// Habit represents a simplified habit structure
type Habit struct {
	ID    string
	Title string
}

// EntryResult represents the result of entry collection
type EntryResult struct {
	Value  interface{}
	Status string
}

// ModalState represents the current state of a modal.
type ModalState int

const (
	// ModalOpening is when the modal is being opened.
	ModalOpening ModalState = iota
	// ModalActive is when the modal is fully open and active.
	ModalActive
	// ModalClosing is when the modal is being closed.
	ModalClosing
	// ModalClosed is when the modal is fully closed.
	ModalClosed
)

// BaseModal provides common modal functionality.
type BaseModal struct {
	state  ModalState
	result interface{}
}

// NewBaseModal creates a new base modal.
func NewBaseModal() *BaseModal {
	return &BaseModal{
		state: ModalOpening,
	}
}

// IsOpen returns true if the modal is open.
func (bm *BaseModal) IsOpen() bool {
	return bm.state == ModalOpening || bm.state == ModalActive
}

// IsClosed returns true if the modal is closed.
func (bm *BaseModal) IsClosed() bool {
	return bm.state == ModalClosed
}

// GetResult returns the modal result.
func (bm *BaseModal) GetResult() interface{} {
	return bm.result
}

// SetResult sets the modal result.
func (bm *BaseModal) SetResult(result interface{}) {
	bm.result = result
}

// Open opens the modal.
func (bm *BaseModal) Open() {
	bm.state = ModalActive
}

// Close closes the modal.
func (bm *BaseModal) Close() {
	bm.state = ModalClosed
}

// EntryFormModal represents a modal for collecting habit entries
type EntryFormModal struct {
	*BaseModal
	habit        Habit
	fieldInput   entry.EntryFieldInput
	form         *huh.Form
	result       *EntryResult
	inputFactory *entry.EntryFieldInputFactory
}

// NewEntryFormModal creates a new entry form modal with Entry Collection Context
func NewEntryFormModal(habit Habit, entryCollector *ui.EntryCollector) *EntryFormModal {
	// Create models.Habit from prototype Habit
	modelsHabit := models.Habit{
		ID:        habit.ID,
		Title:     habit.Title,
		HabitType: "simple",
	}

	// Get existing entry from collector
	value, notes, achievement, status, hasEntry := entryCollector.GetHabitEntry(habit.ID)

	var existingEntry *entry.ExistingEntry
	if hasEntry {
		existingEntry = &entry.ExistingEntry{
			Value:            value,
			Notes:            notes,
			AchievementLevel: achievement,
		}
		debug.Modal("Using existing entry for habit %s: value=%v, notes=%s, status=%s",
			habit.ID, value, notes, status)
	} else {
		debug.Modal("No existing entry for habit %s", habit.ID)
	}

	// Create real field input factory
	inputFactory := entry.NewEntryFieldInputFactory()

	// Create field input config with existing entry context
	config := entry.EntryFieldInputConfig{
		Habit:         modelsHabit,
		FieldType:     models.FieldType{Type: models.BooleanFieldType},
		ExistingEntry: existingEntry,
		ShowScoring:   true, // Enable scoring for complex state
	}

	// Create field input using factory
	fieldInput, err := inputFactory.CreateInput(config)
	if err != nil {
		debug.Modal("Error creating field input: %v", err)
		return nil
	}

	// Create form using field input
	form := fieldInput.CreateInputForm(modelsHabit)

	return &EntryFormModal{
		BaseModal:    NewBaseModal(),
		habit:        habit,
		fieldInput:   fieldInput,
		form:         form,
		inputFactory: inputFactory,
	}
}

// Init initializes the entry form modal
func (efm *EntryFormModal) Init() tea.Cmd {
	efm.Open()
	return efm.form.Init()
}

// View renders the entry form modal content
func (efm *EntryFormModal) View() string {
	if efm.form.State == huh.StateCompleted && efm.result != nil {
		value := efm.fieldInput.GetStringValue()
		return fmt.Sprintf("You selected: %s", value)
	}

	return efm.form.View()
}

// Modal interface compliance
type Modal interface {
	// Lifecycle
	Init() tea.Cmd
	Update(msg tea.Msg) (Modal, tea.Cmd)
	View() string

	// State
	IsOpen() bool
	IsClosed() bool

	// Integration
	GetResult() interface{}
}

// Update handles messages for the entry form modal - Modal interface compliance
func (efm *EntryFormModal) Update(msg tea.Msg) (Modal, tea.Cmd) {
	msgType := fmt.Sprintf("%T", msg)
	debug.Modal("EntryFormModal.Update(Modal): received %s, form state: %d", msgType, efm.form.State)

	// Process the form using canonical pattern
	oldState := efm.form.State
	var cmd tea.Cmd
	formModel, cmd := efm.form.Update(msg)
	if f, ok := formModel.(*huh.Form); ok {
		efm.form = f
	}

	if efm.form.State != oldState {
		debug.Modal("EntryFormModal(Modal): Form state changed from %d to %d", oldState, efm.form.State)
	}

	// Check if form is complete
	if efm.form.State == huh.StateCompleted {
		debug.Modal("EntryFormModal(Modal): Form completed, closing modal")
		efm.result = &EntryResult{
			Value:  efm.fieldInput.GetValue(),
			Status: "completed",
		}
		efm.SetResult(efm.result)
		efm.Close()
	}

	// Check if form was aborted
	if efm.form.State == huh.StateAborted {
		debug.Modal("EntryFormModal(Modal): Form aborted, closing modal")
		efm.Close()
	}

	return efm, cmd
}

const maxWidth = 80

var (
	red    = lipgloss.AdaptiveColor{Light: "#FE5F86", Dark: "#FE5F86"}
	indigo = lipgloss.AdaptiveColor{Light: "#5A56E0", Dark: "#7571F9"}
	green  = lipgloss.AdaptiveColor{Light: "#02BA84", Dark: "#02BF87"}
)

// Styles holds styling configuration for the prototype
type Styles struct {
	Base,
	HeaderText,
	Status,
	StatusHeader,
	Highlight,
	ErrorHeaderText,
	Help lipgloss.Style
}

// NewStyles creates a new Styles instance with the given renderer
func NewStyles(lg *lipgloss.Renderer) *Styles {
	s := Styles{}
	s.Base = lg.NewStyle().
		Padding(1, 4, 0, 1)
	s.HeaderText = lg.NewStyle().
		Foreground(indigo).
		Bold(true).
		Padding(0, 1, 0, 2)
	s.Status = lg.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(indigo).
		PaddingLeft(1).
		MarginTop(1)
	s.StatusHeader = lg.NewStyle().
		Foreground(green).
		Bold(true)
	s.Highlight = lg.NewStyle().
		Foreground(lipgloss.Color("212"))
	s.ErrorHeaderText = s.HeaderText.
		Foreground(red)
	s.Help = lg.NewStyle().
		Foreground(lipgloss.Color("240"))
	return &s
}

// ModalManager manages the display and interaction of modals.
type ModalManager struct {
	activeModal  Modal
	overlayStyle lipgloss.Style
	dimStyle     lipgloss.Style
	width        int
	height       int
}

// NewModalManager creates a new modal manager.
func NewModalManager(width, height int) *ModalManager {
	return &ModalManager{
		width:  width,
		height: height,
		overlayStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Background(lipgloss.Color("235")).
			Padding(1, 2).
			Margin(1, 2),
		dimStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Faint(true),
	}
}

// HasActiveModal returns true if there's an active modal.
func (mm *ModalManager) HasActiveModal() bool {
	return mm.activeModal != nil && mm.activeModal.IsOpen()
}

// OpenModal opens a modal and returns the initialization command.
func (mm *ModalManager) OpenModal(modal Modal) tea.Cmd {
	mm.activeModal = modal
	return modal.Init()
}

// CloseModal closes the current modal.
func (mm *ModalManager) CloseModal() tea.Cmd {
	if mm.activeModal == nil {
		return nil
	}
	mm.activeModal = nil
	return nil
}

// Update processes messages for the modal manager.
func (mm *ModalManager) Update(msg tea.Msg) tea.Cmd {
	if mm.activeModal == nil {
		return nil
	}

	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		mm.width = msg.Width
		mm.height = msg.Height
		return nil
	}

	// Route all messages to the active modal
	var cmd tea.Cmd
	mm.activeModal, cmd = mm.activeModal.Update(msg)

	// Check if modal closed itself
	if mm.activeModal.IsClosed() {
		return tea.Batch(cmd, mm.CloseModal())
	}

	return cmd
}

// View renders the modal over the background content.
func (mm *ModalManager) View(backgroundView string) string {
	if mm.activeModal == nil || !mm.activeModal.IsOpen() {
		return backgroundView
	}

	return mm.renderWithModal(backgroundView, mm.activeModal.View())
}

// renderWithModal renders the modal overlay on top of the background.
func (mm *ModalManager) renderWithModal(background, modal string) string {
	// Dim the background
	dimmedBg := mm.dimStyle.Render(background)

	// Style the modal
	styledModal := mm.overlayStyle.Render(modal)

	// Center the modal
	centeredModal := lipgloss.Place(
		mm.width, mm.height,
		lipgloss.Center, lipgloss.Center,
		styledModal,
	)

	// Overlay the modal on the background
	return lipgloss.JoinVertical(lipgloss.Left, dimmedBg, centeredModal)
}

// EntryMenuModel simulates the real application's entry menu layer
type EntryMenuModel struct {
	modalManager      *ModalManager
	fieldInputFactory *entry.EntryFieldInputFactory
	entryCollector    *ui.EntryCollector
	habits            []models.Habit
	entries           map[string]models.HabitEntry
	width             int
	height            int
}

// NewEntryMenuModel creates a new entry menu model with Entry Collection Context
func NewEntryMenuModel(width, height int) *EntryMenuModel {
	// Create habits and entries with complex state
	habits := []models.Habit{
		{
			ID:        "test_habit",
			Title:     "Exercise",
			HabitType: "simple",
		},
	}

	// Create existing entries with achievement levels and notes
	achievementLevel := models.AchievementMidi
	entries := map[string]models.HabitEntry{
		"test_habit": {
			HabitID:          "test_habit",
			Value:            true, // Existing boolean value
			AchievementLevel: &achievementLevel,
			Notes:            "Previous completion with notes",
			CreatedAt:        time.Now().Add(-24 * time.Hour), // Yesterday
			UpdatedAt:        nil,
			Status:           models.EntryCompleted,
		},
	}

	// Create entry collector with complex state
	entryCollector := ui.NewEntryCollector("/tmp/test_entries.yml")
	entryCollector.InitializeForMenu(habits, entries)

	return &EntryMenuModel{
		modalManager:      NewModalManager(width, height),
		fieldInputFactory: entry.NewEntryFieldInputFactory(),
		entryCollector:    entryCollector,
		habits:            habits,
		entries:           entries,
		width:             width,
		height:            height,
	}
}

// Update processes messages for the entry menu model
func (em *EntryMenuModel) Update(msg tea.Msg) tea.Cmd {
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		em.width = msg.Width
		em.height = msg.Height
	}

	// Route to modal manager
	return em.modalManager.Update(msg)
}

// View renders the entry menu with modal overlay
func (em *EntryMenuModel) View() string {
	background := "Entry Menu Background"
	return em.modalManager.View(background)
}

// OpenModal opens a modal using the entry menu's modal manager
func (em *EntryMenuModel) OpenModal(modal Modal) tea.Cmd {
	return em.modalManager.OpenModal(modal)
}

// OpenEntryFormModal creates and opens an entry form modal with collector context
func (em *EntryMenuModel) OpenEntryFormModal(habit Habit) tea.Cmd {
	modal := NewEntryFormModal(habit, em.entryCollector)
	return em.modalManager.OpenModal(modal)
}

// HasActiveModal returns true if there's an active modal
func (em *EntryMenuModel) HasActiveModal() bool {
	return em.modalManager.HasActiveModal()
}

// Model represents the main application model for the prototype
type Model struct {
	entryMenu *EntryMenuModel
	lg        *lipgloss.Renderer
	width     int
	height    int
}

// NewModel creates a new Model instance
func NewModel() Model {
	m := Model{width: maxWidth, height: 24}
	m.lg = lipgloss.DefaultRenderer()

	// Setup entry menu model (simulates real app architecture)
	m.entryMenu = NewEntryMenuModel(m.width, m.height)

	return m
}

// Init initializes the model and returns the initial command
func (m Model) Init() tea.Cmd {
	// Open the entry form modal on startup via entry menu with collector context
	habit := Habit{
		ID:    "test_habit",
		Title: "Exercise",
	}
	return m.entryMenu.OpenEntryFormModal(habit)
}

// min returns the minimum of two integers
//
//revive:disable-next-line:redefines-builtin-id -- using min for clarity in prototype code
func min(x, y int) int {
	if x > y {
		return y
	}
	return x
}

// Update processes messages and returns updated model and command
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = min(msg.Width, maxWidth)
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Interrupt
		case "esc", "q":
			return m, tea.Quit
		}
	}

	var cmds []tea.Cmd

	// Process through entry menu model (simulates real app architecture)
	if m.entryMenu.HasActiveModal() {
		debug.General("Main Update: Entry menu has active modal, processing message")
		cmd := m.entryMenu.Update(msg)
		cmds = append(cmds, cmd)

		if !m.entryMenu.HasActiveModal() {
			debug.General("Main Update: Entry menu closed modal, quitting")
			// Quit when the modal is closed.
			cmds = append(cmds, tea.Quit)
		}
	} else {
		debug.General("Main Update: No active modal, ignoring message")
	}

	return m, tea.Batch(cmds...)
}

// View renders the model as a string
func (m Model) View() string {
	// Render via entry menu model (simulates real app architecture)
	return m.entryMenu.View()
}

// getRole function removed - not needed for boolean modal

func main() {
	// Initialize debug logging to /tmp for prototype
	err := debug.GetInstance().Initialize("/tmp")
	if err != nil {
		fmt.Printf("Failed to initialize debug logging: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		_ = debug.GetInstance().Close()
	}()

	debug.General("Starting test modal prototype")

	_, err = tea.NewProgram(NewModel()).Run()
	if err != nil {
		fmt.Println("Oh no:", err)
		// Exit with error code after defer functions complete
		defer os.Exit(1)
	}
}
