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

// TODO: allow checklists to be saved in config data and dynamically displayed:
//
//	iter list add $id
//	iter list edit $id
//	iter list entry // select from menu
//	iter list entry $id
//
// - add list completion as a new goal type
// - can be automatically scored (when all complete)
// - or manually scored

func initialModel() model {
	newModel := model{
		items: []string{
			// "# this is a heading, used to visually group checklist items"
			// "this is a checklist item"

			"# clean station: physical inputs (~5m)",
			"clear desk",
			"clear desk inbox, loose papers, notebook",

			"# clean station: digital inputs (~10m)",
			"process emails (inbox)",
			"phone notifications",
			"browsers (all devices)",
			"editors, apps",
			"review periodic notes",
			"log actions",

			"# straighten & reset (~5m)",
			"desk",
			"digital workspace",

			"# sharpen tools (~5m)",
			"sweep calendar (yesterday)",
			"categorise / prioritise actions",

			"# plan the day (~10m)",
			"list actions (scheduled, wanted)",
			"identify immersive vs process actions",
			"batch process actions",
			"estimate hours free for new actions",
			"schedule / time block",

			"# gather resources",
			"set up for action",
		},
		selected: make(map[int]struct{}),
	}

	// set cursor to index of first non-heading
	for strings.HasPrefix(newModel.items[newModel.cursor], "#") {
		newModel.cursor++
	}
	return newModel
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
				m.cursor--
				for m.cursor > 0 && strings.HasPrefix(m.items[m.cursor], "#") {
					m.cursor--
				}
				// handle case where first (n) item(s) is a heading
				for strings.HasPrefix(m.items[m.cursor], "#") {
					m.cursor++
				}
			}

		// The "down" and "a" keys move the cursor down
		// skip over headings
		case "down", "a":
			if m.cursor < len(m.items)-1 {
				m.cursor++
				for m.cursor < len(m.items)-1 && strings.HasPrefix(m.items[m.cursor], "# ") {
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

	// Styles
	headerStyle := lipgloss.NewStyle().Bold(true).Underline(true).Foreground(lipgloss.Color("63"))
	headingStyle := lipgloss.NewStyle().Bold(true).Underline(true).Foreground(lipgloss.Color("202"))
	itemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	checkedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#3C3C3C"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("201"))

	// The header
	s := headerStyle.Render("Complete the checklist:") + "\n\n"

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
		// TODO: append (12/15) to headings showing the count of checked/total items in the group
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
