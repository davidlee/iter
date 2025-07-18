// Copyright (c) 2025 Vice Project
// This file contains code adapted from the go-srs spaced repetition system.
// Original code: https://github.com/revelaction/go-srs
// Original license: Apache License 2.0
//
// Portions of this file are derived from go-srs's algorithm and database interfaces,
// specifically from algo/algo.go and db/db.go.
// The original go-srs code is licensed under Apache-2.0.

package flotsam

import (
	"errors"
	"time"
)

// SRS Algorithm Interface - adapted from go-srs algo.Algo
// AIDEV-NOTE: algorithm interface adapted from go-srs for flotsam markdown-based storage

// Algorithm represents a spaced repetition algorithm (e.g., SM-2, SM-18, etc.)
type Algorithm interface {
	// ProcessReview updates SRS data based on a review session
	// Returns updated SRS data or error if review is invalid
	ProcessReview(oldData *SRSData, quality Quality) (*SRSData, error)

	// IsDue checks if a card is due for review at the current time
	IsDue(data *SRSData) bool

	// IsDueAt checks if a card is due for review at a specific time
	IsDueAt(data *SRSData, t time.Time) bool

	// GetDueTime returns when the card is next due for review
	GetDueTime(data *SRSData) time.Time

	// GetNextInterval returns the number of days until the next review
	GetNextInterval(data *SRSData) int
}

// SRS Storage Interface - adapted from go-srs db.Handler for flotsam file-based storage
// AIDEV-NOTE: storage interface adapted for flotsam's markdown-first architecture

// SRSStorage handles persistence of SRS data in flotsam notes
type SRSStorage interface {
	// LoadSRSData loads SRS data for a specific note by ID
	// Returns nil if note has no SRS data (new card)
	LoadSRSData(noteID string) (*SRSData, error)

	// SaveSRSData persists SRS data for a specific note
	// Updates the note's frontmatter with new SRS information
	SaveSRSData(noteID string, data *SRSData) error

	// GetDueCards returns all note IDs that are due for review at time t
	GetDueCards(t time.Time) ([]string, error)

	// GetDueCardsByContext returns due cards filtered by context
	GetDueCardsByContext(context string, t time.Time) ([]string, error)

	// ListAllSRSCards returns all note IDs that have SRS data
	ListAllSRSCards() ([]string, error)

	// DeleteSRSData removes SRS data from a note (makes it non-SRS)
	DeleteSRSData(noteID string) error
}

// SRS Manager - combines algorithm and storage for complete SRS functionality
// AIDEV-NOTE: high-level SRS manager for flotsam note review workflows

// SRSManager provides high-level SRS operations for flotsam notes
type SRSManager interface {
	// ReviewNote processes a review for a specific note
	// Returns updated SRS data and any due cards after the review
	ReviewNote(noteID string, quality Quality) (*SRSData, error)

	// GetDueNotes returns all notes due for review
	GetDueNotes() ([]*FlotsamNote, error)

	// GetDueNotesInContext returns due notes filtered by context
	GetDueNotesInContext(context string) ([]*FlotsamNote, error)

	// InitializeSRS enables SRS for a note (first review)
	InitializeSRS(noteID string) error

	// DisableSRS removes SRS from a note
	DisableSRS(noteID string) error

	// GetSRSStats returns statistics about SRS usage
	GetSRSStats() (*SRSStats, error)

	// GetSRSStatsForContext returns statistics for a specific context
	GetSRSStatsForContext(context string) (*SRSStats, error)
}

// FlotsamNote represents a complete flotsam note with SRS capabilities
// AIDEV-NOTE: note representation combining content and SRS metadata
//
//revive:disable-next-line:exported FlotsamNote intentionally descriptive to distinguish from other note types
type FlotsamNote struct {
	// Core note data
	ID       string    `yaml:"id" json:"id"`
	Title    string    `yaml:"title" json:"title"`
	Type     string    `yaml:"type" json:"type"` // idea, flashcard, script, log
	Tags     []string  `yaml:"tags" json:"tags"`
	Created  time.Time `yaml:"created-at" json:"created-at"`
	Modified time.Time `yaml:"-" json:"-"` // File modification time

	// Content
	Body      string   `yaml:"-" json:"-"` // Markdown body content
	Links     []string `yaml:"-" json:"-"` // Extracted [[wikilinks]]
	Backlinks []string `yaml:"-" json:"-"` // Computed reverse links
	FilePath  string   `yaml:"-" json:"-"` // Absolute file path

	// SRS data (optional)
	SRS *SRSData `yaml:"srs,omitempty" json:"srs,omitempty"`
}

// SRSStats provides statistics about SRS usage and performance
type SRSStats struct {
	// Card counts
	TotalCards  int `json:"total_cards"`
	NewCards    int `json:"new_cards"`    // Cards never reviewed
	ReviewCards int `json:"review_cards"` // Cards with review history
	DueCards    int `json:"due_cards"`    // Cards due for review now

	// Review statistics
	TotalReviews     int     `json:"total_reviews"`
	CorrectReviews   int     `json:"correct_reviews"`
	IncorrectReviews int     `json:"incorrect_reviews"`
	SuccessRate      float64 `json:"success_rate"` // Percentage correct

	// Timing statistics
	AverageEasiness  float64 `json:"average_easiness"`
	AverageInterval  float64 `json:"average_interval"`  // Days
	LongestInterval  int     `json:"longest_interval"`  // Days
	ShortestInterval int     `json:"shortest_interval"` // Days

	// Due distribution
	DueToday     int `json:"due_today"`
	DueTomorrow  int `json:"due_tomorrow"`
	DueThisWeek  int `json:"due_this_week"`
	DueThisMonth int `json:"due_this_month"`
	Overdue      int `json:"overdue"`
}

// Common SRS errors
var (
	// ErrNoteNotFound is returned when a note ID doesn't exist
	ErrNoteNotFound = errors.New("note not found")

	// ErrNoSRSData is returned when a note has no SRS data
	ErrNoSRSData = errors.New("note has no SRS data")

	// ErrSRSAlreadyEnabled is returned when trying to initialize SRS on a note that already has it
	ErrSRSAlreadyEnabled = errors.New("SRS already enabled for this note")

	// ErrInvalidContext is returned for invalid context names
	ErrInvalidContext = errors.New("invalid context")

	// ErrStorageFailure is returned for file system or storage errors
	ErrStorageFailure = errors.New("storage operation failed")
)

// SRS Configuration and Options
// AIDEV-NOTE: configuration for SRS behavior and defaults

// SRSConfig holds configuration for SRS behavior
type SRSConfig struct {
	// Algorithm to use (currently only SM-2 supported)
	Algorithm string `yaml:"algorithm" json:"algorithm"`

	// Default quality for new cards
	DefaultQuality Quality `yaml:"default_quality" json:"default_quality"`

	// Maximum cards to review per session
	MaxCardsPerSession int `yaml:"max_cards_per_session" json:"max_cards_per_session"`

	// Whether to include review history in frontmatter
	IncludeHistory bool `yaml:"include_history" json:"include_history"`

	// Auto-enable SRS for new flashcard-type notes
	AutoEnableForFlashcards bool `yaml:"auto_enable_flashcards" json:"auto_enable_flashcards"`

	// Context filtering enabled
	ContextFiltering bool `yaml:"context_filtering" json:"context_filtering"`
}

// DefaultSRSConfig returns the default SRS configuration
func DefaultSRSConfig() *SRSConfig {
	return &SRSConfig{
		Algorithm:               "sm2",
		DefaultQuality:          NoReview,
		MaxCardsPerSession:      50,
		IncludeHistory:          true,
		AutoEnableForFlashcards: true,
		ContextFiltering:        true,
	}
}

// ReviewSession represents an active SRS review session
// AIDEV-NOTE: session management for batched reviews
type ReviewSession struct {
	// Session metadata
	StartTime time.Time `json:"start_time"`
	Context   string    `json:"context"`

	// Cards in this session
	DueCards      []*FlotsamNote `json:"due_cards"`
	ReviewedCards []*FlotsamNote `json:"reviewed_cards"`

	// Progress tracking
	CurrentIndex  int `json:"current_index"`
	TotalCards    int `json:"total_cards"`
	CorrectCount  int `json:"correct_count"`
	ReviewedCount int `json:"reviewed_count"`

	// Session statistics
	SessionStats *SessionStats `json:"session_stats"`
}

// SessionStats tracks statistics for a review session
type SessionStats struct {
	Duration         time.Duration `json:"duration"`
	CardsReviewed    int           `json:"cards_reviewed"`
	CorrectAnswers   int           `json:"correct_answers"`
	IncorrectAnswers int           `json:"incorrect_answers"`
	SuccessRate      float64       `json:"success_rate"`
	AverageTime      time.Duration `json:"average_time_per_card"`
}

// ReviewSessionManager handles review session lifecycle
type ReviewSessionManager interface {
	// StartSession begins a new review session
	StartSession(context string, maxCards int) (*ReviewSession, error)

	// GetCurrentCard returns the current card for review
	GetCurrentCard(session *ReviewSession) (*FlotsamNote, error)

	// SubmitReview processes a review and moves to the next card
	SubmitReview(session *ReviewSession, quality Quality) error

	// CompleteSession finalizes the session and returns statistics
	CompleteSession(session *ReviewSession) (*SessionStats, error)

	// PauseSession saves session state for later resumption
	PauseSession(session *ReviewSession) error

	// ResumeSession loads a previously paused session
	ResumeSession(sessionID string) (*ReviewSession, error)
}
