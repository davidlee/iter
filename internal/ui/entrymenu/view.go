package entrymenu

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"davidlee/vice/internal/models"
)

// ViewRenderer handles the visual rendering of the entry menu interface.
// AIDEV-NOTE: view-renderer; separates presentation logic from model state management
type ViewRenderer struct {
	width  int
	height int
}

// NewViewRenderer creates a new view renderer with specified dimensions.
func NewViewRenderer(width, height int) *ViewRenderer {
	return &ViewRenderer{
		width:  width,
		height: height,
	}
}

// RenderProgressBar renders the progress bar showing completion status.
func (v *ViewRenderer) RenderProgressBar(goals []models.Goal, entries map[string]models.GoalEntry) string {
	if len(goals) == 0 {
		return v.renderEmptyProgress()
	}

	stats := v.calculateProgressStats(goals, entries)
	return v.renderProgressWithStats(stats)
}

// RenderFilters renders the current filter state indicator.
func (v *ViewRenderer) RenderFilters(filterState FilterState) string {
	if filterState == FilterNone {
		return ""
	}

	var filters []string
	if filterState == FilterHideSkipped || filterState == FilterHideSkippedAndPrevious {
		filters = append(filters, "hiding skipped")
	}
	if filterState == FilterHidePrevious || filterState == FilterHideSkippedAndPrevious {
		filters = append(filters, "hiding previous")
	}

	filterText := "Filters: " + strings.Join(filters, ", ")
	return filterStyle.Render(filterText)
}

// RenderReturnBehavior renders the current return behavior indicator.
func (v *ViewRenderer) RenderReturnBehavior(behavior ReturnBehavior) string {
	var behaviorText string
	switch behavior {
	case ReturnToMenu:
		behaviorText = "Return: menu"
	case ReturnToNextGoal:
		behaviorText = "Return: next goal"
	default:
		behaviorText = "Return: menu"
	}

	return returnBehaviorStyle.Render(behaviorText)
}

// RenderHeader renders the complete header section with progress, filters, and behavior.
func (v *ViewRenderer) RenderHeader(goals []models.Goal, entries map[string]models.GoalEntry, filterState FilterState, returnBehavior ReturnBehavior) string {
	var headerParts []string

	// Progress bar
	progressBar := v.RenderProgressBar(goals, entries)
	if progressBar != "" {
		headerParts = append(headerParts, progressBar)
	}

	// Status line with filters and return behavior
	var statusParts []string
	
	filters := v.RenderFilters(filterState)
	if filters != "" {
		statusParts = append(statusParts, filters)
	}
	
	returnBehaviorText := v.RenderReturnBehavior(returnBehavior)
	statusParts = append(statusParts, returnBehaviorText)

	if len(statusParts) > 0 {
		statusLine := strings.Join(statusParts, " | ")
		headerParts = append(headerParts, statusLine)
	}

	return strings.Join(headerParts, "\n")
}

// ProgressStats holds calculated progress statistics.
type ProgressStats struct {
	Total     int
	Completed int
	Failed    int
	Skipped   int
	Attempted int
	Remaining int
}

// calculateProgressStats computes progress statistics from goals and entries.
func (v *ViewRenderer) calculateProgressStats(goals []models.Goal, entries map[string]models.GoalEntry) ProgressStats {
	stats := ProgressStats{
		Total: len(goals),
	}

	for _, goal := range goals {
		if entry, hasEntry := entries[goal.ID]; hasEntry {
			stats.Attempted++
			switch entry.Status {
			case models.EntryCompleted:
				stats.Completed++
			case models.EntryFailed:
				stats.Failed++
			case models.EntrySkipped:
				stats.Skipped++
			}
		}
	}

	stats.Remaining = stats.Total - stats.Attempted
	return stats
}

// renderProgressWithStats renders a detailed progress bar with statistics.
func (v *ViewRenderer) renderProgressWithStats(stats ProgressStats) string {
	// Calculate percentages
	completedPct := float64(stats.Completed) / float64(stats.Total) * 100

	// Create progress bar visual
	progressBar := v.renderProgressBarVisual(completedPct, stats.Total)

	// Create status text
	statusText := fmt.Sprintf(
		"Progress: %d/%d completed (%.1f%%) | %d failed | %d skipped | %d remaining",
		stats.Completed, stats.Total, completedPct,
		stats.Failed, stats.Skipped, stats.Remaining,
	)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		progressBar,
		progressStyle.Render(statusText),
	)
}

// renderProgressBarVisual creates a visual progress bar representation.
func (v *ViewRenderer) renderProgressBarVisual(completedPct float64, _ int) string {
	if v.width <= 0 {
		return ""
	}

	// Calculate bar width (leave space for brackets and text)
	barWidth := v.width - 20
	if barWidth < 10 {
		barWidth = 10
	}

	// Calculate filled portion
	filledWidth := int(float64(barWidth) * completedPct / 100)
	if filledWidth > barWidth {
		filledWidth = barWidth
	}

	// Create bar components
	filled := strings.Repeat("█", filledWidth)
	empty := strings.Repeat("░", barWidth-filledWidth)

	// Style the bar
	filledBar := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")). // gold
		Render(filled)

	emptyBar := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")). // dark grey
		Render(empty)

	// Combine with brackets
	progressBar := fmt.Sprintf("[%s%s] %.1f%%", filledBar, emptyBar, completedPct)

	return progressBarStyle.Render(progressBar)
}

// renderEmptyProgress renders progress display when no goals are present.
func (v *ViewRenderer) renderEmptyProgress() string {
	return progressStyle.Render("No goals configured")
}

// Styling definitions for the view renderer.
var (
	// Progress bar styling
	progressBarStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")). // white
		Bold(true)

	progressStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("12")). // bright blue
		Bold(true)

	// Filter styling
	filterStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("11")). // bright yellow
		Italic(true)

	// Return behavior styling
	returnBehaviorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("14")). // bright cyan
		Italic(true)
)