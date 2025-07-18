// Package models defines the data structures for the vice application.
// This file contains flotsam note data structures for ZK-compatible note management with SRS.
package models

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/davidlee/vice/internal/flotsam"
)

// FlotsamFrontmatter represents the YAML frontmatter structure for flotsam notes.
// This struct defines the ZK-compatible schema that gets serialized to frontmatter.
// AIDEV-NOTE: ZK-compatible frontmatter with flotsam SRS extensions
type FlotsamFrontmatter struct {
	// ZK standard fields (required for ZK compatibility)
	ID      string    `yaml:"id"`             // ZK 4-char alphanum ID
	Title   string    `yaml:"title"`          // Note title
	Created time.Time `yaml:"created-at"`     // ZK timestamp format
	Tags    []string  `yaml:"tags,omitempty"` // ZK tag array

	// DEPRECATED: Backward compatibility fields
	Type FlotsamType `yaml:"type,omitempty"` // DEPRECATED: Use vice:type:* tags instead
	SRS  *flotsam.SRSData `yaml:"srs,omitempty"` // DEPRECATED: Use SRS database instead
}

// DEPRECATED: FlotsamType enum - use tag-based behavior system instead
// Behaviors are now determined by tags: vice:type:idea, vice:type:flashcard, etc.
// This type is kept for backward compatibility with repository layer.
type FlotsamType string

// DEPRECATED: Flotsam note types - use vice:type:* tags instead
const (
	IdeaType      FlotsamType = "idea"      // Free-form idea capture
	FlashcardType FlotsamType = "flashcard" // Question/answer cards for SRS
	ScriptType    FlotsamType = "script"    // Executable scripts and commands
	LogType       FlotsamType = "log"       // Journal entries and logs
)

// DEPRECATED: FlotsamType methods - use tag-based behavior system instead
// These methods are kept for backward compatibility with existing tests.

// Validate validates a FlotsamType value.
func (ft FlotsamType) Validate() error {
	switch ft {
	case IdeaType, FlashcardType, ScriptType, LogType:
		return nil
	default:
		return fmt.Errorf("invalid flotsam type '%s': must be one of: idea, flashcard, script, log", string(ft))
	}
}

// String returns the string representation of FlotsamType.
func (ft FlotsamType) String() string {
	return string(ft)
}

// IsEmpty returns true if the FlotsamType is empty (not set).
func (ft FlotsamType) IsEmpty() bool {
	return ft == ""
}

// DefaultType returns the default flotsam type when none is specified.
func DefaultType() FlotsamType {
	return IdeaType
}

// FlotsamNote represents a complete flotsam note with both frontmatter and content.
// This is a simplified wrapper around flotsam.FlotsamNote for models-level compatibility.
// AIDEV-NOTE: simplified models-level interface - main logic in flotsam package
type FlotsamNote struct {
	// Core note data (no embedding)
	ID       string    `yaml:"id"`
	Title    string    `yaml:"title"`
	Created  time.Time `yaml:"created-at"`
	Tags     []string  `yaml:"tags,omitempty"`

	// Runtime fields (not in frontmatter)
	Body     string    `yaml:"-"`
	FilePath string    `yaml:"-"`
	Modified time.Time `yaml:"-"`

	// DEPRECATED: Backward compatibility fields
	Type      string           `yaml:"-"` // DEPRECATED: Use vice:type:* tags instead
	Links     []string         `yaml:"-"` // DEPRECATED: Use zk delegation instead
	Backlinks []string         `yaml:"-"` // DEPRECATED: Use zk delegation instead
	SRS       *flotsam.SRSData `yaml:"-"` // DEPRECATED: Use SRS database instead
}

// DEPRECATED: FlotsamCollection - use flotsam.Collection instead
// This type is kept for backward compatibility with repository layer.
type FlotsamCollection struct {
	Version     string        `yaml:"version"`           // Schema version for future migrations
	CreatedDate string        `yaml:"created_date"`      // Collection creation date
	Context     string        `yaml:"context,omitempty"` // Vice context for isolation
	Notes       []FlotsamNote `yaml:"notes"`             // Collection of notes

	// Collection-level metadata
	TotalNotes int  `yaml:"-"` // Computed: total note count
	SRSEnabled bool `yaml:"-"` // Computed: any notes have SRS data
}

// Note: FlotsamType methods removed - use tag-based behavior system instead

// NewFlotsamFrontmatter creates a new FlotsamFrontmatter with required fields.
// AIDEV-NOTE: constructor ensures ZK compatibility with required fields
func NewFlotsamFrontmatter(id, title string) *FlotsamFrontmatter {
	return &FlotsamFrontmatter{
		ID:      id,
		Title:   title,
		Created: time.Now(),
		Tags:    make([]string, 0),
		Type:    DefaultType(), // DEPRECATED: Use vice:type:* tags instead
	}
}

// NewFlotsamNote creates a new FlotsamNote with the given frontmatter.
func NewFlotsamNote(frontmatter *FlotsamFrontmatter, body, filepath string) *FlotsamNote {
	note := &FlotsamNote{
		ID:       frontmatter.ID,
		Title:    frontmatter.Title,
		Tags:     frontmatter.Tags,
		Created:  frontmatter.Created,
		Modified: time.Now(),
		Body:     body,
		FilePath: filepath,
		
		// Initialize deprecated fields
		Type:      "idea", // Default type
		Links:     make([]string, 0),
		Backlinks: make([]string, 0),
		SRS:       nil,
	}

	return note
}

// DEPRECATED: NewFlotsamCollection - use flotsam.LoadAllNotes() instead
// This constructor is kept for backward compatibility with repository layer.
func NewFlotsamCollection(context string) *FlotsamCollection {
	return &FlotsamCollection{
		Version:     "1.0",
		CreatedDate: time.Now().Format("2006-01-02"),
		Context:     context,
		Notes:       make([]FlotsamNote, 0),
		TotalNotes:  0,
		SRSEnabled:  false,
	}
}

// GetFrontmatter extracts the frontmatter from a FlotsamNote.
// AIDEV-NOTE: bridge method to convert between note and frontmatter representations
func (fn *FlotsamNote) GetFrontmatter() *FlotsamFrontmatter {
	return &FlotsamFrontmatter{
		ID:      fn.ID,
		Title:   fn.Title,
		Created: fn.Created,
		Tags:    fn.Tags,
		Type:    FlotsamType(fn.Type), // DEPRECATED: Use vice:type:* tags instead
		SRS:     fn.SRS,               // DEPRECATED: Use SRS database instead
	}
}

// UpdateFromFrontmatter updates the note fields from frontmatter data.
func (fn *FlotsamNote) UpdateFromFrontmatter(frontmatter *FlotsamFrontmatter) {
	fn.ID = frontmatter.ID
	fn.Title = frontmatter.Title
	fn.Created = frontmatter.Created
	fn.Tags = frontmatter.Tags
	fn.Type = string(frontmatter.Type) // DEPRECATED: Use vice:type:* tags instead
	fn.SRS = frontmatter.SRS           // DEPRECATED: Use SRS database instead
	fn.Modified = time.Now()
}

// HasSRS returns true if the note has SRS data configured.
// Uses tag-based detection for Unix interop approach, with backward compatibility.
func (fn *FlotsamNote) HasSRS() bool {
	return fn.HasTag("vice:srs") || fn.SRS != nil
}

// IsFlashcard returns true if the note is configured as a flashcard.
// Uses tag-based detection for Unix interop approach.
func (fn *FlotsamNote) IsFlashcard() bool {
	return fn.HasTag("vice:type:flashcard")
}

// HasTag returns true if the note has the specified tag.
func (fn *FlotsamNote) HasTag(tag string) bool {
	for _, t := range fn.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// HasType returns true if the note has the specified type tag.
func (fn *FlotsamNote) HasType(noteType string) bool {
	return fn.HasTag("vice:type:" + noteType)
}

// DEPRECATED: ValidateType - use tag-based validation instead
// This method is kept for backward compatibility with repository layer.
func (fn *FlotsamNote) ValidateType() error {
	// No-op: type validation is now handled by tags
	return nil
}

// Validate validates the entire FlotsamNote structure.
// AIDEV-NOTE: comprehensive validation following habit.go patterns
func (fn *FlotsamNote) Validate() error {
	return fn.validateInternal()
}

// validateInternal performs the core validation logic for FlotsamNote.
func (fn *FlotsamNote) validateInternal() error {
	// Validate ID format (ZK-compatible: 4-char alphanum)
	if !isValidFlotsamID(fn.ID) {
		return fmt.Errorf("flotsam ID '%s' is invalid: must be 4-character alphanumeric (ZK-compatible)", fn.ID)
	}

	// Title is required
	if strings.TrimSpace(fn.Title) == "" {
		return fmt.Errorf("title is required")
	}

	// Note: Type validation removed - tags handle behavior

	// Note: SRS data validation removed - stored in SQLite database

	// Created time should not be zero
	if fn.Created.IsZero() {
		return fmt.Errorf("created timestamp is required")
	}

	// Modified time should not be before created time
	if !fn.Modified.IsZero() && fn.Modified.Before(fn.Created) {
		return fmt.Errorf("modified time cannot be before created time")
	}

	return nil
}

// Validate validates the FlotsamFrontmatter structure.
func (ff *FlotsamFrontmatter) Validate() error {
	// Validate ID format
	if !isValidFlotsamID(ff.ID) {
		return fmt.Errorf("flotsam ID '%s' is invalid: must be 4-character alphanumeric (ZK-compatible)", ff.ID)
	}

	// Title is required
	if strings.TrimSpace(ff.Title) == "" {
		return fmt.Errorf("title is required")
	}

	// Created time is required
	if ff.Created.IsZero() {
		return fmt.Errorf("created timestamp is required")
	}

	// DEPRECATED: Validate type for backward compatibility
	if ff.Type != "" {
		if err := ff.Type.Validate(); err != nil {
			return fmt.Errorf("invalid type: %w", err)
		}
	}

	// Note: SRS validation removed - use tag-based behaviors

	return nil
}

// DEPRECATED: FlotsamCollection validation - use flotsam.Collection instead
// This validation is kept for backward compatibility with repository layer.
func (fc *FlotsamCollection) Validate() error {
	// Context is required for isolation
	if strings.TrimSpace(fc.Context) == "" {
		return fmt.Errorf("context is required for collection isolation")
	}

	// Validate each note in the collection
	idsSeen := make(map[string]bool)
	for i, note := range fc.Notes {
		// Validate individual note
		if err := note.validateInternal(); err != nil {
			return fmt.Errorf("note %d validation failed: %w", i, err)
		}

		// Check for duplicate IDs within collection
		if idsSeen[note.ID] {
			return fmt.Errorf("duplicate note ID '%s' found in collection", note.ID)
		}
		idsSeen[note.ID] = true
	}

	return nil
}

// isValidFlotsamID validates flotsam ID format (ZK-compatible: 4-char alphanum).
// AIDEV-NOTE: follows ZK ID generation pattern from internal/flotsam/zk_id.go
func isValidFlotsamID(id string) bool {
	if id == "" {
		return false
	}
	// ZK-compatible: exactly 4 characters, alphanumeric, lowercase
	matched, _ := regexp.MatchString(`^[a-z0-9]{4}$`, id)
	return matched
}

// Note: validateSRSData removed - SRS validation handled in flotsam package

// DEPRECATED: FlotsamCollection methods - use flotsam.Collection instead
// These methods are kept for backward compatibility with repository layer.

// AddNote adds a note to the collection and updates metadata.
func (fc *FlotsamCollection) AddNote(note FlotsamNote) {
	fc.Notes = append(fc.Notes, note)
	fc.computeMetadata()
}

// RemoveNote removes a note by ID from the collection.
func (fc *FlotsamCollection) RemoveNote(id string) bool {
	for i, note := range fc.Notes {
		if note.ID == id {
			fc.Notes = append(fc.Notes[:i], fc.Notes[i+1:]...)
			fc.computeMetadata()
			return true
		}
	}
	return false
}

// GetNote retrieves a note by ID from the collection.
func (fc *FlotsamCollection) GetNote(id string) (*FlotsamNote, bool) {
	for i := range fc.Notes {
		if fc.Notes[i].ID == id {
			return &fc.Notes[i], true
		}
	}
	return nil, false
}

// GetNotesByType returns all notes of the specified type.
func (fc *FlotsamCollection) GetNotesByType(noteType FlotsamType) []FlotsamNote {
	var notes []FlotsamNote
	for _, note := range fc.Notes {
		if note.HasType(string(noteType)) {
			notes = append(notes, note)
		}
	}
	return notes
}

// GetSRSNotes returns all notes that have SRS data configured.
func (fc *FlotsamCollection) GetSRSNotes() []FlotsamNote {
	var notes []FlotsamNote
	for _, note := range fc.Notes {
		if note.HasSRS() {
			notes = append(notes, note)
		}
	}
	return notes
}

// computeMetadata updates the collection's computed metadata fields.
func (fc *FlotsamCollection) computeMetadata() {
	fc.TotalNotes = len(fc.Notes)

	// Check if any notes have SRS enabled
	fc.SRSEnabled = false
	for _, note := range fc.Notes {
		if note.HasSRS() {
			fc.SRSEnabled = true
			break
		}
	}
}
