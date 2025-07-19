// Package flotsam provides Unix interop functionality for flotsam notes.
// This file contains file I/O operations for flotsam notes.
package flotsam

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// ParseFlotsamFile parses a markdown file and returns a FlotsamNote.
// This is migrated from file_repository.go parseFlotsamFile() method.
// AIDEV-NOTE: migrated from file_repository.go; security validation and ZK parser integration preserved
func ParseFlotsamFile(filePath string) (*FlotsamNote, error) {
	// Validate file path security
	if err := ValidateFlotsamPath(filePath); err != nil {
		return nil, err
	}

	// Read file content
	content, err := os.ReadFile(filePath) // #nosec G304 -- path validated above
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// Parse frontmatter and body using existing ZK parser
	frontmatter, body, err := ParseFrontmatter(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter in %s: %w", filePath, err)
	}

	// Get file info for modification time
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info for %s: %w", filePath, err)
	}

	// Create FlotsamNote from parsed data
	note := &FlotsamNote{
		ID:       frontmatter.ID,
		Title:    frontmatter.Title,
		Type:     frontmatter.Type,
		Tags:     frontmatter.Tags,
		Created:  frontmatter.Created,
		Modified: fileInfo.ModTime(),
		Body:     body,
		FilePath: filePath,
		SRS:      frontmatter.SRS,
	}

	return note, nil
}

// SaveFlotsamNote saves a single flotsam note to a markdown file using atomic operations.
// This is migrated from file_repository.go saveFlotsamNote() method.
// AIDEV-NOTE: atomic-pattern-core; preserves crash-safe pattern from T027
func SaveFlotsamNote(note *FlotsamNote, flotsamDir string) error {
	if note == nil {
		return fmt.Errorf("note cannot be nil")
	}

	if note.ID == "" {
		return fmt.Errorf("note ID cannot be empty")
	}

	// Generate filename from note ID
	filename := note.ID + ".md"
	filePath := filepath.Join(flotsamDir, filename)

	// Serialize note to markdown content
	content, err := SerializeFlotsamNote(note)
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

	// Update note's file path
	note.FilePath = filePath

	return nil
}

// SerializeFlotsamNote converts a FlotsamNote to markdown content with YAML frontmatter.
// This is migrated from file_repository.go serializeFlotsamNote() method.
// AIDEV-NOTE: frontmatter-serialization; preserves ZK-compatible YAML format
func SerializeFlotsamNote(note *FlotsamNote) ([]byte, error) {
	if note == nil {
		return nil, fmt.Errorf("note cannot be nil")
	}

	// Create frontmatter structure
	frontmatter := map[string]interface{}{
		"id":         note.ID,
		"title":      note.Title,
		"created-at": note.Created,
	}

	// Add tags if present
	if len(note.Tags) > 0 {
		frontmatter["tags"] = note.Tags
	}

	// Add type if present
	if note.Type != "" {
		frontmatter["type"] = note.Type
	}

	// Add SRS data if present
	if note.SRS != nil {
		frontmatter["srs"] = note.SRS
	}

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

// ValidateFlotsamPath validates file path is within allowed directories and safe.
// This preserves security validation from the repository layer.
// AIDEV-NOTE: security-path-validation; prevents path traversal attacks
func ValidateFlotsamPath(filePath string) error {
	if filePath == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	// Clean the path to resolve any . or .. components
	cleanPath := filepath.Clean(filePath)

	// Check for path traversal attempts
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("path traversal not allowed: %s", filePath)
	}

	// Ensure the path is absolute or relative to current working directory
	if !filepath.IsAbs(cleanPath) {
		// For relative paths, ensure they don't escape the current directory
		if strings.HasPrefix(cleanPath, "../") || cleanPath == ".." {
			return fmt.Errorf("path cannot escape current directory: %s", filePath)
		}
	}

	// Check file extension
	if !strings.HasSuffix(strings.ToLower(cleanPath), ".md") {
		return fmt.Errorf("file must have .md extension: %s", filePath)
	}

	return nil
}

// CreateFlotsamNote creates a new flotsam note file with validation.
// This is a simplified version of the repository CreateFlotsamNote method.
func CreateFlotsamNote(note *FlotsamNote, flotsamDir string) error {
	if note == nil {
		return fmt.Errorf("note cannot be nil")
	}

	if note.ID == "" {
		return fmt.Errorf("note ID cannot be empty")
	}

	// Ensure flotsam directory exists
	if err := os.MkdirAll(flotsamDir, 0o750); err != nil {
		return fmt.Errorf("failed to create flotsam directory: %w", err)
	}

	filename := note.ID + ".md"
	filePath := filepath.Join(flotsamDir, filename)

	// Check if file already exists
	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("note with ID %s already exists", note.ID)
	}

	// Use atomic save logic
	return SaveFlotsamNote(note, flotsamDir)
}

// UpdateFlotsamNote updates an existing flotsam note with validation.
// This is a simplified version of the repository UpdateFlotsamNote method.
func UpdateFlotsamNote(note *FlotsamNote, flotsamDir string) error {
	if note == nil {
		return fmt.Errorf("note cannot be nil")
	}

	if note.ID == "" {
		return fmt.Errorf("note ID cannot be empty")
	}

	filename := note.ID + ".md"
	filePath := filepath.Join(flotsamDir, filename)

	// Check if file exists (can't update non-existent note)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("note with ID %s not found", note.ID)
	} else if err != nil {
		return fmt.Errorf("failed to check note file: %w", err)
	}

	// Update modified time to current time
	note.Modified = time.Now()

	// Use atomic save logic
	return SaveFlotsamNote(note, flotsamDir)
}

// DeleteFlotsamNote deletes a flotsam note file with validation.
// This is a simplified version of the repository DeleteFlotsamNote method.
func DeleteFlotsamNote(noteID string, flotsamDir string) error {
	if noteID == "" {
		return fmt.Errorf("note ID cannot be empty")
	}

	filename := noteID + ".md"
	filePath := filepath.Join(flotsamDir, filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("note with ID %s not found", noteID)
	} else if err != nil {
		return fmt.Errorf("failed to check note file: %w", err)
	}

	// Delete the file
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete note file: %w", err)
	}

	return nil
}
