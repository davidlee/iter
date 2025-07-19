package srs

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDatabase(t *testing.T) {
	tempDir := t.TempDir()
	viceDir := filepath.Join(tempDir, ".vice")
	err := os.MkdirAll(viceDir, 0750) //nolint:gosec // Test directory permissions
	require.NoError(t, err)

	db, err := NewDatabase(tempDir, "test-context")
	require.NoError(t, err)
	defer func() { _ = db.Close() }() //nolint:errcheck // Test cleanup

	assert.Equal(t, "test-context", db.context)
	assert.Equal(t, filepath.Join(tempDir, ".vice", "flotsam.db"), db.dbPath)

	// Verify database file was created
	_, err = os.Stat(db.dbPath)
	assert.NoError(t, err)
}

func TestDatabaseSchema(t *testing.T) {
	tempDir := t.TempDir()
	viceDir := filepath.Join(tempDir, ".vice")
	err := os.MkdirAll(viceDir, 0750) //nolint:gosec // Test directory permissions
	require.NoError(t, err)

	db, err := NewDatabase(tempDir, "test-context")
	require.NoError(t, err)
	defer func() { _ = db.Close() }() //nolint:errcheck // Test cleanup

	// Verify table exists and has correct columns
	query := `
		SELECT name FROM sqlite_master 
		WHERE type='table' AND name='srs_reviews'
	`
	var tableName string
	err = db.db.QueryRow(query).Scan(&tableName)
	require.NoError(t, err)
	assert.Equal(t, "srs_reviews", tableName)

	// Verify column structure
	columns, err := getTableColumns(db, "srs_reviews")
	require.NoError(t, err)

	expectedColumns := []string{
		"note_path", "note_id", "context", "easiness",
		"consecutive_correct", "due_date", "total_reviews",
		"created_at", "last_reviewed",
	}

	for _, col := range expectedColumns {
		assert.Contains(t, columns, col, "Missing column: %s", col)
	}
}

func TestCreateAndGetSRSNote(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }() //nolint:errcheck // Test cleanup

	// Create test SRS note
	notePath := "test/note.md"
	noteID := "test-123"
	context := "test-context"

	initialData := &SRSData{
		Easiness:           2.5,
		ConsecutiveCorrect: 0,
		Due:                time.Now().Add(24 * time.Hour).Unix(),
		TotalReviews:       0,
	}

	err := db.CreateSRSNote(notePath, noteID, context, initialData)
	require.NoError(t, err)

	// Retrieve and verify
	data, err := db.GetSRSData(notePath)
	require.NoError(t, err)

	assert.Equal(t, initialData.Easiness, data.Easiness)
	assert.Equal(t, initialData.ConsecutiveCorrect, data.ConsecutiveCorrect)
	assert.Equal(t, initialData.Due, data.Due)
	assert.Equal(t, initialData.TotalReviews, data.TotalReviews)
}

func TestGetDueNotes(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }() //nolint:errcheck // Test cleanup

	context := "test-context"
	now := time.Now()

	// Create notes with different due dates
	testNotes := []struct {
		path    string
		id      string
		dueDate time.Time
	}{
		{"overdue.md", "over-1", now.Add(-24 * time.Hour)},
		{"due-now.md", "due-1", now.Add(-1 * time.Minute)},
		{"future.md", "future-1", now.Add(24 * time.Hour)},
	}

	for _, note := range testNotes {
		data := &SRSData{
			Easiness:           2.5,
			ConsecutiveCorrect: 0,
			Due:                note.dueDate.Unix(),
			TotalReviews:       0,
		}
		err := db.CreateSRSNote(note.path, note.id, context, data)
		require.NoError(t, err)
	}

	// Get due notes
	dueNotes, err := db.GetDueNotes(context)
	require.NoError(t, err)

	// Should return 2 notes (overdue and due-now), not the future one
	assert.Len(t, dueNotes, 2)

	// Verify they're sorted by due date (oldest first)
	assert.Equal(t, "overdue.md", dueNotes[0].NotePath)
	assert.Equal(t, "due-now.md", dueNotes[1].NotePath)
}

func TestUpdateReview(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }() //nolint:errcheck // Test cleanup

	// Create initial note
	notePath := "test/review.md"
	noteID := "review-123"
	context := "test-context"

	initialData := &SRSData{
		Easiness:           2.5,
		ConsecutiveCorrect: 0,
		Due:                time.Now().Unix(),
		TotalReviews:       0,
	}

	err := db.CreateSRSNote(notePath, noteID, context, initialData)
	require.NoError(t, err)

	// Update after review
	updatedData := &SRSData{
		Easiness:           2.6,
		ConsecutiveCorrect: 1,
		Due:                time.Now().Add(3 * 24 * time.Hour).Unix(),
		TotalReviews:       1, // Note: UpdateReview increments this
	}

	err = db.UpdateReview(notePath, updatedData)
	require.NoError(t, err)

	// Verify updates
	data, err := db.GetSRSData(notePath)
	require.NoError(t, err)

	assert.Equal(t, updatedData.Easiness, data.Easiness)
	assert.Equal(t, updatedData.ConsecutiveCorrect, data.ConsecutiveCorrect)
	assert.Equal(t, updatedData.Due, data.Due)
	assert.Equal(t, 1, data.TotalReviews) // Should be incremented
}

func TestDeleteSRSNote(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }() //nolint:errcheck // Test cleanup

	// Create test note
	notePath := "test/delete.md"
	noteID := "delete-123"
	context := "test-context"

	initialData := &SRSData{
		Easiness:           2.5,
		ConsecutiveCorrect: 0,
		Due:                time.Now().Unix(),
		TotalReviews:       0,
	}

	err := db.CreateSRSNote(notePath, noteID, context, initialData)
	require.NoError(t, err)

	// Verify it exists
	_, err = db.GetSRSData(notePath)
	require.NoError(t, err)

	// Delete it
	err = db.DeleteSRSNote(notePath)
	require.NoError(t, err)

	// Verify it's gone
	_, err = db.GetSRSData(notePath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "note not found")
}

func TestGetStats(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }() //nolint:errcheck // Test cleanup

	context := "test-context"
	now := time.Now()

	// Create test notes
	testData := []struct {
		path     string
		id       string
		easiness float64
		due      time.Time
		reviews  int
	}{
		{"note1.md", "id1", 2.5, now.Add(-1 * time.Hour), 0},
		{"note2.md", "id2", 2.8, now.Add(-1 * time.Hour), 5},
		{"note3.md", "id3", 3.0, now.Add(24 * time.Hour), 10},
	}

	for _, note := range testData {
		data := &SRSData{
			Easiness:           note.easiness,
			ConsecutiveCorrect: 0,
			Due:                note.due.Unix(),
			TotalReviews:       note.reviews,
		}
		err := db.CreateSRSNote(note.path, note.id, context, data)
		require.NoError(t, err)
	}

	// Get stats
	stats, err := db.GetStats(context)
	require.NoError(t, err)

	assert.Equal(t, int64(3), stats["total_notes"])
	assert.Equal(t, int64(2), stats["due_notes"]) // 2 notes are due

	// Check averages (approximate due to floating point)
	avgEasiness := stats["avg_easiness"].(float64)
	assert.InDelta(t, 2.77, avgEasiness, 0.1) // (2.5 + 2.8 + 3.0) / 3

	avgReviews := stats["avg_reviews"].(float64)
	assert.InDelta(t, 5.0, avgReviews, 0.1) // (0 + 5 + 10) / 3
}

// Helper functions

func setupTestDB(t *testing.T) *Database {
	tempDir := t.TempDir()
	viceDir := filepath.Join(tempDir, ".vice")
	err := os.MkdirAll(viceDir, 0750) //nolint:gosec // Test directory permissions
	require.NoError(t, err)

	db, err := NewDatabase(tempDir, "test-context")
	require.NoError(t, err)

	return db
}

func getTableColumns(db *Database, tableName string) ([]string, error) {
	query := `PRAGMA table_info(` + tableName + `)`
	rows, err := db.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }() //nolint:errcheck // Test cleanup

	var columns []string
	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, pk int
		var defaultValue interface{}

		err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk)
		if err != nil {
			return nil, err
		}

		columns = append(columns, name)
	}

	return columns, rows.Err()
}
