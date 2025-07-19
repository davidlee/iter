// Package flotsam provides Unix interop functionality for flotsam notes.
// This file contains in-memory collection operations for performance-critical scenarios.
package flotsam

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Collection represents an in-memory collection of flotsam notes with search indices.
// This is used for performance-critical operations like search-as-you-type.
// AIDEV-NOTE: performance-fallback; in-memory collection for when zk shell-out is too slow for interactive UX
//
//revive:disable-next-line:exported Collection is descriptive enough in flotsam package context
type Collection struct {
	Notes []FlotsamNote

	// Search indices for performance
	noteMap  map[string]*FlotsamNote   // Fast lookup by ID
	titleIdx map[string][]*FlotsamNote // Fast title search
	tagIdx   map[string][]*FlotsamNote // Fast tag search

	// Metadata
	Context    string
	LoadedAt   time.Time
	TotalNotes int
	SRSEnabled bool
}

// LoadAllNotes loads all flotsam notes from the context flotsam directory into memory.
// This is used for performance-critical operations when zk shell-out is too slow.
// AIDEV-NOTE: migrated from file_repository.go LoadFlotsam() method
func LoadAllNotes(contextDir string) (*Collection, error) {
	flotsamDir := filepath.Join(contextDir, "flotsam")

	// Create collection
	collection := &Collection{
		Notes:    make([]FlotsamNote, 0),
		noteMap:  make(map[string]*FlotsamNote),
		titleIdx: make(map[string][]*FlotsamNote),
		tagIdx:   make(map[string][]*FlotsamNote),
		Context:  filepath.Base(contextDir),
		LoadedAt: time.Now(),
	}

	// Check if flotsam directory exists
	if _, err := os.Stat(flotsamDir); os.IsNotExist(err) {
		// Directory doesn't exist, return empty collection
		return collection, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to check flotsam directory: %w", err)
	}

	// Walk the flotsam directory to find .md files
	err := filepath.WalkDir(flotsamDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-markdown files
		if d.IsDir() || !strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
			return nil
		}

		// Parse the markdown file
		note, parseErr := ParseFlotsamFile(path)
		if parseErr != nil {
			// Log parsing error but continue with other files
			return parseErr
		}

		// Add note to collection
		collection.addNote(*note)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk flotsam directory: %w", err)
	}

	// AIDEV-NOTE: T041-backlink-removal; backlink computation removed - delegate to zk instead
	// Backlinks now handled by zk delegation: `zk list --linked-by <note>`

	// Update metadata
	collection.computeMetadata()

	return collection, nil
}

// addNote adds a note to the collection and updates indices.
func (c *Collection) addNote(note FlotsamNote) {
	c.Notes = append(c.Notes, note)

	// Update indices
	notePtr := &c.Notes[len(c.Notes)-1]
	c.noteMap[note.ID] = notePtr

	// Title index (case-insensitive)
	titleLower := strings.ToLower(note.Title)
	c.titleIdx[titleLower] = append(c.titleIdx[titleLower], notePtr)

	// Tag index
	for _, tag := range note.Tags {
		tagLower := strings.ToLower(tag)
		c.tagIdx[tagLower] = append(c.tagIdx[tagLower], notePtr)
	}
}

// SearchByTitle performs fast title search using the in-memory index.
// This is used for search-as-you-type functionality.
func (c *Collection) SearchByTitle(query string) []*FlotsamNote {
	if query == "" {
		return []*FlotsamNote{}
	}

	query = strings.ToLower(strings.TrimSpace(query))
	var results []*FlotsamNote

	// Search in title index
	for title, notes := range c.titleIdx {
		if strings.Contains(title, query) {
			results = append(results, notes...)
		}
	}

	// Remove duplicates
	seen := make(map[string]bool)
	uniqueResults := make([]*FlotsamNote, 0)
	for _, note := range results {
		if !seen[note.ID] {
			seen[note.ID] = true
			uniqueResults = append(uniqueResults, note)
		}
	}

	return uniqueResults
}

// FilterByTags filters notes by tags using the in-memory index.
func (c *Collection) FilterByTags(tags []string) []*FlotsamNote {
	if len(tags) == 0 {
		return []*FlotsamNote{}
	}

	var results []*FlotsamNote
	seen := make(map[string]bool)

	for _, tag := range tags {
		tagLower := strings.ToLower(strings.TrimSpace(tag))
		if notes, exists := c.tagIdx[tagLower]; exists {
			for _, note := range notes {
				if !seen[note.ID] {
					seen[note.ID] = true
					results = append(results, note)
				}
			}
		}
	}

	return results
}

// FilterByType filters notes by type using tag-based behavior.
func (c *Collection) FilterByType(noteType string) []*FlotsamNote {
	typeTag := "vice:type:" + noteType
	return c.FilterByTags([]string{typeTag})
}

// GetNotesByDue filters notes that are due for review (requires SRS data).
// This combines in-memory collection with SRS database queries.
func (c *Collection) GetNotesByDue(_ time.Time) []*FlotsamNote {
	// This method would need SRS database integration
	// For now, return notes with vice:srs tag
	return c.FilterByTags([]string{"vice:srs"})
}

// GetNoteByID retrieves a note by ID using the fast lookup index.
func (c *Collection) GetNoteByID(id string) (*FlotsamNote, bool) {
	note, exists := c.noteMap[id]
	return note, exists
}

// AIDEV-NOTE: T041-deprecated; computeBacklinks removed - use zk delegation instead
// Backlink computation now handled by zk commands:
// - `zk list --linked-by <note>` for backlinks
// - `zk list --link-to <note>` for outbound links

// computeMetadata updates the collection's computed metadata fields.
func (c *Collection) computeMetadata() {
	c.TotalNotes = len(c.Notes)

	// Check if any notes have SRS enabled
	c.SRSEnabled = false
	for _, note := range c.Notes {
		if note.HasSRS() {
			c.SRSEnabled = true
			break
		}
	}
}
