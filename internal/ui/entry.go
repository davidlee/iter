// Package ui provides interactive user interface components for the vice application.
package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"

	"davidlee/vice/internal/models"
	"davidlee/vice/internal/parser"
	"davidlee/vice/internal/scoring"
	"davidlee/vice/internal/storage"
	"davidlee/vice/internal/ui/entry"
)

// EntryCollector handles the interactive collection of today's habit entries.
// AIDEV-NOTE: T010-entry-system-complete; All habit collection flows with field input components and scoring integration
// Architecture: Uses habit collection flows from internal/ui/entry/ package with complete scoring engine integration
type EntryCollector struct {
	goalParser    *parser.HabitParser
	entryStorage  *storage.EntryStorage
	scoringEngine *scoring.Engine
	flowFactory   *entry.HabitCollectionFlowFactory
	habits        []models.Habit
	entries       map[string]interface{}              // Stores raw values for all habit types
	achievements  map[string]*models.AchievementLevel // Stores achievement levels for elastic habits
	notes         map[string]string
	statuses      map[string]models.EntryStatus // T012/2.1-enhanced: Stores entry completion status for skip functionality
}

// NewEntryCollector creates a new entry collector instance.
func NewEntryCollector(checklistsPath string) *EntryCollector {
	scoringEngine := scoring.NewEngine()
	fieldInputFactory := entry.NewEntryFieldInputFactory()
	flowFactory := entry.NewHabitCollectionFlowFactory(fieldInputFactory, scoringEngine, checklistsPath)

	return &EntryCollector{
		goalParser:    parser.NewHabitParser(),
		entryStorage:  storage.NewEntryStorage(),
		scoringEngine: scoringEngine,
		flowFactory:   flowFactory,
		entries:       make(map[string]interface{}),
		achievements:  make(map[string]*models.AchievementLevel),
		notes:         make(map[string]string),
		statuses:      make(map[string]models.EntryStatus),
	}
}

// CollectTodayEntries runs the interactive UI to collect today's habit entries.
func (ec *EntryCollector) CollectTodayEntries(habitsFile, entriesFile string) error {
	// Load habit schema
	schema, err := ec.goalParser.LoadFromFile(habitsFile)
	if err != nil {
		return fmt.Errorf("failed to load habits: %w", err)
	}

	// Get all habits (simple, elastic, and informational)
	ec.habits = schema.Habits
	if len(ec.habits) == 0 {
		return fmt.Errorf("no habits found in %s", habitsFile)
	}

	// Load existing entries for today (if any)
	if err := ec.loadExistingEntries(entriesFile); err != nil {
		return fmt.Errorf("failed to load existing entries: %w", err)
	}

	// Display welcome message
	ec.displayWelcome()

	// Collect entries for each habit
	for _, habit := range ec.habits {
		if err := ec.collectHabitEntry(habit); err != nil {
			return fmt.Errorf("failed to collect entry for habit %s: %w", habit.ID, err)
		}
	}

	// Save entries
	if err := ec.saveEntries(entriesFile); err != nil {
		return fmt.Errorf("failed to save entries: %w", err)
	}

	// Display completion message
	ec.displayCompletion()

	return nil
}

// loadExistingEntries loads any existing entries for today.
func (ec *EntryCollector) loadExistingEntries(entriesFile string) error {
	today := time.Now().Format("2006-01-02")

	dayEntry, err := ec.entryStorage.GetDayEntry(entriesFile, today)
	if err != nil {
		// No existing entries for today, which is fine
		return nil //nolint:nilerr // No entries for today is expected
	}

	// Load existing entries into our maps
	for _, goalEntry := range dayEntry.Habits {
		ec.entries[goalEntry.HabitID] = goalEntry.Value
		ec.notes[goalEntry.HabitID] = goalEntry.Notes
		ec.statuses[goalEntry.HabitID] = goalEntry.Status

		// Load achievement level for elastic habits
		if goalEntry.AchievementLevel != nil {
			ec.achievements[goalEntry.HabitID] = goalEntry.AchievementLevel
		}
	}

	return nil
}

// AIDEV-NOTE: T010/4.1-scoring-integration-complete; uses habit collection flows with full scoring engine integration
// collectHabitEntry collects the entry for a single habit using the appropriate collection flow.
func (ec *EntryCollector) collectHabitEntry(habit models.Habit) error {
	// Create existing entry data from our maps
	var existing *entry.ExistingEntry
	if value, hasValue := ec.entries[habit.ID]; hasValue {
		existing = &entry.ExistingEntry{
			Value:            value,
			Notes:            ec.notes[habit.ID],
			AchievementLevel: ec.achievements[habit.ID],
		}
	}

	// Create the appropriate collection flow for this habit type
	flow, err := ec.flowFactory.CreateFlow(string(habit.HabitType))
	if err != nil {
		return fmt.Errorf("failed to create collection flow for habit %s: %w", habit.ID, err)
	}

	// Use the flow to collect the entry with full scoring integration
	result, err := flow.CollectEntry(habit, existing)
	if err != nil {
		return fmt.Errorf("failed to collect entry for habit %s: %w", habit.ID, err)
	}

	// Store the results in our maps
	ec.entries[habit.ID] = result.Value
	ec.notes[habit.ID] = result.Notes
	ec.statuses[habit.ID] = result.Status

	// Store achievement level if present (for elastic habits)
	if result.AchievementLevel != nil {
		ec.achievements[habit.ID] = result.AchievementLevel
	}

	return nil
}

// saveEntries saves all collected entries to the entries file.
func (ec *EntryCollector) saveEntries(entriesFile string) error {
	today := time.Now().Format("2006-01-02")

	// Create habit entries from collected data
	var goalEntries []models.HabitEntry
	for _, habit := range ec.habits {
		value, exists := ec.entries[habit.ID]
		if !exists {
			continue // Skip habits that weren't processed
		}

		goalEntry := models.HabitEntry{
			HabitID:          habit.ID,
			Value:            value,
			AchievementLevel: ec.achievements[habit.ID], // Will be nil for simple/informational habits
			Notes:            ec.notes[habit.ID],
			Status:           ec.statuses[habit.ID], // Use collected status
		}
		goalEntry.MarkCreated()

		goalEntries = append(goalEntries, goalEntry)
	}

	// Create day entry
	dayEntry := models.DayEntry{
		Date:   today,
		Habits: goalEntries,
	}

	// Save to storage
	if err := ec.entryStorage.UpdateDayEntry(entriesFile, dayEntry); err != nil {
		return fmt.Errorf("failed to update day entry: %w", err)
	}

	return nil
}

// displayWelcome shows a welcome message with today's date.
func (ec *EntryCollector) displayWelcome() {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("14")). // Bright cyan
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2).
		Margin(1, 0)

	today := time.Now().Format("Monday, January 2, 2006")
	welcome := fmt.Sprintf("🎯 Habit Tracker - %s", today)

	goalCountStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")). // Gray
		Margin(0, 0, 1, 0)

	goalCount := goalCountStyle.Render(fmt.Sprintf("Ready to track %d habits for today!", len(ec.habits)))

	fmt.Println(headerStyle.Render(welcome))
	fmt.Println(goalCount)
}

// displayCompletion shows a completion message with summary.
func (ec *EntryCollector) displayCompletion() {
	completedCount := 0
	totalCount := len(ec.habits)

	// Count completions based on habit type and value
	for goalID, value := range ec.entries {
		// Find the habit to determine how to interpret completion
		var habit *models.Habit
		for i := range ec.habits {
			if ec.habits[i].ID == goalID {
				habit = &ec.habits[i]
				break
			}
		}

		if habit == nil {
			continue
		}

		// Determine if this habit is "completed" based on its type
		switch habit.HabitType {
		case models.SimpleHabit:
			// Simple habits: check boolean value
			if boolVal, ok := value.(bool); ok && boolVal {
				completedCount++
			}
		case models.ElasticHabit:
			// Elastic habits: consider any achievement level as completion
			if achievementLevel := ec.achievements[goalID]; achievementLevel != nil && *achievementLevel != models.AchievementNone {
				completedCount++
			}
		case models.InformationalHabit:
			// Informational habits: any non-empty value counts as completion
			if value != nil && fmt.Sprintf("%v", value) != "" {
				completedCount++
			}
		}
	}

	// Choose appropriate styling based on completion rate
	var completionStyle lipgloss.Style
	var emoji string

	completionRate := float64(completedCount) / float64(totalCount)
	switch {
	case completionRate == 1.0:
		completionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")) // Bright green
		emoji = "🎉"
	case completionRate >= 0.7:
		completionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11")) // Bright yellow
		emoji = "💪"
	case completionRate >= 0.5:
		completionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("12")) // Bright blue
		emoji = "👍"
	default:
		completionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8")) // Gray
		emoji = "🤗"
	}

	summaryStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2).
		Margin(1, 0)

	summary := fmt.Sprintf("%s Completed %d out of %d habits today!", emoji, completedCount, totalCount)

	// Add motivational message
	var message string
	switch {
	case completionRate == 1.0:
		message = "Perfect day! You're building amazing habits! ✨"
	case completionRate >= 0.7:
		message = "Great job! You're making excellent progress! 🚀"
	case completionRate >= 0.5:
		message = "Good work! Every step counts towards your habits! 📈"
	default:
		message = "Tomorrow is a new opportunity to build your habits! 🌅"
	}

	messageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Italic(true).
		Margin(1, 0, 0, 0)

	fmt.Println(summaryStyle.Render(completionStyle.Render(summary)))
	fmt.Println(messageStyle.Render(message))

	// Show saved location
	savedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Faint(true)

	fmt.Println(savedStyle.Render("✅ Entries saved successfully!"))
}

// timePtr creates a pointer to a time.Time value.
func timePtr(t time.Time) *time.Time {
	return &t
}

// CollectSingleHabitEntry collects an entry for a single habit, used by the entry menu interface.
// AIDEV-NOTE: T018/3.1-entry-integration; main integration point for menu→entry flow
// This method is called when user presses Enter in entry menu to collect entry for selected habit
func (ec *EntryCollector) CollectSingleHabitEntry(habit models.Habit) error {
	return ec.collectHabitEntry(habit)
}

// GetHabitEntry returns the current entry data for a habit.
// AIDEV-NOTE: T018/3.1-state-sync; used by menu to sync state after entry collection
func (ec *EntryCollector) GetHabitEntry(goalID string) (interface{}, string, *models.AchievementLevel, models.EntryStatus, bool) {
	value, hasValue := ec.entries[goalID]
	notes := ec.notes[goalID]
	achievement := ec.achievements[goalID]
	status, hasStatus := ec.statuses[goalID]

	return value, notes, achievement, status, hasValue && hasStatus
}

// InitializeForMenu initializes the EntryCollector with habits and existing entries for menu usage.
// AIDEV-NOTE: T018/3.1-menu-setup; critical setup for menu integration - must be called before habit selection
// Converts HabitEntry format to internal collector format (interface{} values)
func (ec *EntryCollector) InitializeForMenu(habits []models.Habit, entries map[string]models.HabitEntry) {
	ec.habits = habits

	// Initialize maps
	ec.entries = make(map[string]interface{})
	ec.achievements = make(map[string]*models.AchievementLevel)
	ec.notes = make(map[string]string)
	ec.statuses = make(map[string]models.EntryStatus)

	// Load existing entries into collector format
	for _, entry := range entries {
		ec.entries[entry.HabitID] = entry.Value
		ec.notes[entry.HabitID] = entry.Notes
		ec.statuses[entry.HabitID] = entry.Status
		if entry.AchievementLevel != nil {
			ec.achievements[entry.HabitID] = entry.AchievementLevel
		}
	}
}

// SaveEntriesToFile saves the current entries to the specified file.
// AIDEV-NOTE: T018/3.2-auto-save; called after each habit completion for automatic persistence
// Reuses existing saveEntries() method for consistency with main entry flow
func (ec *EntryCollector) SaveEntriesToFile(entriesFile string) error {
	return ec.saveEntries(entriesFile)
}

// StoreEntryResult stores an entry result from modal processing into the collector.
// AIDEV-NOTE: T024-modal-integration; stores modal results in collector for menu state sync
func (ec *EntryCollector) StoreEntryResult(goalID string, result *entry.EntryResult) {
	ec.entries[goalID] = result.Value
	ec.notes[goalID] = result.Notes
	ec.statuses[goalID] = result.Status

	// Store achievement level if present (for elastic habits)
	if result.AchievementLevel != nil {
		ec.achievements[goalID] = result.AchievementLevel
	}
}

// Testing helpers - these methods are only used in tests

// SetHabitsForTesting sets the habits for testing purposes.
func (ec *EntryCollector) SetHabitsForTesting(habits []models.Habit) {
	ec.habits = habits
}

// SetEntryForTesting sets an entry for testing purposes.
func (ec *EntryCollector) SetEntryForTesting(goalID string, value interface{}, achievementLevel *models.AchievementLevel, notes string) {
	ec.entries[goalID] = value
	ec.notes[goalID] = notes
	if achievementLevel != nil {
		ec.achievements[goalID] = achievementLevel
	}

	// Determine status based on value and achievement level
	if value == nil {
		ec.statuses[goalID] = models.EntrySkipped
	} else if boolVal, ok := value.(bool); ok {
		if boolVal {
			ec.statuses[goalID] = models.EntryCompleted
		} else {
			ec.statuses[goalID] = models.EntryFailed
		}
	} else {
		// Non-boolean values default to completed
		ec.statuses[goalID] = models.EntryCompleted
	}
}

// SaveEntriesForTesting saves entries for testing purposes.
func (ec *EntryCollector) SaveEntriesForTesting(entriesFile string) error {
	return ec.saveEntries(entriesFile)
}

// LoadExistingEntriesForTesting loads existing entries for testing purposes.
func (ec *EntryCollector) LoadExistingEntriesForTesting(entriesFile string) error {
	return ec.loadExistingEntries(entriesFile)
}

// GetEntriesForTesting returns the entries map for testing purposes.
func (ec *EntryCollector) GetEntriesForTesting() map[string]interface{} {
	return ec.entries
}

// GetAchievementsForTesting returns the achievements map for testing purposes.
func (ec *EntryCollector) GetAchievementsForTesting() map[string]*models.AchievementLevel {
	return ec.achievements
}
