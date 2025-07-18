// Package flotsam provides Unix interop functionality for flotsam notes.
// This file contains hybrid search operations with zk fallback logic.
package flotsam

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// SearchMode determines which search strategy to use.
type SearchMode int

const (
	// SearchModeAuto automatically selects the best search strategy.
	SearchModeAuto SearchMode = iota
	// SearchModeInMemory forces in-memory search using loaded collection.
	SearchModeInMemory
	// SearchModeZK forces zk delegation for search operations.
	SearchModeZK
)

// SearchOptions configures search behavior.
type SearchOptions struct {
	Mode        SearchMode
	Interactive bool // Whether this is for interactive/real-time search
	ContextDir  string
	Tags        []string
	Limit       int
}

// SearchNotes performs hybrid search with adaptive performance selection.
// This implements the core strategy: Unix interop for most operations,
// in-memory collection for performance-critical scenarios.
// AIDEV-NOTE: hybrid-search-core; implements Unix interop + performance fallback strategy
func SearchNotes(query string, options SearchOptions) ([]*FlotsamNote, error) {
	// Determine search mode
	mode := options.Mode
	if mode == SearchModeAuto {
		mode = selectOptimalSearchMode(query, options)
	}

	switch mode {
	case SearchModeInMemory:
		return searchInMemory(query, options)
	case SearchModeZK:
		return searchViaZK(query, options)
	default:
		return nil, fmt.Errorf("unknown search mode: %d", mode)
	}
}

// selectOptimalSearchMode chooses the best search strategy based on context.
func selectOptimalSearchMode(query string, options SearchOptions) SearchMode {
	// Use in-memory search for interactive/real-time scenarios
	if options.Interactive && len(query) > 0 {
		return SearchModeInMemory
	}

	// Use zk for one-off searches, batch operations, and empty queries
	return SearchModeZK
}

// searchInMemory performs search using in-memory collection.
// This is used for performance-critical operations like search-as-you-type.
func searchInMemory(query string, options SearchOptions) ([]*FlotsamNote, error) {
	// Load collection into memory
	collection, err := LoadAllNotes(options.ContextDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load collection: %w", err)
	}

	var results []*FlotsamNote

	// If query is empty, use tag filtering
	if query == "" && len(options.Tags) > 0 {
		results = collection.FilterByTags(options.Tags)
	} else {
		// Search by title
		results = collection.SearchByTitle(query)

		// Filter by tags if specified
		if len(options.Tags) > 0 {
			results = filterResultsByTags(results, options.Tags)
		}
	}

	// Apply limit
	if options.Limit > 0 && len(results) > options.Limit {
		results = results[:options.Limit]
	}

	return results, nil
}

// searchViaZK performs search by delegating to zk.
// This is used for one-off searches and batch operations.
func searchViaZK(query string, options SearchOptions) ([]*FlotsamNote, error) {
	// Check if zk is available
	if !isZKAvailable() {
		// Fallback to in-memory search
		return searchInMemory(query, options)
	}

	// Build zk command
	args := []string{"list", "--format", "json", "--no-pager", "--quiet"}

	// Add tag filters
	if len(options.Tags) > 0 {
		tagQuery := strings.Join(options.Tags, " AND ")
		args = append(args, "--tag", tagQuery)
	}

	// Add text search if query provided
	if query != "" {
		args = append(args, "--match", query)
	}

	// Add limit
	if options.Limit > 0 {
		args = append(args, "--limit", fmt.Sprintf("%d", options.Limit))
	}

	// Execute zk command
	result, err := zkShellOut("list", args[1:]...)
	if err != nil {
		// Fallback to in-memory search on zk failure
		return searchInMemory(query, options)
	}

	// Parse JSON response
	var zkNotes []zkNote
	if err := json.Unmarshal([]byte(result), &zkNotes); err != nil {
		return nil, fmt.Errorf("failed to parse zk output: %w", err)
	}

	// Convert to FlotsamNote format
	notes := make([]*FlotsamNote, 0, len(zkNotes))
	for _, zkNote := range zkNotes {
		note := &FlotsamNote{
			ID:       zkNote.ID,
			Title:    zkNote.Title,
			Created:  zkNote.Created,
			Tags:     zkNote.Tags,
			FilePath: zkNote.AbsPath,
		}
		notes = append(notes, note)
	}

	return notes, nil
}

// filterResultsByTags filters search results by tags.
func filterResultsByTags(results []*FlotsamNote, tags []string) []*FlotsamNote {
	if len(tags) == 0 {
		return results
	}

	filtered := make([]*FlotsamNote, 0)
	for _, note := range results {
		if noteHasTags(note, tags) {
			filtered = append(filtered, note)
		}
	}
	return filtered
}

// noteHasTags checks if a note has any of the specified tags.
func noteHasTags(note *FlotsamNote, tags []string) bool {
	noteTagsLower := make(map[string]bool)
	for _, tag := range note.Tags {
		noteTagsLower[strings.ToLower(tag)] = true
	}

	for _, tag := range tags {
		if noteTagsLower[strings.ToLower(tag)] {
			return true
		}
	}
	return false
}

// isZKAvailable checks if zk binary is available.
func isZKAvailable() bool {
	_, err := exec.LookPath("zk")
	return err == nil
}

// zkShellOut executes a zk command and returns the output.
// This is a basic implementation for zk command execution.
func zkShellOut(cmd string, args ...string) (string, error) {
	// Prepare command
	zkArgs := append([]string{cmd}, args...)
	execCmd := exec.Command("zk", zkArgs...) // #nosec G204 -- zk is a safe command with validated args

	// Execute command
	output, err := execCmd.Output()
	if err != nil {
		return "", fmt.Errorf("zk command failed: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// zkNote represents a note as returned by zk's JSON format.
type zkNote struct {
	ID      string    `json:"id"`
	Title   string    `json:"title"`
	Created time.Time `json:"created"`
	Tags    []string  `json:"tags"`
	AbsPath string    `json:"absPath"`
}

// SearchByTitleInteractive performs interactive title search optimized for real-time use.
// This is a convenience function for search-as-you-type scenarios.
func SearchByTitleInteractive(query string, contextDir string) ([]*FlotsamNote, error) {
	options := SearchOptions{
		Mode:        SearchModeInMemory, // Force in-memory for performance
		Interactive: true,
		ContextDir:  contextDir,
		Limit:       50, // Reasonable limit for interactive display
	}

	return SearchNotes(query, options)
}

// SearchByTagsZK performs tag-based search using zk delegation.
// This is a convenience function for batch operations.
func SearchByTagsZK(tags []string, contextDir string) ([]*FlotsamNote, error) {
	options := SearchOptions{
		Mode:        SearchModeZK, // Force zk delegation
		Interactive: false,
		ContextDir:  contextDir,
		Tags:        tags,
	}

	return SearchNotes("", options)
}
