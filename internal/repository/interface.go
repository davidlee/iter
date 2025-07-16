// Package repository provides data access abstractions for the vice application.
package repository

import (
	"time"

	"davidlee/vice/internal/models"
)

// DataRepository defines the interface for data access operations.
// AIDEV-NOTE: T028/2.1-repository-pattern; abstraction for context-aware data loading with clear migration path
// AIDEV-NOTE: T028-repository-interface; enables staged evolution from simple file access to sophisticated lazy loading
type DataRepository interface {
	// Context management
	GetCurrentContext() string
	SwitchContext(context string) error
	ListAvailableContexts() []string

	// Habit data operations
	LoadHabits() (*models.Schema, error)
	SaveHabits(schema *models.Schema) error

	// Entry data operations
	LoadEntries(date time.Time) (*models.EntryLog, error)
	SaveEntries(entries *models.EntryLog) error

	// Checklist operations
	LoadChecklists() (*models.ChecklistSchema, error)
	SaveChecklists(checklists *models.ChecklistSchema) error
	LoadChecklistEntries() (*models.ChecklistEntriesSchema, error)
	SaveChecklistEntries(entries *models.ChecklistEntriesSchema) error

	// Data lifecycle management
	UnloadAllData() error
	IsDataLoaded() bool
}

// Error represents errors from repository operations.
type Error struct {
	Operation string
	Context   string
	Err       error
}

func (e *Error) Error() string {
	return "repository error in " + e.Operation + " for context '" + e.Context + "': " + e.Err.Error()
}

func (e *Error) Unwrap() error {
	return e.Err
}
