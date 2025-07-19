// Package models provides validation tests for flotsam data structures.
package models

import (
	"strings"
	"testing"
	"time"
)

func TestFlotsamNoteValidate(t *testing.T) {
	tests := []struct {
		name    string
		note    FlotsamNote
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid note with tags",
			note: FlotsamNote{
				ID:      "abc1",
				Title:   "Test Note",
				Tags:    []string{"vice:type:idea", "concept"},
				Created: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "valid note minimal",
			note: FlotsamNote{
				ID:      "a1b2",
				Title:   "Minimal Note",
				Created: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "invalid ID - too short",
			note: FlotsamNote{
				ID:      "ab1",
				Title:   "Test Note",
				Created: time.Now(),
			},
			wantErr: true,
			errMsg:  "flotsam ID 'ab1' is invalid: must be 4-character alphanumeric",
		},
		{
			name: "invalid ID - too long",
			note: FlotsamNote{
				ID:      "abc12",
				Title:   "Test Note",
				Created: time.Now(),
			},
			wantErr: true,
			errMsg:  "flotsam ID 'abc12' is invalid: must be 4-character alphanumeric",
		},
		{
			name: "invalid ID - uppercase",
			note: FlotsamNote{
				ID:      "ABC1",
				Title:   "Test Note",
				Created: time.Now(),
			},
			wantErr: true,
			errMsg:  "flotsam ID 'ABC1' is invalid: must be 4-character alphanumeric",
		},
		{
			name: "invalid ID - special characters",
			note: FlotsamNote{
				ID:      "ab-1",
				Title:   "Test Note",
				Created: time.Now(),
			},
			wantErr: true,
			errMsg:  "flotsam ID 'ab-1' is invalid: must be 4-character alphanumeric",
		},
		{
			name: "empty title",
			note: FlotsamNote{
				ID:      "abc1",
				Title:   "",
				Created: time.Now(),
			},
			wantErr: true,
			errMsg:  "title is required",
		},
		{
			name: "whitespace-only title",
			note: FlotsamNote{
				ID:      "abc1",
				Title:   "   ",
				Created: time.Now(),
			},
			wantErr: true,
			errMsg:  "title is required",
		},
		{
			name: "missing created time",
			note: FlotsamNote{
				ID:    "abc1",
				Title: "Test Note",
			},
			wantErr: true,
			errMsg:  "created timestamp is required",
		},
		{
			name: "modified before created",
			note: FlotsamNote{
				ID:       "abc1",
				Title:    "Test Note",
				Created:  time.Now(),
				Modified: time.Now().Add(-1 * time.Hour),
			},
			wantErr: true,
			errMsg:  "modified time cannot be before created time",
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

func TestFlotsamNoteTagBehavior(t *testing.T) {
	tests := []struct {
		name      string
		note      FlotsamNote
		checkSRS  bool
		checkFC   bool
		checkType string
	}{
		{
			name: "flashcard with SRS",
			note: FlotsamNote{
				ID:    "abc1",
				Title: "Test Flashcard",
				Tags:  []string{"vice:srs", "vice:type:flashcard"},
			},
			checkSRS:  true,
			checkFC:   true,
			checkType: "flashcard",
		},
		{
			name: "idea note with SRS",
			note: FlotsamNote{
				ID:    "abc2",
				Title: "Test Idea",
				Tags:  []string{"vice:type:idea", "concept"},
			},
			checkSRS:  true,
			checkFC:   false,
			checkType: "idea",
		},
		{
			name: "note with SRS but not flashcard",
			note: FlotsamNote{
				ID:    "abc3",
				Title: "Test Script",
				Tags:  []string{"vice:srs", "vice:type:script"},
			},
			checkSRS:  true,
			checkFC:   false,
			checkType: "script",
		},
		{
			name: "note without vice tags",
			note: FlotsamNote{
				ID:    "abc4",
				Title: "Regular Note",
				Tags:  []string{"general", "notes"},
			},
			checkSRS:  false,
			checkFC:   false,
			checkType: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.note.HasSRS(); got != tt.checkSRS {
				t.Errorf("HasSRS() = %v, want %v", got, tt.checkSRS)
			}
			if got := tt.note.IsFlashcard(); got != tt.checkFC {
				t.Errorf("IsFlashcard() = %v, want %v", got, tt.checkFC)
			}
			if tt.checkType != "" {
				if got := tt.note.HasType(tt.checkType); !got {
					t.Errorf("HasType(%q) = %v, want true", tt.checkType, got)
				}
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
				Tags:    []string{"vice:type:idea"},
				Created: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "invalid ID",
			frontmatter: FlotsamFrontmatter{
				ID:      "invalid",
				Title:   "Test Note",
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
				Created: time.Now(),
			},
			wantErr: true,
			errMsg:  "title is required",
		},
		{
			name: "missing created time",
			frontmatter: FlotsamFrontmatter{
				ID:    "abc1",
				Title: "Test Note",
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
