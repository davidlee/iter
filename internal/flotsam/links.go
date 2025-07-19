// Package flotsam provides Unix interop functionality for flotsam notes.
// This file contains zk delegation for link operations.
package flotsam

import (
	"fmt"
	"strings"
)

// AIDEV-NOTE: T041-zk-delegation; link operations delegated to zk for Unix interop

// GetBacklinks returns the notes that link to the specified note.
// This delegates to zk: `zk list --linked-by <note> --format path`
func GetBacklinks(notePath string) ([]string, error) {
	output, err := zkShellOut("list", "--linked-by", notePath, "--format", "path")
	if err != nil {
		return nil, fmt.Errorf("failed to get backlinks: %w", err)
	}

	return parseZKPathOutput(output), nil
}

// GetOutboundLinks returns the notes that the specified note links to.
// This delegates to zk: `zk list --link-to <note> --format path`
func GetOutboundLinks(notePath string) ([]string, error) {
	output, err := zkShellOut("list", "--link-to", notePath, "--format", "path")
	if err != nil {
		return nil, fmt.Errorf("failed to get outbound links: %w", err)
	}

	return parseZKPathOutput(output), nil
}

// GetLinkedNotes returns both backlinks and outbound links for a note.
// This provides a comprehensive view of note relationships via zk.
func GetLinkedNotes(notePath string) (backlinks []string, outbound []string, err error) {
	backlinks, err = GetBacklinks(notePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get backlinks: %w", err)
	}

	outbound, err = GetOutboundLinks(notePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get outbound links: %w", err)
	}

	return backlinks, outbound, nil
}

// parseZKPathOutput parses zk command output that returns paths (one per line).
func parseZKPathOutput(output string) []string {
	lines := strings.Split(strings.TrimSpace(output), "\n")

	var result []string
	for _, line := range lines {
		if line = strings.TrimSpace(line); line != "" {
			result = append(result, line)
		}
	}

	return result
}
