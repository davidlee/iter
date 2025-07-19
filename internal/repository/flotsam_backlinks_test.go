package repository

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/davidlee/vice/internal/config"
	"github.com/davidlee/vice/internal/models"
)

// TestFlotsamBacklinks tests backlink computation in LoadFlotsam
// AIDEV-NOTE: T027/4.2.2-backlink-testing; comprehensive test validating bidirectional link relationships
func TestFlotsamBacklinks(t *testing.T) {
	// Create a temporary directory for test
	tmpDir, err := os.MkdirTemp("", "test_flotsam_backlinks")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("Failed to clean up temp directory: %v", err)
		}
	}()

	// Set up test environment using helper function
	viceEnv := &config.ViceEnv{
		ConfigDir:   filepath.Join(tmpDir, "config"),
		DataDir:     filepath.Join(tmpDir, "data"),
		StateDir:    filepath.Join(tmpDir, "state"),
		CacheDir:    filepath.Join(tmpDir, "cache"),
		Context:     "test",
		ContextData: filepath.Join(tmpDir, "data", "test"),
		Contexts:    []string{"test"},
	}

	if err := viceEnv.EnsureDirectories(); err != nil {
		t.Fatalf("Failed to create test directories: %v", err)
	}

	// Create repository
	repo := NewFileRepository(viceEnv)

	// Create flotsam directory
	flotsamDir := viceEnv.GetFlotsamDir()
	if err := os.MkdirAll(flotsamDir, 0o750); err != nil {
		t.Fatalf("Failed to create flotsam directory: %v", err)
	}

	// Create test notes with links
	// Note A links to B and C
	noteA := `---
id: aaaa
title: Note A
created-at: 2024-01-01T00:00:00Z
type: idea
---

This note links to [[bbbb]] and [[cccc]].
`

	// Note B links to C
	noteB := `---
id: bbbb
title: Note B
created-at: 2024-01-01T00:00:00Z
type: idea
---

This note links to [[cccc]].
`

	// Note C doesn't link to anything
	noteC := `---
id: cccc
title: Note C
created-at: 2024-01-01T00:00:00Z
type: idea
---

This note has no outbound links.
`

	// Write test files
	testFiles := map[string]string{
		"aaaa.md": noteA,
		"bbbb.md": noteB,
		"cccc.md": noteC,
	}

	for filename, content := range testFiles {
		filePath := filepath.Join(flotsamDir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0o600); err != nil {
			t.Fatalf("Failed to write test file %s: %v", filename, err)
		}
	}

	// Load flotsam collection (this should compute backlinks)
	// AIDEV-NOTE: test-integration-point; LoadFlotsam() automatically computes backlinks via computeBacklinks()
	collection, err := repo.LoadFlotsam()
	if err != nil {
		t.Fatalf("Failed to load flotsam: %v", err)
	}

	// Verify collection loaded correctly
	if len(collection.Notes) != 3 {
		t.Fatalf("Expected 3 notes, got %d", len(collection.Notes))
	}

	// Create map for easy note lookup
	noteMap := make(map[string]*models.FlotsamNote)
	for i := range collection.Notes {
		note := &collection.Notes[i]
		noteMap[note.ID] = note
	}

	// AIDEV-NOTE: Backlink computation removed in T041/2.3 - now delegated to zk
	// This test verifies that notes were loaded correctly, but backlink computation
	// is no longer done in-memory. See internal/flotsam/links.go for zk delegation.

	// Verify notes were loaded correctly
	if noteA := noteMap["aaaa"]; noteA == nil {
		t.Error("Note A should be loaded")
	}
	if noteB := noteMap["bbbb"]; noteB == nil {
		t.Error("Note B should be loaded")
	}
	if noteC := noteMap["cccc"]; noteC == nil {
		t.Error("Note C should be loaded")
	}

	// NOTE: Backlink testing moved to internal/flotsam/links_test.go
	// where zk delegation functions are tested

	// Verify outbound links are also correctly parsed
	if noteA := noteMap["aaaa"]; noteA != nil {
		if len(noteA.Links) != 2 {
			t.Errorf("Note A should have 2 outbound links, got %d: %v", len(noteA.Links), noteA.Links)
		}
	}

	if noteB := noteMap["bbbb"]; noteB != nil {
		if len(noteB.Links) != 1 {
			t.Errorf("Note B should have 1 outbound link, got %d: %v", len(noteB.Links), noteB.Links)
		}
	}

	if noteC := noteMap["cccc"]; noteC != nil {
		if len(noteC.Links) != 0 {
			t.Errorf("Note C should have 0 outbound links, got %d: %v", len(noteC.Links), noteC.Links)
		}
	}
}

// TestFlotsamBacklinksEmptyCollection tests backlink computation with empty collection
func TestFlotsamBacklinksEmptyCollection(t *testing.T) {
	// Create a temporary directory for test
	tmpDir, err := os.MkdirTemp("", "test_flotsam_empty")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("Failed to clean up temp directory: %v", err)
		}
	}()

	// Set up test environment using helper function
	viceEnv := &config.ViceEnv{
		ConfigDir:   filepath.Join(tmpDir, "config"),
		DataDir:     filepath.Join(tmpDir, "data"),
		StateDir:    filepath.Join(tmpDir, "state"),
		CacheDir:    filepath.Join(tmpDir, "cache"),
		Context:     "test",
		ContextData: filepath.Join(tmpDir, "data", "test"),
		Contexts:    []string{"test"},
	}

	if err := viceEnv.EnsureDirectories(); err != nil {
		t.Fatalf("Failed to create test directories: %v", err)
	}

	// Create repository
	repo := NewFileRepository(viceEnv)

	// Load flotsam collection from non-existent directory
	collection, err := repo.LoadFlotsam()
	if err != nil {
		t.Fatalf("Failed to load flotsam: %v", err)
	}

	// Verify empty collection
	if len(collection.Notes) != 0 {
		t.Errorf("Expected 0 notes in empty collection, got %d", len(collection.Notes))
	}
}
