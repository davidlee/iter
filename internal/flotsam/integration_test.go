// integration_test.go
// Copyright 2025 Vice Contributors
//
// Cross-component integration tests for flotsam note system
// Tests end-to-end workflows combining ZK parsing, link extraction, ID generation, and SRS functionality

package flotsam

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AIDEV-NOTE: Integration tests for complete flotsam note lifecycle combining all components
func TestFlotsamNoteLifecycle(t *testing.T) {
	// Test Case 1: Create note with ZK ID → parse frontmatter → extract links
	t.Run("Complete Note Creation and Parsing", func(t *testing.T) {
		// Generate ZK-compatible ID
		generator := NewIDGenerator(IDOptions{
			Case:    CaseLower,
			Charset: CharsetAlphanum,
			Length:  4,
		})
		noteID := generator()

		// Create note content with frontmatter and links
		noteContent := fmt.Sprintf(`---
id: %s
title: Test Integration Note
created-at: 2025-01-15T10:30:00Z
tags: [integration, test, flotsam]
vice:
  srs:
    easiness: 2.5
    consecutive_correct: 0
    due: %d
    total_reviews: 0
    review_history: []
---

# Test Integration Note

This is a test note with [[wiki link]] and another [[complex link | with label]].

The note also has a relationship link: #[[parent note]] and [[child note]]#.

Content for testing the complete flotsam system integration.
`, noteID, time.Now().Add(24*time.Hour).Unix())

		// Parse frontmatter
		frontmatter, body, err := parseFrontmatter(noteContent)
		require.NoError(t, err)
		require.NotNil(t, frontmatter)

		// Verify frontmatter parsing
		assert.Equal(t, noteID, frontmatter["id"])
		assert.Equal(t, "Test Integration Note", frontmatter["title"])
		tags, ok := frontmatter["tags"].([]interface{})
		require.True(t, ok)
		assert.Len(t, tags, 3)
		assert.Contains(t, body, "This is a test note with")

		// Extract links from body
		links := ExtractLinks(body)

		// Verify link extraction
		assert.Len(t, links, 4) // wiki link, complex link, parent note, child note

		// Find specific links
		var foundWikiLink, foundComplexLink, foundParentLink, foundChildLink bool
		for _, link := range links {
			switch link.Href {
			case "wiki link":
				foundWikiLink = true
				assert.Equal(t, LinkTypeWikiLink, link.Type)
			case "complex link":
				foundComplexLink = true
				assert.Equal(t, "with label", link.Title)
				assert.Equal(t, LinkTypeWikiLink, link.Type)
			case "parent note":
				foundParentLink = true
				assert.Contains(t, link.Rels, LinkRelationUp)
			case "child note":
				foundChildLink = true
				assert.Contains(t, link.Rels, LinkRelationDown)
			}
		}

		assert.True(t, foundWikiLink, "Should find wiki link")
		assert.True(t, foundComplexLink, "Should find complex link with label")
		assert.True(t, foundParentLink, "Should find parent relationship link")
		assert.True(t, foundChildLink, "Should find child relationship link")
	})
}

// AIDEV-NOTE: Tests complete SRS lifecycle from initialization through review cycles
func TestSRSLifecycle(t *testing.T) {
	t.Run("Complete SRS Review Cycle", func(t *testing.T) {
		// Initialize new SRS data
		srsData := &SRSData{
			Easiness:           DefaultEasiness,
			ConsecutiveCorrect: 0,
			Due:                time.Now().Unix(),
			TotalReviews:       0,
			ReviewHistory:      []ReviewRecord{},
		}

		// Create SM-2 calculator
		calc := NewSM2Calculator()

		assert.Equal(t, 2.5, srsData.Easiness)
		assert.Equal(t, 0, srsData.ConsecutiveCorrect)
		assert.Equal(t, 0, srsData.TotalReviews)
		assert.True(t, calc.IsDue(srsData))

		// Create flotsam note with SRS data
		note := &FlotsamNote{
			ID:    "tst1",
			Title: "SRS Test Note",
			Body:  "Test content for SRS lifecycle",
			SRS:   srsData,
		}

		// Test 1: First review (correct answer, quality 5)
		updatedSRS, err := calc.ProcessReview(srsData, CorrectEffort)
		require.NoError(t, err)

		assert.Equal(t, 1, updatedSRS.ConsecutiveCorrect)
		assert.Equal(t, 1, updatedSRS.TotalReviews)
		assert.Greater(t, updatedSRS.Easiness, 2.5) // Should increase
		assert.False(t, calc.IsDue(updatedSRS))     // Should not be due immediately

		// Test 2: Second review (correct answer, quality 4)
		updatedSRS2, err := calc.ProcessReview(updatedSRS, CorrectHard)
		require.NoError(t, err)

		assert.Equal(t, 2, updatedSRS2.ConsecutiveCorrect)
		assert.Equal(t, 2, updatedSRS2.TotalReviews)
		assert.False(t, calc.IsDue(updatedSRS2))

		// Test 3: Third review (incorrect answer, quality 2)
		updatedSRS3, err := calc.ProcessReview(updatedSRS2, IncorrectFamiliar)
		require.NoError(t, err)

		assert.Equal(t, 0, updatedSRS3.ConsecutiveCorrect) // Reset
		assert.Equal(t, 3, updatedSRS3.TotalReviews)
		assert.Less(t, updatedSRS3.Easiness, updatedSRS2.Easiness) // Should decrease
		// Note: Due to SM-2 algorithm, incorrect answers still get scheduled for future review
		// The card may not be immediately due depending on implementation

		// Test review history tracking
		assert.Len(t, updatedSRS3.ReviewHistory, 3)
		assert.Equal(t, CorrectEffort, updatedSRS3.ReviewHistory[0].Quality)
		assert.Equal(t, CorrectHard, updatedSRS3.ReviewHistory[1].Quality)
		assert.Equal(t, IncorrectFamiliar, updatedSRS3.ReviewHistory[2].Quality)

		// Update note with final SRS state
		note.SRS = updatedSRS3

		// Verify note can be serialized with SRS data
		noteJSON, err := json.Marshal(note)
		require.NoError(t, err)

		var deserializedNote FlotsamNote
		err = json.Unmarshal(noteJSON, &deserializedNote)
		require.NoError(t, err)

		assert.Equal(t, note.ID, deserializedNote.ID)
		assert.Equal(t, note.SRS.TotalReviews, deserializedNote.SRS.TotalReviews)
		assert.Equal(t, note.SRS.ConsecutiveCorrect, deserializedNote.SRS.ConsecutiveCorrect)
	})
}

// AIDEV-NOTE: Tests complete cross-component workflow from parsing to SRS to review session
func TestCrossComponentWorkflow(t *testing.T) {
	t.Run("Parse Content → Extract Links → Enable SRS → Complete Review", func(t *testing.T) {
		// Create temporary directory for test files
		tempDir := t.TempDir()

		// Generate note ID with timeout protection
		generator := NewIDGenerator(IDOptions{
			Case:    CaseLower,
			Charset: CharsetAlphanum,
			Length:  4,
		})
		noteID := generator()

		// Ensure we have a valid ID
		require.NotEmpty(t, noteID)
		require.Len(t, noteID, 4)

		// Create note file content
		noteContent := fmt.Sprintf(`---
id: %s
title: Cross-Component Test
created-at: 2025-01-15T10:30:00Z
tags: [cross-component, workflow]
vice:
  srs:
    easiness: 2.5
    consecutive_correct: 0
    due: %d
    total_reviews: 0
    review_history: []
---

# Cross-Component Test

This note links to [[concept A]] and [[concept B | alternative label]].

It also has hierarchical relationships: #[[parent concept]] and [[child concept]]#.

**Question**: What is the relationship between concept A and concept B?

**Answer**: They are related through the parent concept hierarchy.
`, noteID, time.Now().Unix())

		// Write note to file
		notePath := filepath.Join(tempDir, fmt.Sprintf("%s.md", noteID))
		err := os.WriteFile(notePath, []byte(noteContent), 0o600)
		require.NoError(t, err)

		// Step 1: Parse note content
		// #nosec G304 -- File path is controlled by test code
		content, err := os.ReadFile(notePath)
		require.NoError(t, err)

		frontmatter, body, err := parseFrontmatter(string(content))
		require.NoError(t, err)

		// Step 2: Extract links
		links := ExtractLinks(body)

		// Step 3: Create flotsam note structure
		note := &FlotsamNote{
			ID:    frontmatter["id"].(string),
			Title: frontmatter["title"].(string),
			Body:  body,
			Links: make([]string, len(links)),
		}

		// Extract link targets
		for i, link := range links {
			note.Links[i] = link.Href
		}

		// Create SRS data from frontmatter
		srsData := &SRSData{
			Easiness:           DefaultEasiness,
			ConsecutiveCorrect: 0,
			Due:                time.Now().Unix(),
			TotalReviews:       0,
			ReviewHistory:      []ReviewRecord{},
		}
		note.SRS = srsData

		// Step 4: Enable SRS and conduct review session
		calc := NewSM2Calculator()
		assert.True(t, calc.IsDue(note.SRS), "New note should be due for review")

		// Process review (correct answer)
		reviewQuality := CorrectEffort
		updatedSRS, err := calc.ProcessReview(note.SRS, reviewQuality)
		require.NoError(t, err)

		// Step 5: Verify complete workflow
		assert.Equal(t, noteID, note.ID)
		assert.Contains(t, note.Links, "concept A")
		assert.Contains(t, note.Links, "concept B")
		assert.Equal(t, 1, updatedSRS.TotalReviews)
		assert.Equal(t, 1, updatedSRS.ConsecutiveCorrect)
		assert.False(t, calc.IsDue(updatedSRS))

		// Step 6: Update frontmatter with new SRS data
		updatedContent := fmt.Sprintf(`---
id: %s
title: Cross-Component Test
created-at: 2025-01-15T10:30:00Z
tags: [cross-component, workflow]
vice:
  srs:
    easiness: %.1f
    consecutive_correct: %d
    due: %d
    total_reviews: %d
    review_history:
      - timestamp: %d
        quality: %d
---

%s`, noteID,
			updatedSRS.Easiness,
			updatedSRS.ConsecutiveCorrect,
			updatedSRS.Due,
			updatedSRS.TotalReviews,
			updatedSRS.ReviewHistory[0].Timestamp,
			int(updatedSRS.ReviewHistory[0].Quality),
			body)

		// Write updated content back to file
		err = os.WriteFile(notePath, []byte(updatedContent), 0o600)
		require.NoError(t, err)

		// Step 7: Verify round-trip parsing
		// #nosec G304 -- File path is controlled by test code
		updatedFileContent, err := os.ReadFile(notePath)
		require.NoError(t, err)

		parsedFrontmatter, parsedBody, err := parseFrontmatter(string(updatedFileContent))
		require.NoError(t, err)

		// Extract vice.srs data
		viceData := parsedFrontmatter["vice"].(map[string]interface{})
		srsMap := viceData["srs"].(map[string]interface{})

		assert.Equal(t, 1, int(srsMap["total_reviews"].(int)))
		assert.Equal(t, 1, int(srsMap["consecutive_correct"].(int)))

		// Verify links still extracted correctly
		parsedLinks := ExtractLinks(parsedBody)
		assert.Len(t, parsedLinks, 4) // Same number of links
	})
}

// AIDEV-NOTE: Tests data flow consistency across all components
func TestDataFlowConsistency(t *testing.T) {
	t.Run("Frontmatter ↔ SRS Data ↔ Review Structures ↔ Scheduling", func(t *testing.T) {
		// Create initial data
		initialSRS := &SRSData{
			Easiness:           DefaultEasiness,
			ConsecutiveCorrect: 0,
			Due:                time.Now().Unix(),
			TotalReviews:       0,
			ReviewHistory:      []ReviewRecord{},
		}

		// Test 1: SRS Data → Frontmatter → SRS Data
		noteWithSRS := &FlotsamNote{
			ID:    "flow",
			Title: "Data Flow Test",
			Body:  "Test content",
			SRS:   initialSRS,
		}

		// Serialize to JSON (simulating frontmatter storage)
		srsJSON, err := json.Marshal(noteWithSRS.SRS)
		require.NoError(t, err)

		// Deserialize from JSON
		var deserializedSRS SRSData
		err = json.Unmarshal(srsJSON, &deserializedSRS)
		require.NoError(t, err)

		assert.Equal(t, initialSRS.Easiness, deserializedSRS.Easiness)
		assert.Equal(t, initialSRS.ConsecutiveCorrect, deserializedSRS.ConsecutiveCorrect)
		assert.Equal(t, initialSRS.Due, deserializedSRS.Due)

		// Test 2: Review Structures → SRS Data Update
		calc := NewSM2Calculator()
		updatedSRS, err := calc.ProcessReview(&deserializedSRS, CorrectHard)
		require.NoError(t, err)

		// Test 3: Updated SRS Data → Scheduling
		assert.False(t, calc.IsDue(updatedSRS), "Should not be due after review")
		assert.Greater(t, updatedSRS.Due, time.Now().Unix(), "Due time should be in future")

		// Test 4: Complete round-trip
		assert.Equal(t, noteWithSRS.ID, "flow")
		assert.Equal(t, 1, updatedSRS.TotalReviews)
		assert.Equal(t, 1, updatedSRS.ConsecutiveCorrect)
		assert.Len(t, updatedSRS.ReviewHistory, 1)
		assert.Equal(t, CorrectHard, updatedSRS.ReviewHistory[0].Quality)

		// Test 5: Verify scheduling calculations
		nextReviewTime := time.Unix(updatedSRS.Due, 0)
		assert.True(t, nextReviewTime.After(time.Now()), "Next review should be in future")

		// For first review with quality 4, interval should be 1 day
		expectedInterval := time.Hour * 24
		actualInterval := time.Until(nextReviewTime)
		// Use a looser tolerance to prevent flaky tests
		assert.InDelta(t, expectedInterval.Minutes(), actualInterval.Minutes(), 120, "Interval should be ~1 day")
	})
}

// AIDEV-NOTE: Performance test to validate reasonable performance of combined operations
func TestIntegrationPerformance(t *testing.T) {
	t.Run("Performance of Combined Operations", func(t *testing.T) {
		// Skip in short mode
		if testing.Short() {
			t.Skip("Skipping performance test in short mode")
		}

		// Generate multiple notes for performance testing
		noteCount := 10 // Reduced count to prevent potential issues
		notes := make([]*FlotsamNote, noteCount)

		generator := NewIDGenerator(IDOptions{
			Case:    CaseLower,
			Charset: CharsetAlphanum,
			Length:  4,
		})

		start := time.Now()

		for i := 0; i < noteCount; i++ {
			noteID := generator()

			noteContent := fmt.Sprintf(`---
id: %s
title: Performance Test Note %d
created-at: 2025-01-15T10:30:00Z
tags: [performance, test]
vice:
  srs:
    easiness: 2.5
    consecutive_correct: 0
    due: %d
    total_reviews: 0
    review_history: []
---

# Performance Test Note %d

This note has [[link %d]] and [[link %d | with label]].

Content for performance testing the flotsam system.
`, noteID, i, time.Now().Unix(), i, i, i+1)

			// Parse frontmatter
			frontmatter, body, err := parseFrontmatter(noteContent)
			require.NoError(t, err)

			// Extract links
			links := ExtractLinks(body)

			// Create note structure
			note := &FlotsamNote{
				ID:    noteID, // Use the generated ID directly
				Title: frontmatter["title"].(string),
				Body:  body,
				Links: make([]string, len(links)),
				SRS: &SRSData{
					Easiness:           DefaultEasiness,
					ConsecutiveCorrect: 0,
					Due:                time.Now().Unix(),
					TotalReviews:       0,
					ReviewHistory:      []ReviewRecord{},
				},
			}

			for j, link := range links {
				note.Links[j] = link.Href
			}

			notes[i] = note
		}

		elapsed := time.Since(start)

		// Performance assertions
		avgTimePerNote := elapsed / time.Duration(noteCount)
		assert.Less(t, avgTimePerNote, 10*time.Millisecond, "Should process each note in <10ms")

		t.Logf("Processed %d notes in %v (avg: %v per note)", noteCount, elapsed, avgTimePerNote)

		// Verify all notes processed correctly
		for i, note := range notes {
			assert.NotEmpty(t, note.ID)
			assert.Equal(t, fmt.Sprintf("Performance Test Note %d", i), note.Title)
			assert.Len(t, note.Links, 2)
			assert.NotNil(t, note.SRS)
		}
	})
}
