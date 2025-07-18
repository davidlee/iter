package habitconfig

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/davidlee/vice/internal/models"
)

// HabitItem represents a habit as a list item for the bubbles/list component.
type HabitItem struct {
	Habit models.Habit
}

// FilterValue returns the value used for filtering this item.
// We filter on title and habit type for better search experience.
func (g HabitItem) FilterValue() string {
	return fmt.Sprintf("%s %s", g.Habit.Title, g.Habit.HabitType)
}

// Title returns the primary display text for the list item.
// Format: "emoji title" for clean visual grouping.
func (g HabitItem) Title() string {
	emoji := g.getHabitTypeEmoji()
	return fmt.Sprintf("%s %s", emoji, g.Habit.Title)
}

// Description returns the secondary display text for the list item.
// Format: "   description" with spacing to align with title text after emoji.
func (g HabitItem) Description() string {
	if g.Habit.Description == "" {
		return ""
	}
	return fmt.Sprintf("   %s", g.Habit.Description)
}

// getHabitTypeEmoji returns the emoji representing the habit type.
func (g HabitItem) getHabitTypeEmoji() string {
	switch g.Habit.HabitType {
	case models.SimpleHabit:
		return "✅" // Simple boolean habits
	case models.ElasticHabit:
		return "🎯" // Multi-tier achievement habits
	case models.InformationalHabit:
		return "📊" // Data collection habits
	case models.ChecklistHabit:
		return "📝" // Checklist completion habits
	default:
		return "❓" // Unknown habit type
	}
}

// HabitListKeyMap defines the keybindings for the habit list interface.
// AIDEV-NOTE: keybinding-architecture; centralized key management enables user configurability
// This struct enables dynamic keybinding configuration and consistent help text generation.
type HabitListKeyMap struct {
	// Navigation
	Up     key.Binding
	Down   key.Binding
	Select key.Binding

	// Modal actions
	ShowDetail key.Binding
	CloseModal key.Binding

	// Future operations (prepared but not yet implemented)
	Edit   key.Binding
	Delete key.Binding
	Search key.Binding

	// Exit
	Quit key.Binding
}

// DefaultHabitListKeyMap returns the default keybindings for the habit list.
func DefaultHabitListKeyMap() HabitListKeyMap {
	return HabitListKeyMap{
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
			key.WithHelp("enter", "select"),
		),

		// Modal actions
		ShowDetail: key.NewBinding(
			key.WithKeys("enter", " "),
			key.WithHelp("enter/space", "show details"),
		),
		CloseModal: key.NewBinding(
			key.WithKeys("esc", "q"),
			key.WithHelp("esc/q", "close"),
		),

		// Future operations
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit habit"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete habit"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),

		// Exit
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}

// ShortHelp returns the short help for the keybindings.
func (k HabitListKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.ShowDetail, k.Edit, k.Delete, k.Quit}
}

// FullHelp returns the full help for the keybindings.
func (k HabitListKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.ShowDetail, k.CloseModal},
		{k.Edit, k.Delete, k.Search, k.Quit},
	}
}

// HabitListModel represents the state of the habit list UI.
type HabitListModel struct {
	list                   list.Model
	habits                 []models.Habit
	width                  int
	height                 int
	showModal              bool
	keys                   HabitListKeyMap
	selectedHabitForEdit   string // ID of habit selected for editing (triggers quit)
	selectedHabitForDelete string // ID of habit selected for deletion (triggers quit)
}

// NewHabitListModel creates a new habit list model with the provided habits.
func NewHabitListModel(habits []models.Habit) *HabitListModel {
	// Convert habits to list items
	items := make([]list.Item, len(habits))
	for i, habit := range habits {
		items[i] = HabitItem{Habit: habit}
	}

	// Create list with default delegate for clean vertical layout
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Habits"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	// AIDEV-NOTE: help-integration; AdditionalShortHelpKeys integrates custom keys with bubbles/list help
	// Set additional keybindings for the list help
	keyMap := DefaultHabitListKeyMap()
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{keyMap.ShowDetail, keyMap.Edit, keyMap.Delete}
	}

	// AIDEV-NOTE: quit-and-return-pattern; Phase 3 edit/delete operations use this pattern
	// Operations set selectedHabitFor* fields and quit, parent handles action, returns to refreshed list

	return &HabitListModel{
		list:   l,
		habits: habits,
		keys:   DefaultHabitListKeyMap(),
	}
}

// WithKeyMap allows customizing the keybindings for the habit list.
// This enables future user configurability of key mappings.
func (m *HabitListModel) WithKeyMap(keyMap HabitListKeyMap) *HabitListModel {
	m.keys = keyMap
	return m
}

// Init implements the tea.Model interface.
func (m *HabitListModel) Init() tea.Cmd {
	return nil
}

// Update implements the tea.Model interface.
func (m *HabitListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 2) // Account for margins
	case tea.KeyMsg:
		// AIDEV-NOTE: modal-key-isolation; modal keys processed first to prevent interference
		// Handle modal-specific keys first
		if m.showModal {
			if key.Matches(msg, m.keys.CloseModal) {
				m.showModal = false
				return m, nil
			}
			// In modal mode, don't process other keys
			return m, nil
		}

		// Handle main interface keys
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.ShowDetail):
			if len(m.habits) > 0 {
				m.showModal = true
				return m, nil
			}
		case key.Matches(msg, m.keys.Edit):
			if len(m.habits) > 0 {
				selectedHabit := m.getSelectedHabit()
				if selectedHabit != nil {
					m.selectedHabitForEdit = selectedHabit.ID
					return m, tea.Quit // Exit to trigger edit
				}
			}
			return m, nil
		case key.Matches(msg, m.keys.Delete):
			if len(m.habits) > 0 {
				selectedHabit := m.getSelectedHabit()
				if selectedHabit != nil {
					m.selectedHabitForDelete = selectedHabit.ID
					return m, tea.Quit // Exit to trigger delete
				}
			}
			return m, nil
		case key.Matches(msg, m.keys.Search):
			// TODO: Phase 4.1 - Implement search functionality
			return m, nil
		}
	}

	// Only update list when not showing modal
	if !m.showModal {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	return m, nil
}

// View implements the tea.Model interface.
func (m *HabitListModel) View() string {
	if m.width == 0 {
		return "Initializing..."
	}

	// Show modal overlay if active
	if m.showModal {
		return m.renderModalView()
	}

	// Main list view
	listView := m.list.View()

	// Item count
	itemCount := fmt.Sprintf("\n%d habits", len(m.habits))

	// Legend
	legend := "\n" + legendStyle.Render(
		"✅ Simple  🎯 Elastic  📊 Info  📝 Checklist",
	)

	return listView + itemCount + legend
}

// getSelectedHabit returns the currently selected habit.
func (m *HabitListModel) getSelectedHabit() *models.Habit {
	if len(m.habits) == 0 {
		return nil
	}

	selectedIndex := m.list.Index()
	if selectedIndex < 0 || selectedIndex >= len(m.habits) {
		return nil
	}

	return &m.habits[selectedIndex]
}

// renderModalView renders the habit detail modal overlay.
func (m *HabitListModel) renderModalView() string {
	habit := m.getSelectedHabit()
	if habit == nil {
		return "No habit selected"
	}

	// Modal content
	content := m.renderHabitDetails(habit)

	// Modal dimensions (centered)
	modalWidth := 80
	modalHeight := 20
	if m.width > 0 && m.width < modalWidth {
		modalWidth = m.width - 4
	}

	// Create modal box with border
	modal := modalStyle.
		Width(modalWidth).
		Height(modalHeight).
		Render(content)

	// Overlay on background with centered placement
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modal, lipgloss.WithWhitespaceChars(" "), lipgloss.WithWhitespaceForeground(lipgloss.Color("8")))
}

// renderHabitDetails renders the detailed habit information for the modal.
func (m *HabitListModel) renderHabitDetails(habit *models.Habit) string {
	var details []string

	// Title with emoji
	emoji := getHabitTypeEmojiForHabit(habit.HabitType)
	title := modalTitleStyle.Render(fmt.Sprintf("%s %s", emoji, habit.Title))
	details = append(details, title)

	if habit.Description != "" {
		details = append(details, "", modalFieldStyle.Render("Description:"))
		details = append(details, habit.Description)
	}

	// Habit details
	details = append(details, "", modalFieldStyle.Render("Details:"))
	details = append(details, fmt.Sprintf("ID: %s", habit.ID))
	details = append(details, fmt.Sprintf("Type: %s", habit.HabitType))
	details = append(details, fmt.Sprintf("Field: %s", habit.FieldType.Type))

	if habit.ScoringType != "" {
		details = append(details, fmt.Sprintf("Scoring: %s", habit.ScoringType))
	}

	// Habit-type specific details
	switch habit.HabitType {
	case models.ElasticHabit:
		if habit.MiniCriteria != nil || habit.MidiCriteria != nil || habit.MaxiCriteria != nil {
			details = append(details, "", modalFieldStyle.Render("Achievement Levels:"))
			if habit.MiniCriteria != nil {
				details = append(details, fmt.Sprintf("Mini: %s", renderCriteria(habit.MiniCriteria)))
			}
			if habit.MidiCriteria != nil {
				details = append(details, fmt.Sprintf("Midi: %s", renderCriteria(habit.MidiCriteria)))
			}
			if habit.MaxiCriteria != nil {
				details = append(details, fmt.Sprintf("Maxi: %s", renderCriteria(habit.MaxiCriteria)))
			}
		}
	case models.InformationalHabit:
		if habit.Direction != "" {
			details = append(details, fmt.Sprintf("Direction: %s", habit.Direction))
		}
	}

	if habit.Criteria != nil {
		details = append(details, "", modalFieldStyle.Render("Criteria:"))
		details = append(details, renderCriteria(habit.Criteria))
	}

	// UI prompts
	if habit.Prompt != "" {
		details = append(details, "", modalFieldStyle.Render("Prompt:"))
		details = append(details, habit.Prompt)
	}

	if habit.HelpText != "" {
		details = append(details, "", modalFieldStyle.Render("Help:"))
		details = append(details, habit.HelpText)
	}

	// AIDEV-NOTE: dynamic-help-text; avoid hardcoded keys, use Help().Key for configurability
	// Footer with dynamic keybinding help
	closeHelp := m.keys.CloseModal.Help()
	footerText := fmt.Sprintf("Press %s to close", closeHelp.Key)
	details = append(details, "", modalFooterStyle.Render(footerText))

	return strings.Join(details, "\n")
}

// GetSelectedHabitForEdit returns the ID of the habit selected for editing (if any)
func (m *HabitListModel) GetSelectedHabitForEdit() string {
	return m.selectedHabitForEdit
}

// GetSelectedHabitForDelete returns the ID of the habit selected for deletion (if any)
func (m *HabitListModel) GetSelectedHabitForDelete() string {
	return m.selectedHabitForDelete
}

// getHabitTypeEmojiForHabit returns emoji for habit type (helper for modal).
func getHabitTypeEmojiForHabit(habitType models.HabitType) string {
	switch habitType {
	case models.SimpleHabit:
		return "✅"
	case models.ElasticHabit:
		return "🎯"
	case models.InformationalHabit:
		return "📊"
	case models.ChecklistHabit:
		return "📝"
	default:
		return "❓"
	}
}

// AIDEV-NOTE: criteria-rendering; comprehensive criteria display supporting all condition types
// renderCriteria renders criteria information for display.
func renderCriteria(criteria *models.Criteria) string {
	if criteria == nil {
		return "None"
	}

	var parts []string

	if criteria.Description != "" {
		parts = append(parts, criteria.Description)
	}

	if criteria.Condition != nil {
		condition := criteria.Condition
		var conditionParts []string

		if condition.GreaterThan != nil {
			conditionParts = append(conditionParts, fmt.Sprintf("> %.2f", *condition.GreaterThan))
		}
		if condition.GreaterThanOrEqual != nil {
			conditionParts = append(conditionParts, fmt.Sprintf(">= %.2f", *condition.GreaterThanOrEqual))
		}
		if condition.LessThan != nil {
			conditionParts = append(conditionParts, fmt.Sprintf("< %.2f", *condition.LessThan))
		}
		if condition.LessThanOrEqual != nil {
			conditionParts = append(conditionParts, fmt.Sprintf("<= %.2f", *condition.LessThanOrEqual))
		}
		if condition.Equals != nil {
			conditionParts = append(conditionParts, fmt.Sprintf("= %t", *condition.Equals))
		}
		if condition.Before != "" {
			conditionParts = append(conditionParts, fmt.Sprintf("before %s", condition.Before))
		}
		if condition.After != "" {
			conditionParts = append(conditionParts, fmt.Sprintf("after %s", condition.After))
		}

		if len(conditionParts) > 0 {
			parts = append(parts, strings.Join(conditionParts, " and "))
		}
	}

	if len(parts) == 0 {
		return "No conditions specified"
	}

	return strings.Join(parts, " - ")
}

// Styling definitions
var (
	titleStyle = lipgloss.NewStyle().
			MarginLeft(0).
			Bold(true).
			Foreground(lipgloss.Color("15")).  // White text
			Background(lipgloss.Color("205")). // Bright background
			Padding(0, 1)                      // Inverted styling for emphasis

	paginationStyle = list.DefaultStyles().PaginationStyle.
			PaddingLeft(4)

	helpStyle = list.DefaultStyles().HelpStyle.
			PaddingLeft(4).
			PaddingBottom(1)

	legendStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243")). // Muted text for legend
			PaddingLeft(4).
			Italic(true)

	// Modal styling
	modalStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")). // Purple border
			Background(lipgloss.Color("0")).        // Black background
			Padding(1, 2)

	modalTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")). // White text
			Background(lipgloss.Color("62")). // Purple background
			Padding(0, 1).
			MarginBottom(1)

	modalFieldStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("14")) // Cyan text

	modalFooterStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("243")). // Muted text
				Italic(true).
				MarginTop(1)
)
