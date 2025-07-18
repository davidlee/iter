// Package models provides validation tests for flotsam data structures.
package models

import (
	"strings"
	"testing"
	"time"

	"github.com/davidlee/vice/internal/flotsam"
)

func TestFlotsamNoteValidate(t *testing.T) {
	tests := []struct {
		name    string
		note    FlotsamNote
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid note",
			note: FlotsamNote{
				FlotsamNote: flotsam.FlotsamNote{
					ID:       "abc1",
					Title:    "Test Note",
					Type:     "idea",
					Created:  time.Now(),
					Modified: time.Now(),
				},
			},
			wantErr: false,
		},
		{
			name: "invalid ID - too short",
			note: FlotsamNote{
				FlotsamNote: flotsam.FlotsamNote{
					ID:      "ab1",
					Title:   "Test Note",
					Type:    "idea",
					Created: time.Now(),
				},
			},
			wantErr: true,
			errMsg:  "flotsam ID 'ab1' is invalid: must be 4-character alphanumeric",
		},
		{
			name: "invalid ID - too long",
			note: FlotsamNote{
				FlotsamNote: flotsam.FlotsamNote{
					ID:      "abc12",
					Title:   "Test Note",
					Type:    "idea",
					Created: time.Now(),
				},
			},
			wantErr: true,
			errMsg:  "flotsam ID 'abc12' is invalid: must be 4-character alphanumeric",
		},
		{
			name: "invalid ID - uppercase",
			note: FlotsamNote{
				FlotsamNote: flotsam.FlotsamNote{
					ID:      "ABC1",
					Title:   "Test Note",
					Type:    "idea",
					Created: time.Now(),
				},
			},
			wantErr: true,
			errMsg:  "flotsam ID 'ABC1' is invalid: must be 4-character alphanumeric",
		},
		{
			name: "invalid ID - special characters",
			note: FlotsamNote{
				FlotsamNote: flotsam.FlotsamNote{
					ID:      "ab-1",
					Title:   "Test Note",
					Type:    "idea",
					Created: time.Now(),
				},
			},
			wantErr: true,
			errMsg:  "flotsam ID 'ab-1' is invalid: must be 4-character alphanumeric",
		},
		{
			name: "empty title",
			note: FlotsamNote{
				FlotsamNote: flotsam.FlotsamNote{
					ID:      "abc1",
					Title:   "",
					Type:    "idea",
					Created: time.Now(),
				},
			},
			wantErr: true,
			errMsg:  "title is required",
		},
		{
			name: "whitespace-only title",
			note: FlotsamNote{
				FlotsamNote: flotsam.FlotsamNote{
					ID:      "abc1",
					Title:   "   ",
					Type:    "idea",
					Created: time.Now(),
				},
			},
			wantErr: true,
			errMsg:  "title is required",
		},
		{
			name: "invalid type",
			note: FlotsamNote{
				FlotsamNote: flotsam.FlotsamNote{
					ID:      "abc1",
					Title:   "Test Note",
					Type:    "invalid_type",
					Created: time.Now(),
				},
			},
			wantErr: true,
			errMsg:  "invalid type",
		},
		{
			name: "missing created time",
			note: FlotsamNote{
				FlotsamNote: flotsam.FlotsamNote{
					ID:    "abc1",
					Title: "Test Note",
					Type:  "idea",
				},
			},
			wantErr: true,
			errMsg:  "created timestamp is required",
		},
		{
			name: "modified before created",
			note: FlotsamNote{
				FlotsamNote: flotsam.FlotsamNote{
					ID:       "abc1",
					Title:    "Test Note",
					Type:     "idea",
					Created:  time.Now(),
					Modified: time.Now().Add(-1 * time.Hour),
				},
			},
			wantErr: true,
			errMsg:  "modified time cannot be before created time",
		},
		{
			name: "valid note with SRS data",
			note: FlotsamNote{
				FlotsamNote: flotsam.FlotsamNote{
					ID:      "abc1",
					Title:   "Test Flashcard",
					Type:    "flashcard",
					Created: time.Now(),
					SRS: &flotsam.SRSData{
						Easiness:           2.5,
						ConsecutiveCorrect: 1,
						Due:                time.Now().Unix(),
						TotalReviews:       1,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid SRS data - easiness too low",
			note: FlotsamNote{
				FlotsamNote: flotsam.FlotsamNote{
					ID:      "abc1",
					Title:   "Test Flashcard",
					Type:    "flashcard",
					Created: time.Now(),
					SRS: &flotsam.SRSData{
						Easiness:           1.0,
						ConsecutiveCorrect: 1,
						Due:                time.Now().Unix(),
						TotalReviews:       1,
					},
				},
			},
			wantErr: true,
			errMsg:  "invalid SRS data: easiness 1.00 out of bounds",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.note.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Validate() expected error, got nil")
					return
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Validate() error = %v, want to contain %v", err, tt.errMsg)
				}
			} else if err != nil {
				t.Errorf("Validate() unexpected error = %v", err)
			}
		})
	}
}

func TestFlotsamFrontmatterValidate(t *testing.T) {
	tests := []struct {
		name        string
		frontmatter FlotsamFrontmatter
		wantErr     bool
		errMsg      string
	}{
		{
			name: "valid frontmatter",
			frontmatter: FlotsamFrontmatter{
				ID:      "abc1",
				Title:   "Test Note",
				Type:    IdeaType,
				Created: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "invalid ID",
			frontmatter: FlotsamFrontmatter{
				ID:      "invalid",
				Title:   "Test Note",
				Type:    IdeaType,
				Created: time.Now(),
			},
			wantErr: true,
			errMsg:  "flotsam ID 'invalid' is invalid",
		},
		{
			name: "empty title",
			frontmatter: FlotsamFrontmatter{
				ID:      "abc1",
				Title:   "",
				Type:    IdeaType,
				Created: time.Now(),
			},
			wantErr: true,
			errMsg:  "title is required",
		},
		{
			name: "invalid type",
			frontmatter: FlotsamFrontmatter{
				ID:      "abc1",
				Title:   "Test Note",
				Type:    FlotsamType("invalid"),
				Created: time.Now(),
			},
			wantErr: true,
			errMsg:  "invalid type",
		},
		{
			name: "missing created time",
			frontmatter: FlotsamFrontmatter{
				ID:    "abc1",
				Title: "Test Note",
				Type:  IdeaType,
			},
			wantErr: true,
			errMsg:  "created timestamp is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.frontmatter.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Validate() expected error, got nil")
					return
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Validate() error = %v, want to contain %v", err, tt.errMsg)
				}
			} else if err != nil {
				t.Errorf("Validate() unexpected error = %v", err)
			}
		})
	}
}

func TestFlotsamCollectionValidate(t *testing.T) {
	validNote1 := FlotsamNote{
		FlotsamNote: flotsam.FlotsamNote{
			ID:      "abc1",
			Title:   "Note 1",
			Type:    "idea",
			Created: time.Now(),
		},
	}
	validNote2 := FlotsamNote{
		FlotsamNote: flotsam.FlotsamNote{
			ID:      "abc2",
			Title:   "Note 2",
			Type:    "idea",
			Created: time.Now(),
		},
	}
	invalidNote := FlotsamNote{
		FlotsamNote: flotsam.FlotsamNote{
			ID:    "abc3",
			Title: "", // Invalid: empty title
			Type:  "idea",
		},
	}

	tests := []struct {
		name       string
		collection FlotsamCollection
		wantErr    bool
		errMsg     string
	}{
		{
			name: "valid collection",
			collection: FlotsamCollection{
				Context: "test",
				Notes:   []FlotsamNote{validNote1, validNote2},
			},
			wantErr: false,
		},
		{
			name: "empty context",
			collection: FlotsamCollection{
				Context: "",
				Notes:   []FlotsamNote{validNote1},
			},
			wantErr: true,
			errMsg:  "context is required for collection isolation",
		},
		{
			name: "whitespace-only context",
			collection: FlotsamCollection{
				Context: "   ",
				Notes:   []FlotsamNote{validNote1},
			},
			wantErr: true,
			errMsg:  "context is required for collection isolation",
		},
		{
			name: "invalid note in collection",
			collection: FlotsamCollection{
				Context: "test",
				Notes:   []FlotsamNote{validNote1, invalidNote},
			},
			wantErr: true,
			errMsg:  "note 1 validation failed",
		},
		{
			name: "duplicate IDs",
			collection: FlotsamCollection{
				Context: "test",
				Notes:   []FlotsamNote{validNote1, validNote1}, // Same note twice
			},
			wantErr: true,
			errMsg:  "duplicate note ID 'abc1' found in collection",
		},
		{
			name: "empty collection",
			collection: FlotsamCollection{
				Context: "test",
				Notes:   []FlotsamNote{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.collection.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Validate() expected error, got nil")
					return
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Validate() error = %v, want to contain %v", err, tt.errMsg)
				}
			} else if err != nil {
				t.Errorf("Validate() unexpected error = %v", err)
			}
		})
	}
}

func TestIsValidFlotsamID(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want bool
	}{
		{"valid lowercase alphanum", "abc1", true},
		{"valid all letters", "abcd", true},
		{"valid all numbers", "1234", true},
		{"valid mixed", "a1b2", true},
		{"empty string", "", false},
		{"too short", "abc", false},
		{"too long", "abc12", false},
		{"uppercase", "ABC1", false},
		{"special chars", "ab-1", false},
		{"underscore", "ab_1", false},
		{"space", "ab 1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidFlotsamID(tt.id); got != tt.want {
				t.Errorf("isValidFlotsamID(%q) = %v, want %v", tt.id, got, tt.want)
			}
		})
	}
}

func TestValidateSRSData(t *testing.T) {
	tests := []struct {
		name    string
		srs     *flotsam.SRSData
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil SRS data",
			srs:     nil,
			wantErr: false,
		},
		{
			name: "valid SRS data",
			srs: &flotsam.SRSData{
				Easiness:           2.5,
				ConsecutiveCorrect: 1,
				Due:                time.Now().Unix(),
				TotalReviews:       1,
			},
			wantErr: false,
		},
		{
			name: "easiness too low",
			srs: &flotsam.SRSData{
				Easiness:           1.0,
				ConsecutiveCorrect: 0,
				Due:                time.Now().Unix(),
				TotalReviews:       0,
			},
			wantErr: true,
			errMsg:  "easiness 1.00 out of bounds",
		},
		{
			name: "easiness too high",
			srs: &flotsam.SRSData{
				Easiness:           5.0,
				ConsecutiveCorrect: 0,
				Due:                time.Now().Unix(),
				TotalReviews:       0,
			},
			wantErr: true,
			errMsg:  "easiness 5.00 out of bounds",
		},
		{
			name: "negative consecutive correct",
			srs: &flotsam.SRSData{
				Easiness:           2.5,
				ConsecutiveCorrect: -1,
				Due:                time.Now().Unix(),
				TotalReviews:       0,
			},
			wantErr: true,
			errMsg:  "consecutive_correct -1 must be non-negative",
		},
		{
			name: "negative total reviews",
			srs: &flotsam.SRSData{
				Easiness:           2.5,
				ConsecutiveCorrect: 0,
				Due:                time.Now().Unix(),
				TotalReviews:       -1,
			},
			wantErr: true,
			errMsg:  "total_reviews -1 must be non-negative",
		},
		{
			name: "negative due timestamp",
			srs: &flotsam.SRSData{
				Easiness:           2.5,
				ConsecutiveCorrect: 0,
				Due:                -1,
				TotalReviews:       0,
			},
			wantErr: true,
			errMsg:  "due timestamp -1 must be positive",
		},
		{
			name: "total reviews less than consecutive correct",
			srs: &flotsam.SRSData{
				Easiness:           2.5,
				ConsecutiveCorrect: 5,
				Due:                time.Now().Unix(),
				TotalReviews:       3,
			},
			wantErr: true,
			errMsg:  "total_reviews 3 cannot be less than consecutive_correct 5",
		},
		{
			name: "boundary values - minimum easiness",
			srs: &flotsam.SRSData{
				Easiness:           1.3,
				ConsecutiveCorrect: 0,
				Due:                1,
				TotalReviews:       0,
			},
			wantErr: false,
		},
		{
			name: "boundary values - maximum easiness",
			srs: &flotsam.SRSData{
				Easiness:           4.0,
				ConsecutiveCorrect: 0,
				Due:                1,
				TotalReviews:       0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSRSData(tt.srs)
			if tt.wantErr {
				if err == nil {
					t.Errorf("validateSRSData() expected error, got nil")
					return
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("validateSRSData() error = %v, want to contain %v", err, tt.errMsg)
				}
			} else if err != nil {
				t.Errorf("validateSRSData() unexpected error = %v", err)
			}
		})
	}
}
