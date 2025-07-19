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
	err := os.MkdirAll(viceDir, 0o750) //nolint:gosec // Test directory permissions
	require.NoError(t, err)

	db, err := NewDatabase(tempDir, "test-context")
	require.NoError(t, err)
	defer func() { _ = db.Close() }() //nolint:errcheck // Test cleanup

	assert.Equal(t, "test-context", db.context)
	assert.Equal(t, filepath.Join(tempDir, "flotsam", ".vice", "flotsam.db"), db.dbPath)

	// Verify database file was created
	_, err = os.Stat(db.dbPath)
	assert.NoError(t, err)
}

func TestDatabaseSchema(t *testing.T) {
	tempDir := t.TempDir()
	viceDir := filepath.Join(tempDir, ".vice")
	err := os.MkdirAll(viceDir, 0o750) //nolint:gosec // Test directory permissions
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
	err := os.MkdirAll(viceDir, 0o750) //nolint:gosec // Test directory permissions
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

// Cache invalidation tests

func TestCacheManager(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }() //nolint:errcheck // Test cleanup

	tempDir := t.TempDir()

	// Create flotsam directory
	flotsamDir := filepath.Join(tempDir, "flotsam")
	err := os.MkdirAll(flotsamDir, 0o750) //nolint:gosec // Test directory permissions
	require.NoError(t, err)

	cacheManager := db.GetCacheManager(tempDir)
	assert.Equal(t, flotsamDir, cacheManager.flotsamDir)
	assert.Equal(t, tempDir, cacheManager.contextDir)
}

func TestValidateCache_NewCache(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }() //nolint:errcheck // Test cleanup

	tempDir := t.TempDir()
	flotsamDir := filepath.Join(tempDir, "flotsam")
	err := os.MkdirAll(flotsamDir, 0o750) //nolint:gosec // Test directory permissions
	require.NoError(t, err)

	cacheManager := db.GetCacheManager(tempDir)

	// First validation should trigger refresh (cache miss)
	err = cacheManager.ValidateCache()
	require.NoError(t, err)

	// Verify cache metadata was created
	cachedMtime, err := cacheManager.getCachedDirMtime()
	require.NoError(t, err)
	assert.True(t, cachedMtime.Unix() > 0)
}

func TestValidateCache_UpToDate(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }() //nolint:errcheck // Test cleanup

	tempDir := t.TempDir()
	flotsamDir := filepath.Join(tempDir, "flotsam")
	err := os.MkdirAll(flotsamDir, 0o750) //nolint:gosec // Test directory permissions
	require.NoError(t, err)

	cacheManager := db.GetCacheManager(tempDir)

	// Initial cache
	err = cacheManager.ValidateCache()
	require.NoError(t, err)

	initialMtime, err := cacheManager.getCachedDirMtime()
	require.NoError(t, err)

	// Second validation should not change anything (cache hit)
	err = cacheManager.ValidateCache()
	require.NoError(t, err)

	currentMtime, err := cacheManager.getCachedDirMtime()
	require.NoError(t, err)

	// Cache metadata should be unchanged
	assert.Equal(t, initialMtime.Unix(), currentMtime.Unix())
}

func TestValidateCache_DirectoryChanged(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }() //nolint:errcheck // Test cleanup

	tempDir := t.TempDir()
	flotsamDir := filepath.Join(tempDir, "flotsam")
	err := os.MkdirAll(flotsamDir, 0o750) //nolint:gosec // Test directory permissions
	require.NoError(t, err)

	cacheManager := db.GetCacheManager(tempDir)

	// Initial cache
	err = cacheManager.ValidateCache()
	require.NoError(t, err)

	initialMtime, err := cacheManager.getCachedDirMtime()
	require.NoError(t, err)

	// Wait a bit to ensure different mtime
	time.Sleep(100 * time.Millisecond)

	// Modify directory by adding a file
	testFile := filepath.Join(flotsamDir, "test.md")
	err = os.WriteFile(testFile, []byte("test content"), 0o600) //nolint:gosec // Test file permissions
	require.NoError(t, err)

	// Check actual directory mtime before validation
	actualDirMtime, err := cacheManager.getCurrentDirMtime()
	require.NoError(t, err)
	t.Logf("Initial cached mtime: %v", initialMtime)
	t.Logf("Actual dir mtime after file write: %v", actualDirMtime)

	// Validation should detect change and refresh
	err = cacheManager.ValidateCache()
	require.NoError(t, err)

	newMtime, err := cacheManager.getCachedDirMtime()
	require.NoError(t, err)
	t.Logf("New cached mtime: %v", newMtime)

	// Cache should be updated (either newer than initial, or same second as actual dir mtime)
	assert.True(t, newMtime.After(initialMtime) || newMtime.Unix() == actualDirMtime.Unix())
}

func TestRefreshCache(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }() //nolint:errcheck // Test cleanup

	tempDir := t.TempDir()
	flotsamDir := filepath.Join(tempDir, "flotsam")
	err := os.MkdirAll(flotsamDir, 0o750) //nolint:gosec // Test directory permissions
	require.NoError(t, err)

	// Create test markdown files
	testFiles := []string{"note1.md", "note2.md", "other.txt"}
	for _, file := range testFiles {
		filePath := filepath.Join(flotsamDir, file)
		err = os.WriteFile(filePath, []byte("content"), 0o600) //nolint:gosec // Test file permissions
		require.NoError(t, err)
	}

	cacheManager := db.GetCacheManager(tempDir)

	// Refresh cache
	err = cacheManager.RefreshCache()
	require.NoError(t, err)

	// Verify cache metadata was updated
	cachedMtime, err := cacheManager.getCachedDirMtime()
	require.NoError(t, err)
	assert.True(t, cachedMtime.Unix() > 0)

	// Verify we can get current directory mtime
	currentMtime, err := cacheManager.getCurrentDirMtime()
	require.NoError(t, err)
	assert.True(t, currentMtime.Unix() > 0)
}

func TestInvalidateCache(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }() //nolint:errcheck // Test cleanup

	tempDir := t.TempDir()
	flotsamDir := filepath.Join(tempDir, "flotsam")
	err := os.MkdirAll(flotsamDir, 0o750) //nolint:gosec // Test directory permissions
	require.NoError(t, err)

	cacheManager := db.GetCacheManager(tempDir)

	// Initial cache
	err = cacheManager.RefreshCache()
	require.NoError(t, err)

	initialMtime, err := cacheManager.getCachedDirMtime()
	require.NoError(t, err)
	assert.True(t, initialMtime.Unix() > 0)

	// Invalidate cache
	err = cacheManager.InvalidateCache()
	require.NoError(t, err)

	// Cache should be marked as invalid (epoch time)
	invalidatedMtime, err := cacheManager.getCachedDirMtime()
	require.NoError(t, err)
	assert.Equal(t, int64(0), invalidatedMtime.Unix())
}

func TestValidateCache_NonexistentDirectory(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }() //nolint:errcheck // Test cleanup

	tempDir := t.TempDir()
	// Don't create flotsam directory

	cacheManager := db.GetCacheManager(tempDir)

	// Validation should handle nonexistent directory gracefully
	err := cacheManager.ValidateCache()
	require.NoError(t, err)

	// Cache should be set with epoch time
	cachedMtime, err := cacheManager.getCachedDirMtime()
	require.NoError(t, err)
	assert.Equal(t, int64(0), cachedMtime.Unix())
}

func TestCacheMetadataSchema(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }() //nolint:errcheck // Test cleanup

	// Verify cache_metadata table exists
	query := `
		SELECT name FROM sqlite_master 
		WHERE type='table' AND name='cache_metadata'
	`
	var tableName string
	err := db.db.QueryRow(query).Scan(&tableName)
	require.NoError(t, err)
	assert.Equal(t, "cache_metadata", tableName)

	// Verify column structure
	columns, err := getTableColumns(db, "cache_metadata")
	require.NoError(t, err)

	expectedColumns := []string{"context", "last_sync", "flotsam_dir_mtime"}
	for _, col := range expectedColumns {
		assert.Contains(t, columns, col, "Missing column: %s", col)
	}
}

// Database path determination tests

func TestDetermineDatabasePath_DefaultNotebook(t *testing.T) {
	tempDir := t.TempDir()

	dbPath, err := determineDatabasePath(tempDir)
	require.NoError(t, err)

	expectedPath := filepath.Join(tempDir, "flotsam", ".vice", "flotsam.db")
	assert.Equal(t, expectedPath, dbPath)
}

func TestDetermineDatabasePath_ContextDir(t *testing.T) {
	tempDir := t.TempDir()

	dbPath, err := determineDatabasePath(tempDir)
	require.NoError(t, err)

	expectedPath := filepath.Join(tempDir, "flotsam", ".vice", "flotsam.db")
	assert.Equal(t, expectedPath, dbPath)
}

func TestNewDatabase_CreatesViceDirectory(t *testing.T) {
	tempDir := t.TempDir()

	db, err := NewDatabase(tempDir, "test-context")
	require.NoError(t, err)
	defer func() { _ = db.Close() }() //nolint:errcheck // Test cleanup

	// Verify .vice directory was created in notebook directory
	viceDir := filepath.Join(tempDir, "flotsam", ".vice")
	info, err := os.Stat(viceDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	// Verify database file was created
	assert.FileExists(t, db.dbPath)
}

func TestNewDatabase_NotebookDirectoryPlacement(t *testing.T) {
	tempDir := t.TempDir()

	db, err := NewDatabase(tempDir, "test-context")
	require.NoError(t, err)
	defer func() { _ = db.Close() }() //nolint:errcheck // Test cleanup

	// Database should be placed in flotsam/.vice/ within context
	expectedPath := filepath.Join(tempDir, "flotsam", ".vice", "flotsam.db")
	assert.Equal(t, expectedPath, db.dbPath)

	// Verify .vice directory was created in notebook directory
	viceDir := filepath.Join(tempDir, "flotsam", ".vice")
	info, err := os.Stat(viceDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}
