// Package repository provides data access abstractions for the vice application.
package repository

import (
	"database/sql"
	"time"

	"github.com/davidlee/vice/internal/models"
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

	// Deprecated: Flotsam operations (T027 integration)
	// AIDEV-NOTE: T041-deprecated; repository abstraction layer scheduled for removal
	// AIDEV-NOTE: T041-unix-interop; replaced by direct flotsam package usage + zk delegation
	// Use flotsam.LoadAllNotes() and flotsam.Collection instead

	// Deprecated: Collection operations - use flotsam.LoadAllNotes() instead
	LoadFlotsam() (*models.FlotsamCollection, error)
	SaveFlotsam(collection *models.FlotsamCollection) error

	// Deprecated: Individual note CRUD operations - use flotsam package directly
	CreateFlotsamNote(note *models.FlotsamNote) error
	GetFlotsamNote(id string) (*models.FlotsamNote, error)
	UpdateFlotsamNote(note *models.FlotsamNote) error
	DeleteFlotsamNote(id string) error

	// Deprecated: Search and query operations - use zk delegation or flotsam.SearchNotes()
	SearchFlotsam(query string) ([]*models.FlotsamNote, error)
	GetFlotsamByType(noteType models.FlotsamType) ([]*models.FlotsamNote, error)
	GetFlotsamByTag(tag string) ([]*models.FlotsamNote, error)

	// Deprecated: SRS operations - use flotsam.Collection methods instead
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
