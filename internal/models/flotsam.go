// Package models defines the data structures for the vice application.
// This file contains flotsam note data structures for ZK-compatible note management with SRS.
package models

import (
	"fmt"
	"time"

	"davidlee/vice/internal/flotsam"
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
