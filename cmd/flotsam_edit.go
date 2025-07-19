package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/davidlee/vice/internal/config"
	"github.com/davidlee/vice/internal/flotsam"
)

// Edit command flags
var editInteractive bool // force interactive selection even with note ID

// flotsamEditCmd represents the flotsam edit command
// AIDEV-NOTE: T041/5.3-edit-cmd; implements ZK delegation with interactive selection per ADR-008
// AIDEV-NOTE: supports both interactive picker (no args) and direct note ID editing
var flotsamEditCmd = &cobra.Command{
	Use:   "edit [note-id]",
	Short: "Edit flotsam notes via ZK integration",
	Long: `Edit flotsam notes using ZK's editor integration with interactive selection.

When called without arguments, opens an interactive picker showing all vice-typed
notes. When given a note ID, resolves the path and opens that specific note.

Follows the ZK-first enrichment pattern (ADR-008) for note discovery and delegates
actual editing to ZK's editor integration, which respects ZK_EDITOR, VISUAL, and 
EDITOR environment variables.

Examples:
  vice flotsam edit                 # Interactive picker of all vice-typed notes
  vice flotsam edit abc1            # Edit note with ID 'abc1'
  vice flotsam edit --interactive   # Force interactive mode even with note ID`,
	RunE: runFlotsamEdit,
}

func init() {
	flotsamCmd.AddCommand(flotsamEditCmd)

	// Interactive mode flag
	flotsamEditCmd.Flags().BoolVar(&editInteractive, "interactive", false, "force interactive selection")
}

// runFlotsamEdit executes the flotsam edit command using ZK delegation
func runFlotsamEdit(_ *cobra.Command, args []string) error {
	env := GetViceEnv()

	// Check ZK availability first
	if !env.IsZKAvailable() {
		return fmt.Errorf("zk not available - install from https://github.com/zk-org/zk")
	}

	// Mode 1: Interactive selection (no args or --interactive flag)
	if len(args) == 0 || editInteractive {
		return runInteractiveEdit(env)
	}

	// Mode 2: Direct note ID editing
	noteID := args[0]
	return runDirectEdit(env, noteID)
}

// runInteractiveEdit opens ZK's interactive picker for all vice-typed notes
func runInteractiveEdit(env *config.ViceEnv) error {
	// Use ZK's interactive mode with tag filtering for all vice:type:* notes
	// This leverages ZK's built-in fuzzy finder and respects user's editor preferences
	return env.ZKEdit("--interactive", "--tag", "vice:type:*")
}

// runDirectEdit resolves a note ID to its path and opens it for editing
func runDirectEdit(env *config.ViceEnv, noteID string) error {
	// Step 1: Discover all vice-typed notes (ZK-first pattern per ADR-008)
	notes, err := flotsam.GetAllViceNotes(env)
	if err != nil {
		return fmt.Errorf("failed to query vice-typed notes: %w", err)
	}

	// Step 2: Find notes matching the given ID
	matchingPaths := findNotesByID(notes, noteID)

	if len(matchingPaths) == 0 {
		return fmt.Errorf("no notes found with ID '%s'", noteID)
	}

	if len(matchingPaths) == 1 {
		// Single match - edit directly
		return env.ZKEdit(matchingPaths[0])
	}

	// Multiple matches - open all in editor (let ZK handle multi-file editing)
	fmt.Printf("Found %d notes matching ID '%s', opening all:\n", len(matchingPaths), noteID)
	for _, path := range matchingPaths {
		fmt.Printf("  %s\n", path)
	}

	return env.ZKEdit(matchingPaths...)
}

// findNotesByID searches for notes containing the given ID in their filename
// This is a simple implementation that can be enhanced later with more sophisticated matching
func findNotesByID(notePaths []string, noteID string) []string {
	var matches []string

	for _, path := range notePaths {
		// Extract filename from path
		parts := strings.Split(path, "/")
		filename := parts[len(parts)-1]

		// Remove .md extension for comparison
		filename = strings.TrimSuffix(filename, ".md")

		// Check if filename starts with the note ID
		// This handles ZK-style naming like "abc1-my-note.md"
		if strings.HasPrefix(filename, noteID) {
			matches = append(matches, path)
			continue
		}

		// Also check for exact ID match anywhere in filename
		// This handles cases where ID might not be at the start
		if strings.Contains(filename, noteID) {
			matches = append(matches, path)
		}
	}

	return matches
}
