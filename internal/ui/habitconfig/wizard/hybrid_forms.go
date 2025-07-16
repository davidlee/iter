package wizard

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// AIDEV-NOTE: Hybrid form components for embedding huh forms within bubbletea applications
// This provides the best of both worlds:
// - Rich bubbletea applications with navigation, progress tracking, and complex state
// - Simple, well-tested huh forms for individual steps
// Use this pattern when you need both the simplicity of huh and the power of bubbletea

// HybridFormModel wraps a huh form within a bubbletea model for embedding
type HybridFormModel struct {
	form         *huh.Form
	title        string
	description  string
	showProgress bool
	currentStep  int
	totalSteps   int
	width        int
	height       int
	complete     bool
	cancelled    bool
}

// NewHybridFormModel creates a new hybrid form model
func NewHybridFormModel(form *huh.Form, title, description string) *HybridFormModel {
	return &HybridFormModel{
		form:        form,
		title:       title,
		description: description,
		width:       80,
		height:      24,
	}
}

// WithProgress adds progress tracking to the hybrid form
func (m *HybridFormModel) WithProgress(current, total int) *HybridFormModel {
	m.showProgress = true
	m.currentStep = current
	m.totalSteps = total
	return m
}

// WithSize sets the display size for the hybrid form
func (m *HybridFormModel) WithSize(width, height int) *HybridFormModel {
	m.width = width
	m.height = height
	return m
}

// Init initializes the hybrid form model
func (m *HybridFormModel) Init() tea.Cmd {
	return m.form.Init()
}

// Update handles messages for the hybrid form
func (m *HybridFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.cancelled = true
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		if m.form.State == huh.StateCompleted {
			m.complete = true
			return m, tea.Quit
		}
	}

	return m, cmd
}

// View renders the hybrid form with optional progress and styling
func (m *HybridFormModel) View() string {
	if m.width == 0 {
		m.width = 80
	}

	var content string

	// Header with title and description
	if m.title != "" {
		titleStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("12")).
			Margin(0, 0, 1, 0)
		content += titleStyle.Render(m.title) + "\n"
	}

	if m.description != "" {
		descStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Margin(0, 0, 1, 0)
		content += descStyle.Render(m.description) + "\n"
	}

	// Progress indicator
	if m.showProgress {
		content += m.renderProgress() + "\n"
	}

	// Form content
	content += m.form.View()

	// Help text
	if !m.complete && !m.cancelled {
		helpStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Margin(1, 0, 0, 0)
		content += helpStyle.Render("Press Ctrl+C or Esc to cancel")
	}

	// Container styling
	containerStyle := lipgloss.NewStyle().
		Width(m.width).
		Padding(1, 2)

	return containerStyle.Render(content)
}

// IsComplete returns true if the form is completed
func (m *HybridFormModel) IsComplete() bool {
	return m.complete
}

// IsCancelled returns true if the form was cancelled
func (m *HybridFormModel) IsCancelled() bool {
	return m.cancelled
}

// GetForm returns the underlying huh form for data access
func (m *HybridFormModel) GetForm() *huh.Form {
	return m.form
}

// renderProgress renders a progress indicator
func (m *HybridFormModel) renderProgress() string {
	if m.totalSteps <= 1 {
		return ""
	}

	progress := float64(m.currentStep) / float64(m.totalSteps-1)
	if progress > 1.0 {
		progress = 1.0
	}

	// Progress bar
	progressWidth := 40
	filledWidth := int(progress * float64(progressWidth))
	if filledWidth > progressWidth {
		filledWidth = progressWidth
	}

	filled := strings.Repeat("█", filledWidth)
	empty := strings.Repeat("░", progressWidth-filledWidth)

	progressStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("12"))
	emptyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	progressBar := progressStyle.Render(filled) + emptyStyle.Render(empty)

	// Step counter
	stepText := fmt.Sprintf("Step %d of %d", m.currentStep+1, m.totalSteps)
	stepStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		MarginLeft(2)

	return progressBar + stepStyle.Render(stepText)
}

// HybridFormRunner provides utilities for running hybrid forms
type HybridFormRunner struct{}

// NewHybridFormRunner creates a new hybrid form runner
func NewHybridFormRunner() *HybridFormRunner {
	return &HybridFormRunner{}
}

// RunFormWithProgress runs a huh form within a bubbletea context with progress tracking
func (r *HybridFormRunner) RunFormWithProgress(form *huh.Form, title, description string, current, total int) error {
	model := NewHybridFormModel(form, title, description).
		WithProgress(current, total)

	program := tea.NewProgram(model)
	finalModel, err := program.Run()
	if err != nil {
		return fmt.Errorf("failed to run hybrid form: %w", err)
	}

	if hybridModel, ok := finalModel.(*HybridFormModel); ok {
		if hybridModel.IsCancelled() {
			return fmt.Errorf("form was cancelled")
		}
		if !hybridModel.IsComplete() {
			return fmt.Errorf("form was not completed")
		}
		return nil
	}

	return fmt.Errorf("unexpected model type returned")
}

// RunForm runs a huh form within a bubbletea context
func (r *HybridFormRunner) RunForm(form *huh.Form, title, description string) error {
	model := NewHybridFormModel(form, title, description)

	program := tea.NewProgram(model)
	finalModel, err := program.Run()
	if err != nil {
		return fmt.Errorf("failed to run hybrid form: %w", err)
	}

	if hybridModel, ok := finalModel.(*HybridFormModel); ok {
		if hybridModel.IsCancelled() {
			return fmt.Errorf("form was cancelled")
		}
		if !hybridModel.IsComplete() {
			return fmt.Errorf("form was not completed")
		}
		return nil
	}

	return fmt.Errorf("unexpected model type returned")
}
