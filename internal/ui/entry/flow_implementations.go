package entry

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"davidlee/vice/internal/models"
)

// AIDEV-NOTE: flow-implementations; concrete implementations of habit collection flow methods
// Provides scoring, feedback display, and notes collection for each habit type flow
// AIDEV-NOTE: T010/3.1-3.3-complete; Simple/Elastic/Informational habit collection flows fully implemented and tested
// Key methods: performAutomaticScoring (elastic conversion), determineManualAchievement (field-type aware), displayDirectionFeedback (direction-aware), collectOptionalNotes

// Simple Habit Flow Implementations

// AIDEV-NOTE: T016-fix; proper habit type handling prevents "not an elastic habit" errors
// This method was rewritten to use dedicated ScoreSimpleHabit() instead of masquerading simple habits as elastic
func (f *SimpleHabitCollectionFlow) performAutomaticScoring(habit models.Habit, value interface{}) (*models.AchievementLevel, error) {
	if f.scoringEngine == nil {
		return nil, fmt.Errorf("scoring engine not available")
	}

	// Use dedicated simple habit scoring method (T016 fix)
	// OLD APPROACH: Tried to fake simple habits as elastic habits, causing type validation errors
	// NEW APPROACH: Each habit type uses its own appropriate scoring method
	scoreResult, err := f.scoringEngine.ScoreSimpleHabit(&habit, value)
	if err != nil {
		return nil, fmt.Errorf("scoring failed: %w", err)
	}

	return &scoreResult.AchievementLevel, nil
}

func (f *SimpleHabitCollectionFlow) determineManualAchievement(habit models.Habit, value interface{}) *models.AchievementLevel {
	// For manual simple habits, determine pass/fail based on field type
	switch habit.FieldType.Type {
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

func (f *SimpleHabitCollectionFlow) collectOptionalNotes(_ models.Habit, _ interface{}, existing *ExistingEntry) (string, error) {
	return collectStandardOptionalNotes(existing)
}

// Elastic Habit Flow Implementations

func (f *ElasticHabitCollectionFlow) displayCriteriaInformation(habit models.Habit) {
	if !habit.RequiresAutomaticScoring() {
		return
	}

	criteriaStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")). // Bright green
		Faint(true).
		Margin(1, 0)

	var parts []string
	if habit.MiniCriteria != nil {
		if value := extractCriteriaDisplayValue(habit.MiniCriteria); value != "" {
			parts = append(parts, fmt.Sprintf("Mini: %s", value))
		}
	}
	if habit.MidiCriteria != nil {
		if value := extractCriteriaDisplayValue(habit.MidiCriteria); value != "" {
			parts = append(parts, fmt.Sprintf("Midi: %s", value))
		}
	}
	if habit.MaxiCriteria != nil {
		if value := extractCriteriaDisplayValue(habit.MaxiCriteria); value != "" {
			parts = append(parts, fmt.Sprintf("Maxi: %s", value))
		}
	}

	if len(parts) > 0 {
		criteriaInfo := criteriaStyle.Render("🎯 Achievement Criteria: " + strings.Join(parts, " • "))
		fmt.Println(criteriaInfo)
	}
}

func (f *ElasticHabitCollectionFlow) performElasticScoring(habit models.Habit, value interface{}) (*models.AchievementLevel, error) {
	if f.scoringEngine == nil {
		return nil, fmt.Errorf("scoring engine not available")
	}

	// Use existing scoring engine for elastic habits with three-tier criteria
	scoreResult, err := f.scoringEngine.ScoreElasticHabit(&habit, value)
	if err != nil {
		return nil, fmt.Errorf("elastic scoring failed: %w", err)
	}

	return &scoreResult.AchievementLevel, nil
}

func (f *ElasticHabitCollectionFlow) collectManualAchievementLevel(habit models.Habit, value interface{}) (*models.AchievementLevel, error) {
	level := models.AchievementNone

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[models.AchievementLevel]().
				Title(fmt.Sprintf("Select achievement level for %s (value: %v):", habit.Title, value)).
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

func (f *ElasticHabitCollectionFlow) displayAchievementResult(habit models.Habit, _ interface{}, level *models.AchievementLevel) {
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

	achievementMsg := fmt.Sprintf("%s %s Achievement: %s", emoji, habit.Title, levelName)

	fmt.Println()
	fmt.Println(style.Render(achievementMsg))
	fmt.Println(style.Render(message))
	fmt.Println()
}

func (f *ElasticHabitCollectionFlow) collectOptionalNotes(_ models.Habit, _ interface{}, existing *ExistingEntry) (string, error) {
	return collectStandardOptionalNotes(existing)
}

// Informational Habit Flow Implementations

func (f *InformationalHabitCollectionFlow) displayInformationalContext(_ models.Habit) {
	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("14")). // Bright cyan
		Faint(true).
		Margin(1, 0)

	contextMsg := "ℹ️  This is an informational habit - for tracking data only (no scoring)"
	fmt.Println(infoStyle.Render(contextMsg))
}

// AIDEV-NOTE: direction-aware-feedback; displays value with directional context for informational habits
// Supports higher_better (green 📈), lower_better (blue 📉), neutral (gray 📊) with contextual hints
func (f *InformationalHabitCollectionFlow) displayDirectionFeedback(habit models.Habit, value interface{}) {
	// Display direction-aware feedback based on habit.Direction configuration
	var style lipgloss.Style
	var emoji string
	var directionHint string

	// Direction-specific styling and feedback
	switch strings.ToLower(habit.Direction) {
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

func (f *InformationalHabitCollectionFlow) collectOptionalNotes(_ models.Habit, _ interface{}, existing *ExistingEntry) (string, error) {
	return collectStandardOptionalNotes(existing)
}

// Checklist Habit Flow Implementations

func (f *ChecklistHabitCollectionFlow) displayChecklistContext(habit models.Habit, existing *ExistingEntry) {
	contextStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("13")). // Bright magenta
		Faint(true).
		Margin(1, 0)

	var contextMsg string
	if existing != nil && existing.Value != nil {
		if items, ok := existing.Value.([]string); ok {
			// Load actual checklist data to get total item count
			checklist, err := f.loadChecklistData(habit)
			if err != nil {
				// Fallback to item count if checklist loading fails
				total := 3 // Default fallback
				completed := len(items)
				contextMsg = fmt.Sprintf("📋 Checklist Progress: %d/%d items completed (loading failed)", completed, total)
			} else {
				total := checklist.GetTotalItemCount()
				completed := len(items)
				contextMsg = fmt.Sprintf("📋 Checklist Progress: %d/%d items completed", completed, total)
			}
		}
	} else {
		contextMsg = "📋 Complete the checklist items below"
	}

	fmt.Println(contextStyle.Render(contextMsg))
}

func (f *ChecklistHabitCollectionFlow) performChecklistScoring(habit models.Habit, value interface{}) (*models.AchievementLevel, error) {
	if f.scoringEngine == nil {
		return nil, fmt.Errorf("scoring engine not available")
	}

	// For checklist habits, validate selected items and perform criteria-based scoring
	selectedItems, ok := value.([]string)
	if !ok {
		return nil, fmt.Errorf("invalid checklist value type: %T", value)
	}

	// Load actual checklist data to get total item count
	checklist, err := f.loadChecklistData(habit)
	if err != nil {
		return nil, fmt.Errorf("failed to load checklist data for scoring: %w", err)
	}

	// Calculate completion metrics
	completed := len(selectedItems)
	total := checklist.GetTotalItemCount()

	if total == 0 {
		level := models.AchievementNone
		return &level, nil
	}

	// For automatic scoring, use criteria-based evaluation
	if habit.Criteria != nil && habit.Criteria.Condition != nil && habit.Criteria.Condition.ChecklistCompletion != nil {
		// Criteria-based scoring using ChecklistCompletionCondition
		condition := habit.Criteria.Condition.ChecklistCompletion

		// Currently only "all" criteria is supported
		if condition.RequiredItems == "all" {
			if completed >= total {
				level := models.AchievementMaxi
				return &level, nil
			}
			level := models.AchievementNone
			return &level, nil
		}
		return nil, fmt.Errorf("unsupported checklist completion criteria: %s", condition.RequiredItems)
	}

	// Fallback to percentage-based scoring if no criteria specified
	percentage := float64(completed) / float64(total)

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

func (f *ChecklistHabitCollectionFlow) collectManualAchievementLevel(habit models.Habit, value interface{}) (*models.AchievementLevel, error) {
	// Similar to elastic habits but with checklist-specific context
	level := models.AchievementNone

	// Calculate completion percentage for context
	var completionInfo string
	if items, ok := value.([]string); ok {
		completed := len(items)

		// Load actual checklist data to get total item count
		checklist, err := f.loadChecklistData(habit)
		if err != nil {
			// Better error handling - show error context to user
			completionInfo = fmt.Sprintf("(%d items selected, checklist data unavailable: %s)", completed, err.Error())
		} else {
			total := checklist.GetTotalItemCount()
			if total == 0 {
				completionInfo = "(empty checklist)"
			} else {
				percentage := float64(completed) / float64(total) * 100
				completionInfo = fmt.Sprintf("(%d/%d items = %.0f%% complete)", completed, total, percentage)
			}
		}
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[models.AchievementLevel]().
				Title(fmt.Sprintf("Select achievement level for %s %s:", habit.Title, completionInfo)).
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

func (f *ChecklistHabitCollectionFlow) displayCompletionProgress(habit models.Habit, value interface{}, level *models.AchievementLevel) {
	if items, ok := value.([]string); ok {
		completed := len(items)

		// Load actual checklist data to get total item count
		checklist, err := f.loadChecklistData(habit)
		var total int
		if err != nil {
			// Fallback to item count if checklist loading fails
			total = 3 // Default fallback
		} else {
			total = checklist.GetTotalItemCount()
		}

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

func (f *ChecklistHabitCollectionFlow) collectOptionalNotes(_ models.Habit, _ interface{}, existing *ExistingEntry) (string, error) {
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
