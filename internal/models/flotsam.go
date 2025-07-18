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

	// Flotsam extensions
	Type FlotsamType `yaml:"type,omitempty"` // idea|flashcard|script|log

	// SRS data (flotsam extension, ignored by ZK)
	SRS *flotsam.SRSData `yaml:"srs,omitempty"` // Spaced repetition data
}

// FlotsamType represents the type of flotsam note.
type FlotsamType string

// Flotsam note types define the behavior and intended use of notes.
const (
	IdeaType      FlotsamType = "idea"      // Free-form idea capture
	FlashcardType FlotsamType = "flashcard" // Question/answer cards for SRS
	ScriptType    FlotsamType = "script"    // Executable scripts and commands
	LogType       FlotsamType = "log"       // Journal entries and logs
)

// FlotsamNote represents a complete flotsam note with both frontmatter and content.
// This embeds the flotsam package's FlotsamNote for compatibility while providing
// a models-level interface that integrates with Vice's data patterns.
// AIDEV-NOTE: bridge between flotsam package and Vice models layer
type FlotsamNote struct {
	// Embed the core flotsam note structure
	flotsam.FlotsamNote

	// Additional Vice-specific fields can be added here if needed
}

// FlotsamCollection represents a collection of flotsam notes with metadata.
// This follows the pattern of other Vice collections (Schema, ChecklistSchema).
type FlotsamCollection struct {
	Version     string        `yaml:"version"`           // Schema version for future migrations
	CreatedDate string        `yaml:"created_date"`      // Collection creation date
	Context     string        `yaml:"context,omitempty"` // Vice context for isolation
	Notes       []FlotsamNote `yaml:"notes"`             // Collection of notes

	// Collection-level metadata
	TotalNotes int  `yaml:"-"` // Computed: total note count
	SRSEnabled bool `yaml:"-"` // Computed: any notes have SRS data
}

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

// NewFlotsamFrontmatter creates a new FlotsamFrontmatter with required fields.
// AIDEV-NOTE: constructor ensures ZK compatibility with required fields
func NewFlotsamFrontmatter(id, title string) *FlotsamFrontmatter {
	return &FlotsamFrontmatter{
		ID:      id,
		Title:   title,
		Created: time.Now(),
		Type:    DefaultType(),
		Tags:    make([]string, 0),
	}
}

// NewFlotsamNote creates a new FlotsamNote with the given frontmatter.
func NewFlotsamNote(frontmatter *FlotsamFrontmatter, body, filepath string) *FlotsamNote {
	note := &FlotsamNote{
		FlotsamNote: flotsam.FlotsamNote{
			ID:       frontmatter.ID,
			Title:    frontmatter.Title,
			Type:     string(frontmatter.Type),
			Tags:     frontmatter.Tags,
			Created:  frontmatter.Created,
			Modified: time.Now(),
			Body:     body,
			FilePath: filepath,
			SRS:      frontmatter.SRS,
		},
	}

	return note
}

// NewFlotsamCollection creates a new FlotsamCollection with metadata.
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
		Type:    FlotsamType(fn.Type),
		SRS:     fn.SRS,
	}
}

// UpdateFromFrontmatter updates the note fields from frontmatter data.
func (fn *FlotsamNote) UpdateFromFrontmatter(frontmatter *FlotsamFrontmatter) {
	fn.ID = frontmatter.ID
	fn.Title = frontmatter.Title
	fn.Created = frontmatter.Created
	fn.Tags = frontmatter.Tags
	fn.Type = string(frontmatter.Type)
	fn.SRS = frontmatter.SRS
	fn.Modified = time.Now()
}

// HasSRS returns true if the note has SRS data configured.
func (fn *FlotsamNote) HasSRS() bool {
	return fn.SRS != nil
}

// IsFlashcard returns true if the note is configured as a flashcard.
func (fn *FlotsamNote) IsFlashcard() bool {
	return FlotsamType(fn.Type) == FlashcardType
}

// ValidateType validates the note's type field.
func (fn *FlotsamNote) ValidateType() error {
	if fn.Type == "" {
		fn.Type = string(DefaultType())
		return nil
	}

	return FlotsamType(fn.Type).Validate()
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

	// Validate type (uses existing ValidateType with defaults)
	if err := fn.ValidateType(); err != nil {
		return fmt.Errorf("invalid type: %w", err)
	}

	// Validate SRS data if present
	if fn.SRS != nil {
		if err := validateSRSData(fn.SRS); err != nil {
			return fmt.Errorf("invalid SRS data: %w", err)
		}
	}

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

	// Validate type
	if err := ff.Type.Validate(); err != nil {
		return fmt.Errorf("invalid type: %w", err)
	}

	// Created time is required
	if ff.Created.IsZero() {
		return fmt.Errorf("created timestamp is required")
	}

	// Validate SRS data if present
	if ff.SRS != nil {
		if err := validateSRSData(ff.SRS); err != nil {
			return fmt.Errorf("invalid SRS data: %w", err)
		}
	}

	return nil
}

// Validate validates the FlotsamCollection structure.
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

// validateSRSData validates SRS data structure and bounds.
// AIDEV-NOTE: follows go-srs SM-2 algorithm constraints from internal/flotsam/srs_sm2.go
func validateSRSData(srs *flotsam.SRSData) error {
	if srs == nil {
		return nil // SRS data is optional
	}

	// Easiness bounds from SM-2 algorithm (MinEasiness = 1.3, typical max = 4.0)
	if srs.Easiness < 1.3 || srs.Easiness > 4.0 {
		return fmt.Errorf("easiness %.2f out of bounds: must be between 1.3 and 4.0", srs.Easiness)
	}

	// ConsecutiveCorrect must be non-negative
	if srs.ConsecutiveCorrect < 0 {
		return fmt.Errorf("consecutive_correct %d must be non-negative", srs.ConsecutiveCorrect)
	}

	// TotalReviews must be non-negative
	if srs.TotalReviews < 0 {
		return fmt.Errorf("total_reviews %d must be non-negative", srs.TotalReviews)
	}

	// Due timestamp must be positive (Unix timestamp)
	if srs.Due < 0 {
		return fmt.Errorf("due timestamp %d must be positive", srs.Due)
	}

	// Logical consistency: TotalReviews should be >= ConsecutiveCorrect
	if srs.TotalReviews < srs.ConsecutiveCorrect {
		return fmt.Errorf("total_reviews %d cannot be less than consecutive_correct %d", srs.TotalReviews, srs.ConsecutiveCorrect)
	}

	return nil
}

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
		if FlotsamType(note.Type) == noteType {
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
// AIDEV-NOTE: maintains collection statistics for UI and performance
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
