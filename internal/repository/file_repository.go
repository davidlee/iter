// Package repository provides file-based data repository implementation.
package repository

import (
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/davidlee/vice/internal/config"
	"github.com/davidlee/vice/internal/flotsam"
	init_pkg "github.com/davidlee/vice/internal/init"
	"github.com/davidlee/vice/internal/models"
	"github.com/davidlee/vice/internal/parser"
	"github.com/davidlee/vice/internal/storage"
	"gopkg.in/yaml.v3"
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

// Flotsam operations implementation (T027)
// AIDEV-NOTE: T027/3.2-flotsam-repository; implements files-first architecture per ADR-002

// LoadFlotsam loads all flotsam notes from the context flotsam directory.
// AIDEV-NOTE: T027/3.2.1-load-flotsam; scans .md files and parses ZK-compatible frontmatter
// AIDEV-NOTE: collection-loading-pattern; returns empty collection if directory doesn't exist (not an error)
// AIDEV-NOTE: performance-scanning; uses filepath.WalkDir for efficient recursive scanning
func (r *FileRepository) LoadFlotsam() (*models.FlotsamCollection, error) {
	// Get flotsam directory path
	flotsamDir := r.viceEnv.GetFlotsamDir()

	// Create collection for this context
	collection := models.NewFlotsamCollection(r.viceEnv.Context)

	// Check if flotsam directory exists
	if _, err := os.Stat(flotsamDir); os.IsNotExist(err) {
		// Directory doesn't exist, return empty collection
		return collection, nil
	} else if err != nil {
		return nil, &Error{
			Operation: "LoadFlotsam",
			Context:   r.viceEnv.Context,
			Err:       fmt.Errorf("failed to check flotsam directory: %w", err),
		}
	}

	// Walk the flotsam directory to find .md files
	err := filepath.WalkDir(flotsamDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-markdown files
		if d.IsDir() || !strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
			return nil
		}

		// Parse the markdown file
		note, parseErr := r.parseFlotsamFile(path)
		if parseErr != nil {
			// Log parsing error but continue with other files
			// TODO: Consider adding structured logging
			return parseErr
		}

		// Add note to collection
		collection.AddNote(*note)

		return nil
	})

	if err != nil {
		return nil, &Error{
			Operation: "LoadFlotsam",
			Context:   r.viceEnv.Context,
			Err:       fmt.Errorf("failed to walk flotsam directory: %w", err),
		}
	}

	// Compute backlinks across the entire collection (context-scoped)
	// AIDEV-NOTE: T027/4.2.2-backlink-integration; backlinks computed after note loading for complete collection
	r.computeBacklinks(collection)

	return collection, nil
}

// parseFlotsamFile parses a markdown file and returns a FlotsamNote.
// AIDEV-NOTE: T027/3.2-file-parsing; uses ZK parser for frontmatter + goldmark for links
// AIDEV-NOTE: parsing-pipeline; frontmatter → links → models bridge → validation (core parsing flow)
// AIDEV-NOTE: security-path-validation; validates file path is within flotsam directory to prevent traversal
func (r *FileRepository) parseFlotsamFile(filePath string) (*models.FlotsamNote, error) {
	// Validate file path is within flotsam directory for security
	flotsamDir := r.viceEnv.GetFlotsamDir()
	if !strings.HasPrefix(filePath, flotsamDir) {
		return nil, fmt.Errorf("file path %s is outside flotsam directory", filePath)
	}

	// Read file content
	content, err := os.ReadFile(filePath) // #nosec G304 -- path validated above
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// Parse frontmatter and body using ZK parser
	frontmatter, body, err := flotsam.ParseFrontmatter(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter in %s: %w", filePath, err)
	}

	// Extract links from body content
	linkStructs := flotsam.ExtractLinks(body)

	// Convert []Link to []string (just the href/target for simplicity)
	links := make([]string, 0, len(linkStructs))
	for _, link := range linkStructs {
		links = append(links, link.Href)
	}

	// Get file info for modification time
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info for %s: %w", filePath, err)
	}

	// AIDEV-NOTE: T041-model-simplification; models.FlotsamNote no longer embeds flotsam.FlotsamNote
	// Create FlotsamNote from parsed data using new flat structure
	note := &models.FlotsamNote{
		ID:       frontmatter.ID,
		Title:    frontmatter.Title,
		Tags:     frontmatter.Tags,
		Created:  frontmatter.Created,
		Modified: fileInfo.ModTime(),
		Body:     body,
		FilePath: filePath,
		
		// DEPRECATED: Backward compatibility fields
		Type:      frontmatter.Type, // Will be replaced by vice:type:* tags
		Links:     links,
		SRS:       frontmatter.SRS,
		Backlinks: make([]string, 0), // Will be computed later
	}

	// Validate and set defaults for type
	if err := note.ValidateType(); err != nil {
		return nil, fmt.Errorf("invalid note type in %s: %w", filePath, err)
	}

	return note, nil
}

// computeBacklinks computes backlinks for all notes in a collection.
// AIDEV-NOTE: T027/4.2.2-backlink-computation; context-scoped backlink index using ZK algorithm
// AIDEV-NOTE: backlink-algorithm; builds reverse link map from all note content in collection
func (r *FileRepository) computeBacklinks(collection *models.FlotsamCollection) {
	// Build a map of note ID -> content for backlink computation
	noteContents := make(map[string]string)
	for _, note := range collection.Notes {
		noteContents[note.ID] = note.Body
	}

	// Use ZK's BuildBacklinkIndex to compute reverse links
	// AIDEV-NOTE: zk-algorithm-reuse; leverages proven ZK backlink computation for context-scoped links
	backlinkIndex := flotsam.BuildBacklinkIndex(noteContents)

	// Update each note with its computed backlinks
	for i := range collection.Notes {
		note := &collection.Notes[i]
		if backlinks, exists := backlinkIndex[note.ID]; exists {
			note.Backlinks = backlinks
		} else {
			note.Backlinks = []string{} // Ensure empty slice rather than nil
			// AIDEV-NOTE: empty-slice-pattern; consistent with Vice patterns to use empty slice vs nil
		}
	}
}

// SaveFlotsam saves a flotsam collection to markdown files.
// AIDEV-NOTE: T027/3.2.2-save-flotsam; implements atomic file operations per ADR-002
func (r *FileRepository) SaveFlotsam(collection *models.FlotsamCollection) error {
	if collection == nil {
		return &Error{
			Operation: "SaveFlotsam",
			Context:   r.viceEnv.Context,
			Err:       fmt.Errorf("collection cannot be nil"),
		}
	}

	// Ensure flotsam directory exists
	if err := r.EnsureFlotsamDir(); err != nil {
		return &Error{
			Operation: "SaveFlotsam",
			Context:   r.viceEnv.Context,
			Err:       fmt.Errorf("failed to ensure flotsam directory: %w", err),
		}
	}

	flotsamDir := r.viceEnv.GetFlotsamDir()

	// Save each note as an individual markdown file
	for _, note := range collection.Notes {
		if err := r.saveFlotsamNote(&note, flotsamDir); err != nil {
			return &Error{
				Operation: "SaveFlotsam",
				Context:   r.viceEnv.Context,
				Err:       fmt.Errorf("failed to save note %s: %w", note.ID, err),
			}
		}
	}

	return nil
}

// saveFlotsamNote saves a single flotsam note to a markdown file using atomic operations.
// AIDEV-NOTE: T027/3.2.2-atomic-save; uses temp file + rename for atomic operations
// AIDEV-NOTE: atomic-pattern-core; this is the crash-safe pattern used throughout flotsam for all writes
// AIDEV-NOTE: filename-pattern-zk; uses note.ID + ".md" for ZK compatibility (e.g. "6ub6.md")
func (r *FileRepository) saveFlotsamNote(note *models.FlotsamNote, flotsamDir string) error {
	// Generate filename from note ID
	filename := note.ID + ".md"
	filePath := filepath.Join(flotsamDir, filename)

	// Serialize note to markdown content
	content, err := r.serializeFlotsamNote(note)
	if err != nil {
		return fmt.Errorf("failed to serialize note: %w", err)
	}

	// Write to temporary file first (atomic operation pattern)
	tempPath := filePath + ".tmp"

	// Write content to temp file
	if err := os.WriteFile(tempPath, content, 0o600); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Atomically rename temp file to final location
	if err := os.Rename(tempPath, filePath); err != nil {
		// Clean up temp file on failure
		_ = os.Remove(tempPath)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// serializeFlotsamNote converts a FlotsamNote to markdown content with YAML frontmatter.
// AIDEV-NOTE: T027/3.2.2-serialization; converts models.FlotsamNote to markdown format
func (r *FileRepository) serializeFlotsamNote(note *models.FlotsamNote) ([]byte, error) {
	// Extract frontmatter from note
	frontmatter := note.GetFrontmatter()

	// Convert frontmatter to YAML
	frontmatterYAML, err := yaml.Marshal(frontmatter)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal frontmatter: %w", err)
	}

	// Build complete markdown content
	var content strings.Builder
	content.WriteString("---\n")
	content.Write(frontmatterYAML)
	content.WriteString("---\n")

	// Add body content (ensure it starts with newline)
	if note.Body != "" {
		if !strings.HasPrefix(note.Body, "\n") {
			content.WriteString("\n")
		}
		content.WriteString(note.Body)
	}

	// Ensure file ends with newline
	if !strings.HasSuffix(content.String(), "\n") {
		content.WriteString("\n")
	}

	return []byte(content.String()), nil
}

// CreateFlotsamNote creates a new flotsam note file.
// AIDEV-NOTE: T027/3.2.3-crud-create; atomic file creation with existence check
func (r *FileRepository) CreateFlotsamNote(note *models.FlotsamNote) error {
	if note == nil {
		return &Error{
			Operation: "CreateFlotsamNote",
			Context:   r.viceEnv.Context,
			Err:       fmt.Errorf("note cannot be nil"),
		}
	}

	if note.ID == "" {
		return &Error{
			Operation: "CreateFlotsamNote",
			Context:   r.viceEnv.Context,
			Err:       fmt.Errorf("note ID cannot be empty"),
		}
	}

	// Ensure flotsam directory exists
	if err := r.EnsureFlotsamDir(); err != nil {
		return &Error{
			Operation: "CreateFlotsamNote",
			Context:   r.viceEnv.Context,
			Err:       fmt.Errorf("failed to ensure flotsam directory: %w", err),
		}
	}

	flotsamDir := r.viceEnv.GetFlotsamDir()
	filename := note.ID + ".md"
	filePath := filepath.Join(flotsamDir, filename)

	// Check if file already exists
	if _, err := os.Stat(filePath); err == nil {
		return &Error{
			Operation: "CreateFlotsamNote",
			Context:   r.viceEnv.Context,
			Err:       fmt.Errorf("note with ID %s already exists", note.ID),
		}
	}

	// Use the same atomic save logic as SaveFlotsam
	if err := r.saveFlotsamNote(note, flotsamDir); err != nil {
		return &Error{
			Operation: "CreateFlotsamNote",
			Context:   r.viceEnv.Context,
			Err:       fmt.Errorf("failed to save note: %w", err),
		}
	}

	return nil
}

// GetFlotsamNote retrieves a flotsam note by ID.
// AIDEV-NOTE: T027/3.2.3-crud-read; single note retrieval with existence check
func (r *FileRepository) GetFlotsamNote(id string) (*models.FlotsamNote, error) {
	if id == "" {
		return nil, &Error{
			Operation: "GetFlotsamNote",
			Context:   r.viceEnv.Context,
			Err:       fmt.Errorf("note ID cannot be empty"),
		}
	}

	flotsamDir := r.viceEnv.GetFlotsamDir()
	filename := id + ".md"
	filePath := filepath.Join(flotsamDir, filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, &Error{
			Operation: "GetFlotsamNote",
			Context:   r.viceEnv.Context,
			Err:       fmt.Errorf("note with ID %s not found", id),
		}
	} else if err != nil {
		return nil, &Error{
			Operation: "GetFlotsamNote",
			Context:   r.viceEnv.Context,
			Err:       fmt.Errorf("failed to check note file: %w", err),
		}
	}

	// Use existing parseFlotsamFile method
	note, err := r.parseFlotsamFile(filePath)
	if err != nil {
		return nil, &Error{
			Operation: "GetFlotsamNote",
			Context:   r.viceEnv.Context,
			Err:       fmt.Errorf("failed to parse note: %w", err),
		}
	}

	return note, nil
}

// UpdateFlotsamNote updates an existing flotsam note.
// AIDEV-NOTE: T027/3.2.3-crud-update; atomic update with existence check
func (r *FileRepository) UpdateFlotsamNote(note *models.FlotsamNote) error {
	if note == nil {
		return &Error{
			Operation: "UpdateFlotsamNote",
			Context:   r.viceEnv.Context,
			Err:       fmt.Errorf("note cannot be nil"),
		}
	}

	if note.ID == "" {
		return &Error{
			Operation: "UpdateFlotsamNote",
			Context:   r.viceEnv.Context,
			Err:       fmt.Errorf("note ID cannot be empty"),
		}
	}

	flotsamDir := r.viceEnv.GetFlotsamDir()
	filename := note.ID + ".md"
	filePath := filepath.Join(flotsamDir, filename)

	// Check if file exists (can't update non-existent note)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return &Error{
			Operation: "UpdateFlotsamNote",
			Context:   r.viceEnv.Context,
			Err:       fmt.Errorf("note with ID %s not found", note.ID),
		}
	} else if err != nil {
		return &Error{
			Operation: "UpdateFlotsamNote",
			Context:   r.viceEnv.Context,
			Err:       fmt.Errorf("failed to check note file: %w", err),
		}
	}

	// Update modified time to current time
	note.Modified = time.Now()

	// Use the same atomic save logic as SaveFlotsam
	if err := r.saveFlotsamNote(note, flotsamDir); err != nil {
		return &Error{
			Operation: "UpdateFlotsamNote",
			Context:   r.viceEnv.Context,
			Err:       fmt.Errorf("failed to save updated note: %w", err),
		}
	}

	return nil
}

// DeleteFlotsamNote deletes a flotsam note file.
// AIDEV-NOTE: T027/3.2.3-crud-delete; file deletion with existence check
func (r *FileRepository) DeleteFlotsamNote(id string) error {
	if id == "" {
		return &Error{
			Operation: "DeleteFlotsamNote",
			Context:   r.viceEnv.Context,
			Err:       fmt.Errorf("note ID cannot be empty"),
		}
	}

	flotsamDir := r.viceEnv.GetFlotsamDir()
	filename := id + ".md"
	filePath := filepath.Join(flotsamDir, filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return &Error{
			Operation: "DeleteFlotsamNote",
			Context:   r.viceEnv.Context,
			Err:       fmt.Errorf("note with ID %s not found", id),
		}
	} else if err != nil {
		return &Error{
			Operation: "DeleteFlotsamNote",
			Context:   r.viceEnv.Context,
			Err:       fmt.Errorf("failed to check note file: %w", err),
		}
	}

	// Delete the file
	if err := os.Remove(filePath); err != nil {
		return &Error{
			Operation: "DeleteFlotsamNote",
			Context:   r.viceEnv.Context,
			Err:       fmt.Errorf("failed to delete note file: %w", err),
		}
	}

	return nil
}

// SearchFlotsam searches flotsam notes by query.
func (r *FileRepository) SearchFlotsam(_ string) ([]*models.FlotsamNote, error) {
	// TODO: Implement in later subtask
	return nil, &Error{
		Operation: "SearchFlotsam",
		Context:   r.viceEnv.Context,
		Err:       fmt.Errorf("not yet implemented"),
	}
}

// GetFlotsamByType returns flotsam notes of a specific type.
func (r *FileRepository) GetFlotsamByType(_ models.FlotsamType) ([]*models.FlotsamNote, error) {
	// TODO: Implement in later subtask
	return nil, &Error{
		Operation: "GetFlotsamByType",
		Context:   r.viceEnv.Context,
		Err:       fmt.Errorf("not yet implemented"),
	}
}

// GetFlotsamByTag returns flotsam notes with a specific tag.
func (r *FileRepository) GetFlotsamByTag(_ string) ([]*models.FlotsamNote, error) {
	// TODO: Implement in later subtask
	return nil, &Error{
		Operation: "GetFlotsamByTag",
		Context:   r.viceEnv.Context,
		Err:       fmt.Errorf("not yet implemented"),
	}
}

// GetDueFlotsamNotes returns flotsam notes due for SRS review.
// AIDEV-NOTE: uses SM-2 algorithm to check due dates per ADR-005 quality scale
func (r *FileRepository) GetDueFlotsamNotes() ([]*models.FlotsamNote, error) {
	// Load all flotsam notes
	collection, err := r.LoadFlotsam()
	if err != nil {
		return nil, &Error{
			Operation: "GetDueFlotsamNotes",
			Context:   r.viceEnv.Context,
			Err:       fmt.Errorf("failed to load flotsam collection: %w", err),
		}
	}

	// Create SM-2 calculator for due date checking
	calc := flotsam.NewSM2Calculator()

	// Filter notes that are due for review
	var dueNotes []*models.FlotsamNote
	for _, note := range collection.Notes {
		// Convert models.FlotsamNote.SRS to flotsam.SRSData for algorithm
		var srsData *flotsam.SRSData
		if note.SRS != nil {
			srsData = &flotsam.SRSData{
				Easiness:           note.SRS.Easiness,
				ConsecutiveCorrect: note.SRS.ConsecutiveCorrect,
				Due:                note.SRS.Due,
				TotalReviews:       note.SRS.TotalReviews,
			}
		}

		// Check if note is due using SM-2 algorithm
		if calc.IsDue(srsData) {
			dueNotes = append(dueNotes, &note)
		}
	}

	return dueNotes, nil
}

// GetFlotsamWithSRS returns flotsam notes that have SRS enabled.
// AIDEV-NOTE: filters notes with SRS data per ADR-002 files-first architecture
func (r *FileRepository) GetFlotsamWithSRS() ([]*models.FlotsamNote, error) {
	// Load all flotsam notes
	collection, err := r.LoadFlotsam()
	if err != nil {
		return nil, &Error{
			Operation: "GetFlotsamWithSRS",
			Context:   r.viceEnv.Context,
			Err:       fmt.Errorf("failed to load flotsam collection: %w", err),
		}
	}

	// Filter notes that have SRS data
	var srsNotes []*models.FlotsamNote
	for _, note := range collection.Notes {
		if note.HasSRS() {
			srsNotes = append(srsNotes, &note)
		}
	}

	return srsNotes, nil
}

// GetFlotsamDir returns the context-aware flotsam directory path.
func (r *FileRepository) GetFlotsamDir() (string, error) {
	return r.viceEnv.GetFlotsamDir(), nil
}

// EnsureFlotsamDir ensures the flotsam directory exists.
func (r *FileRepository) EnsureFlotsamDir() error {
	flotsamDir := r.viceEnv.GetFlotsamDir()
	if err := os.MkdirAll(flotsamDir, 0o750); err != nil {
		return &Error{
			Operation: "EnsureFlotsamDir",
			Context:   r.viceEnv.Context,
			Err:       fmt.Errorf("failed to create flotsam directory: %w", err),
		}
	}
	return nil
}

// GetFlotsamCacheDB returns the flotsam SQLite cache database connection.
func (r *FileRepository) GetFlotsamCacheDB() (*sql.DB, error) {
	// TODO: Implement in later subtask (cache implementation)
	return nil, &Error{
		Operation: "GetFlotsamCacheDB",
		Context:   r.viceEnv.Context,
		Err:       fmt.Errorf("not yet implemented"),
	}
}
