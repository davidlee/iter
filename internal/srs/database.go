// Package srs provides SRS database functionality for flotsam Unix interop.
// AIDEV-NOTE: minimal SRS database for Unix interop - stores scheduling data separate from note content
package srs

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite driver for database/sql
)

// Database manages SRS scheduling data in SQLite.
// AIDEV-NOTE: database location strategy - .vice/flotsam.db separate from zk notebook.db
type Database struct {
	db      *sql.DB
	dbPath  string
	context string
}

//revive:disable-next-line:exported SRSNote prefixed for clarity with flotsam package types
type SRSNote struct {
	NotePath           string     `json:"note_path"`
	NoteID             string     `json:"note_id"`
	Context            string     `json:"context"`
	Easiness           float64    `json:"easiness"`
	ConsecutiveCorrect int        `json:"consecutive_correct"`
	DueDate            time.Time  `json:"due_date"`
	TotalReviews       int        `json:"total_reviews"`
	CreatedAt          time.Time  `json:"created_at"`
	LastReviewed       *time.Time `json:"last_reviewed,omitempty"`
}

//revive:disable-next-line:exported SRSData matches flotsam.SRSData for algorithm compatibility
type SRSData struct {
	Easiness           float64 `json:"easiness"`
	ConsecutiveCorrect int     `json:"consecutive_correct"`
	Due                int64   `json:"due"` // Unix timestamp
	TotalReviews       int     `json:"total_reviews"`
}

// NewDatabase creates a new SRS database connection.
// AIDEV-NOTE: database path convention - finds notebook root and places .vice/ alongside .zk/
func NewDatabase(contextDir, context string) (*Database, error) {
	dbPath, err := determineDatabasePath(contextDir)
	if err != nil {
		return nil, fmt.Errorf("failed to determine database path: %w", err)
	}

	// Ensure .vice directory exists
	viceDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(viceDir, 0750); err != nil {
		return nil, fmt.Errorf("failed to create .vice directory: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open SRS database: %w", err)
	}

	srsDB := &Database{
		db:      db,
		dbPath:  dbPath,
		context: context,
	}

	if err := srsDB.ensureSchema(); err != nil {
		_ = db.Close() //nolint:errcheck // Error already being returned
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return srsDB, nil
}

// GetCacheManager creates a cache manager for this database.
func (d *Database) GetCacheManager(contextDir string) *CacheManager {
	return NewCacheManager(d, contextDir)
}

// Close closes the database connection.
func (d *Database) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

// ensureSchema creates the SRS database schema if it doesn't exist.
// AIDEV-NOTE: minimal schema per flotsam.md specification - note_path primary key, SM-2 fields
func (d *Database) ensureSchema() error {
	// Create main table
	tableSchema := `
		CREATE TABLE IF NOT EXISTS srs_reviews (
			note_path TEXT PRIMARY KEY,
			note_id TEXT NOT NULL,
			context TEXT NOT NULL,
			
			-- SM-2 algorithm fields
			easiness REAL NOT NULL DEFAULT 2.5,
			consecutive_correct INTEGER NOT NULL DEFAULT 0,
			due_date INTEGER NOT NULL,
			total_reviews INTEGER NOT NULL DEFAULT 0,
			
			-- Metadata
			created_at INTEGER NOT NULL,
			last_reviewed INTEGER
		);
	`

	_, err := d.db.Exec(tableSchema)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	// Create cache metadata table for mtime tracking
	cacheMetadataSchema := `
		CREATE TABLE IF NOT EXISTS cache_metadata (
			context TEXT PRIMARY KEY,
			last_sync INTEGER NOT NULL,
			flotsam_dir_mtime INTEGER NOT NULL
		);
	`

	_, err = d.db.Exec(cacheMetadataSchema)
	if err != nil {
		return fmt.Errorf("failed to create cache metadata table: %w", err)
	}

	// Create performance indexes
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_srs_due_date ON srs_reviews (due_date);`,
		`CREATE INDEX IF NOT EXISTS idx_srs_context ON srs_reviews (context);`,
		`CREATE INDEX IF NOT EXISTS idx_srs_context_due ON srs_reviews (context, due_date);`,
	}

	for _, indexSQL := range indexes {
		_, err := d.db.Exec(indexSQL)
		if err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}

// GetDueNotes returns all notes due for review in the given context.
func (d *Database) GetDueNotes(contextName string) ([]SRSNote, error) {
	query := `
		SELECT note_path, note_id, context, easiness, consecutive_correct, 
		       due_date, total_reviews, created_at, last_reviewed
		FROM srs_reviews 
		WHERE context = ? AND due_date <= ?
		ORDER BY due_date ASC
	`

	now := time.Now().Unix()
	rows, err := d.db.Query(query, contextName, now)
	if err != nil {
		return nil, fmt.Errorf("failed to query due notes: %w", err)
	}
	defer func() { _ = rows.Close() }() //nolint:errcheck // Defer cleanup

	var notes []SRSNote
	for rows.Next() {
		var note SRSNote
		var dueDate, createdAt int64
		var lastReviewed sql.NullInt64

		err := rows.Scan(
			&note.NotePath, &note.NoteID, &note.Context,
			&note.Easiness, &note.ConsecutiveCorrect,
			&dueDate, &note.TotalReviews,
			&createdAt, &lastReviewed,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan note: %w", err)
		}

		// Convert Unix timestamps to time.Time
		note.DueDate = time.Unix(dueDate, 0)
		note.CreatedAt = time.Unix(createdAt, 0)
		if lastReviewed.Valid {
			t := time.Unix(lastReviewed.Int64, 0)
			note.LastReviewed = &t
		}

		notes = append(notes, note)
	}

	return notes, rows.Err()
}

// UpdateReview updates SRS data after a review session.
func (d *Database) UpdateReview(notePath string, data *SRSData) error {
	query := `
		UPDATE srs_reviews 
		SET easiness = ?, consecutive_correct = ?, due_date = ?, 
		    total_reviews = total_reviews + 1, last_reviewed = ?
		WHERE note_path = ?
	`

	now := time.Now().Unix()
	_, err := d.db.Exec(query, data.Easiness, data.ConsecutiveCorrect,
		data.Due, now, notePath)
	if err != nil {
		return fmt.Errorf("failed to update review: %w", err)
	}

	return nil
}

// GetSRSData retrieves SRS data for a specific note.
func (d *Database) GetSRSData(notePath string) (*SRSData, error) {
	query := `
		SELECT easiness, consecutive_correct, due_date, total_reviews
		FROM srs_reviews 
		WHERE note_path = ?
	`

	var data SRSData
	err := d.db.QueryRow(query, notePath).Scan(
		&data.Easiness, &data.ConsecutiveCorrect,
		&data.Due, &data.TotalReviews,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("note not found: %s", notePath)
		}
		return nil, fmt.Errorf("failed to get SRS data: %w", err)
	}

	return &data, nil
}

// CreateSRSNote creates a new SRS note entry.
func (d *Database) CreateSRSNote(notePath, noteID, context string, initialData *SRSData) error {
	query := `
		INSERT INTO srs_reviews 
		(note_path, note_id, context, easiness, consecutive_correct, 
		 due_date, total_reviews, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now().Unix()
	_, err := d.db.Exec(query, notePath, noteID, context,
		initialData.Easiness, initialData.ConsecutiveCorrect,
		initialData.Due, initialData.TotalReviews, now)
	if err != nil {
		return fmt.Errorf("failed to create SRS note: %w", err)
	}

	return nil
}

// DeleteSRSNote removes a note from SRS tracking.
func (d *Database) DeleteSRSNote(notePath string) error {
	query := `DELETE FROM srs_reviews WHERE note_path = ?`

	_, err := d.db.Exec(query, notePath)
	if err != nil {
		return fmt.Errorf("failed to delete SRS note: %w", err)
	}

	return nil
}

// GetStats returns SRS statistics for a context.
func (d *Database) GetStats(contextName string) (map[string]interface{}, error) {
	query := `
		SELECT 
			COUNT(*) as total_notes,
			COUNT(CASE WHEN due_date <= ? THEN 1 END) as due_notes,
			AVG(easiness) as avg_easiness,
			AVG(total_reviews) as avg_reviews
		FROM srs_reviews 
		WHERE context = ?
	`

	now := time.Now().Unix()
	var totalNotes, dueNotes int64
	var avgEasiness, avgReviews float64

	err := d.db.QueryRow(query, now, contextName).Scan(
		&totalNotes, &dueNotes, &avgEasiness, &avgReviews,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	stats := map[string]interface{}{
		"total_notes":  totalNotes,
		"due_notes":    dueNotes,
		"avg_easiness": avgEasiness,
		"avg_reviews":  avgReviews,
	}

	return stats, nil
}

// CacheManager handles mtime-based cache invalidation for SRS database.
// AIDEV-NOTE: mtime-based cache invalidation for Unix interop - directory-level checks for performance
type CacheManager struct {
	db          *Database
	contextDir  string
	flotsamDir  string
}

// NewCacheManager creates a new cache manager for the given database and context.
func NewCacheManager(db *Database, contextDir string) *CacheManager {
	flotsamDir := filepath.Join(contextDir, "flotsam")
	return &CacheManager{
		db:         db,
		contextDir: contextDir,
		flotsamDir: flotsamDir,
	}
}

// ValidateCache checks if the cache is up-to-date and refreshes if necessary.
// AIDEV-NOTE: fast directory-level mtime check before expensive file scanning
func (c *CacheManager) ValidateCache() error {
	// 1. Get cached directory mtime
	cachedMtime, err := c.getCachedDirMtime()
	if err != nil {
		// Cache miss - do full refresh
		return c.RefreshCache()
	}

	// 2. Check current directory mtime
	currentMtime, err := c.getCurrentDirMtime()
	if err != nil {
		// Directory doesn't exist or error - cache is invalid
		return c.RefreshCache()
	}

	// 3. If directory unchanged, cache is valid
	if !currentMtime.After(cachedMtime) {
		return nil
	}

	// 4. Directory changed - refresh cache
	return c.RefreshCache()
}

// RefreshCache scans files and updates cache with current state.
// AIDEV-NOTE: file-level granular refresh for precise cache updates
func (c *CacheManager) RefreshCache() error {
	// 1. Scan flotsam directory for markdown files (for future file-level caching)
	pattern := filepath.Join(c.flotsamDir, "*.md")
	_, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to scan flotsam files: %w", err)
	}

	// 2. Get current directory mtime
	currentDirMtime, err := c.getCurrentDirMtime()
	if err != nil {
		// If directory doesn't exist, set to epoch time
		currentDirMtime = time.Unix(0, 0)
	}

	// 3. Update cache metadata
	err = c.updateCacheMetadata(currentDirMtime)
	if err != nil {
		return fmt.Errorf("failed to update cache metadata: %w", err)
	}

	// 4. For Unix interop, we don't implement file-level caching yet
	// The SRS database will be the source of truth for SRS data
	// File content parsing is delegated to zk or direct file reads
	
	return nil
}

// getCachedDirMtime retrieves the cached directory modification time.
func (c *CacheManager) getCachedDirMtime() (time.Time, error) {
	query := `SELECT flotsam_dir_mtime FROM cache_metadata WHERE context = ?`
	
	var mtimeUnix int64
	err := c.db.db.QueryRow(query, c.db.context).Scan(&mtimeUnix)
	if err != nil {
		return time.Time{}, err
	}
	
	return time.Unix(mtimeUnix, 0), nil
}

// getCurrentDirMtime gets the current modification time of the flotsam directory.
func (c *CacheManager) getCurrentDirMtime() (time.Time, error) {
	info, err := os.Stat(c.flotsamDir)
	if err != nil {
		return time.Time{}, err
	}
	
	return info.ModTime(), nil
}

// updateCacheMetadata updates the cache metadata with current timestamps.
func (c *CacheManager) updateCacheMetadata(dirMtime time.Time) error {
	query := `
		INSERT OR REPLACE INTO cache_metadata 
		(context, last_sync, flotsam_dir_mtime)
		VALUES (?, ?, ?)
	`
	
	now := time.Now().Unix()
	_, err := c.db.db.Exec(query, c.db.context, now, dirMtime.Unix())
	if err != nil {
		return fmt.Errorf("failed to update cache metadata: %w", err)
	}
	
	return nil
}

// InvalidateCache marks the cache as invalid by setting old timestamps.
func (c *CacheManager) InvalidateCache() error {
	oldTime := time.Unix(0, 0)
	return c.updateCacheMetadata(oldTime)
}

// determineDatabasePath implements ADR-004 database placement strategy.
// AIDEV-NOTE: places .vice/flotsam.db alongside .zk/ in notebook root, or in contextDir if no zk
func determineDatabasePath(contextDir string) (string, error) {
	// 1. Check if contextDir is or contains a zk notebook
	notebookRoot, isZKNotebook := findZKNotebookRoot(contextDir)
	
	if isZKNotebook {
		// Place .vice/flotsam.db alongside .zk/notebook.db
		return filepath.Join(notebookRoot, ".vice", "flotsam.db"), nil
	}
	
	// 2. No zk notebook found - use contextDir/.vice/flotsam.db
	return filepath.Join(contextDir, ".vice", "flotsam.db"), nil
}

// findZKNotebookRoot searches for .zk directory in contextDir and parent directories.
// Returns (notebookRoot, true) if found, ("", false) if not found.
func findZKNotebookRoot(startDir string) (string, bool) {
	currentDir := startDir
	
	for {
		// Check if .zk directory exists in current directory
		zkDir := filepath.Join(currentDir, ".zk")
		if info, err := os.Stat(zkDir); err == nil && info.IsDir() {
			// Found .zk directory - this is the notebook root
			return currentDir, true
		}
		
		// Move up one directory
		parentDir := filepath.Dir(currentDir)
		
		// Stop if we've reached the root or can't go up further
		if parentDir == currentDir || parentDir == "/" {
			break
		}
		
		currentDir = parentDir
	}
	
	return "", false
}
