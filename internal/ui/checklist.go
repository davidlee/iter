package ui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	// "github.com/charmbracelet/lipgloss"
	// "davidlee/iter/internal/models"
)

type model struct {
	items    []string
	cursor   int
	selected map[int]struct{}
}

func initialModel() model {
	return model{
		items: []string{
			"check email",
			"check slack",
			"check calendar",
		},
		selected: make(map[int]struct{}),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

		// The "up" and "e" keys move the cursor up
		case "up", "e":
			if m.cursor > 0 {
				m.cursor--
			}

		// The "down" and "a" keys move the cursor down
		case "down", "a":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}

		// The "enter" key and the spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case "enter", " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m model) View() string {
	// The header
	s := "Complete checklist items:\n\n"

	// Iterate over our items
	for i, item := range m.items {

		// Is the cursor pointing at this item?
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor!
		}

		// Is this choice selected?
		checked := " " // not selected
		if _, ok := m.selected[i]; ok {
			checked = "x" // selected!
		}

		// Render the row
		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, item)
	}

	// The footer
	s += "\nPress q to quit.\n"

	// Send the UI for rendering
	return s
}

func NewChecklistScreen() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

//
// func UseStuff() {
// 	var words string
// 	var huhs := lipgloss.NewStyle()
//   var form := huh.NewForm()
//
// 	return nil
// }
