// Package ui provides interactive user interface components for the vice application.
package ui

import (
	"github.com/davidlee/vice/internal/models"
	"github.com/davidlee/vice/internal/scoring"
)

// HabitEntryHandler defines the interface for collecting entries for different habit types.
type HabitEntryHandler interface {
	// CollectEntry collects user input for a habit and returns the result.
	// It handles the complete UI flow including input collection, scoring (if applicable),
	// and achievement display.
	CollectEntry(habit models.Habit, existing *ExistingEntry) (*EntryResult, error)
}

// ExistingEntry represents existing data for a habit entry that can be updated.
type ExistingEntry struct {
	Value            interface{}              // The current value (any type based on field type)
	Notes            string                   // Any existing notes
	AchievementLevel *models.AchievementLevel // Achievement level for elastic habits
}

// EntryResult represents the complete result of collecting an entry for a habit.
type EntryResult struct {
	Value            interface{}              // The collected value (any type based on field type)
	AchievementLevel *models.AchievementLevel // Achievement level for elastic habits (nil for simple habits)
	Notes            string                   // Any notes collected from the user
}

// CreateHabitHandler creates the appropriate handler for a given habit type.
// Returns a handler that can collect entries for the specific habit type.
// AIDEV-NOTE: handler-factory-pattern; current implementation creates habit-specific handlers but needs bubbletea integration (see T010)
func CreateHabitHandler(habit models.Habit, scoringEngine *scoring.Engine) HabitEntryHandler {
	switch habit.HabitType {
	case models.SimpleHabit:
		return NewSimpleHabitHandler()
	case models.ElasticHabit:
		return NewElasticHabitHandler(scoringEngine)
	case models.InformationalHabit:
		return NewInformationalHabitHandler()
	case models.ChecklistHabit:
		// AIDEV-TODO: add ChecklistHabitHandler for checklist habits (T010 implementation)
		return NewSimpleHabitHandler() // Temporary fallback
	default:
		// Fallback to simple habit handler for unknown types
		return NewSimpleHabitHandler()
	}
}
