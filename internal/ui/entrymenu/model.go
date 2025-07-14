// Package entrymenu provides an interactive entry menu interface for habit tracking.
// AIDEV-NOTE: entry-menu-package; combines goal browsing with direct entry collection
package entrymenu

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"davidlee/vice/internal/models"
	"davidlee/vice/internal/ui"
)

// EntryMenuItem represents a goal as a menu item for entry collection.
// AIDEV-NOTE: entry-menu-item; extends GoalItem pattern with entry status tracking
//revive:disable-next-line:exported // descriptive name preferred over stuttering avoidance
type EntryMenuItem struct {
	Goal         models.Goal
	EntryStatus  models.EntryStatus
	HasEntry     bool
	Value        interface{}
	AchievementLevel *models.AchievementLevel
}

// FilterValue returns the value used for filtering this item.
func (e EntryMenuItem) FilterValue() string {
	return fmt.Sprintf("%s %s", e.Goal.Title, e.Goal.GoalType)
}

// Title returns the primary display text with status indicator.
func (e EntryMenuItem) Title() string {
	emoji := e.getGoalTypeEmoji()
	statusColor := e.getStatusColor()
	status := e.getStatusIndicator()
	
	titleStyle := lipgloss.NewStyle().Foreground(statusColor)
	return fmt.Sprintf("%s %s %s", emoji, titleStyle.Render(e.Goal.Title), status)
}

// Description returns the secondary display text.
func (e EntryMenuItem) Description() string {
	if e.Goal.Description == "" {
		return ""
	}
	return fmt.Sprintf("   %s", e.Goal.Description)
}

// getGoalTypeEmoji returns the emoji representing the goal type.
func (e EntryMenuItem) getGoalTypeEmoji() string {
	switch e.Goal.GoalType {
	case models.SimpleGoal:
		return "✅"
	case models.ElasticGoal:
		return "🎯"
	case models.InformationalGoal:
		return "📊"
	case models.ChecklistGoal:
		return "📝"
	default:
		return "❓"
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
		return lipgloss.Color("88")  // dark red - failed
	case models.EntrySkipped:
		return lipgloss.Color("240") // dark grey - skipped
	default:
		return lipgloss.Color("250") // light grey - incomplete
	}
}

// getStatusIndicator returns a visual indicator of entry status.
func (e EntryMenuItem) getStatusIndicator() string {
	if !e.HasEntry {
		return "○" // incomplete
	}
	
	switch e.EntryStatus {
	case models.EntryCompleted:
		return "●" // success
	case models.EntryFailed:
		return "✗" // failed
	case models.EntrySkipped:
		return "–" // skipped
	default:
		return "○" // incomplete
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
//revive:disable-next-line:exported // descriptive name preferred over stuttering avoidance
type EntryMenuKeyMap struct {
	// Navigation
	Up     key.Binding
	Down   key.Binding
	Select key.Binding

	// Entry menu specific
	ToggleReturnBehavior key.Binding
	FilterSkipped        key.Binding
	FilterPrevious       key.Binding

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

		// Entry menu specific
		ToggleReturnBehavior: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "toggle return behavior"),
		),
		FilterSkipped: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "filter skipped"),
		),
		FilterPrevious: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "filter previous"),
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
	viewRenderer   *ViewRenderer
	
	// Navigation state
	selectedGoalID string  // ID of goal selected for entry
	shouldQuit     bool    // Flag to quit the menu
}

// NewEntryMenuModel creates a new entry menu model with the provided goals and entries.
func NewEntryMenuModel(goals []models.Goal, entries map[string]models.GoalEntry, collector *ui.EntryCollector) *EntryMenuModel {
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
		return []key.Binding{keyMap.ToggleReturnBehavior, keyMap.FilterSkipped, keyMap.FilterPrevious}
	}

	return &EntryMenuModel{
		list:           l,
		goals:          goals,
		entries:        entries,
		keys:           keyMap,
		filterState:    FilterNone,
		returnBehavior: ReturnToMenu,
		entryCollector: collector,
		viewRenderer:   NewViewRenderer(0, 0), // Will be updated on first WindowSizeMsg
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
	}
}

// createMenuItems converts goals and entries into menu items.
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

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			m.shouldQuit = true
			return m, tea.Quit
		case key.Matches(msg, m.keys.Select):
			if len(m.goals) > 0 {
				selected := m.list.SelectedItem()
				if item, ok := selected.(EntryMenuItem); ok {
					m.selectedGoalID = item.Goal.ID
					// Entry collection will be handled by parent
					return m, nil
				}
			}
		case key.Matches(msg, m.keys.ToggleReturnBehavior):
			m.toggleReturnBehavior()
			return m, nil
		case key.Matches(msg, m.keys.FilterSkipped):
			m.toggleSkippedFilter()
			return m, nil
		case key.Matches(msg, m.keys.FilterPrevious):
			m.togglePreviousFilter()
			return m, nil
		}
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
	
	header := m.viewRenderer.RenderHeader(m.goals, m.entries, m.filterState, m.returnBehavior)
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		m.list.View(),
	)
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
	m.applyFilter()
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
	m.applyFilter()
}

// applyFilter applies the current filter state to the list.
func (m *EntryMenuModel) applyFilter() {
	allItems := createMenuItems(m.goals, m.entries)
	filteredItems := make([]list.Item, 0, len(allItems))
	
	for _, item := range allItems {
		menuItem := item.(EntryMenuItem)
		
		// Apply filter logic
		if m.shouldFilterOut(menuItem) {
			continue
		}
		
		filteredItems = append(filteredItems, item)
	}
	
	m.list.SetItems(filteredItems)
}

// shouldFilterOut determines if a menu item should be filtered out.
func (m *EntryMenuModel) shouldFilterOut(item EntryMenuItem) bool {
	hideSkipped := m.filterState == FilterHideSkipped || m.filterState == FilterHideSkippedAndPrevious
	hidePrevious := m.filterState == FilterHidePrevious || m.filterState == FilterHideSkippedAndPrevious
	
	if hideSkipped && item.HasEntry && item.EntryStatus == models.EntrySkipped {
		return true
	}
	
	if hidePrevious && item.HasEntry && (item.EntryStatus == models.EntryCompleted || item.EntryStatus == models.EntryFailed) {
		return true
	}
	
	return false
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