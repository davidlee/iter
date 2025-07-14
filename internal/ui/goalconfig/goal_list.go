package goalconfig

import (
	"fmt"
	"strings"

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

// GoalListModel represents the state of the goal list UI.
type GoalListModel struct {
	list      list.Model
	goals     []models.Goal
	width     int
	height    int
	showModal bool
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

	return &GoalListModel{
		list:  l,
		goals: goals,
	}
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
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc":
			if m.showModal {
				m.showModal = false
				return m, nil
			}
			return m, tea.Quit
		case "enter", " ":
			if !m.showModal && len(m.goals) > 0 {
				m.showModal = true
				return m, nil
			}
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

	// Footer
	details = append(details, "", modalFooterStyle.Render("Press ESC to close"))

	return strings.Join(details, "\n")
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
