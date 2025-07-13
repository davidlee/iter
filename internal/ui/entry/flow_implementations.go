package entry

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"davidlee/iter/internal/models"
)

// AIDEV-NOTE: flow-implementations; concrete implementations of goal collection flow methods
// Provides scoring, feedback display, and notes collection for each goal type flow
// AIDEV-NOTE: T010/3.1-3.3-complete; Simple/Elastic/Informational goal collection flows fully implemented and tested
// Key methods: performAutomaticScoring (elastic conversion), determineManualAchievement (field-type aware), displayDirectionFeedback (direction-aware), collectOptionalNotes

// Simple Goal Flow Implementations

func (f *SimpleGoalCollectionFlow) performAutomaticScoring(goal models.Goal, value interface{}) (*models.AchievementLevel, error) {
	if f.scoringEngine == nil {
		return nil, fmt.Errorf("scoring engine not available")
	}

	// For simple goals, treat them as single-criteria elastic goals
	// Convert simple goal criteria to elastic format for scoring
	elasticGoal := goal
	if goal.Criteria != nil {
		// Use the single criteria as mini criteria for elastic scoring
		elasticGoal.MiniCriteria = goal.Criteria
	}

	scoreResult, err := f.scoringEngine.ScoreElasticGoal(&elasticGoal, value)
	if err != nil {
		return nil, fmt.Errorf("scoring failed: %w", err)
	}

	// For simple goals, convert elastic result to pass/fail
	level := models.AchievementNone
	if scoreResult.MetMini {
		level = models.AchievementMini
	}

	return &level, nil
}

func (f *SimpleGoalCollectionFlow) determineManualAchievement(goal models.Goal, value interface{}) *models.AchievementLevel {
	// For manual simple goals, determine pass/fail based on field type
	switch goal.FieldType.Type {
	case models.BooleanFieldType:
		if boolVal, ok := value.(bool); ok {
			if boolVal {
				level := models.AchievementMini
				return &level
			}
			level := models.AchievementNone
			return &level
		}
	case models.TextFieldType:
		// Text fields require manual scoring (per T009 design decisions)
		// Default to Mini if text is provided
		if textVal, ok := value.(string); ok && strings.TrimSpace(textVal) != "" {
			level := models.AchievementMini
			return &level
		}
	default:
		// Other field types default to Mini if value is provided
		if value != nil {
			level := models.AchievementMini
			return &level
		}
	}

	// Default to None if no value or false boolean
	level := models.AchievementNone
	return &level
}

func (f *SimpleGoalCollectionFlow) collectOptionalNotes(_ models.Goal, _ interface{}, existing *ExistingEntry) (string, error) {
	return collectStandardOptionalNotes(existing)
}

// Elastic Goal Flow Implementations

func (f *ElasticGoalCollectionFlow) displayCriteriaInformation(goal models.Goal) {
	if !goal.RequiresAutomaticScoring() {
		return
	}

	criteriaStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")). // Bright green
		Faint(true).
		Margin(1, 0)

	var parts []string
	if goal.MiniCriteria != nil {
		if value := extractCriteriaDisplayValue(goal.MiniCriteria); value != "" {
			parts = append(parts, fmt.Sprintf("Mini: %s", value))
		}
	}
	if goal.MidiCriteria != nil {
		if value := extractCriteriaDisplayValue(goal.MidiCriteria); value != "" {
			parts = append(parts, fmt.Sprintf("Midi: %s", value))
		}
	}
	if goal.MaxiCriteria != nil {
		if value := extractCriteriaDisplayValue(goal.MaxiCriteria); value != "" {
			parts = append(parts, fmt.Sprintf("Maxi: %s", value))
		}
	}

	if len(parts) > 0 {
		criteriaInfo := criteriaStyle.Render("🎯 Achievement Criteria: " + strings.Join(parts, " • "))
		fmt.Println(criteriaInfo)
	}
}

func (f *ElasticGoalCollectionFlow) performElasticScoring(goal models.Goal, value interface{}) (*models.AchievementLevel, error) {
	if f.scoringEngine == nil {
		return nil, fmt.Errorf("scoring engine not available")
	}

	// Use existing scoring engine for elastic goals with three-tier criteria
	scoreResult, err := f.scoringEngine.ScoreElasticGoal(&goal, value)
	if err != nil {
		return nil, fmt.Errorf("elastic scoring failed: %w", err)
	}

	return &scoreResult.AchievementLevel, nil
}

func (f *ElasticGoalCollectionFlow) collectManualAchievementLevel(goal models.Goal, value interface{}) (*models.AchievementLevel, error) {
	level := models.AchievementNone

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[models.AchievementLevel]().
				Title(fmt.Sprintf("Select achievement level for %s (value: %v):", goal.Title, value)).
				Description("Choose the achievement level that best represents your performance").
				Options(
					huh.NewOption("None - No achievement", models.AchievementNone),
					huh.NewOption("Mini - Minimal achievement", models.AchievementMini),
					huh.NewOption("Midi - Moderate achievement", models.AchievementMidi),
					huh.NewOption("Maxi - Maximum achievement", models.AchievementMaxi),
				).
				Value(&level),
		),
	)

	if err := form.Run(); err != nil {
		return nil, fmt.Errorf("manual achievement level form failed: %w", err)
	}

	return &level, nil
}

func (f *ElasticGoalCollectionFlow) displayAchievementResult(goal models.Goal, _ interface{}, level *models.AchievementLevel) {
	if level == nil {
		return
	}

	// Achievement display styling based on level
	var style lipgloss.Style
	var emoji string
	var levelName string
	var message string

	switch *level {
	case models.AchievementMaxi:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true) // Bright green
		emoji = "🌟"
		levelName = "MAXI"
		message = "Outstanding achievement!"
	case models.AchievementMidi:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true) // Bright yellow
		emoji = "🎯"
		levelName = "MIDI"
		message = "Great progress!"
	case models.AchievementMini:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true) // Bright blue
		emoji = "✨"
		levelName = "MINI"
		message = "Good start!"
	default:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("8")) // Gray
		emoji = "📝"
		levelName = "NONE"
		message = "Entry recorded"
	}

	achievementMsg := fmt.Sprintf("%s %s Achievement: %s", emoji, goal.Title, levelName)

	fmt.Println()
	fmt.Println(style.Render(achievementMsg))
	fmt.Println(style.Render(message))
	fmt.Println()
}

func (f *ElasticGoalCollectionFlow) collectOptionalNotes(_ models.Goal, _ interface{}, existing *ExistingEntry) (string, error) {
	return collectStandardOptionalNotes(existing)
}

// Informational Goal Flow Implementations

func (f *InformationalGoalCollectionFlow) displayInformationalContext(_ models.Goal) {
	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("14")). // Bright cyan
		Faint(true).
		Margin(1, 0)

	contextMsg := "ℹ️  This is an informational goal - for tracking data only (no scoring)"
	fmt.Println(infoStyle.Render(contextMsg))
}

// AIDEV-NOTE: direction-aware-feedback; displays value with directional context for informational goals
// Supports higher_better (green 📈), lower_better (blue 📉), neutral (gray 📊) with contextual hints
func (f *InformationalGoalCollectionFlow) displayDirectionFeedback(goal models.Goal, value interface{}) {
	// Display direction-aware feedback based on goal.Direction configuration
	var style lipgloss.Style
	var emoji string
	var directionHint string

	// Direction-specific styling and feedback
	switch strings.ToLower(goal.Direction) {
	case "higher_better":
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("10")) // Bright green
		emoji = "📈"
		directionHint = " (higher is better)"
	case "lower_better":
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("12")) // Bright blue
		emoji = "📉"
		directionHint = " (lower is better)"
	case "neutral", "":
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("8")) // Gray
		emoji = "📊"
		directionHint = ""
	default:
		// Unknown direction - fall back to neutral
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("8")) // Gray
		emoji = "📊"
		directionHint = ""
	}

	style = style.Faint(true)
	feedback := fmt.Sprintf("%s Recorded: %v%s", emoji, value, directionHint)
	fmt.Println(style.Render(feedback))
}

func (f *InformationalGoalCollectionFlow) collectOptionalNotes(_ models.Goal, _ interface{}, existing *ExistingEntry) (string, error) {
	return collectStandardOptionalNotes(existing)
}

// Checklist Goal Flow Implementations

func (f *ChecklistGoalCollectionFlow) displayChecklistContext(_ models.Goal, existing *ExistingEntry) {
	contextStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("13")). // Bright magenta
		Faint(true).
		Margin(1, 0)

	var contextMsg string
	if existing != nil && existing.Value != nil {
		if items, ok := existing.Value.([]string); ok {
			total := len([]string{"Item 1", "Item 2", "Item 3"}) // TODO: Get actual total from checklist definition
			completed := len(items)
			contextMsg = fmt.Sprintf("📋 Checklist Progress: %d/%d items completed", completed, total)
		}
	} else {
		contextMsg = "📋 Complete the checklist items below"
	}

	fmt.Println(contextStyle.Render(contextMsg))
}

func (f *ChecklistGoalCollectionFlow) performChecklistScoring(_ models.Goal, value interface{}) (*models.AchievementLevel, error) {
	if f.scoringEngine == nil {
		return nil, fmt.Errorf("scoring engine not available")
	}

	// For checklist goals, calculate completion percentage and score accordingly
	selectedItems, ok := value.([]string)
	if !ok {
		return nil, fmt.Errorf("invalid checklist value type: %T", value)
	}

	// Calculate completion percentage
	completed := len(selectedItems)
	total := len([]string{"Item 1", "Item 2", "Item 3"}) // TODO: Get actual total from checklist definition

	if total == 0 {
		level := models.AchievementNone
		return &level, nil
	}

	percentage := float64(completed) / float64(total)

	// Determine achievement level based on completion percentage
	// This is a simplified scoring - real implementation would use goal criteria
	var level models.AchievementLevel
	switch {
	case percentage >= 1.0:
		level = models.AchievementMaxi
	case percentage >= 0.75:
		level = models.AchievementMidi
	case percentage >= 0.5:
		level = models.AchievementMini
	default:
		level = models.AchievementNone
	}

	return &level, nil
}

func (f *ChecklistGoalCollectionFlow) collectManualAchievementLevel(goal models.Goal, value interface{}) (*models.AchievementLevel, error) {
	// Similar to elastic goals but with checklist-specific context
	level := models.AchievementNone

	// Calculate completion percentage for context
	var completionInfo string
	if items, ok := value.([]string); ok {
		completed := len(items)
		total := len([]string{"Item 1", "Item 2", "Item 3"}) // TODO: Get actual total from checklist definition
		percentage := float64(completed) / float64(total) * 100
		completionInfo = fmt.Sprintf("(%d/%d items = %.0f%% complete)", completed, total, percentage)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[models.AchievementLevel]().
				Title(fmt.Sprintf("Select achievement level for %s %s:", goal.Title, completionInfo)).
				Description("Choose the achievement level based on your checklist completion").
				Options(
					huh.NewOption("None - Minimal completion", models.AchievementNone),
					huh.NewOption("Mini - Basic completion", models.AchievementMini),
					huh.NewOption("Midi - Good completion", models.AchievementMidi),
					huh.NewOption("Maxi - Excellent completion", models.AchievementMaxi),
				).
				Value(&level),
		),
	)

	if err := form.Run(); err != nil {
		return nil, fmt.Errorf("manual achievement level form failed: %w", err)
	}

	return &level, nil
}

func (f *ChecklistGoalCollectionFlow) displayCompletionProgress(_ models.Goal, value interface{}, level *models.AchievementLevel) {
	if items, ok := value.([]string); ok {
		completed := len(items)
		total := len([]string{"Item 1", "Item 2", "Item 3"}) // TODO: Get actual total from checklist definition
		percentage := float64(completed) / float64(total) * 100

		progressStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("13")). // Bright magenta
			Bold(true)

		progressMsg := fmt.Sprintf("📋 Checklist Complete: %d/%d items (%.0f%%)", completed, total, percentage)

		if level != nil {
			achievementMsg := ""
			switch *level {
			case models.AchievementMaxi:
				achievementMsg = " 🌟 Excellent work!"
			case models.AchievementMidi:
				achievementMsg = " 🎯 Good progress!"
			case models.AchievementMini:
				achievementMsg = " ✨ Nice start!"
			}
			progressMsg += achievementMsg
		}

		fmt.Println()
		fmt.Println(progressStyle.Render(progressMsg))
		fmt.Println()
	}
}

func (f *ChecklistGoalCollectionFlow) collectOptionalNotes(_ models.Goal, _ interface{}, existing *ExistingEntry) (string, error) {
	return collectStandardOptionalNotes(existing)
}

// Common Helper Functions

func collectStandardOptionalNotes(existing *ExistingEntry) (string, error) {
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
				Description("Optional notes about this entry (press Enter when done)").
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

func extractCriteriaDisplayValue(criteria *models.Criteria) string {
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
