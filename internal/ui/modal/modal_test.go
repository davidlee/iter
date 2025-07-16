package modal

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestModalManager_Creation tests modal manager creation and initialization.
func TestModalManager_Creation(t *testing.T) {
	mm := NewModalManager(80, 24)

	if mm.width != 80 {
		t.Errorf("Expected width 80, got %d", mm.width)
	}

	if mm.height != 24 {
		t.Errorf("Expected height 24, got %d", mm.height)
	}

	if mm.HasActiveModal() {
		t.Error("Expected no active modal initially")
	}
}

// TestModalManager_SetDimensions tests dimension updates.
func TestModalManager_SetDimensions(t *testing.T) {
	mm := NewModalManager(80, 24)
	mm.SetDimensions(100, 30)

	if mm.width != 100 {
		t.Errorf("Expected width 100, got %d", mm.width)
	}

	if mm.height != 30 {
		t.Errorf("Expected height 30, got %d", mm.height)
	}
}

// TestModalManager_ViewWithoutModal tests view rendering without modal.
func TestModalManager_ViewWithoutModal(t *testing.T) {
	mm := NewModalManager(80, 24)
	background := "test background"

	result := mm.View(background)

	if result != background {
		t.Errorf("Expected background unchanged, got %s", result)
	}
}

// TestBaseModal_States tests base modal state transitions.
func TestBaseModal_States(t *testing.T) {
	bm := NewBaseModal()

	// Initial state
	if bm.GetState() != ModalOpening {
		t.Errorf("Expected initial state ModalOpening, got %v", bm.GetState())
	}

	if !bm.IsOpen() {
		t.Error("Expected modal to be open initially")
	}

	if bm.IsClosed() {
		t.Error("Expected modal not to be closed initially")
	}

	// Open state
	bm.Open()
	if bm.GetState() != ModalActive {
		t.Errorf("Expected state ModalActive after Open(), got %v", bm.GetState())
	}

	if !bm.IsOpen() {
		t.Error("Expected modal to be open after Open()")
	}

	if bm.IsClosed() {
		t.Error("Expected modal not to be closed after Open()")
	}

	// Close state
	bm.Close()
	if bm.GetState() != ModalClosed {
		t.Errorf("Expected state ModalClosed after Close(), got %v", bm.GetState())
	}

	if bm.IsOpen() {
		t.Error("Expected modal not to be open after Close()")
	}

	if !bm.IsClosed() {
		t.Error("Expected modal to be closed after Close()")
	}
}

// TestBaseModal_Result tests modal result handling.
func TestBaseModal_Result(t *testing.T) {
	bm := NewBaseModal()

	// Initial result should be nil
	if bm.GetResult() != nil {
		t.Error("Expected initial result to be nil")
	}

	// Set and get result
	testResult := "test result"
	bm.SetResult(testResult)

	if bm.GetResult() != testResult {
		t.Errorf("Expected result %s, got %v", testResult, bm.GetResult())
	}
}

// MockModal is a test modal implementation.
type MockModal struct {
	*BaseModal
	keyHandled bool
	viewCalled bool
	initCalled bool
}

// NewMockModal creates a new mock modal.
func NewMockModal() *MockModal {
	return &MockModal{
		BaseModal: NewBaseModal(),
	}
}

// Init implements Modal interface.
func (mm *MockModal) Init() tea.Cmd {
	mm.initCalled = true
	mm.Open()
	return nil
}

// Update implements Modal interface.
func (mm *MockModal) Update(msg tea.Msg) (Modal, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		return mm.HandleKey(keyMsg)
	}
	return mm, nil
}

// View implements Modal interface.
func (mm *MockModal) View() string {
	mm.viewCalled = true
	return "mock modal view"
}

// HandleKey implements Modal interface.
func (mm *MockModal) HandleKey(msg tea.KeyMsg) (Modal, tea.Cmd) {
	mm.keyHandled = true
	if msg.String() == "esc" {
		mm.Close()
	}
	return mm, nil
}

// TestModalManager_OpenModal tests modal opening.
func TestModalManager_OpenModal(t *testing.T) {
	mm := NewModalManager(80, 24)
	modal := NewMockModal()

	cmd := mm.OpenModal(modal)

	if !mm.HasActiveModal() {
		t.Error("Expected active modal after OpenModal()")
	}

	if cmd == nil {
		t.Error("Expected command from OpenModal()")
	}

	if !modal.initCalled {
		t.Error("Expected modal Init() to be called")
	}
}

// TestModalManager_CloseModal tests modal closing.
func TestModalManager_CloseModal(t *testing.T) {
	mm := NewModalManager(80, 24)
	modal := NewMockModal()

	// Open modal first
	mm.OpenModal(modal)

	// Set a result
	testResult := "test result"
	modal.SetResult(testResult)

	// Close modal
	cmd := mm.CloseModal()

	if mm.HasActiveModal() {
		t.Error("Expected no active modal after CloseModal()")
	}

	if cmd == nil {
		t.Error("Expected command from CloseModal()")
	}
}

// TestModalManager_KeyRouting tests key routing to modal.
func TestModalManager_KeyRouting(t *testing.T) {
	mm := NewModalManager(80, 24)
	modal := NewMockModal()

	mm.OpenModal(modal)

	// Send key message
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")}
	cmd := mm.Update(keyMsg)

	if !modal.keyHandled {
		t.Error("Expected key to be handled by modal")
	}

	if cmd != nil {
		t.Error("Expected nil command for normal key")
	}
}

// TestModalManager_EscapeKeyClosing tests ESC key closing modal.
func TestModalManager_EscapeKeyClosing(t *testing.T) {
	mm := NewModalManager(80, 24)
	modal := NewMockModal()

	mm.OpenModal(modal)

	// Send ESC key
	escMsg := tea.KeyMsg{Type: tea.KeyEsc}
	cmd := mm.Update(escMsg)

	if mm.HasActiveModal() {
		t.Error("Expected modal to be closed after ESC key")
	}

	if cmd == nil {
		t.Error("Expected command from ESC key handling")
	}
}

// TestModalManager_ViewWithModal tests view rendering with modal.
func TestModalManager_ViewWithModal(t *testing.T) {
	mm := NewModalManager(80, 24)
	modal := NewMockModal()

	background := "test background"

	// Without modal
	result := mm.View(background)
	if result != background {
		t.Errorf("Expected background without modal, got %s", result)
	}

	// With modal
	mm.OpenModal(modal)
	result = mm.View(background)

	if result == background {
		t.Error("Expected different view with modal")
	}

	if !modal.viewCalled {
		t.Error("Expected modal View() to be called")
	}
}
