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

	// Progress bar with right-aligned return behavior
	progressSection := v.renderProgressWithReturnBehavior(goals, entries, returnBehavior)
	if progressSection != "" {
		headerParts = append(headerParts, progressSection)
	}

	// Filters on separate line if present
	filters := v.RenderFilters(filterState)
	if filters != "" {
		headerParts = append(headerParts, filters)
	}

	return strings.Join(headerParts, "\n")
}

// renderProgressWithReturnBehavior renders progress bar with right-aligned return behavior.
// AIDEV-NOTE: layout-improvement; T018 user-requested right-alignment of return behavior text
func (v *ViewRenderer) renderProgressWithReturnBehavior(goals []models.Goal, entries map[string]models.GoalEntry, returnBehavior ReturnBehavior) string {
	if len(goals) == 0 {
		return progressStyle.Render("No goals configured")
	}

	stats := v.calculateProgressStats(goals, entries)
	completedPct := float64(stats.Completed) / float64(stats.Total) * 100
	
	// Create visual progress bar
	progressBarVisual := v.renderProgressBarVisual(completedPct)
	
	// Create progress text
	progressText := fmt.Sprintf(
		"Progress: %d/%d completed (%.1f%%) | %d failed | %d skipped | %d remaining",
		stats.Completed, stats.Total, completedPct,
		stats.Failed, stats.Skipped, stats.Remaining,
	)
	
	// Create return behavior text
	var returnText string
	switch returnBehavior {
	case ReturnToMenu:
		returnText = "Return: menu"
	case ReturnToNextGoal:
		returnText = "Return: next goal"
	default:
		returnText = "Return: menu"
	}
	
	// Try to fit progress text and return behavior on same line if width allows
	totalLength := len(progressText) + len(returnText) + 3 // 3 spaces between
	var statusLine string
	
	if v.width > 0 && totalLength <= v.width {
		// Fit on same line with right alignment
		spacesNeeded := v.width - len(progressText) - len(returnText)
		if spacesNeeded < 1 {
			spacesNeeded = 1
		}
		spacing := strings.Repeat(" ", spacesNeeded)
		
		leftSide := progressStyle.Render(progressText)
		rightSide := returnBehaviorStyle.Render(returnText)
		
		statusLine = leftSide + spacing + rightSide
	} else {
		// Separate lines if doesn't fit
		statusLine = progressStyle.Render(progressText)
	}
	
	// Combine progress bar visual and status line
	return lipgloss.JoinVertical(
		lipgloss.Left,
		progressBarVisual,
		statusLine,
	)
}

// renderProgressBarVisual creates a visual progress bar representation.
func (v *ViewRenderer) renderProgressBarVisual(completedPct float64) string {
	if v.width <= 0 {
		return ""
	}

	// Calculate bar width (leave space for brackets and percentage)
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