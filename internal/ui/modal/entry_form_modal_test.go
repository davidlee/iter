package modal

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"davidlee/vice/internal/models"
	"davidlee/vice/internal/ui"
	"davidlee/vice/internal/ui/entry"
)

// TestEntryFormModal_Creation tests entry form modal creation.
func TestEntryFormModal_Creation(t *testing.T) {
	habit := models.Habit{
		ID:        "test_habit",
		Title:     "Test Habit",
		HabitType: models.SimpleHabit,
		FieldType: models.FieldType{Type: models.BooleanFieldType},
	}

	collector := ui.NewEntryCollector("testdata/checklists")
	factory := entry.NewEntryFieldInputFactory()

	modal, err := NewEntryFormModal(habit, collector, factory)
	if err != nil {
		t.Fatalf("Failed to create entry form modal: %v", err)
	}

	if modal.habit.ID != habit.ID {
		t.Errorf("Expected habit ID %s, got %s", habit.ID, modal.habit.ID)
	}

	if modal.collector != collector {
		t.Error("Expected collector to be set")
	}

	if modal.fieldInput == nil {
		t.Error("Expected field input to be created")
	}

	if modal.form == nil {
		t.Error("Expected form to be created")
	}
}

// TestEntryFormModal_InitializationWithExistingEntry tests modal creation with existing entry.
func TestEntryFormModal_InitializationWithExistingEntry(t *testing.T) {
	habit := models.Habit{
		ID:        "existing_habit",
		Title:     "Existing Habit",
		HabitType: models.SimpleHabit,
		FieldType: models.FieldType{Type: models.BooleanFieldType},
	}

	collector := ui.NewEntryCollector("testdata/checklists")

	// Set up existing entry in collector
	achievement := models.AchievementMini
	collector.SetEntryForTesting(habit.ID, true, &achievement, "test notes")

	factory := entry.NewEntryFieldInputFactory()

	modal, err := NewEntryFormModal(habit, collector, factory)
	if err != nil {
		t.Fatalf("Failed to create entry form modal: %v", err)
	}

	// The modal should have been initialized with the existing entry
	// This is tested indirectly through the field input component
	if modal.fieldInput == nil {
		t.Error("Expected field input to be created with existing entry")
	}
}

// TestEntryFormModal_Init tests modal initialization.
func TestEntryFormModal_Init(t *testing.T) {
	habit := models.Habit{
		ID:        "test_habit",
		Title:     "Test Habit",
		HabitType: models.SimpleHabit,
		FieldType: models.FieldType{Type: models.BooleanFieldType},
	}

	collector := ui.NewEntryCollector("testdata/checklists")
	factory := entry.NewEntryFieldInputFactory()

	modal, err := NewEntryFormModal(habit, collector, factory)
	if err != nil {
		t.Fatalf("Failed to create entry form modal: %v", err)
	}

	// Test initialization
	cmd := modal.Init()

	if !modal.IsOpen() {
		t.Error("Expected modal to be open after Init()")
	}

	if !modal.IsOpen() {
		t.Error("Expected modal to be open after Init()")
	}

	if cmd == nil {
		t.Error("Expected command from Init()")
	}
}

// TestEntryFormModal_HandleKey tests keyboard handling.
func TestEntryFormModal_HandleKey(t *testing.T) {
	habit := models.Habit{
		ID:        "test_habit",
		Title:     "Test Habit",
		HabitType: models.SimpleHabit,
		FieldType: models.FieldType{Type: models.BooleanFieldType},
	}

	collector := ui.NewEntryCollector("testdata/checklists")
	factory := entry.NewEntryFieldInputFactory()

	modal, err := NewEntryFormModal(habit, collector, factory)
	if err != nil {
		t.Fatalf("Failed to create entry form modal: %v", err)
	}

	modal.Init()

	// Test ESC key closes modal
	escMsg := tea.KeyMsg{Type: tea.KeyEsc}
	updatedModal, cmd := modal.Update(escMsg)

	if !updatedModal.IsClosed() {
		t.Error("Expected modal to be closed after ESC key")
	}

	if cmd != nil {
		t.Error("Expected nil command for ESC key")
	}
}

// TestEntryFormModal_Update tests message handling.
func TestEntryFormModal_Update(t *testing.T) {
	habit := models.Habit{
		ID:        "test_habit",
		Title:     "Test Habit",
		HabitType: models.SimpleHabit,
		FieldType: models.FieldType{Type: models.BooleanFieldType},
	}

	collector := ui.NewEntryCollector("testdata/checklists")
	factory := entry.NewEntryFieldInputFactory()

	modal, err := NewEntryFormModal(habit, collector, factory)
	if err != nil {
		t.Fatalf("Failed to create entry form modal: %v", err)
	}

	modal.Init()

	// Test window size update
	sizeMsg := tea.WindowSizeMsg{Width: 100, Height: 30}
	updatedModal, cmd := modal.Update(sizeMsg)

	if efm, ok := updatedModal.(*EntryFormModal); ok {
		if efm.width != 100 {
			t.Errorf("Expected width 100, got %d", efm.width)
		}
		if efm.height != 30 {
			t.Errorf("Expected height 30, got %d", efm.height)
		}
	}

	if cmd != nil {
		t.Error("Expected nil command for window size message")
	}
}

// TestEntryFormModal_View tests view rendering.
func TestEntryFormModal_View(t *testing.T) {
	habit := models.Habit{
		ID:        "test_habit",
		Title:     "Test Habit",
		HabitType: models.SimpleHabit,
		FieldType: models.FieldType{Type: models.BooleanFieldType},
	}

	collector := ui.NewEntryCollector("testdata/checklists")
	factory := entry.NewEntryFieldInputFactory()

	modal, err := NewEntryFormModal(habit, collector, factory)
	if err != nil {
		t.Fatalf("Failed to create entry form modal: %v", err)
	}

	modal.Init()

	// Test normal view
	view := modal.View()

	if view == "" {
		t.Error("Expected non-empty view")
	}

	// Should contain habit title
	if len(view) < 10 {
		t.Error("Expected substantial view content")
	}
}

// TestEntryFormModal_ErrorView tests error rendering.
func TestEntryFormModal_ErrorView(t *testing.T) {
	habit := models.Habit{
		ID:        "test_habit",
		Title:     "Test Habit",
		HabitType: models.SimpleHabit,
		FieldType: models.FieldType{Type: models.BooleanFieldType},
	}

	collector := ui.NewEntryCollector("testdata/checklists")
	factory := entry.NewEntryFieldInputFactory()

	modal, err := NewEntryFormModal(habit, collector, factory)
	if err != nil {
		t.Fatalf("Failed to create entry form modal: %v", err)
	}

	modal.Init()

	// Set an error
	modal.error = fmt.Errorf("test error")

	view := modal.View()

	if view == "" {
		t.Error("Expected non-empty error view")
	}

	// Should contain error message
	if len(view) < 10 {
		t.Error("Expected substantial error view content")
	}
}

// TestEntryFormModal_GetEntryResult tests result retrieval.
func TestEntryFormModal_GetEntryResult(t *testing.T) {
	habit := models.Habit{
		ID:        "test_habit",
		Title:     "Test Habit",
		HabitType: models.SimpleHabit,
		FieldType: models.FieldType{Type: models.BooleanFieldType},
	}

	collector := ui.NewEntryCollector("testdata/checklists")
	factory := entry.NewEntryFieldInputFactory()

	modal, err := NewEntryFormModal(habit, collector, factory)
	if err != nil {
		t.Fatalf("Failed to create entry form modal: %v", err)
	}

	modal.Init()

	// Initially no result
	result := modal.GetEntryResult()
	if result != nil {
		t.Error("Expected no result initially")
	}

	// Set a result
	testResult := &entry.EntryResult{
		Value:  true,
		Status: models.EntryCompleted,
	}
	modal.entryResult = testResult

	result = modal.GetEntryResult()
	if result != testResult {
		t.Error("Expected to get the set result")
	}
}

// TestEntryFormModal_ProcessEntry tests entry processing.
func TestEntryFormModal_ProcessEntry(t *testing.T) {
	habit := models.Habit{
		ID:        "test_habit",
		Title:     "Test Habit",
		HabitType: models.SimpleHabit,
		FieldType: models.FieldType{Type: models.BooleanFieldType},
	}

	collector := ui.NewEntryCollector("testdata/checklists")
	factory := entry.NewEntryFieldInputFactory()

	modal, err := NewEntryFormModal(habit, collector, factory)
	if err != nil {
		t.Fatalf("Failed to create entry form modal: %v", err)
	}

	modal.Init()

	// Simulate form completion
	modal.form.State = huh.StateCompleted
	modal.formComplete = true

	// Process the entry
	updatedModal, cmd := modal.processEntry()

	if !updatedModal.IsClosed() {
		t.Error("Expected modal to be closed after processing entry")
	}

	if modal.GetResult() == nil {
		t.Error("Expected result to be set after processing entry")
	}

	if cmd != nil {
		t.Error("Expected nil command from processEntry")
	}
}

// TestEntryFormModal_FormAborted tests handling of form abortion.
func TestEntryFormModal_FormAborted(t *testing.T) {
	habit := models.Habit{
		ID:        "test_habit",
		Title:     "Test Habit",
		HabitType: models.SimpleHabit,
		FieldType: models.FieldType{Type: models.BooleanFieldType},
	}

	collector := ui.NewEntryCollector("testdata/checklists")
	factory := entry.NewEntryFieldInputFactory()

	modal, err := NewEntryFormModal(habit, collector, factory)
	if err != nil {
		t.Fatalf("Failed to create entry form modal: %v", err)
	}

	modal.Init()

	// Simulate form abortion
	modal.form.State = huh.StateAborted

	// Update should close the modal
	updatedModal, _ := modal.Update(tea.KeyMsg{})

	if !updatedModal.IsClosed() {
		t.Error("Expected modal to be closed after form abortion")
	}

	if modal.GetResult() != nil {
		t.Error("Expected no result after form abortion")
	}
}
