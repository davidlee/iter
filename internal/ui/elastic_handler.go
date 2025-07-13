package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"davidlee/iter/internal/models"
	"davidlee/iter/internal/scoring"
)

// ElasticGoalHandler handles entry collection for elastic goals with mini/midi/maxi achievement levels.
type ElasticGoalHandler struct {
	scoringEngine *scoring.Engine
}

// NewElasticGoalHandler creates a new elastic goal handler with scoring integration.
func NewElasticGoalHandler(scoringEngine *scoring.Engine) *ElasticGoalHandler {
	return &ElasticGoalHandler{
		scoringEngine: scoringEngine,
	}
}

// AIDEV-NOTE: elastic-goal-entry-handler; current bubbletea+huh implementation pattern for field type adaptation (reference for T010)
// CollectEntry collects an entry for an elastic goal including automatic scoring and achievement display.
func (h *ElasticGoalHandler) CollectEntry(goal models.Goal, existing *ExistingEntry) (*EntryResult, error) {
	// Prepare the form title with goal information
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")). // Bright blue
		Margin(1, 0)

	_ = titleStyle.Render(goal.Title) // Title styling available for future use

	// Prepare description if available
	var description string
	if goal.Description != "" {
		descStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")). // Gray
			Italic(true)
		description = descStyle.Render(goal.Description)
	}

	// Add criteria information to description for motivation
	if goal.RequiresAutomaticScoring() {
		criteriaInfo := h.formatCriteriaInfo(goal)
		if criteriaInfo != "" {
			if description != "" {
				description += "\n"
			}
			description += criteriaInfo
		}
	}
	_ = description // Description used in field type collection

	// Collect value based on field type
	value, err := h.collectValueByFieldType(goal, existing)
	if err != nil {
		return nil, fmt.Errorf("failed to collect value: %w", err)
	}

	// Score the value if automatic scoring is enabled
	var achievementLevel *models.AchievementLevel
	if goal.RequiresAutomaticScoring() {
		scoreResult, err := h.scoringEngine.ScoreElasticGoal(&goal, value)
		if err != nil {
			// Fall back to manual scoring if automatic scoring fails
			manualLevel, err := h.collectManualAchievementLevel(goal, value)
			if err != nil {
				return nil, fmt.Errorf("automatic scoring failed and manual scoring failed: %w", err)
			}
			achievementLevel = manualLevel
		} else {
			achievementLevel = &scoreResult.AchievementLevel
			// Display the achievement result
			h.displayAchievementResult(scoreResult, goal)
		}
	} else {
		// Manual scoring for elastic goals without automatic criteria
		manualLevel, err := h.collectManualAchievementLevel(goal, value)
		if err != nil {
			return nil, fmt.Errorf("failed to collect manual achievement level: %w", err)
		}
		achievementLevel = manualLevel
	}

	// Collect optional notes
	notes, err := h.collectOptionalNotes(goal, value, existing)
	if err != nil {
		return nil, fmt.Errorf("failed to collect notes: %w", err)
	}

	return &EntryResult{
		Value:            value,
		AchievementLevel: achievementLevel,
		Notes:            notes,
	}, nil
}

// collectValueByFieldType collects a value based on the goal's field type.
func (h *ElasticGoalHandler) collectValueByFieldType(goal models.Goal, existing *ExistingEntry) (interface{}, error) {
	switch goal.FieldType.Type {
	case models.BooleanFieldType:
		return h.collectBooleanValue(goal, existing)
	case models.UnsignedIntFieldType, models.UnsignedDecimalFieldType, models.DecimalFieldType:
		return h.collectNumericValue(goal, existing)
	case models.DurationFieldType:
		return h.collectDurationValue(goal, existing)
	case models.TimeFieldType:
		return h.collectTimeValue(goal, existing)
	case models.TextFieldType:
		return h.collectTextValue(goal, existing)
	default:
		return nil, fmt.Errorf("unsupported field type: %s", goal.FieldType.Type)
	}
}

// collectBooleanValue collects a boolean value using a confirmation dialog.
func (h *ElasticGoalHandler) collectBooleanValue(goal models.Goal, existing *ExistingEntry) (bool, error) {
	var currentValue bool
	var hasExisting bool
	if existing != nil && existing.Value != nil {
		if boolVal, ok := existing.Value.(bool); ok {
			currentValue = boolVal
			hasExisting = true
		}
	}

	var completed bool
	prompt := goal.Prompt
	if prompt == "" {
		prompt = fmt.Sprintf("Did you complete: %s?", goal.Title)
	}

	if hasExisting {
		status := "❌ No"
		if currentValue {
			status = "✅ Yes"
		}
		prompt = fmt.Sprintf("%s (currently: %s)", prompt, status)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(prompt).
				Value(&completed).
				Affirmative("Yes").
				Negative("No"),
		),
	)

	if err := form.Run(); err != nil {
		return false, fmt.Errorf("boolean form failed: %w", err)
	}

	return completed, nil
}

// collectNumericValue collects a numeric value with validation.
func (h *ElasticGoalHandler) collectNumericValue(goal models.Goal, existing *ExistingEntry) (float64, error) {
	var currentValue string
	if existing != nil && existing.Value != nil {
		currentValue = fmt.Sprintf("%v", existing.Value)
	}

	valueStr := currentValue
	prompt := goal.Prompt
	if prompt == "" {
		unit := goal.FieldType.Unit
		if unit != "" {
			prompt = fmt.Sprintf("Enter value for %s (%s):", goal.Title, unit)
		} else {
			prompt = fmt.Sprintf("Enter value for %s:", goal.Title)
		}
	}

	if currentValue != "" {
		prompt = fmt.Sprintf("%s (current: %s)", prompt, currentValue)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(prompt).
				Value(&valueStr).
				Placeholder("Enter numeric value"),
		),
	)

	if err := form.Run(); err != nil {
		return 0, fmt.Errorf("numeric form failed: %w", err)
	}

	value, err := strconv.ParseFloat(strings.TrimSpace(valueStr), 64)
	if err != nil {
		return 0, fmt.Errorf("invalid numeric value: %w", err)
	}

	return value, nil
}

// collectDurationValue collects a duration value with format hints.
func (h *ElasticGoalHandler) collectDurationValue(goal models.Goal, existing *ExistingEntry) (string, error) {
	var currentValue string
	if existing != nil && existing.Value != nil {
		currentValue = fmt.Sprintf("%v", existing.Value)
	}

	valueStr := currentValue
	prompt := goal.Prompt
	if prompt == "" {
		prompt = fmt.Sprintf("Enter duration for %s:", goal.Title)
	}

	if currentValue != "" {
		prompt = fmt.Sprintf("%s (current: %s)", prompt, currentValue)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(prompt).
				Description("Examples: 30 (minutes), 1h30m, 1:30:00").
				Value(&valueStr).
				Placeholder("30m"),
		),
	)

	if err := form.Run(); err != nil {
		return "", fmt.Errorf("duration form failed: %w", err)
	}

	return strings.TrimSpace(valueStr), nil
}

// collectTimeValue collects a time value with format hints.
func (h *ElasticGoalHandler) collectTimeValue(goal models.Goal, existing *ExistingEntry) (string, error) {
	var currentValue string
	if existing != nil && existing.Value != nil {
		currentValue = fmt.Sprintf("%v", existing.Value)
	}

	valueStr := currentValue
	prompt := goal.Prompt
	if prompt == "" {
		prompt = fmt.Sprintf("Enter time for %s:", goal.Title)
	}

	if currentValue != "" {
		prompt = fmt.Sprintf("%s (current: %s)", prompt, currentValue)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(prompt).
				Description("Format: HH:MM (24-hour format)").
				Value(&valueStr).
				Placeholder("14:30"),
		),
	)

	if err := form.Run(); err != nil {
		return "", fmt.Errorf("time form failed: %w", err)
	}

	return strings.TrimSpace(valueStr), nil
}

// collectTextValue collects a text value.
func (h *ElasticGoalHandler) collectTextValue(goal models.Goal, existing *ExistingEntry) (string, error) {
	var currentValue string
	if existing != nil && existing.Value != nil {
		currentValue = fmt.Sprintf("%v", existing.Value)
	}

	valueStr := currentValue
	prompt := goal.Prompt
	if prompt == "" {
		prompt = fmt.Sprintf("Enter text for %s:", goal.Title)
	}

	if currentValue != "" {
		prompt = fmt.Sprintf("%s (current: %s)", prompt, currentValue)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(prompt).
				Value(&valueStr).
				Placeholder("Enter your response"),
		),
	)

	if err := form.Run(); err != nil {
		return "", fmt.Errorf("text form failed: %w", err)
	}

	return strings.TrimSpace(valueStr), nil
}

// collectManualAchievementLevel allows manual selection of achievement level.
func (h *ElasticGoalHandler) collectManualAchievementLevel(goal models.Goal, value interface{}) (*models.AchievementLevel, error) {
	level := models.AchievementNone

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[models.AchievementLevel]().
				Title(fmt.Sprintf("Select achievement level for %s (value: %v):", goal.Title, value)).
				Options(
					huh.NewOption("None", models.AchievementNone),
					huh.NewOption("Mini", models.AchievementMini),
					huh.NewOption("Midi", models.AchievementMidi),
					huh.NewOption("Maxi", models.AchievementMaxi),
				).
				Value(&level),
		),
	)

	if err := form.Run(); err != nil {
		return nil, fmt.Errorf("manual achievement level form failed: %w", err)
	}

	return &level, nil
}

// displayAchievementResult shows the scoring result with appropriate styling.
func (h *ElasticGoalHandler) displayAchievementResult(result *scoring.ScoreResult, _ models.Goal) {
	// Choose styling based on achievement level
	var style lipgloss.Style
	var emoji string
	var levelName string

	switch result.AchievementLevel {
	case models.AchievementMaxi:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true) // Bright green
		emoji = "🌟"
		levelName = "MAXI"
	case models.AchievementMidi:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true) // Bright yellow
		emoji = "🎯"
		levelName = "MIDI"
	case models.AchievementMini:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true) // Bright blue
		emoji = "✨"
		levelName = "MINI"
	default:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("8")) // Gray
		emoji = "📝"
		levelName = "NONE"
	}

	message := fmt.Sprintf("%s Achievement Level: %s", emoji, levelName)

	fmt.Println()
	fmt.Println(style.Render(message))

	// Show which levels were met for detailed feedback
	if result.MetMini || result.MetMidi || result.MetMaxi {
		details := "Levels achieved: "
		var achieved []string
		if result.MetMini {
			achieved = append(achieved, "Mini")
		}
		if result.MetMidi {
			achieved = append(achieved, "Midi")
		}
		if result.MetMaxi {
			achieved = append(achieved, "Maxi")
		}
		details += strings.Join(achieved, ", ")

		detailStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Faint(true)
		fmt.Println(detailStyle.Render(details))
	}
	fmt.Println()
}

// formatCriteriaInfo formats the criteria information for display as motivation.
func (h *ElasticGoalHandler) formatCriteriaInfo(goal models.Goal) string {
	if !goal.RequiresAutomaticScoring() {
		return ""
	}

	var parts []string
	criteriaStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Faint(true)

	if goal.MiniCriteria != nil {
		if value := extractDisplayValue(goal.MiniCriteria); value != "" {
			parts = append(parts, fmt.Sprintf("Mini: %s", value))
		}
	}
	if goal.MidiCriteria != nil {
		if value := extractDisplayValue(goal.MidiCriteria); value != "" {
			parts = append(parts, fmt.Sprintf("Midi: %s", value))
		}
	}
	if goal.MaxiCriteria != nil {
		if value := extractDisplayValue(goal.MaxiCriteria); value != "" {
			parts = append(parts, fmt.Sprintf("Maxi: %s", value))
		}
	}

	if len(parts) > 0 {
		return criteriaStyle.Render("🎯 Criteria: " + strings.Join(parts, " • "))
	}
	return ""
}

// extractDisplayValue extracts a display-friendly value from criteria.
func extractDisplayValue(criteria *models.Criteria) string {
	if criteria == nil || criteria.Condition == nil {
		return ""
	}

	cond := criteria.Condition
	if cond.GreaterThanOrEqual != nil {
		return fmt.Sprintf("≥%.0f", *cond.GreaterThanOrEqual)
	}
	if cond.GreaterThan != nil {
		return fmt.Sprintf(">%.0f", *cond.GreaterThan)
	}
	if cond.LessThanOrEqual != nil {
		return fmt.Sprintf("≤%.0f", *cond.LessThanOrEqual)
	}
	if cond.LessThan != nil {
		return fmt.Sprintf("<%.0f", *cond.LessThan)
	}
	if cond.Equals != nil {
		return fmt.Sprintf("=%v", *cond.Equals)
	}
	return ""
}

// collectOptionalNotes allows the user to optionally add notes for an elastic goal.
func (h *ElasticGoalHandler) collectOptionalNotes(_ models.Goal, _ interface{}, existing *ExistingEntry) (string, error) {
	// Get existing notes
	var existingNotes string
	if existing != nil {
		existingNotes = existing.Notes
	}

	// Ask if user wants to add notes
	var wantNotes bool
	notesPrompt := "Add notes for this entry?"
	if existingNotes != "" {
		notesPrompt = fmt.Sprintf("Update notes? (current: %s)", existingNotes)
	}

	notesForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(notesPrompt).
				Value(&wantNotes).
				Affirmative("Yes").
				Negative("Skip"),
		),
	)

	if err := notesForm.Run(); err != nil {
		return "", fmt.Errorf("notes prompt failed: %w", err)
	}

	if !wantNotes {
		return existingNotes, nil // Return existing notes unchanged
	}

	// Collect the notes
	var notes string
	if existingNotes != "" {
		notes = existingNotes // Pre-populate with existing notes
	}

	notesInputForm := huh.NewForm(
		huh.NewGroup(
			huh.NewText().
				Title("Notes:").
				Description("Optional notes about this goal (press Enter when done)").
				Value(&notes).
				Placeholder("How did it go? Any observations?"),
		),
	)

	if err := notesInputForm.Run(); err != nil {
		return "", fmt.Errorf("notes input failed: %w", err)
	}

	// Return the notes (trimmed)
	return strings.TrimSpace(notes), nil
}
