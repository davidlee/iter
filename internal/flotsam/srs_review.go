// Copyright (c) 2025 Vice Project
// This file contains code adapted from the go-srs spaced repetition system.
// Original code: https://github.com/revelaction/go-srs
// Original license: Apache License 2.0
//
// Portions of this file are derived from go-srs's review data structures,
// specifically from review/review.go.
// The original go-srs code is licensed under Apache-2.0.

package flotsam

import (
	"errors"
	"fmt"
	"time"
)

// Review data structures adapted for flotsam's note-based architecture
// AIDEV-NOTE: review structures adapted from go-srs for flotsam note workflows

// FlotsamReview represents a review session for flotsam notes
// Adapted from go-srs Review but uses note IDs instead of card/deck IDs
//
//revive:disable-next-line:exported FlotsamReview intentionally descriptive to distinguish from other review types
type FlotsamReview struct {
	// Context identifies the collection of notes (replaces DeckId concept)
	Context string `json:"context" yaml:"context"`

	// SessionID for tracking review sessions
	SessionID string `json:"session_id" yaml:"session_id"`

	// Review timestamp
	Timestamp time.Time `json:"timestamp" yaml:"timestamp"`

	// Items being reviewed in this session
	Items []FlotsamReviewItem `json:"items" yaml:"items"`

	// Session metadata
	TotalDuration time.Duration `json:"total_duration,omitempty" yaml:"total_duration,omitempty"`
	Completed     bool          `json:"completed" yaml:"completed"`
}

// FlotsamReviewItem represents a single note review
// Adapted from go-srs ReviewItem but uses note IDs instead of card IDs
//
//revive:disable-next-line:exported FlotsamReviewItem intentionally descriptive to distinguish from other review types
type FlotsamReviewItem struct {
	// NoteID replaces CardId - references flotsam note by ID
	NoteID string `json:"note_id" yaml:"note_id"`

	// Quality rating given by user (0-6 scale from go-srs)
	Quality Quality `json:"quality" yaml:"quality"`

	// Time taken for this specific review
	ReviewTime time.Duration `json:"review_time,omitempty" yaml:"review_time,omitempty"`

	// Timestamp when this item was reviewed
	ReviewedAt time.Time `json:"reviewed_at" yaml:"reviewed_at"`

	// Previous SRS data before this review (for rollback/analysis)
	PreviousSRSData *SRSData `json:"previous_srs_data,omitempty" yaml:"previous_srs_data,omitempty"`

	// Updated SRS data after this review
	UpdatedSRSData *SRSData `json:"updated_srs_data" yaml:"updated_srs_data"`
}

// FlotsamDue represents notes that are due for review
// Adapted from go-srs Due structure for flotsam notes
//
//revive:disable-next-line:exported FlotsamDue intentionally descriptive to distinguish from other due types
type FlotsamDue struct {
	// Context for which notes are due
	Context string `json:"context" yaml:"context"`

	// Timestamp when this due list was generated
	GeneratedAt time.Time `json:"generated_at" yaml:"generated_at"`

	// Notes that are due for review
	Items []FlotsamDueItem `json:"items" yaml:"items"`

	// Summary statistics
	TotalDue int `json:"total_due" yaml:"total_due"`
	Overdue  int `json:"overdue" yaml:"overdue"`
	NewCards int `json:"new_cards" yaml:"new_cards"`
}

// FlotsamDueItem represents a single note that's due for review
// Adapted from go-srs DueItem for flotsam notes
//
//revive:disable-next-line:exported FlotsamDueItem intentionally descriptive to distinguish from other due item types
type FlotsamDueItem struct {
	// NoteID replaces CardId
	NoteID string `json:"note_id" yaml:"note_id"`

	// When this note is due (may be in the past for overdue)
	DueAt time.Time `json:"due_at" yaml:"due_at"`

	// How overdue this note is (0 if not overdue)
	OverdueDays int `json:"overdue_days" yaml:"overdue_days"`

	// Note metadata for display/selection
	NoteTitle string `json:"note_title,omitempty" yaml:"note_title,omitempty"`
	NoteType  string `json:"note_type,omitempty" yaml:"note_type,omitempty"`

	// SRS metadata
	IsNewCard       bool    `json:"is_new_card" yaml:"is_new_card"`
	CurrentEasiness float64 `json:"current_easiness,omitempty" yaml:"current_easiness,omitempty"`
	ReviewCount     int     `json:"review_count,omitempty" yaml:"review_count,omitempty"`
}

// Review validation and utility functions
// AIDEV-NOTE: validation logic adapted from go-srs for flotsam workflows

// Validate checks if the FlotsanReview is valid
func (r *FlotsamReview) Validate() error {
	if r.Context == "" {
		return ErrInvalidContext
	}

	if len(r.Items) == 0 {
		return errors.New("review must contain at least one item")
	}

	// Validate each review item
	for i, item := range r.Items {
		if err := item.Validate(); err != nil {
			return fmt.Errorf("review item %d invalid: %w", i, err)
		}
	}

	// Check for duplicate note IDs in the same review
	noteIDs := make(map[string]bool)
	for _, item := range r.Items {
		if noteIDs[item.NoteID] {
			return fmt.Errorf("duplicate note ID in review: %s", item.NoteID)
		}
		noteIDs[item.NoteID] = true
	}

	return nil
}

// Validate checks if the FlotsamReviewItem is valid
func (item *FlotsamReviewItem) Validate() error {
	if item.NoteID == "" {
		return errors.New("note ID cannot be empty")
	}

	if err := item.Quality.Validate(); err != nil {
		return fmt.Errorf("invalid quality: %w", err)
	}

	if item.UpdatedSRSData == nil {
		return errors.New("updated SRS data is required")
	}

	return nil
}

// GetReviewCount returns the total number of items in the review
func (r *FlotsamReview) GetReviewCount() int {
	return len(r.Items)
}

// GetCorrectCount returns the number of correct answers (quality >= 4)
func (r *FlotsamReview) GetCorrectCount() int {
	count := 0
	for _, item := range r.Items {
		if item.Quality.IsCorrect() {
			count++
		}
	}
	return count
}

// GetIncorrectCount returns the number of incorrect answers (quality < 4)
func (r *FlotsamReview) GetIncorrectCount() int {
	return r.GetReviewCount() - r.GetCorrectCount()
}

// GetSuccessRate returns the percentage of correct answers
func (r *FlotsamReview) GetSuccessRate() float64 {
	if r.GetReviewCount() == 0 {
		return 0
	}
	return float64(r.GetCorrectCount()) / float64(r.GetReviewCount()) * 100
}

// GetAverageQuality returns the average quality rating for the review
func (r *FlotsamReview) GetAverageQuality() float64 {
	if len(r.Items) == 0 {
		return 0
	}

	total := 0
	for _, item := range r.Items {
		total += int(item.Quality)
	}

	return float64(total) / float64(len(r.Items))
}

// GetTotalReviewTime returns the total time spent on all reviews
func (r *FlotsamReview) GetTotalReviewTime() time.Duration {
	if r.TotalDuration > 0 {
		return r.TotalDuration
	}

	// Calculate from individual item times if total not set
	var total time.Duration
	for _, item := range r.Items {
		total += item.ReviewTime
	}

	return total
}

// GetAverageReviewTime returns the average time per item
func (r *FlotsamReview) GetAverageReviewTime() time.Duration {
	if len(r.Items) == 0 {
		return 0
	}

	return r.GetTotalReviewTime() / time.Duration(len(r.Items))
}

// HasNewCards returns true if any items in the review are new cards
func (r *FlotsamReview) HasNewCards() bool {
	for _, item := range r.Items {
		if item.PreviousSRSData == nil {
			return true
		}
	}
	return false
}

// GetNewCardCount returns the number of new cards in the review
func (r *FlotsamReview) GetNewCardCount() int {
	count := 0
	for _, item := range r.Items {
		if item.PreviousSRSData == nil {
			count++
		}
	}
	return count
}

// Validate checks if the FlotsamDue is valid
func (d *FlotsamDue) Validate() error {
	if d.Context == "" {
		return ErrInvalidContext
	}

	if len(d.Items) != d.TotalDue {
		return fmt.Errorf("item count %d does not match total due %d", len(d.Items), d.TotalDue)
	}

	// Validate each due item
	for i, item := range d.Items {
		if err := item.Validate(); err != nil {
			return fmt.Errorf("due item %d invalid: %w", i, err)
		}
	}

	return nil
}

// Validate checks if the FlotsamDueItem is valid
func (item *FlotsamDueItem) Validate() error {
	if item.NoteID == "" {
		return errors.New("note ID cannot be empty")
	}

	if item.OverdueDays < 0 {
		return errors.New("overdue days cannot be negative")
	}

	return nil
}

// GetOverdueItems returns only the overdue items from the due list
func (d *FlotsamDue) GetOverdueItems() []FlotsamDueItem {
	var overdue []FlotsamDueItem
	for _, item := range d.Items {
		if item.OverdueDays > 0 {
			overdue = append(overdue, item)
		}
	}
	return overdue
}

// GetNewCardItems returns only the new card items from the due list
func (d *FlotsamDue) GetNewCardItems() []FlotsamDueItem {
	var newCards []FlotsamDueItem
	for _, item := range d.Items {
		if item.IsNewCard {
			newCards = append(newCards, item)
		}
	}
	return newCards
}

// GetDueToday returns items that are due today (not overdue, not future)
func (d *FlotsamDue) GetDueToday() []FlotsamDueItem {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	tomorrow := today.AddDate(0, 0, 1)

	var dueToday []FlotsamDueItem
	for _, item := range d.Items {
		if !item.DueAt.Before(today) && item.DueAt.Before(tomorrow) && item.OverdueDays == 0 {
			dueToday = append(dueToday, item)
		}
	}
	return dueToday
}

// SortByDueDate sorts the due items by due date (earliest first)
func (d *FlotsamDue) SortByDueDate() {
	// Simple bubble sort - could use sort.Slice for better performance
	n := len(d.Items)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if d.Items[j].DueAt.After(d.Items[j+1].DueAt) {
				d.Items[j], d.Items[j+1] = d.Items[j+1], d.Items[j]
			}
		}
	}
}

// SortByOverdue sorts the due items by overdue days (most overdue first)
func (d *FlotsamDue) SortByOverdue() {
	n := len(d.Items)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if d.Items[j].OverdueDays < d.Items[j+1].OverdueDays {
				d.Items[j], d.Items[j+1] = d.Items[j+1], d.Items[j]
			}
		}
	}
}

// CreateFlotsamReview creates a new review session
func CreateFlotsamReview(context, sessionID string) *FlotsamReview {
	return &FlotsamReview{
		Context:   context,
		SessionID: sessionID,
		Timestamp: time.Now(),
		Items:     make([]FlotsamReviewItem, 0),
		Completed: false,
	}
}

// AddReviewItem adds a review item to the review session
func (r *FlotsamReview) AddReviewItem(noteID string, quality Quality, reviewTime time.Duration, previousSRS, updatedSRS *SRSData) {
	item := FlotsamReviewItem{
		NoteID:          noteID,
		Quality:         quality,
		ReviewTime:      reviewTime,
		ReviewedAt:      time.Now(),
		PreviousSRSData: previousSRS,
		UpdatedSRSData:  updatedSRS,
	}

	r.Items = append(r.Items, item)
}

// CompleteReview marks the review as completed and calculates final statistics
func (r *FlotsamReview) CompleteReview() {
	r.Completed = true
	r.TotalDuration = r.GetTotalReviewTime()
}

// CreateFlotsamDue creates a new due list for a context
func CreateFlotsamDue(context string) *FlotsamDue {
	return &FlotsamDue{
		Context:     context,
		GeneratedAt: time.Now(),
		Items:       make([]FlotsamDueItem, 0),
		TotalDue:    0,
		Overdue:     0,
		NewCards:    0,
	}
}

// AddDueItem adds a due item to the due list
func (d *FlotsamDue) AddDueItem(noteID string, dueAt time.Time, isNewCard bool, title, noteType string, easiness float64, reviewCount int) {
	now := time.Now()
	overdueDays := 0
	if dueAt.Before(now) && !isNewCard {
		overdueDays = int(now.Sub(dueAt).Hours() / 24)
	}

	item := FlotsamDueItem{
		NoteID:          noteID,
		DueAt:           dueAt,
		OverdueDays:     overdueDays,
		NoteTitle:       title,
		NoteType:        noteType,
		IsNewCard:       isNewCard,
		CurrentEasiness: easiness,
		ReviewCount:     reviewCount,
	}

	d.Items = append(d.Items, item)
	d.TotalDue++

	if overdueDays > 0 {
		d.Overdue++
	}

	if isNewCard {
		d.NewCards++
	}
}
