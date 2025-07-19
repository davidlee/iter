// Package srs provides SRS database functionality for flotsam Unix interop.
// AIDEV-NOTE: minimal SRS database for Unix interop - stores scheduling data separate from note content
package srs

import (
	"database/sql"
	"fmt"
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
// AIDEV-NOTE: database path convention - contextDir/.vice/flotsam.db for Unix interop
func NewDatabase(contextDir, context string) (*Database, error) {
	dbPath := filepath.Join(contextDir, ".vice", "flotsam.db")

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
