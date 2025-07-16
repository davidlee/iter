package modal

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"davidlee/vice/internal/models"
	"davidlee/vice/internal/ui"
	"davidlee/vice/internal/ui/entry"
)

// TestEntryFormModal_BasicIntegration tests basic modal functionality without teatest.
func TestEntryFormModal_BasicIntegration(t *testing.T) {
	habit := models.Habit{
		ID:        "test_habit",
		Title:     "Test Habit",
		HabitType: models.SimpleHabit,
		FieldType: models.FieldType{Type: models.BooleanFieldType},
		Prompt:    "Did you complete this habit?",
	}

	collector := ui.NewEntryCollector("testdata/checklists")
	factory := entry.NewEntryFieldInputFactory()

	modal, err := NewEntryFormModal(habit, collector, factory)
	if err != nil {
		t.Fatalf("Failed to create entry form modal: %v", err)
	}

	// Create a modal manager
	modalManager := NewModalManager(80, 24)

	// Test opening modal
	cmd := modalManager.OpenModal(modal)
	if cmd == nil {
		t.Error("Expected command from OpenModal")
	}

	if !modalManager.HasActiveModal() {
		t.Error("Expected active modal after OpenModal")
	}

	// Test ESC key handling
	escMsg := tea.KeyMsg{Type: tea.KeyEsc}
	modalManager.Update(escMsg)

	if modalManager.HasActiveModal() {
		t.Error("Expected modal to be closed after ESC key")
	}
}

// TestModalModel is a simple test model that wraps the modal manager.
type TestModalModel struct {
	modalManager *ModalManager
	modal        Modal
}

func (m *TestModalModel) Init() tea.Cmd {
	return m.modalManager.OpenModal(m.modal)
}

func (m *TestModalModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.modalManager.SetDimensions(msg.Width, msg.Height)
		return m, nil
	case ModalClosedMsg:
		// Modal closed, clean up
		return m, nil
	default:
		// Route all other messages to modal manager
		cmd := m.modalManager.Update(msg)
		return m, cmd
	}
}

func (m *TestModalModel) View() string {
	backgroundView := "Background View"
	return m.modalManager.View(backgroundView)
}

// TestModalManager_Integration tests the modal manager integration.
func TestModalManager_Integration(t *testing.T) {
	habit := models.Habit{
		ID:        "integration_habit",
		Title:     "Integration Test Habit",
		HabitType: models.SimpleHabit,
		FieldType: models.FieldType{Type: models.BooleanFieldType},
	}

	collector := ui.NewEntryCollector("testdata/checklists")
	factory := entry.NewEntryFieldInputFactory()

	modal, err := NewEntryFormModal(habit, collector, factory)
	if err != nil {
		t.Fatalf("Failed to create entry form modal: %v", err)
	}

	modalManager := NewModalManager(80, 24)

	// Test opening modal
	cmd := modalManager.OpenModal(modal)
	if cmd == nil {
		t.Error("Expected command from OpenModal")
	}

	if !modalManager.HasActiveModal() {
		t.Error("Expected active modal after OpenModal")
	}

	// Test view rendering
	backgroundView := "Test Background"
	view := modalManager.View(backgroundView)

	if view == backgroundView {
		t.Error("Expected view to be different with modal")
	}

	// Test modal closing
	cmd = modalManager.CloseModal()
	if cmd == nil {
		t.Error("Expected command from CloseModal")
	}

	if modalManager.HasActiveModal() {
		t.Error("Expected no active modal after CloseModal")
	}
}

// TestEntryFormModal_FormIntegration tests form integration within modal.
func TestEntryFormModal_FormIntegration(t *testing.T) {
	habit := models.Habit{
		ID:        "form_test_habit",
		Title:     "Form Test Habit",
		HabitType: models.SimpleHabit,
		FieldType: models.FieldType{Type: models.BooleanFieldType},
	}

	collector := ui.NewEntryCollector("testdata/checklists")
	factory := entry.NewEntryFieldInputFactory()

	modal, err := NewEntryFormModal(habit, collector, factory)
	if err != nil {
		t.Fatalf("Failed to create entry form modal: %v", err)
	}

	// Initialize modal
	cmd := modal.Init()
	if cmd == nil {
		t.Error("Expected command from Init")
	}

	if !modal.IsOpen() {
		t.Error("Expected modal to be open after Init")
	}

	// Test form view
	view := modal.View()
	if view == "" {
		t.Error("Expected non-empty view")
	}

	// Should contain habit title
	if !strings.Contains(view, habit.Title) {
		t.Error("Expected view to contain habit title")
	}

	// Test ESC key handling
	updatedModal, cmd := modal.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if !updatedModal.IsClosed() {
		t.Error("Expected modal to be closed after ESC")
	}

	if cmd != nil {
		t.Error("Expected nil command from ESC handling")
	}
}
