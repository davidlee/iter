// Package repository provides file-based data repository implementation.
package repository

import (
	"fmt"
	"time"

	"davidlee/vice/internal/config"
	init_pkg "davidlee/vice/internal/init"
	"davidlee/vice/internal/models"
	"davidlee/vice/internal/parser"
	"davidlee/vice/internal/storage"
)

// FileRepository implements DataRepository with file-based storage.
// AIDEV-NOTE: T028/2.1-simple-repository; "turn off and on again" approach for context switching
// AIDEV-NOTE: T028-race-condition-avoidance; complete data unload prevents T024-style concurrency issues
type FileRepository struct {
	viceEnv         *config.ViceEnv
	habitParser     *parser.HabitParser
	entryStorage    *storage.EntryStorage
	fileInitializer *init_pkg.FileInitializer

	// Simple state tracking - no complex caching
	dataLoaded              bool
	currentSchema           *models.Schema
	currentEntries          *models.EntryLog
	currentChecklists       *models.ChecklistSchema
	currentChecklistEntries *models.ChecklistEntriesSchema
}

// NewFileRepository creates a new file-based repository.
func NewFileRepository(viceEnv *config.ViceEnv) *FileRepository {
	return &FileRepository{
		viceEnv:         viceEnv,
		habitParser:     parser.NewHabitParser(),
		entryStorage:    storage.NewEntryStorage(),
		fileInitializer: init_pkg.NewFileInitializer(),
		dataLoaded:      false,
	}
}

// GetCurrentContext returns the active context name.
func (r *FileRepository) GetCurrentContext() string {
	return r.viceEnv.Context
}

// SwitchContext switches to a new context with complete data unload.
// AIDEV-NOTE: T028/2.1-turn-off-on-again; unloads all data, switches context, data loads on next access
// AIDEV-NOTE: T028-context-validation; ensures context exists before switching to prevent invalid states
func (r *FileRepository) SwitchContext(context string) error {
	// Validate context exists in available contexts
	available := r.ListAvailableContexts()
	contextValid := false
	for _, ctx := range available {
		if ctx == context {
			contextValid = true
			break
		}
	}
	if !contextValid {
		return &Error{
			Operation: "SwitchContext",
			Context:   context,
			Err:       fmt.Errorf("context '%s' not found in available contexts %v", context, available),
		}
	}

	// Unload all current data
	if err := r.UnloadAllData(); err != nil {
		return &Error{
			Operation: "SwitchContext",
			Context:   context,
			Err:       fmt.Errorf("failed to unload data: %w", err),
		}
	}

	// Update ViceEnv context
	r.viceEnv.Context = context
	r.viceEnv.ContextData = r.viceEnv.DataDir + "/" + context

	// Ensure new context directory exists
	if err := r.viceEnv.EnsureDirectories(); err != nil {
		return &Error{
			Operation: "SwitchContext",
			Context:   context,
			Err:       fmt.Errorf("failed to create context directories: %w", err),
		}
	}

	return nil
}

// ListAvailableContexts returns all contexts defined in config.toml.
func (r *FileRepository) ListAvailableContexts() []string {
	return r.viceEnv.Contexts
}

// LoadHabits loads the habit schema for the current context.
// AIDEV-NOTE: T028/2.2-file-init; automatically ensures context files exist before loading
func (r *FileRepository) LoadHabits() (*models.Schema, error) {
	if r.currentSchema != nil && r.dataLoaded {
		return r.currentSchema, nil
	}

	// Ensure context files exist before loading
	if err := r.fileInitializer.EnsureContextFiles(r.viceEnv); err != nil {
		return nil, &Error{
			Operation: "LoadHabits",
			Context:   r.viceEnv.Context,
			Err:       fmt.Errorf("failed to ensure context files: %w", err),
		}
	}

	habitsPath := r.viceEnv.GetHabitsFile()
	schema, err := r.habitParser.LoadFromFile(habitsPath)
	if err != nil {
		return nil, &Error{
			Operation: "LoadHabits",
			Context:   r.viceEnv.Context,
			Err:       err,
		}
	}

	r.currentSchema = schema
	r.dataLoaded = true
	return schema, nil
}

// SaveHabits saves the habit schema for the current context.
func (r *FileRepository) SaveHabits(schema *models.Schema) error {
	habitsPath := r.viceEnv.GetHabitsFile()
	if err := r.habitParser.SaveToFile(schema, habitsPath); err != nil {
		return &Error{
			Operation: "SaveHabits",
			Context:   r.viceEnv.Context,
			Err:       err,
		}
	}

	r.currentSchema = schema
	return nil
}

// LoadEntries loads entries for the specified date in the current context.
// AIDEV-NOTE: T028/2.2-file-init; automatically ensures context files exist before loading
func (r *FileRepository) LoadEntries(_ time.Time) (*models.EntryLog, error) {
	// Ensure context files exist before loading
	if err := r.fileInitializer.EnsureContextFiles(r.viceEnv); err != nil {
		return nil, &Error{
			Operation: "LoadEntries",
			Context:   r.viceEnv.Context,
			Err:       fmt.Errorf("failed to ensure context files: %w", err),
		}
	}

	entriesPath := r.viceEnv.GetEntriesFile()
	entries, err := r.entryStorage.LoadFromFile(entriesPath)
	if err != nil {
		return nil, &Error{
			Operation: "LoadEntries",
			Context:   r.viceEnv.Context,
			Err:       err,
		}
	}

	r.currentEntries = entries
	return entries, nil
}

// SaveEntries saves entries for the current context.
func (r *FileRepository) SaveEntries(entries *models.EntryLog) error {
	entriesPath := r.viceEnv.GetEntriesFile()
	if err := r.entryStorage.SaveToFile(entries, entriesPath); err != nil {
		return &Error{
			Operation: "SaveEntries",
			Context:   r.viceEnv.Context,
			Err:       err,
		}
	}

	r.currentEntries = entries
	return nil
}

// LoadChecklists loads checklist templates for the current context.
// AIDEV-NOTE: T028/2.2-file-init; automatically ensures context files exist before loading
func (r *FileRepository) LoadChecklists() (*models.ChecklistSchema, error) {
	// Ensure context files exist before loading
	if err := r.fileInitializer.EnsureContextFiles(r.viceEnv); err != nil {
		return nil, &Error{
			Operation: "LoadChecklists",
			Context:   r.viceEnv.Context,
			Err:       fmt.Errorf("failed to ensure context files: %w", err),
		}
	}

	checklistsPath := r.viceEnv.GetChecklistsFile()

	// Use checklist parser - need to implement this based on existing patterns
	checklistParser := parser.NewChecklistParser()
	checklists, err := checklistParser.LoadFromFile(checklistsPath)
	if err != nil {
		return nil, &Error{
			Operation: "LoadChecklists",
			Context:   r.viceEnv.Context,
			Err:       err,
		}
	}

	r.currentChecklists = checklists
	return checklists, nil
}

// SaveChecklists saves checklist templates for the current context.
func (r *FileRepository) SaveChecklists(checklists *models.ChecklistSchema) error {
	checklistsPath := r.viceEnv.GetChecklistsFile()

	checklistParser := parser.NewChecklistParser()
	if err := checklistParser.SaveToFile(checklists, checklistsPath); err != nil {
		return &Error{
			Operation: "SaveChecklists",
			Context:   r.viceEnv.Context,
			Err:       err,
		}
	}

	r.currentChecklists = checklists
	return nil
}

// LoadChecklistEntries loads checklist entry data for the current context.
func (r *FileRepository) LoadChecklistEntries() (*models.ChecklistEntriesSchema, error) {
	entriesPath := r.viceEnv.GetChecklistEntriesFile()

	entriesParser := parser.NewChecklistEntriesParser()
	entries, err := entriesParser.LoadFromFile(entriesPath)
	if err != nil {
		return nil, &Error{
			Operation: "LoadChecklistEntries",
			Context:   r.viceEnv.Context,
			Err:       err,
		}
	}

	r.currentChecklistEntries = entries
	return entries, nil
}

// SaveChecklistEntries saves checklist entry data for the current context.
func (r *FileRepository) SaveChecklistEntries(entries *models.ChecklistEntriesSchema) error {
	entriesPath := r.viceEnv.GetChecklistEntriesFile()

	entriesParser := parser.NewChecklistEntriesParser()
	if err := entriesParser.SaveToFile(entries, entriesPath); err != nil {
		return &Error{
			Operation: "SaveChecklistEntries",
			Context:   r.viceEnv.Context,
			Err:       err,
		}
	}

	r.currentChecklistEntries = entries
	return nil
}

// UnloadAllData clears all cached data - "turn off and on again" approach.
func (r *FileRepository) UnloadAllData() error {
	r.currentSchema = nil
	r.currentEntries = nil
	r.currentChecklists = nil
	r.currentChecklistEntries = nil
	r.dataLoaded = false
	return nil
}

// IsDataLoaded returns whether any data is currently loaded.
func (r *FileRepository) IsDataLoaded() bool {
	return r.dataLoaded && (r.currentSchema != nil || r.currentEntries != nil ||
		r.currentChecklists != nil || r.currentChecklistEntries != nil)
}
