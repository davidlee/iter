package wizard

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// DefaultFormRenderer implements FormRenderer with consistent styling
type DefaultFormRenderer struct {
	// Styles
	titleStyle       lipgloss.Style
	descriptionStyle lipgloss.Style
	progressStyle    lipgloss.Style
	navigationStyle  lipgloss.Style
	errorStyle       lipgloss.Style
	summaryStyle     lipgloss.Style
	breadcrumbStyle  lipgloss.Style
	
	// Layout
	width  int
	height int
}

// NewDefaultFormRenderer creates a new form renderer with default styling
func NewDefaultFormRenderer() *DefaultFormRenderer {
	return &DefaultFormRenderer{
		// Title styling - bright blue, bold
		titleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("12")).
			Margin(1, 0),
			
		// Description styling - gray, italic
		descriptionStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Italic(true).
			Margin(0, 0, 1, 0),
			
		// Progress bar styling - green accent
		progressStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("8")).
			Padding(0, 1).
			Margin(1, 0),
			
		// Navigation styling - subtle border
		navigationStyle: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true, false, false, false).
			BorderForeground(lipgloss.Color("8")).
			Padding(1, 0, 0, 0).
			Margin(1, 0, 0, 0),
			
		// Error styling - bright red, bold
		errorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Bold(true).
			Margin(1, 0),
			
		// Summary styling - cyan
		summaryStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("14")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("8")).
			Padding(1).
			Margin(1, 0),
			
		// Breadcrumb styling
		breadcrumbStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Margin(0, 0, 1, 0),
			
		width:  80,
		height: 24,
	}
}

func (r *DefaultFormRenderer) RenderForm(step HuhFormStep, state WizardState) string {
	var b strings.Builder
	
	// Render breadcrumbs
	b.WriteString(r.renderBreadcrumbs(state))
	b.WriteString("\n")
	
	// Render progress
	b.WriteString(r.RenderProgress(
		state.GetCurrentStep()+1, 
		state.GetTotalSteps(), 
		r.getCompletedStepsList(state),
	))
	b.WriteString("\n")
	
	// Render step title and description
	b.WriteString(r.titleStyle.Render(step.title))
	b.WriteString("\n")
	if step.description != "" {
		b.WriteString(r.descriptionStyle.Render(step.description))
		b.WriteString("\n")
	}
	
	// The actual form content would be rendered by the huh form
	// This is a placeholder for form integration
	b.WriteString("[ Form content will be rendered here ]")
	b.WriteString("\n")
	
	return b.String()
}

func (r *DefaultFormRenderer) RenderProgress(current, total int, completedSteps []int) string {
	// Create progress bar with step indicators
	progressText := fmt.Sprintf("Step %d of %d", current, total)
	
	// Create visual progress bar
	progressPercent := float64(current-1) / float64(total-1) * 100
	if total == 1 {
		progressPercent = 100
	}
	
	barWidth := 40
	filledWidth := int(float64(barWidth) * progressPercent / 100)
	
	var bar strings.Builder
	bar.WriteString("[")
	for i := 0; i < barWidth; i++ {
		if i < filledWidth {
			bar.WriteString("█")
		} else {
			bar.WriteString("░")
		}
	}
	bar.WriteString("]")
	
	progressBar := bar.String()
	progressLine := fmt.Sprintf("%s %s %.0f%%", progressText, progressBar, progressPercent)
	
	return r.progressStyle.Render(progressLine)
}

func (r *DefaultFormRenderer) RenderNavigation(nav NavigationController, state WizardState) string {
	var parts []string
	
	// Back button
	if nav.CanGoBack(state) {
		parts = append(parts, "← Back (b)")
	} else {
		parts = append(parts, "  Back    ")
	}
	
	// Forward/Next button
	if nav.CanGoForward(state) {
		if state.GetCurrentStep() == state.GetTotalSteps()-1 {
			parts = append(parts, "Finish (f) →")
		} else {
			parts = append(parts, "Next (n) →")
		}
	} else {
		parts = append(parts, "  Next     ")
	}
	
	// Cancel option
	parts = append(parts, "Cancel (ctrl+c)")
	
	navText := strings.Join(parts, "  |  ")
	return r.navigationStyle.Render(navText)
}

func (r *DefaultFormRenderer) RenderSummary(state WizardState) string {
	var b strings.Builder
	
	b.WriteString("Goal Configuration Summary:\n\n")
	
	// Basic info
	if basicInfo := state.GetStep(0); basicInfo != nil {
		if data, ok := basicInfo.GetData().(*BasicInfoStepData); ok {
			b.WriteString(fmt.Sprintf("Title: %s\n", data.Title))
			if data.Description != "" {
				b.WriteString(fmt.Sprintf("Description: %s\n", data.Description))
			}
			b.WriteString(fmt.Sprintf("Type: %s\n", data.GoalType))
		}
	}
	
	// Field configuration
	if fieldConfig := state.GetStep(1); fieldConfig != nil {
		if data, ok := fieldConfig.GetData().(*FieldConfigStepData); ok {
			b.WriteString(fmt.Sprintf("Field Type: %s\n", data.FieldType))
			if data.Unit != "" {
				b.WriteString(fmt.Sprintf("Unit: %s\n", data.Unit))
			}
		}
	}
	
	// Scoring configuration
	if scoring := state.GetStep(2); scoring != nil {
		if data, ok := scoring.GetData().(*ScoringStepData); ok {
			b.WriteString(fmt.Sprintf("Scoring: %s\n", data.ScoringType))
		}
	}
	
	return r.summaryStyle.Render(b.String())
}

func (r *DefaultFormRenderer) RenderValidationErrors(errors []ValidationError) string {
	if len(errors) == 0 {
		return ""
	}
	
	var b strings.Builder
	b.WriteString("⚠ Validation Errors:\n")
	
	for _, err := range errors {
		b.WriteString(fmt.Sprintf("• %s", err.Message))
		if err.Field != "" {
			b.WriteString(fmt.Sprintf(" (%s)", err.Field))
		}
		b.WriteString("\n")
	}
	
	return r.errorStyle.Render(b.String())
}

// Helper methods

func (r *DefaultFormRenderer) renderBreadcrumbs(state WizardState) string {
	var parts []string
	
	stepNames := r.getStepNames(state)
	currentStep := state.GetCurrentStep()
	
	for i, name := range stepNames {
		switch {
		case i == currentStep:
			parts = append(parts, fmt.Sprintf("[%s]", name))
		case state.IsStepCompleted(i):
			parts = append(parts, fmt.Sprintf("✓ %s", name))
		default:
			parts = append(parts, name)
		}
	}
	
	breadcrumbs := strings.Join(parts, " → ")
	return r.breadcrumbStyle.Render(breadcrumbs)
}

func (r *DefaultFormRenderer) getStepNames(state WizardState) []string {
	// This would be customized based on the goal type and configuration
	// For now, return generic step names
	totalSteps := state.GetTotalSteps()
	names := make([]string, totalSteps)
	
	for i := 0; i < totalSteps; i++ {
		switch i {
		case 0:
			names[i] = "Basic Info"
		case 1:
			names[i] = "Field Config"
		case 2:
			names[i] = "Scoring"
		case 3:
			names[i] = "Mini Criteria"
		case 4:
			names[i] = "Midi Criteria"
		case 5:
			names[i] = "Maxi Criteria"
		case 6:
			names[i] = "Validation"
		case 7:
			names[i] = "Confirmation"
		default:
			names[i] = fmt.Sprintf("Step %d", i+1)
		}
	}
	
	return names
}

func (r *DefaultFormRenderer) getCompletedStepsList(state WizardState) []int {
	var completed []int
	for i := 0; i < state.GetTotalSteps(); i++ {
		if state.IsStepCompleted(i) {
			completed = append(completed, i)
		}
	}
	return completed
}