// Copyright (c) 2025 Vice Project
// This file contains code adapted from the go-srs spaced repetition system.
// Original code: https://github.com/revelaction/go-srs
// Original license: Apache License 2.0
// 
// Portions of this file are derived from go-srs's SM-2 algorithm implementation,
// specifically from algo/sm2/sm2.go and review/review.go.
// The original go-srs code is licensed under Apache-2.0.

// Package flotsam provides SRS (Spaced Repetition System) implementation using SM-2 algorithm.
// AIDEV-NOTE: SM-2 algorithm implementation adapted from go-srs for flotsam note review
package flotsam

import (
	"encoding/json"
	"errors"
	"math"
	"time"
)

// SRS constants and configuration
const (
	// DefaultEasiness is the starting easiness factor for new cards
	DefaultEasiness = 2.5
	// MinEasiness is the minimum allowed easiness factor
	MinEasiness = 1.3
	
	// Easiness calculation constants from SM-2 algorithm
	EasinessConst     = -0.8
	EasinessLineal    = 0.28
	EasinessQuadratic = 0.02
	
	// DueDateStartDays is the base interval for calculating due dates
	DueDateStartDays = 6
	// CorrectThreshold is the minimum quality rating considered "correct"
	CorrectThreshold = 4
)

// Quality represents the user's self-evaluation of their recall performance.
// AIDEV-NOTE: quality scale from go-srs - 0-6 range where 0=no review, 1-3=incorrect, 4-6=correct
type Quality int

const (
	// NoReview indicates no review was performed (0)
	NoReview Quality = iota
	// IncorrectBlackout indicates total failure to recall (1)
	IncorrectBlackout
	// IncorrectFamiliar indicates incorrect but familiar upon seeing answer (2)
	IncorrectFamiliar
	// IncorrectEasy indicates incorrect but seemed easy upon seeing answer (3)
	IncorrectEasy
	// CorrectHard indicates correct but required significant difficulty (4)
	CorrectHard
	// CorrectEffort indicates correct after some hesitation (5)
	CorrectEffort
	// CorrectEasy indicates correct with perfect recall (6)
	CorrectEasy
)

// Validate checks if the quality value is within valid range
func (q Quality) Validate() error {
	if q > CorrectEasy || q < NoReview {
		return errors.New("invalid quality: must be between 0 and 6")
	}
	return nil
}

// IsCorrect returns true if the quality represents a correct answer
func (q Quality) IsCorrect() bool {
	return q >= CorrectHard
}

// SRSData represents the spaced repetition data stored in flotsam frontmatter.
// AIDEV-NOTE: matches flotsam frontmatter schema from task requirements
type SRSData struct {
	// Easiness factor (default 2.5, minimum 1.3)
	Easiness float64 `yaml:"easiness" json:"easiness"`
	// Number of consecutive correct answers
	ConsecutiveCorrect int `yaml:"consecutive_correct" json:"consecutive_correct"`
	// Unix timestamp when the card is due for review
	Due int64 `yaml:"due" json:"due"`
	// Total number of reviews performed
	TotalReviews int `yaml:"total_reviews" json:"total_reviews"`
	// Optional: Review history for debugging/analysis
	ReviewHistory []ReviewRecord `yaml:"review_history,omitempty" json:"review_history,omitempty"`
}

// ReviewRecord represents a single review session for history tracking.
type ReviewRecord struct {
	// Unix timestamp when review was performed
	Timestamp int64 `yaml:"timestamp" json:"timestamp"`
	// Quality rating given by user (0-6)
	Quality Quality `yaml:"quality" json:"quality"`
}

// SM2Calculator implements the SuperMemo 2 spaced repetition algorithm.
// AIDEV-NOTE: core SM-2 implementation adapted from go-srs for flotsam use
type SM2Calculator struct {
	// Current time for calculations (allows testing with fixed time)
	now time.Time
}

// NewSM2Calculator creates a new SM-2 calculator with the current time
func NewSM2Calculator() *SM2Calculator {
	return &SM2Calculator{now: time.Now()}
}

// NewSM2CalculatorWithTime creates a new SM-2 calculator with a specific time (for testing)
func NewSM2CalculatorWithTime(t time.Time) *SM2Calculator {
	return &SM2Calculator{now: t}
}

// ProcessReview updates SRS data based on a review session.
// AIDEV-NOTE: main SRS algorithm - processes quality rating and updates scheduling
func (calc *SM2Calculator) ProcessReview(oldData *SRSData, quality Quality) (*SRSData, error) {
	if err := quality.Validate(); err != nil {
		return nil, err
	}
	
	var newData SRSData
	
	// Initialize new card if no previous data
	if oldData == nil {
		newData = calc.createNewCard(quality)
	} else {
		newData = calc.updateCard(*oldData, quality)
	}
	
	// Add review to history
	newData.ReviewHistory = append(newData.ReviewHistory, ReviewRecord{
		Timestamp: calc.now.Unix(),
		Quality:   quality,
	})
	
	return &newData, nil
}

// createNewCard initializes SRS data for a brand new card
func (calc *SM2Calculator) createNewCard(quality Quality) SRSData {
	data := SRSData{
		TotalReviews: 1,
	}
	
	if quality == NoReview {
		// No review performed - set defaults
		data.ConsecutiveCorrect = 0
		data.Easiness = DefaultEasiness
		data.Due = calc.now.AddDate(0, 0, 1).Unix() // Due tomorrow
	} else {
		// First review for new card
		data.Easiness = calc.calculateEasiness(DefaultEasiness, quality)
		if quality.IsCorrect() {
			data.ConsecutiveCorrect = 1
		} else {
			data.ConsecutiveCorrect = 0
		}
		data.Due = calc.now.AddDate(0, 0, 1).Unix() // Due tomorrow
	}
	
	return data
}

// updateCard updates existing SRS data based on review performance
func (calc *SM2Calculator) updateCard(oldData SRSData, quality Quality) SRSData {
	newData := SRSData{
		TotalReviews:   oldData.TotalReviews + 1,
		ReviewHistory:  oldData.ReviewHistory, // Will be appended to by caller
	}
	
	// Update easiness factor
	newData.Easiness = calc.calculateEasiness(oldData.Easiness, quality)
	
	// Update consecutive correct count and due date
	if quality.IsCorrect() {
		newData.ConsecutiveCorrect = oldData.ConsecutiveCorrect + 1
		
		// Calculate next due date based on consecutive correct answers
		var days float64
		switch oldData.ConsecutiveCorrect {
		case 0:
			days = 1 // First correct answer: 1 day
		case 1:
			days = float64(DueDateStartDays) // Second correct: 6 days
		default:
			// Subsequent correct answers: exponential growth
			days = float64(DueDateStartDays) * math.Pow(oldData.Easiness, float64(oldData.ConsecutiveCorrect-1))
		}
		
		newData.Due = calc.now.AddDate(0, 0, int(math.Round(days))).Unix()
	} else {
		// Incorrect answer: reset to beginning
		newData.ConsecutiveCorrect = 0
		newData.Due = calc.now.AddDate(0, 0, 1).Unix() // Due tomorrow
	}
	
	return newData
}

// calculateEasiness computes the new easiness factor based on quality rating.
// AIDEV-NOTE: SM-2 easiness formula from go-srs implementation
func (calc *SM2Calculator) calculateEasiness(oldEasiness float64, quality Quality) float64 {
	// Convert quality (0-6) to SM-2 scale: quality 0 maps to -1, quality 1-6 maps to 0-5
	// This matches go-srs's quality conversion: q = quality - 1
	q := float64(quality - 1)
	
	// SM-2 easiness formula with BlueRaja modifications
	newEasiness := oldEasiness + EasinessConst + (EasinessLineal * q) + (EasinessQuadratic * q * q)
	
	// Enforce minimum easiness
	if newEasiness < MinEasiness {
		return MinEasiness
	}
	
	return newEasiness
}

// IsDue checks if a card is due for review based on its SRS data
func (calc *SM2Calculator) IsDue(data *SRSData) bool {
	if data == nil {
		return true // New cards are always due
	}
	return data.Due <= calc.now.Unix()
}

// IsDueAt checks if a card is due for review at a specific time
func (calc *SM2Calculator) IsDueAt(data *SRSData, t time.Time) bool {
	if data == nil {
		return true // New cards are always due
	}
	return data.Due <= t.Unix()
}

// GetDueTime returns when the card is next due for review
func (calc *SM2Calculator) GetDueTime(data *SRSData) time.Time {
	if data == nil {
		return calc.now // New cards are due now
	}
	return time.Unix(data.Due, 0)
}

// GetNextInterval returns the number of days until the next review
func (calc *SM2Calculator) GetNextInterval(data *SRSData) int {
	if data == nil {
		return 0 // New cards have no interval
	}
	
	dueTime := time.Unix(data.Due, 0)
	duration := dueTime.Sub(calc.now)
	days := int(math.Ceil(duration.Hours() / 24))
	
	if days < 0 {
		return 0 // Overdue cards
	}
	
	return days
}

// SerializeSRSData converts SRS data to JSON for storage
func SerializeSRSData(data *SRSData) ([]byte, error) {
	if data == nil {
		return []byte{}, nil
	}
	return json.Marshal(data)
}

// ErrEmptyData is returned when trying to deserialize empty data
var ErrEmptyData = errors.New("empty data provided")

// DeserializeSRSData converts JSON back to SRS data
func DeserializeSRSData(jsonData []byte) (*SRSData, error) {
	if len(jsonData) == 0 {
		return nil, ErrEmptyData
	}
	
	var data SRSData
	err := json.Unmarshal(jsonData, &data)
	if err != nil {
		return nil, err
	}
	
	return &data, nil
}