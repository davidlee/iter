package goalconfig

import (
	"fmt"
	"io"
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
func (g GoalItem) Title() string {
	return g.Goal.Title
}

// Description returns the secondary display text for the list item.
// Format: "ID | Type | Status" for table-like appearance within list.
func (g GoalItem) Description() string {
	// AIDEV-NOTE: Custom delegate will handle full tabular formatting
	status := g.getGoalStatus()
	return fmt.Sprintf("%s | %s | %s", g.Goal.ID, g.Goal.GoalType, status)
}

// getGoalStatus determines the display status of a goal.
// For now, we show basic goal type status - future enhancement could include active/archived.
func (g GoalItem) getGoalStatus() string {
	switch g.Goal.GoalType {
	case models.SimpleGoal:
		if g.Goal.ScoringType != "" {
			return fmt.Sprintf("Simple (%s)", g.Goal.ScoringType)
		}
		return "Simple"
	case models.ElasticGoal:
		return "Elastic"
	case models.InformationalGoal:
		return "Info"
	case models.ChecklistGoal:
		return "Checklist"
	default:
		return "Unknown"
	}
}

// GoalListModel represents the state of the goal list UI.
type GoalListModel struct {
	list   list.Model
	goals  []models.Goal
	width  int
	height int
}

// NewGoalListModel creates a new goal list model with the provided goals.
func NewGoalListModel(goals []models.Goal) *GoalListModel {
	// Convert goals to list items
	items := make([]list.Item, len(goals))
	for i, goal := range goals {
		items[i] = GoalItem{Goal: goal}
	}

	// Create list with default styling
	l := list.New(items, NewGoalItemDelegate(), 0, 0)
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
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View implements the tea.Model interface.
func (m *GoalListModel) View() string {
	if m.width == 0 {
		return "Initializing..."
	}
	return m.list.View()
}

// GoalItemDelegate handles the rendering and interaction for individual goal items.
type GoalItemDelegate struct{}

// NewGoalItemDelegate creates a new goal item delegate.
func NewGoalItemDelegate() *GoalItemDelegate {
	return &GoalItemDelegate{}
}

// Height returns the height of a single item.
func (d *GoalItemDelegate) Height() int {
	return 1
}

// Spacing returns the spacing between items.
func (d *GoalItemDelegate) Spacing() int {
	return 0
}

// Update handles updates for individual items.
func (d *GoalItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
	return nil
}

// Render renders a single goal item with tabular formatting.
func (d *GoalItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	goalItem, ok := listItem.(GoalItem)
	if !ok {
		return
	}

	goal := goalItem.Goal
	isSelected := index == m.Index()

	// Define column widths for tabular appearance
	const (
		idWidth     = 12
		titleWidth  = 30
		typeWidth   = 12
		statusWidth = 15
	)

	// Format columns with fixed widths
	id := truncateOrPad(goal.ID, idWidth)
	title := truncateOrPad(goal.Title, titleWidth)
	goalType := truncateOrPad(string(goal.GoalType), typeWidth)
	status := truncateOrPad(goalItem.getGoalStatus(), statusWidth)

	// Combine columns
	line := fmt.Sprintf("%s %s %s %s", id, title, goalType, status)

	// Apply styling based on selection
	if isSelected {
		line = selectedItemStyle.Render(line)
	} else {
		line = itemStyle.Render(line)
	}

	_, _ = fmt.Fprint(w, line)
}

// truncateOrPad ensures text fits within specified width.
func truncateOrPad(text string, width int) string {
	if len(text) > width {
		if width > 3 {
			return text[:width-3] + "..."
		}
		return text[:width]
	}
	return text + strings.Repeat(" ", width-len(text))
}

// Styling definitions
var (
	titleStyle = lipgloss.NewStyle().
			MarginLeft(2).
			Bold(true).
			Foreground(lipgloss.Color("205"))

	itemStyle = lipgloss.NewStyle().
			PaddingLeft(4)

	selectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(lipgloss.Color("170")).
				Bold(true)

	paginationStyle = list.DefaultStyles().PaginationStyle.
			PaddingLeft(4)

	helpStyle = list.DefaultStyles().HelpStyle.
			PaddingLeft(4).
			PaddingBottom(1)
)
