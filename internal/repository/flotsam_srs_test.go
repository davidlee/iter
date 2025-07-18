// Package repository provides file-based data repository implementation.
package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"davidlee/vice/internal/config"
)

func TestGetDueFlotsamNotes(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "TestGetDueFlotsamNotes")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("Failed to clean up temp dir: %v", err)
		}
	}()

	// Setup test environment
	viceEnv := &config.ViceEnv{
		DataDir: tmpDir,
		Context: "test",
	}
	viceEnv.ContextData = filepath.Join(viceEnv.DataDir, viceEnv.Context)
	repo := NewFileRepository(viceEnv)

	// Create flotsam directory
	flotsamDir := viceEnv.GetFlotsamDir()
	if err := os.MkdirAll(flotsamDir, 0o750); err != nil {
		t.Fatalf("Failed to create flotsam dir: %v", err)
	}

	// Test data: notes with different SRS states
	now := time.Now()
	pastDue := now.Add(-24 * time.Hour).Unix()   // Due yesterday
	futureDue := now.Add(24 * time.Hour).Unix()  // Due tomorrow
	
	testNotes := []struct {
		filename string
		content  string
		expectDue bool
	}{
		{
			filename: "abc1.md",
			content: `---
id: abc1
title: "Due Note"
created-at: 2025-07-18T10:00:00Z
type: flashcard
srs:
  easiness: 2.5
  consecutive_correct: 1
  due: ` + fmt.Sprintf("%d", pastDue) + `
  total_reviews: 1
---
This note is due for review.`,
			expectDue: true,
		},
		{
			filename: "abc2.md", 
			content: `---
id: abc2
title: "Future Note"
created-at: 2025-07-18T10:00:00Z
type: flashcard
srs:
  easiness: 2.5
  consecutive_correct: 1
  due: ` + fmt.Sprintf("%d", futureDue) + `
  total_reviews: 1
---
This note is not due yet.`,
			expectDue: false,
		},
		{
			filename: "abc3.md",
			content: `---
id: abc3
title: "New Note"
created-at: 2025-07-18T10:00:00Z
type: idea
---
This note has no SRS data (new card).`,
			expectDue: true, // New cards are always due
		},
	}

	// Write test notes to filesystem
	for _, note := range testNotes {
		filePath := filepath.Join(flotsamDir, note.filename)
		if err := os.WriteFile(filePath, []byte(note.content), 0o600); err != nil {
			t.Fatalf("Failed to write test note %s: %v", note.filename, err)
		}
	}

	// Test GetDueFlotsamNotes
	dueNotes, err := repo.GetDueFlotsamNotes()
	if err != nil {
		t.Fatalf("GetDueFlotsamNotes failed: %v", err)
	}

	// Verify results
	expectedDueCount := 0
	for _, note := range testNotes {
		if note.expectDue {
			expectedDueCount++
		}
	}

	if len(dueNotes) != expectedDueCount {
		t.Errorf("Expected %d due notes, got %d", expectedDueCount, len(dueNotes))
	}

	// Verify correct notes are marked as due
	dueIDs := make(map[string]bool)
	for _, note := range dueNotes {
		dueIDs[note.ID] = true
	}

	for i, testNote := range testNotes {
		expectedID := []string{"abc1", "abc2", "abc3"}[i]
		if testNote.expectDue && !dueIDs[expectedID] {
			t.Errorf("Expected note %s to be due, but it wasn't", expectedID)
		}
		if !testNote.expectDue && dueIDs[expectedID] {
			t.Errorf("Expected note %s to not be due, but it was", expectedID)
		}
	}
}

func TestGetDueFlotsamNotesEmptyCollection(t *testing.T) {
	// Create temp directory for test  
	tmpDir, err := os.MkdirTemp("", "TestGetDueFlotsamNotesEmptyCollection")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("Failed to clean up temp dir: %v", err)
		}
	}()

	// Setup test environment
	viceEnv := &config.ViceEnv{
		DataDir: tmpDir,
		Context: "test-empty",
	}
	viceEnv.ContextData = filepath.Join(viceEnv.DataDir, viceEnv.Context)
	repo := NewFileRepository(viceEnv)

	// Create empty flotsam directory
	flotsamDir := viceEnv.GetFlotsamDir()
	t.Logf("Using flotsam directory: %s", flotsamDir)
	if err := os.MkdirAll(flotsamDir, 0o750); err != nil {
		t.Fatalf("Failed to create flotsam dir: %v", err)
	}

	// Verify directory is empty
	files, err := os.ReadDir(flotsamDir)
	if err != nil {
		t.Fatalf("Failed to read flotsam dir: %v", err)
	}
	t.Logf("Directory contains %d files", len(files))
	for _, file := range files {
		t.Logf("  - %s", file.Name())
	}

	// Test GetDueFlotsamNotes with empty collection
	dueNotes, err := repo.GetDueFlotsamNotes()
	if err != nil {
		t.Fatalf("GetDueFlotsamNotes failed: %v", err)
	}

	if len(dueNotes) != 0 {
		t.Errorf("Expected 0 due notes in empty collection, got %d", len(dueNotes))
		for i, note := range dueNotes {
			t.Logf("Due note %d: ID=%s, Title=%s, FilePath=%s", i, note.ID, note.Title, note.FilePath)
		}
	}
}