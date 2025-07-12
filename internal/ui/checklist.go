package ui

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
			"# process inboxes",
			"process email inbox",
			"review email starred",
			"check slack saved messages",
			"check slack notifications",
			"check slack priority channels",
			"review calendar",

			"# plan",
			"time block",
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
		// skip over headings
		case "up", "e":
			if m.cursor > 0 {
				if strings.HasPrefix(m.items[m.cursor-1], "#") {
					if m.cursor > 1 {
						m.cursor -= 2
					}
				} else {
					m.cursor--
				}
			}

		// The "down" and "a" keys move the cursor down
		// skip over headings
		case "down", "a":
			if m.cursor < len(m.items)-1 {
				m.cursor++
				if strings.HasPrefix(m.items[m.cursor], "# ") {
					m.cursor++
				}
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
	s := "Complete the checklist:\n\n"

	headingStyle := lipgloss.NewStyle().Bold(true).Underline(true).Foreground(lipgloss.Color("202")) //.Padding(0).Margin(0)
	itemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))                                 //.Padding(0).Margin(0)
	checkedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#3C3C3C"))                        //.Padding(0).Margin(0)
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("201"))                           //.Padding(0).Margin(0)

	// Iterate over our items
	for i, item := range m.items {
		isHeading := strings.HasPrefix(item, "# ")

		// Is the cursor pointing at this item?
		cursor := " " // no cursor
		if m.cursor == i {
			if !isHeading {
				cursor = ">" // cursor!
			}
		}

		// Is this choice selected?
		checked := " " // not selected
		if _, ok := m.selected[i]; ok {
			checked = "x" // selected!
		}

		// Render the row
		if isHeading {
			if i > 0 {
				s += fmt.Sprintf("\n")
			}
			s += fmt.Sprintf("      ")
			text := fmt.Sprintf("%s", strings.TrimLeft(item, "# "))
			s += headingStyle.Render(text)
			s += fmt.Sprintf("\n")
		} else {
			if cursor == ">" {
				text := fmt.Sprintf("%s [%s] %s", cursor, checked, item)
				s += selectedStyle.Render(text)
			} else if checked == "x" {
				text := fmt.Sprintf("%s [%s] %s", cursor, checked, item)
				s += checkedStyle.Render(text)
			} else {
				text := fmt.Sprintf("%s [%s] %s", cursor, checked, item)
				s += itemStyle.Render(text)
			}
			s += "\n"
		}
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
