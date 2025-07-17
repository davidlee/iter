package repository

import (
	"os"
	"path/filepath"
	"testing"

	"davidlee/vice/internal/config"
	"davidlee/vice/internal/models"
)

// TestFlotsamBacklinks tests backlink computation in LoadFlotsam
func TestFlotsamBacklinks(t *testing.T) {
	// Create a temporary directory for test
	tmpDir, err := os.MkdirTemp("", "test_flotsam_backlinks")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

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
	if err := os.MkdirAll(flotsamDir, 0o755); err != nil {
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

	// Verify backlinks
	// Note A should have no backlinks (no one links to it)
	if noteA := noteMap["aaaa"]; noteA != nil {
		if len(noteA.Backlinks) != 0 {
			t.Errorf("Note A should have 0 backlinks, got %d: %v", len(noteA.Backlinks), noteA.Backlinks)
		}
	}

	// Note B should have 1 backlink from A
	if noteB := noteMap["bbbb"]; noteB != nil {
		if len(noteB.Backlinks) != 1 {
			t.Errorf("Note B should have 1 backlink, got %d: %v", len(noteB.Backlinks), noteB.Backlinks)
		} else if noteB.Backlinks[0] != "aaaa" {
			t.Errorf("Note B should have backlink from 'aaaa', got '%s'", noteB.Backlinks[0])
		}
	}

	// Note C should have 2 backlinks from A and B
	if noteC := noteMap["cccc"]; noteC != nil {
		if len(noteC.Backlinks) != 2 {
			t.Errorf("Note C should have 2 backlinks, got %d: %v", len(noteC.Backlinks), noteC.Backlinks)
		} else {
			// Check that both A and B are in backlinks (order doesn't matter)
			found := make(map[string]bool)
			for _, backlink := range noteC.Backlinks {
				found[backlink] = true
			}
			if !found["aaaa"] || !found["bbbb"] {
				t.Errorf("Note C should have backlinks from 'aaaa' and 'bbbb', got %v", noteC.Backlinks)
			}
		}
	}

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
	defer os.RemoveAll(tmpDir)

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