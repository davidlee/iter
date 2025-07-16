// Package modal provides a modal/overlay system for BubbleTea applications.
// AIDEV-NOTE: modal-system; eliminates form.Run() takeover approach with overlay pattern
// AIDEV-NOTE: T024-solution; architectural fix for entry menu bugs via modal overlay system
package modal

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Modal represents a modal dialog that can be displayed over other content.
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

// ModalState represents the current state of a modal.
//revive:disable-next-line:exported -- ModalState name follows Go naming convention
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

// ModalOpenedMsg is sent when a modal is opened.
//revive:disable-next-line:exported -- ModalOpenedMsg name follows Go naming convention
type ModalOpenedMsg struct {
	Modal Modal
}

// ModalClosedMsg is sent when a modal is closed.
//revive:disable-next-line:exported -- ModalClosedMsg name follows Go naming convention
type ModalClosedMsg struct {
	Result interface{}
}

// ModalManager manages the display and interaction of modals.
// AIDEV-NOTE: modal-manager; core component orchestrating modal lifecycle and overlay rendering
//revive:disable-next-line:exported -- ModalManager name follows Go naming convention
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
// AIDEV-NOTE: modal-open; entry point for launching modals from parent components
func (mm *ModalManager) OpenModal(modal Modal) tea.Cmd {
	mm.activeModal = modal
	return tea.Batch(
		modal.Init(),
		func() tea.Msg { return ModalOpenedMsg{Modal: modal} },
	)
}

// CloseModal closes the current modal.
// AIDEV-NOTE: modal-close; returns result and cleans up modal state
func (mm *ModalManager) CloseModal() tea.Cmd {
	if mm.activeModal == nil {
		return nil
	}

	result := mm.activeModal.GetResult()
	mm.activeModal = nil

	return func() tea.Msg {
		return ModalClosedMsg{Result: result}
	}
}

// Update processes messages for the modal manager.
// AIDEV-NOTE: modal-routing; critical keyboard and message routing to active modal
func (mm *ModalManager) Update(msg tea.Msg) tea.Cmd {
	if mm.activeModal == nil {
		return nil
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		mm.width = msg.Width
		mm.height = msg.Height
		return nil

	default:
		// Route all messages to the active modal (including KeyMsg)
		// AIDEV-NOTE: T024-fix; simplified routing, let modal handle all messages directly
		var cmd tea.Cmd
		mm.activeModal, cmd = mm.activeModal.Update(msg)

		// Check if modal closed itself
		if mm.activeModal.IsClosed() {
			return tea.Batch(cmd, mm.CloseModal())
		}

		return cmd
	}
}

// View renders the modal over the background content.
func (mm *ModalManager) View(backgroundView string) string {
	if mm.activeModal == nil || !mm.activeModal.IsOpen() {
		return backgroundView
	}

	return mm.renderWithModal(backgroundView, mm.activeModal.View())
}

// renderWithModal renders the modal overlay on top of the background.
// AIDEV-NOTE: modal-rendering; layered rendering with dimmed background and centered modal
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

// SetDimensions updates the modal manager dimensions.
func (mm *ModalManager) SetDimensions(width, height int) {
	mm.width = width
	mm.height = height
}

// BaseModal provides common modal functionality.
// AIDEV-NOTE: modal-base; shared state management and lifecycle for all modal implementations
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

// GetState returns the current modal state.
func (bm *BaseModal) GetState() ModalState {
	return bm.state
}

// SetState sets the modal state.
func (bm *BaseModal) SetState(state ModalState) {
	bm.state = state
}
