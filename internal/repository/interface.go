// Package repository provides data access abstractions for the vice application.
package repository

import (
	"database/sql"
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

	// Flotsam operations (T027 integration)
	// AIDEV-NOTE: T027/3.1-flotsam-repository; context-aware flotsam note operations with ZK compatibility
	// AIDEV-NOTE: interface-extension-complete; 13 methods added for comprehensive flotsam CRUD and query operations
	// AIDEV-NOTE: t028-integration-pattern; follows same context-aware patterns established in T028

	// Collection operations
	LoadFlotsam() (*models.FlotsamCollection, error)
	SaveFlotsam(collection *models.FlotsamCollection) error

	// Individual note CRUD operations
	CreateFlotsamNote(note *models.FlotsamNote) error
	GetFlotsamNote(id string) (*models.FlotsamNote, error)
	UpdateFlotsamNote(note *models.FlotsamNote) error
	DeleteFlotsamNote(id string) error

	// Search and query operations
	SearchFlotsam(query string) ([]*models.FlotsamNote, error)
	GetFlotsamByType(noteType models.FlotsamType) ([]*models.FlotsamNote, error)
	GetFlotsamByTag(tag string) ([]*models.FlotsamNote, error)

	// SRS operations
	GetDueFlotsamNotes() ([]*models.FlotsamNote, error)
	GetFlotsamWithSRS() ([]*models.FlotsamNote, error)

	// Context and path support (T028 integration)
	GetFlotsamDir() (string, error)
	EnsureFlotsamDir() error
	GetFlotsamCacheDB() (*sql.DB, error)
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
