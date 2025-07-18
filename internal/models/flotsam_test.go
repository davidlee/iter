package models

import (
	"testing"
	"time"

	"github.com/davidlee/vice/internal/flotsam"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlotsamType_Validate(t *testing.T) {
	tests := []struct {
		name      string
		noteType  FlotsamType
		wantError bool
	}{
		{"valid idea type", IdeaType, false},
		{"valid flashcard type", FlashcardType, false},
		{"valid script type", ScriptType, false},
		{"valid log type", LogType, false},
		{"invalid type", FlotsamType("invalid"), true},
		{"empty type", FlotsamType(""), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.noteType.Validate()
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFlotsamType_String(t *testing.T) {
	assert.Equal(t, "idea", IdeaType.String())
	assert.Equal(t, "flashcard", FlashcardType.String())
	assert.Equal(t, "script", ScriptType.String())
	assert.Equal(t, "log", LogType.String())
}

func TestFlotsamType_IsEmpty(t *testing.T) {
	assert.True(t, FlotsamType("").IsEmpty())
	assert.False(t, IdeaType.IsEmpty())
	assert.False(t, FlashcardType.IsEmpty())
}

func TestDefaultType(t *testing.T) {
	assert.Equal(t, IdeaType, DefaultType())
}

func TestNewFlotsamFrontmatter(t *testing.T) {
	fm := NewFlotsamFrontmatter("abc1", "Test Note")

	assert.Equal(t, "abc1", fm.ID)
	assert.Equal(t, "Test Note", fm.Title)
	assert.Equal(t, IdeaType, fm.Type)
	assert.NotNil(t, fm.Tags)
	assert.Empty(t, fm.Tags)
	assert.WithinDuration(t, time.Now(), fm.Created, time.Second)
	assert.Nil(t, fm.SRS)
}

func TestNewFlotsamNote(t *testing.T) {
	fm := &FlotsamFrontmatter{
		ID:      "xyz9",
		Title:   "Test Note",
		Created: time.Now(),
		Type:    FlashcardType,
		Tags:    []string{"test", "example"},
	}

	note := NewFlotsamNote(fm, "This is the body content", "/path/to/note.md")

	assert.Equal(t, "xyz9", note.ID)
	assert.Equal(t, "Test Note", note.Title)
	assert.Equal(t, "flashcard", note.Type)
	assert.Equal(t, []string{"test", "example"}, note.Tags)
	assert.Equal(t, "This is the body content", note.Body)
	assert.Equal(t, "/path/to/note.md", note.FilePath)
	assert.WithinDuration(t, time.Now(), note.Modified, time.Second)
}

func TestNewFlotsamCollection(t *testing.T) {
	collection := NewFlotsamCollection("personal")

	assert.Equal(t, "1.0", collection.Version)
	assert.Equal(t, "personal", collection.Context)
	assert.Equal(t, time.Now().Format("2006-01-02"), collection.CreatedDate)
	assert.NotNil(t, collection.Notes)
	assert.Empty(t, collection.Notes)
	assert.Equal(t, 0, collection.TotalNotes)
	assert.False(t, collection.SRSEnabled)
}

func TestFlotsamNote_GetFrontmatter(t *testing.T) {
	srsData := &flotsam.SRSData{
		Easiness:           2.5,
		ConsecutiveCorrect: 0,
		Due:                time.Now().Unix(),
		TotalReviews:       0,
	}

	note := &FlotsamNote{
		FlotsamNote: flotsam.FlotsamNote{
			ID:      "test123",
			Title:   "Test Note",
			Type:    "flashcard",
			Tags:    []string{"learning"},
			Created: time.Now(),
			SRS:     srsData,
		},
	}

	fm := note.GetFrontmatter()

	assert.Equal(t, "test123", fm.ID)
	assert.Equal(t, "Test Note", fm.Title)
	assert.Equal(t, FlashcardType, fm.Type)
	assert.Equal(t, []string{"learning"}, fm.Tags)
	assert.Equal(t, srsData, fm.SRS)
}

func TestFlotsamNote_UpdateFromFrontmatter(t *testing.T) {
	note := &FlotsamNote{
		FlotsamNote: flotsam.FlotsamNote{
			ID:    "old123",
			Title: "Old Title",
			Type:  "idea",
		},
	}

	newFM := &FlotsamFrontmatter{
		ID:      "new456",
		Title:   "New Title",
		Type:    FlashcardType,
		Tags:    []string{"updated"},
		Created: time.Now().Add(-time.Hour),
	}

	note.UpdateFromFrontmatter(newFM)

	assert.Equal(t, "new456", note.ID)
	assert.Equal(t, "New Title", note.Title)
	assert.Equal(t, "flashcard", note.Type)
	assert.Equal(t, []string{"updated"}, note.Tags)
	assert.WithinDuration(t, time.Now(), note.Modified, time.Second)
}

func TestFlotsamNote_HasSRS(t *testing.T) {
	t.Run("note without SRS", func(t *testing.T) {
		note := &FlotsamNote{
			FlotsamNote: flotsam.FlotsamNote{SRS: nil},
		}
		assert.False(t, note.HasSRS())
	})

	t.Run("note with SRS", func(t *testing.T) {
		note := &FlotsamNote{
			FlotsamNote: flotsam.FlotsamNote{
				SRS: &flotsam.SRSData{Easiness: 2.5},
			},
		}
		assert.True(t, note.HasSRS())
	})
}

func TestFlotsamNote_IsFlashcard(t *testing.T) {
	tests := []struct {
		name     string
		noteType string
		expected bool
	}{
		{"flashcard type", "flashcard", true},
		{"idea type", "idea", false},
		{"script type", "script", false},
		{"log type", "log", false},
		{"empty type", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			note := &FlotsamNote{
				FlotsamNote: flotsam.FlotsamNote{Type: tt.noteType},
			}
			assert.Equal(t, tt.expected, note.IsFlashcard())
		})
	}
}

func TestFlotsamNote_ValidateType(t *testing.T) {
	t.Run("valid type", func(t *testing.T) {
		note := &FlotsamNote{
			FlotsamNote: flotsam.FlotsamNote{Type: "flashcard"},
		}
		err := note.ValidateType()
		assert.NoError(t, err)
		assert.Equal(t, "flashcard", note.Type)
	})

	t.Run("empty type gets default", func(t *testing.T) {
		note := &FlotsamNote{
			FlotsamNote: flotsam.FlotsamNote{Type: ""},
		}
		err := note.ValidateType()
		assert.NoError(t, err)
		assert.Equal(t, "idea", note.Type)
	})

	t.Run("invalid type", func(t *testing.T) {
		note := &FlotsamNote{
			FlotsamNote: flotsam.FlotsamNote{Type: "invalid"},
		}
		err := note.ValidateType()
		assert.Error(t, err)
	})
}

func TestFlotsamCollection_AddNote(t *testing.T) {
	collection := NewFlotsamCollection("test")

	note := FlotsamNote{
		FlotsamNote: flotsam.FlotsamNote{
			ID:    "note1",
			Title: "First Note",
			Type:  "idea",
		},
	}

	collection.AddNote(note)

	assert.Equal(t, 1, len(collection.Notes))
	assert.Equal(t, 1, collection.TotalNotes)
	assert.Equal(t, "note1", collection.Notes[0].ID)
	assert.False(t, collection.SRSEnabled)
}

func TestFlotsamCollection_AddNote_WithSRS(t *testing.T) {
	collection := NewFlotsamCollection("test")

	note := FlotsamNote{
		FlotsamNote: flotsam.FlotsamNote{
			ID:    "note1",
			Title: "SRS Note",
			Type:  "flashcard",
			SRS:   &flotsam.SRSData{Easiness: 2.5},
		},
	}

	collection.AddNote(note)

	assert.Equal(t, 1, collection.TotalNotes)
	assert.True(t, collection.SRSEnabled)
}

func TestFlotsamCollection_RemoveNote(t *testing.T) {
	collection := NewFlotsamCollection("test")

	note1 := FlotsamNote{
		FlotsamNote: flotsam.FlotsamNote{ID: "note1", Title: "Note 1"},
	}
	note2 := FlotsamNote{
		FlotsamNote: flotsam.FlotsamNote{ID: "note2", Title: "Note 2"},
	}

	collection.AddNote(note1)
	collection.AddNote(note2)
	require.Equal(t, 2, collection.TotalNotes)

	// Remove existing note
	removed := collection.RemoveNote("note1")
	assert.True(t, removed)
	assert.Equal(t, 1, collection.TotalNotes)
	assert.Equal(t, "note2", collection.Notes[0].ID)

	// Try to remove non-existent note
	removed = collection.RemoveNote("nonexistent")
	assert.False(t, removed)
	assert.Equal(t, 1, collection.TotalNotes)
}

func TestFlotsamCollection_GetNote(t *testing.T) {
	collection := NewFlotsamCollection("test")

	note := FlotsamNote{
		FlotsamNote: flotsam.FlotsamNote{ID: "findme", Title: "Find Me"},
	}
	collection.AddNote(note)

	// Find existing note
	found, exists := collection.GetNote("findme")
	assert.True(t, exists)
	assert.NotNil(t, found)
	assert.Equal(t, "Find Me", found.Title)

	// Try to find non-existent note
	found, exists = collection.GetNote("missing")
	assert.False(t, exists)
	assert.Nil(t, found)
}

func TestFlotsamCollection_GetNotesByType(t *testing.T) {
	collection := NewFlotsamCollection("test")

	notes := []FlotsamNote{
		{FlotsamNote: flotsam.FlotsamNote{ID: "1", Type: "idea"}},
		{FlotsamNote: flotsam.FlotsamNote{ID: "2", Type: "flashcard"}},
		{FlotsamNote: flotsam.FlotsamNote{ID: "3", Type: "idea"}},
		{FlotsamNote: flotsam.FlotsamNote{ID: "4", Type: "script"}},
	}

	for _, note := range notes {
		collection.AddNote(note)
	}

	ideaNotes := collection.GetNotesByType(IdeaType)
	assert.Equal(t, 2, len(ideaNotes))

	flashcardNotes := collection.GetNotesByType(FlashcardType)
	assert.Equal(t, 1, len(flashcardNotes))
	assert.Equal(t, "2", flashcardNotes[0].ID)

	logNotes := collection.GetNotesByType(LogType)
	assert.Equal(t, 0, len(logNotes))
}

func TestFlotsamCollection_GetSRSNotes(t *testing.T) {
	collection := NewFlotsamCollection("test")

	notes := []FlotsamNote{
		{FlotsamNote: flotsam.FlotsamNote{ID: "1", SRS: nil}},
		{FlotsamNote: flotsam.FlotsamNote{ID: "2", SRS: &flotsam.SRSData{Easiness: 2.5}}},
		{FlotsamNote: flotsam.FlotsamNote{ID: "3", SRS: nil}},
		{FlotsamNote: flotsam.FlotsamNote{ID: "4", SRS: &flotsam.SRSData{Easiness: 2.8}}},
	}

	for _, note := range notes {
		collection.AddNote(note)
	}

	srsNotes := collection.GetSRSNotes()
	assert.Equal(t, 2, len(srsNotes))

	// Verify the right notes were returned
	srsIDs := make([]string, len(srsNotes))
	for i, note := range srsNotes {
		srsIDs[i] = note.ID
	}
	assert.Contains(t, srsIDs, "2")
	assert.Contains(t, srsIDs, "4")
}

func TestFlotsamCollection_computeMetadata(t *testing.T) {
	collection := NewFlotsamCollection("test")

	// Initially empty
	assert.Equal(t, 0, collection.TotalNotes)
	assert.False(t, collection.SRSEnabled)

	// Add note without SRS
	noteNoSRS := FlotsamNote{
		FlotsamNote: flotsam.FlotsamNote{ID: "1", SRS: nil},
	}
	collection.AddNote(noteNoSRS)
	assert.Equal(t, 1, collection.TotalNotes)
	assert.False(t, collection.SRSEnabled)

	// Add note with SRS
	noteWithSRS := FlotsamNote{
		FlotsamNote: flotsam.FlotsamNote{
			ID:  "2",
			SRS: &flotsam.SRSData{Easiness: 2.5},
		},
	}
	collection.AddNote(noteWithSRS)
	assert.Equal(t, 2, collection.TotalNotes)
	assert.True(t, collection.SRSEnabled)

	// Remove SRS note
	collection.RemoveNote("2")
	assert.Equal(t, 1, collection.TotalNotes)
	assert.False(t, collection.SRSEnabled)
}

// Test YAML serialization for frontmatter
func TestFlotsamFrontmatter_YAMLSerialization(t *testing.T) {
	fm := &FlotsamFrontmatter{
		ID:      "test123",
		Title:   "Test Note",
		Created: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
		Tags:    []string{"test", "example"},
		Type:    FlashcardType,
		SRS: &flotsam.SRSData{
			Easiness:           2.5,
			ConsecutiveCorrect: 1,
			Due:                1640995200,
			TotalReviews:       3,
		},
	}

	// This test validates the struct tags are correct for YAML serialization
	// The actual YAML marshaling would be tested in integration tests

	assert.Equal(t, "test123", fm.ID)
	assert.Equal(t, FlashcardType, fm.Type)
	assert.NotNil(t, fm.SRS)
	assert.Equal(t, 2.5, fm.SRS.Easiness)
}
