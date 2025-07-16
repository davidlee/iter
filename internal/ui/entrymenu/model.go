// Package entrymenu provides an interactive entry menu interface for habit tracking.
// AIDEV-NOTE: entry-menu-package; combines goal browsing with direct entry collection
package entrymenu

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"davidlee/vice/internal/debug"
	"davidlee/vice/internal/models"
	"davidlee/vice/internal/ui"
	"davidlee/vice/internal/ui/entry"
	"davidlee/vice/internal/ui/modal"
)

// EntryMenuItem represents a goal as a menu item for entry collection.
// AIDEV-NOTE: entry-menu-item; extends GoalItem pattern with entry status tracking
//
//revive:disable-next-line:exported // descriptive name preferred over stuttering avoidance
type EntryMenuItem struct {
	Goal             models.Goal
	EntryStatus      models.EntryStatus
	HasEntry         bool
	Value            interface{}
	AchievementLevel *models.AchievementLevel
}

// FilterValue returns the value used for filtering this item.
func (e EntryMenuItem) FilterValue() string {
	return fmt.Sprintf("%s %s", e.Goal.Title, e.Goal.GoalType)
}

// Title returns the primary display text with status indicator.
func (e EntryMenuItem) Title() string {
	emoji := e.getGoalStatusEmoji()
	statusColor := e.getStatusColor()

	titleStyle := lipgloss.NewStyle().Foreground(statusColor)
	return fmt.Sprintf("%s %s", emoji, titleStyle.Render(e.Goal.Title))
}

// Description returns the secondary display text.
func (e EntryMenuItem) Description() string {
	if e.Goal.Description == "" {
		return ""
	}
	return fmt.Sprintf("   %s", e.Goal.Description)
}

// getGoalStatusEmoji returns the emoji representing the goal's entry status.
// AIDEV-NOTE: status-emoji-design; T018 user-requested change from goal type to status emojis
func (e EntryMenuItem) getGoalStatusEmoji() string {
	if !e.HasEntry {
		return "☐" // incomplete - empty box
	}

	switch e.EntryStatus {
	case models.EntryCompleted:
		return "✓" // completed - checkmark
	case models.EntryFailed:
		return "✗" // failed - red cross
	case models.EntrySkipped:
		return "~" // skipped - tilde
	default:
		return "☐" // incomplete
	}
}

// getStatusColor returns the color for the goal based on entry status.
func (e EntryMenuItem) getStatusColor() lipgloss.Color {
	if !e.HasEntry {
		return lipgloss.Color("250") // light grey - incomplete
	}

	switch e.EntryStatus {
	case models.EntryCompleted:
		return lipgloss.Color("214") // gold - success
	case models.EntryFailed:
		return lipgloss.Color("88") // dark red - failed
	case models.EntrySkipped:
		return lipgloss.Color("240") // dark grey - skipped
	default:
		return lipgloss.Color("250") // light grey - incomplete
	}
}

// FilterState represents the current filtering state of the menu.
type FilterState int

// Filter states for controlling menu display.
const (
	FilterNone FilterState = iota
	FilterHideSkipped
	FilterHidePrevious
	FilterHideSkippedAndPrevious
)

// ReturnBehavior represents how the menu should behave after entry completion.
type ReturnBehavior int

// Return behaviors for post-entry navigation.
const (
	ReturnToMenu ReturnBehavior = iota
	ReturnToNextGoal
)

// EntryMenuKeyMap defines the keybindings for the entry menu interface.
// AIDEV-NOTE: entry-menu-keybinding; extends GoalListKeyMap with entry-specific actions
//
//revive:disable-next-line:exported // descriptive name preferred over stuttering avoidance
type EntryMenuKeyMap struct {
	// Navigation
	Up     key.Binding
	Down   key.Binding
	Select key.Binding

	// Smart navigation
	NextIncomplete     key.Binding
	PreviousIncomplete key.Binding

	// Entry menu specific
	ToggleReturnBehavior key.Binding
	FilterSkipped        key.Binding
	FilterPrevious       key.Binding
	ClearFilters         key.Binding

	// Exit
	Quit key.Binding
}

// DefaultEntryMenuKeyMap returns the default keybindings for the entry menu.
func DefaultEntryMenuKeyMap() EntryMenuKeyMap {
	return EntryMenuKeyMap{
		// Navigation - vim-style + arrow keys
		Up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("↓/j", "down"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "enter goal"),
		),

		// Smart navigation
		NextIncomplete: key.NewBinding(
			key.WithKeys("n", "tab"),
			key.WithHelp("n/tab", "next incomplete"),
		),
		PreviousIncomplete: key.NewBinding(
			key.WithKeys("N", "shift+tab"),
			key.WithHelp("N/shift+tab", "prev incomplete"),
		),

		// Entry menu specific
		ToggleReturnBehavior: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "toggle return"),
		),
		FilterSkipped: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "toggle skip filter"),
		),
		FilterPrevious: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "toggle prev filter"),
		),
		ClearFilters: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "clear filters"),
		),

		// Exit
		Quit: key.NewBinding(
			key.WithKeys("q", "esc", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}

// EntryMenuModel represents the state of the entry menu interface.
// AIDEV-NOTE: entry-menu-model; adapts GoalListModel patterns for entry workflow
//
//revive:disable-next-line:exported // descriptive name preferred over stuttering avoidance
type EntryMenuModel struct {
	list           list.Model
	goals          []models.Goal
	entries        map[string]models.GoalEntry
	keys           EntryMenuKeyMap
	width          int
	height         int
	filterState    FilterState
	returnBehavior ReturnBehavior
	entryCollector *ui.EntryCollector
	entriesFile    string // Path to entries file for auto-save
	viewRenderer   *ViewRenderer
	navEnhancer    *NavigationEnhancer

	// Modal system for entry editing
	// modalManager      *modal.ModalManager  // TEMPORARILY REMOVED for ModalManager experiment
	directModal       modal.Modal // Direct modal handling like prototype
	fieldInputFactory *entry.EntryFieldInputFactory

	// Navigation state
	selectedGoalID string // ID of goal selected for entry
	shouldQuit     bool   // Flag to quit the menu
}

// NewEntryMenuModel creates a new entry menu model with the provided goals and entries.
func NewEntryMenuModel(goals []models.Goal, entries map[string]models.GoalEntry, collector *ui.EntryCollector, entriesFile string) *EntryMenuModel {
	items := createMenuItems(goals, entries)

	// Create list with default delegate
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Entry Menu"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	// Set additional keybindings for help
	keyMap := DefaultEntryMenuKeyMap()
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			keyMap.NextIncomplete, keyMap.ToggleReturnBehavior,
			keyMap.FilterSkipped, keyMap.FilterPrevious, keyMap.ClearFilters,
		}
	}

	return &EntryMenuModel{
		list:           l,
		goals:          goals,
		entries:        entries,
		keys:           keyMap,
		filterState:    FilterNone,
		returnBehavior: ReturnToMenu,
		entryCollector: collector,
		entriesFile:    entriesFile,
		viewRenderer:   NewViewRenderer(0, 0), // Will be updated on first WindowSizeMsg
		navEnhancer:    NewNavigationEnhancer(),
		// modalManager:      modal.NewModalManager(0, 0), // TEMPORARILY REMOVED for ModalManager experiment
		directModal:       nil, // Direct modal handling like prototype
		fieldInputFactory: entry.NewEntryFieldInputFactory(),
	}
}

// NewEntryMenuModelForTesting creates a headless entry menu model for testing.
func NewEntryMenuModelForTesting(goals []models.Goal, entries map[string]models.GoalEntry) *EntryMenuModel {
	items := createMenuItems(goals, entries)

	// Create minimal list for testing
	l := list.New(items, list.NewDefaultDelegate(), 80, 24)
	l.Title = "Entry Menu"

	return &EntryMenuModel{
		list:           l,
		goals:          goals,
		entries:        entries,
		keys:           DefaultEntryMenuKeyMap(),
		filterState:    FilterNone,
		returnBehavior: ReturnToMenu,
		viewRenderer:   NewViewRenderer(80, 24), // Fixed size for testing
		navEnhancer:    NewNavigationEnhancer(),
		// modalManager:      modal.NewModalManager(80, 24), // TEMPORARILY REMOVED for ModalManager experiment
		directModal:       nil, // Direct modal handling like prototype
		fieldInputFactory: entry.NewEntryFieldInputFactory(),
	}
}

// createMenuItems converts goals and entries into menu items.
// AIDEV-NOTE: T024-bug1-analysis; status display logic - check entry status mapping
func createMenuItems(goals []models.Goal, entries map[string]models.GoalEntry) []list.Item {
	items := make([]list.Item, len(goals))
	for i, goal := range goals {
		entry, hasEntry := entries[goal.ID]
		items[i] = EntryMenuItem{
			Goal:             goal,
			EntryStatus:      entry.Status,
			HasEntry:         hasEntry,
			Value:            entry.Value,
			AchievementLevel: entry.AchievementLevel,
		}
	}
	return items
}

// Init implements the tea.Model interface.
func (m *EntryMenuModel) Init() tea.Cmd {
	return nil
}

// Update implements the tea.Model interface.
func (m *EntryMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 4) // Account for progress bar and margins
		m.viewRenderer = NewViewRenderer(msg.Width, msg.Height)
		// m.modalManager = modal.NewModalManager(msg.Width, msg.Height)  // REMOVED for ModalManager experiment

	case modal.ModalOpenedMsg:
		// Modal opened - no action needed, just continue
		return m, nil

	case modal.ModalClosedMsg:
		// AIDEV-NOTE: T024-bug-fix; modal closed with result, sync menu state and auto-save
		// Handle modal result and update menu state
		debug.EntryMenu("Modal closed for goal %s, result: %v", m.selectedGoalID, msg.Result != nil)
		if result := msg.Result; result != nil {
			if entryResult, ok := result.(*entry.EntryResult); ok {
				debug.EntryMenu("Processing entry result for goal %s: value=%v, status=%v", m.selectedGoalID, entryResult.Value, entryResult.Status)

				// Store the entry result in the collector
				if m.entryCollector != nil {
					m.entryCollector.StoreEntryResult(m.selectedGoalID, entryResult)
				}

				// Update menu state after entry storage
				m.updateEntriesFromCollector()

				// Auto-save entries after collection
				if m.entriesFile != "" && m.entryCollector != nil {
					err := m.entryCollector.SaveEntriesToFile(m.entriesFile)
					if err != nil {
						debug.EntryMenu("Failed to save entries for goal %s: %v", m.selectedGoalID, err)
						// Log error but continue - could add error display later
						_ = err // TODO: Consider adding save error handling UI
					}
				}

				// Smart navigation based on return behavior preference
				if m.returnBehavior == ReturnToNextGoal {
					m.navEnhancer.SelectNextIncompleteGoal(m)
				}
			}
		} else {
			debug.EntryMenu("Modal closed for goal %s with no result (cancelled)", m.selectedGoalID)
		}
		return m, nil

	case DeferredStateSyncMsg:
		// AIDEV-NOTE: T024-fix; Handle deferred state synchronization to prevent modal auto-closing
		debug.EntryMenu("Received deferred state sync message for goal %s", msg.goalID)
		m.processDeferredStateSync(msg)
		return m, nil

	case tea.KeyMsg:
		// AIDEV-NOTE: T024-experiment; direct modal handling replacing ModalManager
		// Route key messages to modal if active (direct handling like prototype)
		if m.directModal != nil && m.directModal.IsOpen() {
			var cmd tea.Cmd
			m.directModal, cmd = m.directModal.Update(msg)

			// Check if modal was closed and sync state (simple cleanup)
			if m.directModal.IsClosed() {
				m.directModal = nil
				syncCmd := m.syncStateAfterEntry()
				return m, tea.Batch(cmd, syncCmd)
			}
			return m, cmd
		}

		switch {
		case key.Matches(msg, m.keys.Quit):
			m.shouldQuit = true
			return m, tea.Quit
		case key.Matches(msg, m.keys.Select):
			if len(m.goals) > 0 {
				selected := m.list.SelectedItem()
				if item, ok := selected.(EntryMenuItem); ok {
					m.selectedGoalID = item.Goal.ID

					// AIDEV-NOTE: T024-modal-integration; replaced form.Run() with modal system to eliminate looping
					// Launch entry form modal instead of direct collector call
					if m.entryCollector != nil {
						debug.EntryMenu("Opening modal for goal %s (type: %s, field: %s)", item.Goal.ID, item.Goal.GoalType, item.Goal.FieldType.Type)

						// Create entry form modal
						entryFormModal, err := modal.NewEntryFormModal(item.Goal, m.entryCollector, m.fieldInputFactory)
						if err != nil {
							debug.EntryMenu("Failed to create modal for goal %s: %v", item.Goal.ID, err)
							// Log error but continue - could add error display later
							_ = err // TODO: Consider adding error handling UI
							return m, nil
						}

						// AIDEV-NOTE: T024-experiment; direct modal opening replacing ModalManager
						// Open modal directly like prototype
						debug.EntryMenu("Opening modal directly for goal %s (bypassing ModalManager)", item.Goal.ID)
						m.directModal = entryFormModal
						cmd := entryFormModal.Init()
						return m, cmd
					}

					return m, nil
				}
			}
		case key.Matches(msg, m.keys.NextIncomplete):
			m.navEnhancer.SelectNextIncompleteGoal(m)
			return m, nil
		case key.Matches(msg, m.keys.PreviousIncomplete):
			m.navEnhancer.SelectPreviousIncompleteGoal(m)
			return m, nil
		case key.Matches(msg, m.keys.ToggleReturnBehavior):
			m.toggleReturnBehavior()
			return m, nil
		case key.Matches(msg, m.keys.FilterSkipped):
			m.toggleSkippedFilter()
			m.navEnhancer.UpdateListAfterFilterChange(m)
			return m, nil
		case key.Matches(msg, m.keys.FilterPrevious):
			m.togglePreviousFilter()
			m.navEnhancer.UpdateListAfterFilterChange(m)
			return m, nil
		case key.Matches(msg, m.keys.ClearFilters):
			m.clearAllFilters()
			m.navEnhancer.UpdateListAfterFilterChange(m)
			return m, nil
		}
	}

	// AIDEV-NOTE: T024-experiment; direct modal handling replacing ModalManager
	// Update modal if active (direct handling like prototype)
	if m.directModal != nil && m.directModal.IsOpen() {
		var cmd tea.Cmd
		m.directModal, cmd = m.directModal.Update(msg)

		// Check if modal was closed and sync state (simple cleanup)
		if m.directModal.IsClosed() {
			m.directModal = nil
			syncCmd := m.syncStateAfterEntry()
			return m, tea.Batch(cmd, syncCmd)
		}
		return m, cmd
	}

	// Update the list component
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View implements the tea.Model interface.
func (m *EntryMenuModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	header := m.viewRenderer.RenderHeader(m.goals, m.entries, m.filterState)
	m.list.Title = "Entry Menu"

	// Get list view with return behavior inserted before help
	listView := m.renderListWithFooter()

	baseView := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		listView,
	)

	// AIDEV-NOTE: T024-experiment; direct modal rendering replacing ModalManager
	// Render with modal overlay if modal is active (direct rendering like prototype)
	if m.directModal != nil && m.directModal.IsOpen() {
		return m.renderWithDirectModal(baseView, m.directModal.View())
	}

	return baseView
}

// syncStateAfterEntry handles state synchronization after modal closes using deferred command pattern
// AIDEV-NOTE: T024-fix; Uses BubbleTea command to defer state sync, preventing timing conflicts with modal closure
func (m *EntryMenuModel) syncStateAfterEntry() tea.Cmd {
	if m.directModal == nil {
		return nil
	}

	// Get the entry result from the modal
	if entryFormModal, ok := m.directModal.(*modal.EntryFormModal); ok {
		result := entryFormModal.GetEntryResult()
		if result != nil {
			debug.EntryMenu("Deferring state sync for goal %s: value=%v, status=%v", m.selectedGoalID, result.Value, result.Status)

			// Return command to defer state synchronization until next BubbleTea cycle
			// This prevents timing conflicts between modal closure and state updates
			return tea.Cmd(func() tea.Msg {
				return DeferredStateSyncMsg{
					goalID: m.selectedGoalID,
					result: result,
				}
			})
		} else {
			debug.EntryMenu("Modal closed for goal %s with no result (cancelled)", m.selectedGoalID)
		}
	}

	return nil
}

// DeferredStateSyncMsg carries entry result data for deferred state synchronization
type DeferredStateSyncMsg struct {
	goalID string
	result *entry.EntryResult
}

// processDeferredStateSync handles deferred state synchronization operations
// AIDEV-NOTE: T024-fix; Separated from modal closure to prevent auto-closing timing conflicts
func (m *EntryMenuModel) processDeferredStateSync(msg DeferredStateSyncMsg) {
	debug.EntryMenu("Processing deferred state sync for goal %s: value=%v, status=%v", msg.goalID, msg.result.Value, msg.result.Status)

	// Store the entry result in the collector
	if m.entryCollector != nil {
		debug.EntryMenu("Executing Entry Storage - StoreEntryResult")
		m.entryCollector.StoreEntryResult(msg.goalID, msg.result)
	}

	// Update menu state after entry storage
	debug.EntryMenu("Executing Menu Updates - updateEntriesFromCollector")
	m.updateEntriesFromCollector()

	// Auto-save entries after collection
	if m.entriesFile != "" && m.entryCollector != nil {
		debug.EntryMenu("Executing Auto-Save - SaveEntriesToFile")
		err := m.entryCollector.SaveEntriesToFile(m.entriesFile)
		if err != nil {
			debug.EntryMenu("Failed to save entries for goal %s: %v", msg.goalID, err)
			_ = err // TODO: Consider adding save error handling UI
		}
	}

	// Smart navigation based on return behavior preference
	if m.returnBehavior == ReturnToNextGoal {
		debug.EntryMenu("Executing Smart Navigation - SelectNextIncompleteGoal")
		m.navEnhancer.SelectNextIncompleteGoal(m)
	}
}

// renderWithDirectModal renders modal overlay directly (for ModalManager experiment)
func (m *EntryMenuModel) renderWithDirectModal(background, modalContent string) string {
	// Simple modal overlay implementation like ModalManager
	dimmedBg := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#888888", Dark: "#444444"}).
		Render(background)

	// Center the modal
	modalBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Padding(1, 2).
		Background(lipgloss.Color("#1a1a1a")).
		Render(modalContent)

	centeredModal := lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modalBox)

	// Overlay the modal on the background
	return lipgloss.JoinVertical(lipgloss.Left, dimmedBg, centeredModal)
}

// renderListWithFooter renders the list with return behavior inserted before help.
// AIDEV-NOTE: footer-layout; robust approach moving return behavior to footer above keybindings
func (m *EntryMenuModel) renderListWithFooter() string {
	// Temporarily disable list help to render it manually
	showHelp := m.list.ShowHelp()
	m.list.SetShowHelp(false)

	listContent := m.list.View()

	// Restore help setting
	m.list.SetShowHelp(showHelp)

	// Create return behavior line
	var returnText string
	switch m.returnBehavior {
	case ReturnToMenu:
		returnText = "Return: menu"
	case ReturnToNextGoal:
		returnText = "Return: next goal"
	default:
		returnText = "Return: menu"
	}

	returnLine := returnBehaviorStyle.Render(returnText)

	// Add help if it was enabled
	var parts []string
	parts = append(parts, listContent)
	parts = append(parts, returnLine)

	if showHelp {
		// Get the list's help text by temporarily restoring help and getting just that part
		m.list.SetShowHelp(true)
		fullView := m.list.View()
		m.list.SetShowHelp(false)

		// Extract help text from the bottom of the full view
		lines := strings.Split(fullView, "\n")
		if len(lines) > 0 {
			helpLine := lines[len(lines)-1]
			if helpLine != "" {
				parts = append(parts, helpLine)
			}
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// ShouldQuit returns true if the menu should quit.
func (m *EntryMenuModel) ShouldQuit() bool {
	return m.shouldQuit
}

// SelectedGoalID returns the ID of the selected goal for entry, or empty string if none.
func (m *EntryMenuModel) SelectedGoalID() string {
	return m.selectedGoalID
}

// ClearSelection clears the selected goal ID.
func (m *EntryMenuModel) ClearSelection() {
	m.selectedGoalID = ""
}

// GetReturnBehavior returns the current return behavior setting.
func (m *EntryMenuModel) GetReturnBehavior() ReturnBehavior {
	return m.returnBehavior
}

// UpdateEntries updates the entries and refreshes the menu items.
func (m *EntryMenuModel) UpdateEntries(entries map[string]models.GoalEntry) {
	m.entries = entries
	items := createMenuItems(m.goals, entries)
	m.list.SetItems(items)
}

// updateEntriesFromCollector updates the entries map with data from the EntryCollector.
// AIDEV-NOTE: T018/3.1-state-sync; CRITICAL method for syncing collector state to menu after entry collection
// AIDEV-NOTE: T024-bug1-analysis; potential source of incorrect completion status display
// Handles type conversion from collector interface{} values to GoalEntry structs for menu display
// This is what makes the menu visual state update after user completes an entry
func (m *EntryMenuModel) updateEntriesFromCollector() {
	if m.entryCollector == nil {
		return
	}

	// Update entries for all goals based on collector state
	for _, goal := range m.goals {
		value, notes, achievement, status, hasEntry := m.entryCollector.GetGoalEntry(goal.ID)
		if hasEntry {
			// Convert to GoalEntry format
			goalEntry := models.GoalEntry{
				GoalID:           goal.ID,
				Status:           status,
				Notes:            notes,
				AchievementLevel: achievement,
				CreatedAt:        time.Now(),
			}

			// Set value based on type
			switch v := value.(type) {
			case string:
				goalEntry.Value = v
			case bool:
				if v {
					goalEntry.Value = "true"
				} else {
					goalEntry.Value = "false"
				}
			case time.Time:
				goalEntry.Value = v.Format("15:04")
			default:
				goalEntry.Value = fmt.Sprintf("%v", v)
			}

			m.entries[goal.ID] = goalEntry
		}
	}

	// Recreate menu items with updated entry data
	items := createMenuItems(m.goals, m.entries)
	m.list.SetItems(items)
}

// toggleReturnBehavior toggles between returning to menu and advancing to next goal.
func (m *EntryMenuModel) toggleReturnBehavior() {
	if m.returnBehavior == ReturnToMenu {
		m.returnBehavior = ReturnToNextGoal
	} else {
		m.returnBehavior = ReturnToMenu
	}
}

// toggleSkippedFilter toggles filtering of skipped goals.
func (m *EntryMenuModel) toggleSkippedFilter() {
	switch m.filterState {
	case FilterNone:
		m.filterState = FilterHideSkipped
	case FilterHideSkipped:
		m.filterState = FilterNone
	case FilterHidePrevious:
		m.filterState = FilterHideSkippedAndPrevious
	case FilterHideSkippedAndPrevious:
		m.filterState = FilterHidePrevious
	}
}

// togglePreviousFilter toggles filtering of previously entered goals.
func (m *EntryMenuModel) togglePreviousFilter() {
	switch m.filterState {
	case FilterNone:
		m.filterState = FilterHidePrevious
	case FilterHidePrevious:
		m.filterState = FilterNone
	case FilterHideSkipped:
		m.filterState = FilterHideSkippedAndPrevious
	case FilterHideSkippedAndPrevious:
		m.filterState = FilterHideSkipped
	}
}

// clearAllFilters clears all active filters.
func (m *EntryMenuModel) clearAllFilters() {
	m.filterState = FilterNone
}

// Styles for the entry menu interface.
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Background(lipgloss.Color("235")).
			Padding(0, 1).
			Bold(true)

	paginationStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))
)
