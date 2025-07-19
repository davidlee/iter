package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/davidlee/vice/internal/config"
	"github.com/davidlee/vice/internal/flotsam"
	"github.com/davidlee/vice/internal/srs"
	"github.com/davidlee/vice/internal/zk"
)

// Add command flags
var (
	addType     string // note type: flashcard, idea, script, log
	addTemplate string // template for note content
	addEdit     bool   // open editor after creation
)

// flotsamAddCmd represents the flotsam add command
// AIDEV-NOTE: T041/6.1b-add-cmd; creates vice-typed notes with SRS integration and auto-init
// AIDEV-NOTE: supports all vice:type:* tags with YAML frontmatter and immediate SRS scheduling
var flotsamAddCmd = &cobra.Command{
	Use:   "add [title]",
	Short: "Create a new flotsam note with vice type tags",
	Long: `Create a new flotsam note with specified type and add to SRS scheduling.

Creates a markdown file with YAML frontmatter including vice:type:* tags and
automatically adds the note to the SRS database for spaced repetition learning.

Auto-initializes flotsam environment (directory + ZK notebook) if needed when
ZK tool is available.

Examples:
  vice flotsam add "What is X?" --type flashcard    # Create flashcard note
  vice flotsam add "Random idea" --type idea        # Create idea note  
  vice flotsam add --type script --edit             # Create script note and edit
  vice flotsam add "Daily log" --type log           # Create log entry`,
	RunE: runFlotsamAdd,
}

func init() {
	flotsamCmd.AddCommand(flotsamAddCmd)

	// Note type and creation options
	flotsamAddCmd.Flags().StringVar(&addType, "type", "idea", "note type (flashcard, idea, script, log)")
	flotsamAddCmd.Flags().StringVar(&addTemplate, "template", "", "template for note content")
	flotsamAddCmd.Flags().BoolVar(&addEdit, "edit", false, "open editor after creating note")
}

// runFlotsamAdd creates a new vice-typed note with SRS integration
// AIDEV-NOTE: T041/6.1b-implementation; complete workflow from creation to SRS scheduling
func runFlotsamAdd(_ *cobra.Command, args []string) error {
	env := GetViceEnv()

	// Auto-initialize flotsam environment if needed
	if err := flotsam.EnsureFlotsamEnvironment(env); err != nil {
		return fmt.Errorf("failed to initialize flotsam environment: %w", err)
	}

	// Validate note type
	validTypes := []string{"flashcard", "idea", "script", "log"}
	if !contains(validTypes, addType) {
		return fmt.Errorf("invalid note type: %s (valid: %s)", addType, strings.Join(validTypes, ", "))
	}

	// Get title from args or use default
	var title string
	if len(args) > 0 {
		title = strings.Join(args, " ")
	} else {
		title = fmt.Sprintf("New %s note", addType)
	}

	// Get ZK notebook instance
	zkNotebook := env.GetFlotsamZK()
	if !zkNotebook.Available() {
		return fmt.Errorf("zk not available - install from https://github.com/zk-org/zk for note creation")
	}

	// Use ZK to create the note and get the path/ID
	notePath, noteID, err := createNoteWithZK(zkNotebook, title, addType, addTemplate)
	if err != nil {
		return fmt.Errorf("failed to create note via ZK: %w", err)
	}

	fmt.Printf("Created note: %s (ID: %s)\n", filepath.Base(notePath), noteID)

	// Add to SRS database
	if err := addToSRSDatabase(notePath, env); err != nil {
		fmt.Printf("Warning: failed to add note to SRS database: %v\n", err)
		// Don't fail the command - note was created successfully
	} else {
		fmt.Printf("Added to SRS scheduling\n")
	}

	// Open editor if requested
	if addEdit {
		if err := zkNotebook.Edit(notePath); err != nil {
			fmt.Printf("Warning: failed to open editor: %v\n", err)
		}
	}

	return nil
}

// createNoteWithZK uses ZK to create a note and returns the path and extracted ID
// AIDEV-NOTE: T041/6.1c-zk-delegation; delegates note creation to ZK for proper ID generation
// AIDEV-NOTE: T041/6.1c-completed; ZK delegation working with unique ID generation via zk new --working-dir
// AIDEV-NOTE: ID uniqueness achieved - ZK generates unique filenames and IDs per note
func createNoteWithZK(zkNotebook *zk.ZKNotebook, title, noteType, template string) (string, string, error) {
	// Use ZK's new command to create the note with proper working directory
	// ZK needs --working-dir to avoid path validation errors
	result, err := zkNotebook.Execute("new",
		"--working-dir", zkNotebook.NotebookDir(),
		"--title", title,
		"--print-path")
	if err != nil {
		return "", "", fmt.Errorf("zk new failed: %w", err)
	}

	if result.ExitCode != 0 {
		return "", "", fmt.Errorf("zk new failed with exit code %d: %s", result.ExitCode, result.Stderr)
	}

	// Extract the path from ZK output
	notePath := strings.TrimSpace(result.Stdout)
	if notePath == "" {
		return "", "", fmt.Errorf("zk new did not return a path")
	}

	// Read the created file to get the ID that ZK generated
	zkContent, err := os.ReadFile(notePath) //nolint:gosec // ZK-generated path is trusted
	if err != nil {
		return "", "", fmt.Errorf("failed to read ZK-created note: %w", err)
	}

	// Extract ID from ZK's frontmatter or filename
	noteID := extractIDFromContent(string(zkContent))
	if noteID == "" {
		// Fallback: extract from filename
		filename := filepath.Base(notePath)
		noteID = strings.TrimSuffix(filename, ".md")
		if strings.Contains(noteID, "-") {
			noteID = strings.Split(noteID, "-")[0] // Take part before first dash
		}
	}

	// Update the file with our vice-specific content while preserving ZK's structure
	updatedContent := createNoteContent(noteID, title, noteType, template)
	if err := os.WriteFile(notePath, []byte(updatedContent), 0o600); err != nil { //nolint:gosec // Standard file permissions
		return "", "", fmt.Errorf("failed to update note with vice content: %w", err)
	}

	return notePath, noteID, nil
}

// extractIDFromContent extracts the ID from ZK-generated content
func extractIDFromContent(content string) string {
	// Look for ZK's ID in the frontmatter
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "id:") {
			// Extract ID value (handle both quoted and unquoted)
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				id := strings.TrimSpace(parts[1])
				id = strings.Trim(id, `"'`) // Remove quotes if present
				return id
			}
		}
	}

	// Fallback: extract from filename if no ID in frontmatter
	// ZK typically uses filename pattern like "20240101-title.md"

	// Last resort: generate a simple ID
	return fmt.Sprintf("%04x", time.Now().Unix()%65536)
}

// Deprecated: generateNoteID is no longer used - ZK now handles ID generation
// AIDEV-NOTE: T041/6.1c-deprecated; replaced by ZK delegation for unique ID generation

// createNoteContent generates markdown content with YAML frontmatter
// AIDEV-NOTE: T041/6.1b-content-generation; follows ZK-compatible frontmatter with vice tags
func createNoteContent(noteID, title, noteType, template string) string {
	now := time.Now()

	// Build tags array
	tags := []string{
		fmt.Sprintf("vice:type:%s", noteType),
	}

	// Format tags for YAML - use quoted format for ZK compatibility
	var tagsYAML string
	if len(tags) == 1 {
		tagsYAML = fmt.Sprintf("tags: ['%s']", tags[0])
	} else {
		tagsYAML = "tags: ["
		for i, tag := range tags {
			if i > 0 {
				tagsYAML += ", "
			}
			tagsYAML += fmt.Sprintf("'%s'", tag)
		}
		tagsYAML += "]"
	}

	// Create frontmatter
	frontmatter := fmt.Sprintf(`---
id: "%s"
title: "%s"
created-at: "%s"
%s
---

`, noteID, title, now.Format(time.RFC3339), tagsYAML)

	// Add content based on type and template
	var body string
	if template != "" {
		body = template
	} else {
		switch noteType {
		case "flashcard":
			body = fmt.Sprintf("# %s\n\n## Question\n\n%s\n\n## Answer\n\n<!-- Add your answer here -->\n", title, title)
		case "idea":
			body = fmt.Sprintf("# %s\n\n<!-- Develop your idea here -->\n", title)
		case "script":
			body = fmt.Sprintf("# %s\n\n```bash\n#!/bin/bash\n# %s\n\n# Add your script here\n```\n", title, title)
		case "log":
			body = fmt.Sprintf("# %s - %s\n\n<!-- Daily log entry -->\n", title, now.Format("2006-01-02"))
		default:
			body = fmt.Sprintf("# %s\n\n<!-- Add content here -->\n", title)
		}
	}

	return frontmatter + body
}

// addToSRSDatabase adds the new note to SRS scheduling
// AIDEV-NOTE: T041/6.1b-srs-integration; immediate SRS scheduling for new notes
func addToSRSDatabase(notePath string, env *config.ViceEnv) error {
	srsDB, err := srs.NewDatabase(env.ContextData, env.Context)
	if err != nil {
		return fmt.Errorf("failed to open SRS database: %w", err)
	}
	defer func() {
		if err := srsDB.Close(); err != nil {
			fmt.Printf("Warning: failed to close SRS database: %v\n", err)
		}
	}()

	// Create SRS entry with default values
	// Note will be due immediately for first review
	initialSRSData := &srs.SRSData{
		Easiness:           2.5,               // Default SM-2 easiness
		ConsecutiveCorrect: 0,                 // New note
		Due:                time.Now().Unix(), // Due immediately for first review
		TotalReviews:       0,                 // New note
	}

	return srsDB.CreateSRSNote(notePath, extractNoteIDFromPath(notePath), env.Context, initialSRSData)
}

// extractNoteIDFromPath extracts note ID from file path
func extractNoteIDFromPath(notePath string) string {
	filename := filepath.Base(notePath)
	filename = strings.TrimSuffix(filename, ".md")

	// Extract first 4 characters as ID (ZK-style)
	if len(filename) >= 4 {
		return filename[:4]
	}

	return filename
}

// contains checks if slice contains string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
