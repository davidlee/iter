package goalconfig

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"davidlee/iter/internal/models"
)

// GoalItem represents a goal as a list item for the bubbles/list component.
type GoalItem struct {
	Goal models.Goal
}

// FilterValue returns the value used for filtering this item.
// We filter on title and goal type for better search experience.
func (g GoalItem) FilterValue() string {
	return fmt.Sprintf("%s %s", g.Goal.Title, g.Goal.GoalType)
}

// Title returns the primary display text for the list item.
// Format: "emoji title" for clean visual grouping.
func (g GoalItem) Title() string {
	emoji := g.getGoalTypeEmoji()
	return fmt.Sprintf("%s %s", emoji, g.Goal.Title)
}

// Description returns the secondary display text for the list item.
// Format: "   description" with spacing to align with title text after emoji.
func (g GoalItem) Description() string {
	if g.Goal.Description == "" {
		return ""
	}
	return fmt.Sprintf("   %s", g.Goal.Description)
}

// getGoalTypeEmoji returns the emoji representing the goal type.
func (g GoalItem) getGoalTypeEmoji() string {
	switch g.Goal.GoalType {
	case models.SimpleGoal:
		return "✅" // Simple boolean goals
	case models.ElasticGoal:
		return "🎯" // Multi-tier achievement goals
	case models.InformationalGoal:
		return "📊" // Data collection goals
	case models.ChecklistGoal:
		return "📝" // Checklist completion goals
	default:
		return "❓" // Unknown goal type
	}
}

// AIDEV-NOTE: keybinding-architecture; centralized key management enables user configurability
// GoalListKeyMap defines the keybindings for the goal list interface.
// This struct enables dynamic keybinding configuration and consistent help text generation.
type GoalListKeyMap struct {
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

// DefaultGoalListKeyMap returns the default keybindings for the goal list.
func DefaultGoalListKeyMap() GoalListKeyMap {
	return GoalListKeyMap{
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
			key.WithHelp("e", "edit goal"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete goal"),
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
func (k GoalListKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.ShowDetail, k.Quit}
}

// FullHelp returns the full help for the keybindings.
func (k GoalListKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.ShowDetail, k.CloseModal},
		{k.Edit, k.Delete, k.Search, k.Quit},
	}
}

// GoalListModel represents the state of the goal list UI.
type GoalListModel struct {
	list                  list.Model
	goals                 []models.Goal
	width                 int
	height                int
	showModal             bool
	keys                  GoalListKeyMap
	selectedGoalForEdit   string // ID of goal selected for editing (triggers quit)
	selectedGoalForDelete string // ID of goal selected for deletion (triggers quit)
}

// NewGoalListModel creates a new goal list model with the provided goals.
func NewGoalListModel(goals []models.Goal) *GoalListModel {
	// Convert goals to list items
	items := make([]list.Item, len(goals))
	for i, goal := range goals {
		items[i] = GoalItem{Goal: goal}
	}

	// Create list with default delegate for clean vertical layout
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Goals"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	// AIDEV-NOTE: help-integration; AdditionalShortHelpKeys integrates custom keys with bubbles/list help
	// Set additional keybindings for the list help
	keyMap := DefaultGoalListKeyMap()
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{keyMap.ShowDetail}
	}

	// AIDEV-NOTE: quit-and-return-pattern; Phase 3 edit/delete operations use this pattern
	// Operations set selectedGoalFor* fields and quit, parent handles action, returns to refreshed list

	return &GoalListModel{
		list:  l,
		goals: goals,
		keys:  DefaultGoalListKeyMap(),
	}
}

// WithKeyMap allows customizing the keybindings for the goal list.
// This enables future user configurability of key mappings.
func (m *GoalListModel) WithKeyMap(keyMap GoalListKeyMap) *GoalListModel {
	m.keys = keyMap
	return m
}

// Init implements the tea.Model interface.
func (m *GoalListModel) Init() tea.Cmd {
	return nil
}

// Update implements the tea.Model interface.
func (m *GoalListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			switch {
			case key.Matches(msg, m.keys.CloseModal):
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
			if len(m.goals) > 0 {
				m.showModal = true
				return m, nil
			}
		case key.Matches(msg, m.keys.Edit):
			if len(m.goals) > 0 {
				selectedGoal := m.getSelectedGoal()
				if selectedGoal != nil {
					m.selectedGoalForEdit = selectedGoal.ID
					return m, tea.Quit // Exit to trigger edit
				}
			}
			return m, nil
		case key.Matches(msg, m.keys.Delete):
			if len(m.goals) > 0 {
				selectedGoal := m.getSelectedGoal()
				if selectedGoal != nil {
					m.selectedGoalForDelete = selectedGoal.ID
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
func (m *GoalListModel) View() string {
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
	itemCount := fmt.Sprintf("\n%d goals", len(m.goals))

	// Legend
	legend := "\n" + legendStyle.Render(
		"✅ Simple  🎯 Elastic  📊 Info  📝 Checklist",
	)

	return listView + itemCount + legend
}

// getSelectedGoal returns the currently selected goal.
func (m *GoalListModel) getSelectedGoal() *models.Goal {
	if len(m.goals) == 0 {
		return nil
	}

	selectedIndex := m.list.Index()
	if selectedIndex < 0 || selectedIndex >= len(m.goals) {
		return nil
	}

	return &m.goals[selectedIndex]
}

// renderModalView renders the goal detail modal overlay.
func (m *GoalListModel) renderModalView() string {
	goal := m.getSelectedGoal()
	if goal == nil {
		return "No goal selected"
	}

	// Modal content
	content := m.renderGoalDetails(goal)

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

	// Center the modal
	x := (m.width - modalWidth) / 2
	y := (m.height - modalHeight) / 2
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}

	// Overlay on background with centered placement
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modal, lipgloss.WithWhitespaceChars(" "), lipgloss.WithWhitespaceForeground(lipgloss.Color("8")))
}

// renderGoalDetails renders the detailed goal information for the modal.
func (m *GoalListModel) renderGoalDetails(goal *models.Goal) string {
	var details []string

	// Title with emoji
	emoji := getGoalTypeEmojiForGoal(goal.GoalType)
	title := modalTitleStyle.Render(fmt.Sprintf("%s %s", emoji, goal.Title))
	details = append(details, title)

	if goal.Description != "" {
		details = append(details, "", modalFieldStyle.Render("Description:"))
		details = append(details, goal.Description)
	}

	// Goal details
	details = append(details, "", modalFieldStyle.Render("Details:"))
	details = append(details, fmt.Sprintf("ID: %s", goal.ID))
	details = append(details, fmt.Sprintf("Type: %s", goal.GoalType))
	details = append(details, fmt.Sprintf("Field: %s", goal.FieldType.Type))

	if goal.ScoringType != "" {
		details = append(details, fmt.Sprintf("Scoring: %s", goal.ScoringType))
	}

	// Goal-type specific details
	switch goal.GoalType {
	case models.ElasticGoal:
		if goal.MiniCriteria != nil || goal.MidiCriteria != nil || goal.MaxiCriteria != nil {
			details = append(details, "", modalFieldStyle.Render("Achievement Levels:"))
			if goal.MiniCriteria != nil {
				details = append(details, fmt.Sprintf("Mini: %s", renderCriteria(goal.MiniCriteria)))
			}
			if goal.MidiCriteria != nil {
				details = append(details, fmt.Sprintf("Midi: %s", renderCriteria(goal.MidiCriteria)))
			}
			if goal.MaxiCriteria != nil {
				details = append(details, fmt.Sprintf("Maxi: %s", renderCriteria(goal.MaxiCriteria)))
			}
		}
	case models.InformationalGoal:
		if goal.Direction != "" {
			details = append(details, fmt.Sprintf("Direction: %s", goal.Direction))
		}
	}

	if goal.Criteria != nil {
		details = append(details, "", modalFieldStyle.Render("Criteria:"))
		details = append(details, renderCriteria(goal.Criteria))
	}

	// UI prompts
	if goal.Prompt != "" {
		details = append(details, "", modalFieldStyle.Render("Prompt:"))
		details = append(details, goal.Prompt)
	}

	if goal.HelpText != "" {
		details = append(details, "", modalFieldStyle.Render("Help:"))
		details = append(details, goal.HelpText)
	}

	// AIDEV-NOTE: dynamic-help-text; avoid hardcoded keys, use Help().Key for configurability
	// Footer with dynamic keybinding help
	closeHelp := m.keys.CloseModal.Help()
	footerText := fmt.Sprintf("Press %s to close", closeHelp.Key)
	details = append(details, "", modalFooterStyle.Render(footerText))

	return strings.Join(details, "\n")
}

// GetSelectedGoalForEdit returns the ID of the goal selected for editing (if any)
func (m *GoalListModel) GetSelectedGoalForEdit() string {
	return m.selectedGoalForEdit
}

// GetSelectedGoalForDelete returns the ID of the goal selected for deletion (if any)
func (m *GoalListModel) GetSelectedGoalForDelete() string {
	return m.selectedGoalForDelete
}

// getGoalTypeEmojiForGoal returns emoji for goal type (helper for modal).
func getGoalTypeEmojiForGoal(goalType models.GoalType) string {
	switch goalType {
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
