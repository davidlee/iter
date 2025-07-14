// Package ui provides interactive user interface components for the vice application.
package ui

import (
	"davidlee/vice/internal/models"
	"davidlee/vice/internal/scoring"
)

// GoalEntryHandler defines the interface for collecting entries for different goal types.
type GoalEntryHandler interface {
	// CollectEntry collects user input for a goal and returns the result.
	// It handles the complete UI flow including input collection, scoring (if applicable),
	// and achievement display.
	CollectEntry(goal models.Goal, existing *ExistingEntry) (*EntryResult, error)
}

// ExistingEntry represents existing data for a goal entry that can be updated.
type ExistingEntry struct {
	Value            interface{}              // The current value (any type based on field type)
	Notes            string                   // Any existing notes
	AchievementLevel *models.AchievementLevel // Achievement level for elastic goals
}

// EntryResult represents the complete result of collecting an entry for a goal.
type EntryResult struct {
	Value            interface{}              // The collected value (any type based on field type)
	AchievementLevel *models.AchievementLevel // Achievement level for elastic goals (nil for simple goals)
	Notes            string                   // Any notes collected from the user
}

// CreateGoalHandler creates the appropriate handler for a given goal type.
// Returns a handler that can collect entries for the specific goal type.
// AIDEV-NOTE: handler-factory-pattern; current implementation creates goal-specific handlers but needs bubbletea integration (see T010)
func CreateGoalHandler(goal models.Goal, scoringEngine *scoring.Engine) GoalEntryHandler {
	switch goal.GoalType {
	case models.SimpleGoal:
		return NewSimpleGoalHandler()
	case models.ElasticGoal:
		return NewElasticGoalHandler(scoringEngine)
	case models.InformationalGoal:
		return NewInformationalGoalHandler()
	case models.ChecklistGoal:
		// AIDEV-TODO: add ChecklistGoalHandler for checklist goals (T010 implementation)
		return NewSimpleGoalHandler() // Temporary fallback
	default:
		// Fallback to simple goal handler for unknown types
		return NewSimpleGoalHandler()
	}
}
