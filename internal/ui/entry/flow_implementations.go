package entry

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"davidlee/iter/internal/models"
	"davidlee/iter/internal/scoring"
)

// AIDEV-NOTE: flow-implementations; concrete implementations of goal collection flow methods
// Provides scoring, feedback display, and notes collection for each goal type flow

// Simple Goal Flow Implementations

func (f *SimpleGoalCollectionFlow) performAutomaticScoring(goal models.Goal, value interface{}) (*models.AchievementLevel, error) {
	if f.scoringEngine == nil {
		return nil, fmt.Errorf("scoring engine not available")
	}
	
	// Use existing scoring engine for simple goals
	// Simple goals have Pass/Fail achievement levels
	scoreResult, err := f.scoringEngine.ScoreSimpleGoal(&goal, value)
	if err != nil {
		return nil, fmt.Errorf("scoring failed: %w", err)
	}
	
	return &scoreResult.AchievementLevel, nil
}

func (f *SimpleGoalCollectionFlow) determineManualAchievement(goal models.Goal, value interface{}) *models.AchievementLevel {
	// For manual simple goals, determine pass/fail based on field type
	switch goal.FieldType.Type {
	case models.BooleanFieldType:
		if boolVal, ok := value.(bool); ok {
			if boolVal {
				level := models.Pass
				return &level
			} else {
				level := models.Fail
				return &level
			}
		}
	case models.TextFieldType:
		// Text fields require manual scoring (per T009 design decisions)
		// Default to Pass if text is provided
		if textVal, ok := value.(string); ok && strings.TrimSpace(textVal) != "" {
			level := models.Pass
			return &level
		}
	default:
		// Other field types default to Pass if value is provided
		if value != nil {
			level := models.Pass
			return &level
		}
	}
	
	// Default to Fail if no value or false boolean
	level := models.Fail
	return &level
}

func (f *SimpleGoalCollectionFlow) collectOptionalNotes(goal models.Goal, value interface{}, existing *ExistingEntry) (string, error) {
	return collectStandardOptionalNotes(goal, value, existing)
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

func (f *ElasticGoalCollectionFlow) displayAchievementResult(goal models.Goal, value interface{}, level *models.AchievementLevel) {
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

func (f *ElasticGoalCollectionFlow) collectOptionalNotes(goal models.Goal, value interface{}, existing *ExistingEntry) (string, error) {
	return collectStandardOptionalNotes(goal, value, existing)
}

// Informational Goal Flow Implementations

func (f *InformationalGoalCollectionFlow) displayInformationalContext(goal models.Goal) {
	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("14")). // Bright cyan
		Faint(true).
		Margin(1, 0)
	
	contextMsg := "ℹ️  This is an informational goal - for tracking data only (no scoring)"
	fmt.Println(infoStyle.Render(contextMsg))
}

func (f *InformationalGoalCollectionFlow) displayDirectionFeedback(goal models.Goal, value interface{}) {
	// Display direction-aware feedback based on goal configuration
	// This would integrate with goal.Direction field if available
	feedbackStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")). // Gray
		Faint(true)
	
	feedback := fmt.Sprintf("📊 Recorded: %v", value)
	fmt.Println(feedbackStyle.Render(feedback))
}

func (f *InformationalGoalCollectionFlow) collectOptionalNotes(goal models.Goal, value interface{}, existing *ExistingEntry) (string, error) {
	return collectStandardOptionalNotes(goal, value, existing)
}

// Checklist Goal Flow Implementations

func (f *ChecklistGoalCollectionFlow) displayChecklistContext(goal models.Goal, existing *ExistingEntry) {
	contextStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("13")). // Bright magenta
		Faint(true).
		Margin(1, 0)
	
	var contextMsg string
	if existing != nil && existing.Value != nil {
		if items, ok := existing.Value.([]string); ok {
			total := len(goal.FieldType.ChecklistItems)
			completed := len(items)
			contextMsg = fmt.Sprintf("📋 Checklist Progress: %d/%d items completed", completed, total)
		}
	} else {
		contextMsg = "📋 Complete the checklist items below"
	}
	
	fmt.Println(contextStyle.Render(contextMsg))
}

func (f *ChecklistGoalCollectionFlow) performChecklistScoring(goal models.Goal, value interface{}) (*models.AchievementLevel, error) {
	if f.scoringEngine == nil {
		return nil, fmt.Errorf("scoring engine not available")
	}
	
	// Use existing scoring engine for checklist goals
	scoreResult, err := f.scoringEngine.ScoreChecklistGoal(&goal, value)
	if err != nil {
		return nil, fmt.Errorf("checklist scoring failed: %w", err)
	}
	
	return &scoreResult.AchievementLevel, nil
}

func (f *ChecklistGoalCollectionFlow) collectManualAchievementLevel(goal models.Goal, value interface{}) (*models.AchievementLevel, error) {
	// Similar to elastic goals but with checklist-specific context
	level := models.AchievementNone
	
	// Calculate completion percentage for context
	var completionInfo string
	if items, ok := value.([]string); ok && goal.FieldType.ChecklistItems != nil {
		completed := len(items)
		total := len(*goal.FieldType.ChecklistItems)
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

func (f *ChecklistGoalCollectionFlow) displayCompletionProgress(goal models.Goal, value interface{}, level *models.AchievementLevel) {
	if items, ok := value.([]string); ok && goal.FieldType.ChecklistItems != nil {
		completed := len(items)
		total := len(*goal.FieldType.ChecklistItems)
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

func (f *ChecklistGoalCollectionFlow) collectOptionalNotes(goal models.Goal, value interface{}, existing *ExistingEntry) (string, error) {
	return collectStandardOptionalNotes(goal, value, existing)
}

// Common Helper Functions

func collectStandardOptionalNotes(goal models.Goal, value interface{}, existing *ExistingEntry) (string, error) {
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